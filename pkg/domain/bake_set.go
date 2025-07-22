package domain

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type BakeSet struct {
	Index       int  // インデックス
	IsTerminate bool // 処理停止フラグ

	OriginalMotionPath string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPath  string `json:"original_model_path"`  // 元モデルパス
	OutputMotionPath   string `json:"-"`                    // 出力モーションパス
	OutputModelPath    string `json:"-"`                    // 出力モデルパス

	OriginalMotionName string `json:"-"` // 元モーション名
	OriginalModelName  string `json:"-"` // 元モーション名
	OutputModelName    string `json:"-"` // 物理焼き込み先モデル名

	OriginalMotion *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel  *pmx.PmxModel  `json:"-"` // 元モデル
	BakedModel     *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	OutputMotion   *vmd.VmdMotion `json:"-"` // 出力結果モーション

	PhysicsTree *PhysicsModel `json:"-"` // 物理ボーンツリー
	OutputTree  *OutputModel  `json:"-"` // 出力ボーンツリー
}

func NewPhysicsSet(index int) *BakeSet {
	return &BakeSet{
		Index:       index,
		PhysicsTree: NewPhysicsModel(),
	}
}

func (ss *BakeSet) CreateOutputModelPath() string {
	if ss.OriginalModel == nil {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	return mfile.CreateOutputPath(ss.OriginalModel.Path(), "BB")
}

func (ss *BakeSet) CreateOutputMotionPath() string {
	if ss.OriginalMotion == nil || ss.BakedModel == nil {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	_, fileName, _ := mfile.SplitPath(ss.BakedModel.Path())

	return mfile.CreateOutputPath(
		ss.OriginalMotion.Path(), fmt.Sprintf("BB_%s", fileName))
}

func (ss *BakeSet) setMotion(originalMotion, outputMotion *vmd.VmdMotion) {
	if originalMotion == nil || outputMotion == nil {
		ss.OriginalMotionPath = ""
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.OutputMotionPath = ""
		ss.OutputMotion = vmd.NewVmdMotion("")

		return
	}

	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion
	ss.OutputMotion = outputMotion
}

func (ss *BakeSet) setModels(originalModel, physicsBakedModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.OriginalModelPath = ""
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.BakedModel = nil
		return
	}

	ss.OriginalModelPath = originalModel.Path()
	ss.OriginalModelName = originalModel.Name()
	ss.OriginalModel = originalModel
	ss.BakedModel = physicsBakedModel
}

// LoadModel モデルを読み込む
func (ss *BakeSet) LoadModel(path string) error {
	if path == "" {
		ss.setModels(nil, nil)
		return nil
	}

	var wg sync.WaitGroup
	var originalModel, physicsBakedModel *pmx.PmxModel

	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		pmxRep := repository.NewPmxRepository(true)
		if data, err := pmxRep.Load(path); err == nil {
			originalModel = data.(*pmx.PmxModel)
			if err := originalModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
			} else {
				// 物理ボーンの名前に接頭辞を追加
				ss.insertPhysicsBonePrefix(originalModel)
				// 物理ボーンを表示枠に追加
				ss.appendPhysicsBoneToDisplaySlots(originalModel)
			}
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		pmxRep := repository.NewPmxRepository(false)
		if data, err := pmxRep.Load(path); err == nil {
			physicsBakedModel = data.(*pmx.PmxModel)
			if err := physicsBakedModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
			} else {
				// 物理ボーンの名前に接頭辞を追加
				ss.insertPhysicsBonePrefix(physicsBakedModel)
				// 物理ボーンを表示枠に追加
				ss.appendPhysicsBoneToDisplaySlots(physicsBakedModel)
				// 物理ボーンの物理剛体を無効化
				ss.fixPhysicsRigidBodies(physicsBakedModel)
			}
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			ss.setModels(nil, nil)
			return err
		}
	}

	ss.setModels(originalModel, physicsBakedModel)
	ss.OutputModelPath = ss.CreateOutputModelPath()

	return nil
}

