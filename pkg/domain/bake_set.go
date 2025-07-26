package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
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

	PhysicsTableModel *PhysicsTableModel `json:"physics_table"` // 物理ボーンツリー
	OutputTableModel  *OutputTableModel  `json:"output_table"`  // 出力定義テーブル
}

func NewPhysicsSet(index int) *BakeSet {
	return &BakeSet{
		Index:             index,
		PhysicsTableModel: NewPhysicsTableModel(),
		OutputTableModel:  NewOutputTableModel(),
	}
}

func (ss *BakeSet) MaxFrame() float32 {
	if ss.OriginalMotion == nil {
		return 0
	}

	return ss.OriginalMotion.MaxFrame()
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

// SetModels ドメインロジックでモデルを設定（公開メソッド）
func (ss *BakeSet) SetModels(originalModel, bakedModel *pmx.PmxModel) error {
	if originalModel == nil {
		ss.setModels(nil, nil)
		return nil
	}

	// ドメインルールの適用
	ss.processPhysicsBones(originalModel)
	ss.processPhysicsBones(bakedModel)

	if bakedModel != nil {
		ss.fixPhysicsRigidBodies(bakedModel)
	}

	ss.setModels(originalModel, bakedModel)
	ss.OutputModelPath = ss.CreateOutputModelPath()

	return nil
}

// ClearModels モデルをクリア（公開メソッド）
func (ss *BakeSet) ClearModels() {
	ss.setModels(nil, nil)
}

// SetMotions ドメインロジックでモーションを設定（公開メソッド）
func (ss *BakeSet) SetMotions(originalMotion, outputMotion *vmd.VmdMotion) error {
	ss.setMotion(originalMotion, outputMotion)
	ss.OutputMotionPath = ss.CreateOutputMotionPath()
	return nil
}

// ClearMotions モーションをクリア（公開メソッド）
func (ss *BakeSet) ClearMotions() {
	ss.setMotion(nil, nil)
}

// processPhysicsBones 物理ボーンの処理（ドメインロジック）
func (ss *BakeSet) processPhysicsBones(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	// 物理ボーンの名前に接頭辞を追加
	ss.insertPhysicsBonePrefix(model)
	// 物理ボーンを表示枠に追加
	ss.appendPhysicsBoneToDisplaySlots(model)
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
		if bone.HasDynamicPhysics() {
			// ボーンINDEXを0埋めして設定
			formattedBoneName := fmt.Sprintf("BB%0*d_%s", digits, boneIndex, bone.Name())
			bone.SetName(ss.encodeName(formattedBoneName, 15))
		}
		return true
	})

	model.Bones.UpdateNameIndexes()
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
func (ss *BakeSet) GetOutputMotionOnlyChecked(records []*OutputBoneRecord) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)

	if ss.OriginalModel == nil || ss.OutputMotion == nil || len(records) == 0 {
		return motions, errors.New(mi18n.T("物理焼き込みセットの元モデルまたは出力モーションが設定されていません"))
	}

	var bakedMotion *vmd.VmdMotion

	// まずは既存モーションに焼き込みボーンを追加挿入する
	var err error
	bakedMotion, err = ss.OriginalMotion.Copy()
	bakedMotion.SetPath(ss.OutputMotionPath)
	if err != nil {
		return motions, fmt.Errorf(mi18n.T("元モーションのコピーに失敗しました: %w"), err)
	}

	keyCounts := make([]int, int(ss.OriginalMotion.MaxFrame()+1+1))
	for _, record := range records {
		if record == nil || record.OutputBoneTreeModel == nil {
			continue
		}

		for f := record.StartFrame; f <= record.EndFrame; f++ {
			for _, boneName := range record.TargetBoneNames {
				bf := ss.OutputMotion.BoneFrames.Get(boneName).Get(f)

				if bf == nil {
					continue
				}

				if bone, err := ss.OriginalModel.Bones.GetByName(boneName); err == nil {
					if bone.HasPhysics() {
						bf.DisablePhysics = true
					}
					bakedMotion.AppendBoneFrame(boneName, bf)
					keyCounts[int(f)]++

					// 次のキーフレ物理有効で登録しておく
					if bone.HasPhysics() {
						bf := ss.OutputMotion.BoneFrames.Get(boneName).Get(f + 1)
						if bf == nil {
							continue
						}
						bf.DisablePhysics = false
						bakedMotion.AppendBoneFrame(boneName, bf)
						keyCounts[int(f)]++
					}
				}
			}
		}
	}

	if mmath.Sum(keyCounts) == 0 {
		return motions, errors.New(mi18n.T("焼き込み対象キーフレームなし"))
	}

	dirPath, fileName, ext := mfile.SplitPath(ss.OutputMotionPath)
	motion := vmd.NewVmdMotion("")
	motion.SetPath(fmt.Sprintf("%s%s_%04d%s", dirPath, fileName, 0, ext))
	motion.MorphFrames, _ = ss.OriginalMotion.MorphFrames.Copy()

	frameCount := 0
	logFrameCount := 0

	for f := 0; f <= len(keyCounts); f++ {
		if f < len(keyCounts)-1 && frameCount+keyCounts[int(f+1)] > vmd.MAX_BONE_FRAMES {
			// キーフレーム数が上限を超える場合は切り替える
			motions = append(motions, motion)

			dirPath, fileName, ext := mfile.SplitPath(ss.OutputMotionPath)
			motion = vmd.NewVmdMotion(fmt.Sprintf("%s%s_%04d%s", dirPath, fileName, f, ext))
			motion.MorphFrames, _ = ss.OriginalMotion.MorphFrames.Copy()

			mlog.I(fmt.Sprintf(mi18n.T("キーフレーム数が上限を超えるため、モーションを切り替えます[%04dF]: %d -> %d"),
				f, frameCount, vmd.MAX_BONE_FRAMES))

			frameCount = 0
			logFrameCount = 0
		}

		if frameCount/100000 > logFrameCount/100000 {
			mlog.I(fmt.Sprintf(mi18n.T("- 物理焼き込み中... [%04dF] %dキーフレーム"), f, frameCount))
			logFrameCount = frameCount
		}

		ss.OriginalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
			if bakedMotion.BoneFrames.Get(bone.Name()).Contains(float32(f)) {
				// キーフレームがある場合は登録
				motion.AppendBoneFrame(bone.Name(), bakedMotion.BoneFrames.Get(bone.Name()).Get(float32(f)))
			}

			return true
		})

		if f < len(keyCounts)-1 {
			frameCount += keyCounts[int(f)]
		}
	}

	motions = append(motions, motion)
	return motions, nil
}