// appendPhysicsBoneToDisplaySlots 物理ボーンを表示枠に追加
func (ss *BakeSet) appendPhysicsBoneToDisplaySlots(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	// 表示枠に追加済みのボーン一覧を取得
	displayedBones := make([]bool, model.Bones.Length())
	model.DisplaySlots.ForEach(func(slotIndex int, slot *pmx.DisplaySlot) bool {
		for _, ref := range slot.References {
			if ref.DisplayType == pmx.DISPLAY_TYPE_BONE {
				displayedBones[ref.DisplayIndex] = true
			}
		}
		return true
	})

	var physicsDisplaySlot *pmx.DisplaySlot

	// 物理ボーンを表示枠に追加
	model.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		if bone.HasPhysics() && !displayedBones[boneIndex] {
			// 物理ボーンで、表示枠に追加されていない場合
			if physicsDisplaySlot == nil {
				// 物理ボーン用の表示枠がまだない場合、作成する
				physicsDisplaySlot = pmx.NewDisplaySlot()
				physicsDisplaySlot.SetName("Physics")
			}

			// 物理ボーンを表示枠に追加
			ref := pmx.NewDisplaySlotReferenceByValues(pmx.DISPLAY_TYPE_BONE, boneIndex)
			physicsDisplaySlot.References = append(physicsDisplaySlot.References, ref)

			// 操作できるようにフラグを設定
			bone.BoneFlag |= pmx.BONE_FLAG_IS_VISIBLE
			bone.BoneFlag |= pmx.BONE_FLAG_CAN_MANIPULATE
			bone.BoneFlag |= pmx.BONE_FLAG_CAN_TRANSLATE
			bone.BoneFlag |= pmx.BONE_FLAG_CAN_ROTATE

			model.Bones.Update(bone)
		}
		return true
	})

	if physicsDisplaySlot != nil {
		// 物理ボーン用の表示枠が作成された場合、モデルに追加する
		model.DisplaySlots.Append(physicsDisplaySlot)
	}
}

// insertPhysicsBonePrefix 物理ボーンの名前に接頭辞を追加
func (ss *BakeSet) insertPhysicsBonePrefix(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	digits := int(math.Log10(float64(model.Bones.Length()))) + 1

	// 物理ボーンの名前に接頭辞を追加
	model.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		if bone.HasPhysics() {
			// ボーンINDEXを0埋めして設定
			formattedBoneName := fmt.Sprintf("PF%0*d_%s", digits, boneIndex, bone.Name())
			bone.SetName(ss.encodeName(formattedBoneName, 15))
		}
		return true
	})
}

func (ss *BakeSet) encodeName(name string, limit int) string {
	// Encode to CP932
	cp932Encoder := japanese.ShiftJIS.NewEncoder()
	cp932Encoded, err := cp932Encoder.String(name)
	if err != nil {
		return name // エンコード失敗時はそのまま返す
	}

	// Decode to Shift_JIS
	shiftJISDecoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(bytes.NewReader([]byte(cp932Encoded)), shiftJISDecoder)
	shiftJISDecoded, err := io.ReadAll(reader)
	if err != nil {
		return name // エンコード失敗時はそのまま返す
	}

	// Encode to Shift_JIS
	shiftJISEncoder := japanese.ShiftJIS.NewEncoder()
	shiftJISEncoded, err := shiftJISEncoder.String(string(shiftJISDecoded))
	if err != nil {
		return name // エンコード失敗時はそのまま返す
	}

	encodedName := []byte(shiftJISEncoded)
	if len(encodedName) <= limit {
		// 指定バイト数に足りない場合は b"\x00" で埋める
		encodedName = append(encodedName, make([]byte, limit-len(encodedName))...)
	}

	// 指定バイト数に切り詰め
	encodedLimitName := encodedName[:limit]

	// VMDは空白込みで入っているので、正規表現で空白以降は削除する
	decodedBytes, err := japanese.ShiftJIS.NewDecoder().Bytes(encodedLimitName)
	if err != nil {
		return name // エンコード失敗時はそのまま返す
	}

	trimBytes := bytes.TrimRight(decodedBytes, "\xfd")                   // PMDで保存したVMDに入ってる
	trimBytes = bytes.TrimRight(trimBytes, "\x00")                       // VMDの末尾空白を除去
	trimBytes = bytes.ReplaceAll(trimBytes, []byte("\x00"), []byte(" ")) // 空白をスペースに変換

	decodedText := string(trimBytes)

	return decodedText
}

func (ss *BakeSet) fixPhysicsRigidBodies(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	// 物理ボーンの剛体を修正
	model.RigidBodies.ForEach(func(rigidBodyIndex int, rigidBody *pmx.RigidBody) bool {
		rigidBody.PhysicsType = pmx.PHYSICS_TYPE_STATIC // 剛体の物理演算を無効にする
		model.RigidBodies.Update(rigidBody)
		return true
	})
}

func (ss *BakeSet) LoadMotion(path string) error {

	if path == "" {
		ss.setMotion(nil, nil)

		return nil
	}

	var wg sync.WaitGroup
	var originalMotion, outputMotion *vmd.VmdMotion
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		vmdRep := repository.NewVmdVpdRepository(false)
		if data, err := vmdRep.Load(path); err == nil {
			originalMotion = data.(*vmd.VmdMotion)
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		vmdRep := repository.NewVmdVpdRepository(true)
		if data, err := vmdRep.Load(path); err == nil {
			outputMotion = data.(*vmd.VmdMotion)
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	ss.setMotion(originalMotion, outputMotion)

	// 出力パスを設定
	ss.OutputMotionPath = ss.CreateOutputMotionPath()

	return nil
}

func (ss *BakeSet) Delete() {
	ss.OriginalMotionPath = ""
	ss.OriginalModelPath = ""
	ss.OutputMotionPath = ""
	ss.OutputModelPath = ""

	ss.OriginalMotionName = ""
	ss.OriginalModelName = ""
	ss.OutputModelName = ""

	ss.OriginalMotion = nil
	ss.OriginalModel = nil
	ss.OutputMotion = nil
}

// 物理ボーンだけ残す
func (ss *BakeSet) GetOutputMotionOnlyChecked(startFrame, endFrame float64) (*vmd.VmdMotion, error) {
	if ss.OriginalModel == nil || ss.OutputMotion == nil {
		return nil, errors.New(mi18n.T("物理焼き込みセットの元モデルまたは出力モーションが設定されていません"))
	}

	if startFrame < 0 || endFrame < 0 || startFrame > endFrame {
		return nil, errors.New(mi18n.T("開始フレームより終了フレームが小さいか、負の値が設定されています"))
	}

	motion := vmd.NewVmdMotion(ss.OutputMotionPath)

	ss.OriginalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		item := ss.OutputTree.AtByBoneIndex(boneIndex)
		if item == nil || !item.(*OutputItem).Checked() {
			// チェックされていないボーンはスキップ
			return true
		}

		for index := startFrame; index <= endFrame; index++ {
			bf := ss.OutputMotion.BoneFrames.Get(bone.Name()).Get(float32(index))
			if bf == nil {
				continue
			}
			bf.DisablePhysics = true // 物理演算を無効にする
			motion.AppendBoneFrame(bone.Name(), bf)
		}

		if bone.HasPhysics() {
			// 最後に物理有効化を入れる
			lastBf := ss.OutputMotion.BoneFrames.Get(bone.Name()).Get(float32(endFrame + 1))
			if lastBf != nil {
				lastBf.DisablePhysics = false // 物理演算を有効にする
				motion.AppendBoneFrame(bone.Name(), lastBf)
			}
		}

		return true
	})

	return motion, nil
}