// SaveBakeSets セット保存用のドメインサービス
func SaveBakeSets(bakeSets []*BakeSet, jsonPath string) error {
	if strings.ToLower(filepath.Ext(jsonPath)) != ".json" {
		jsonPath += ".json"
	}

	output, err := json.Marshal(bakeSets)
	if err != nil {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	if err := os.WriteFile(jsonPath, output, 0644); err != nil {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	mlog.I(mi18n.T("物理焼き込みセット保存成功", map[string]any{"Path": jsonPath}))
	return nil
}

// LoadBakeSets セット読み込み用のドメインサービス
func LoadBakeSets(jsonPath string) ([]*BakeSet, error) {
	input, err := os.ReadFile(jsonPath)
	if err != nil {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return nil, err
	}

	var bakeSets []*BakeSet
	if err := json.Unmarshal(input, &bakeSets); err != nil {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return nil, err
	}

	mlog.I(mi18n.T("物理焼き込みセット読込成功", map[string]any{"Path": jsonPath}))
	return bakeSets, nil
}

// // SetOutputChildrenChecked は指定されたアイテムの子どもを再帰的にチェック状態を設定する
// func (ss *BakeSet) SetOutputChildrenChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
// 	if item == nil || ss.IsOutputUpdatingChildren ||
// 		ss.IsOutputUpdatingPhysics || ss.IsOutputUpdatingIk {
// 		return
// 	}

// 	// 無限ループを防ぐためのフラグ
// 	ss.IsOutputUpdatingChildren = true
// 	defer func() {
// 		ss.IsOutputUpdatingChildren = false
// 	}()

// 	item.(*OutputItem).SetChecked(checked)

// 	// 子どもの数を取得
// 	childCount := item.ChildCount()
// 	for i := range childCount {
// 		child := item.ChildAt(i)
// 		if child == nil {
// 			continue
// 		}

// 		// 子どものチェック状態を設定
// 		if outputItem, ok := child.(*OutputItem); ok {
// 			outputItem.SetChecked(checked)
// 			treeView.SetChecked(outputItem, checked)
// 		}

// 		// 再帰的に孫も処理（フラグを一時的にクリアして再帰呼び出し）
// 		ss.IsOutputUpdatingChildren = false
// 		ss.SetOutputChildrenChecked(treeView, child, checked)
// 		ss.IsOutputUpdatingChildren = true
// 	}
// }

// // SetOutputPhysicsChecked は物理関連ボーンのチェック状態を設定する
// func (ss *BakeSet) SetOutputPhysicsChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
// 	// if ss.IsOutputUpdatingPhysics {
// 	// 	return
// 	// }

// 	// 無限ループを防ぐためのフラグ
// 	ss.IsOutputUpdatingPhysics = true
// 	defer func() {
// 		ss.IsOutputUpdatingPhysics = false
// 	}()

// 	if item == nil {
// 		for i := range treeView.Model().RootCount() {
// 			item := treeView.Model().RootAt(i)
// 			ss.SetOutputPhysicsChecked(treeView, item, checked)
// 		}
// 		return
// 	}

// 	// 子どもの数を取得
// 	for i := range item.ChildCount() {
// 		child := item.ChildAt(i)
// 		if child == nil {
// 			continue
// 		}

// 		// 出力IKボーンのチェック状態を設定
// 		if outputItem, ok := child.(*OutputItem); ok {
// 			if outputItem.AsPhysics() {
// 				outputItem.SetChecked(checked)
// 				treeView.SetChecked(outputItem, checked)
// 			}
// 		}

// 		// 子どもアイテムのチェック状態を設定
// 		ss.SetOutputPhysicsChecked(treeView, child, checked)
// 	}
// }

// // SetOutputIkChecked はIK関連ボーンのチェック状態を設定する
// func (ss *BakeSet) SetOutputIkChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
// 	// if ss.IsOutputUpdatingIk {
// 	// 	return
// 	// }

// 	if item == nil {
// 		for i := range treeView.Model().RootCount() {
// 			item := treeView.Model().RootAt(i)
// 			ss.SetOutputIkChecked(treeView, item, checked)
// 		}
// 		return
// 	}

// 	// 無限ループを防ぐためのフラグ
// 	ss.IsOutputUpdatingIk = true
// 	defer func() {
// 		ss.IsOutputUpdatingIk = false
// 	}()

// 	// 子どもの数を取得
// 	for i := range item.ChildCount() {
// 		child := item.ChildAt(i)
// 		if child == nil {
// 			continue
// 		}

// 		// 出力IKボーンのチェック状態を設定
// 		if outputItem, ok := child.(*OutputItem); ok {
// 			if outputItem.AsIk() {
// 				outputItem.SetChecked(checked)
// 				treeView.SetChecked(outputItem, checked)
// 			}
// 		}

// 		// 子どもアイテムのチェック状態を設定
// 		ss.SetOutputIkChecked(treeView, child, checked)
// 	}
// }
