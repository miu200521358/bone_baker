package usecase

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	pRepository "github.com/miu200521358/bone_baker/pkg/infrastructure/repository"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type LoadUsecase struct {
	fileRepo *pRepository.FileRepository
}

func NewLoadUsecase(fileRepo *pRepository.FileRepository) *LoadUsecase {
	return &LoadUsecase{
		fileRepo: fileRepo,
	}
}

func (uc *LoadUsecase) LoadFile(path string) ([]*entity.BakeSet, []*entity.PhysicsRecord, error) {
	return uc.fileRepo.Load(path)
}

func (uc *LoadUsecase) LoadMotion(bakeSet *entity.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearMotion()
		return nil
	}

	var wg sync.WaitGroup
	var originalMotion, outputMotion *vmd.VmdMotion
	errChan := make(chan error, 2)

	// 元モーション読み込み
	wg.Add(1)
	go func() {
		defer wg.Done()
		rep := repository.NewVmdRepository(true)
		if motion, err := rep.Load(path); err == nil {
			originalMotion = motion.(*vmd.VmdMotion)
		} else {
			errChan <- err
		}
	}()

	// 出力モーション読み込み
	wg.Add(1)
	go func() {
		defer wg.Done()
		rep := repository.NewVmdRepository(true)
		if motion, err := rep.Load(path); err == nil {
			outputMotion = motion.(*vmd.VmdMotion)
		} else {
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

	bakeSet.OriginalMotion = originalMotion
	bakeSet.OriginalMotionPath = path
	bakeSet.OutputMotion = outputMotion

	return nil
}

func (uc *LoadUsecase) LoadModel(bakeSet *entity.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearModel()
		return nil
	}

	var wg sync.WaitGroup
	var originalModel, bakeModel *pmx.PmxModel
	errChan := make(chan error, 2)

	// 元モーション読み込み
	wg.Add(1)
	go func() {
		defer wg.Done()
		rep := repository.NewPmxRepository(true)
		if model, err := rep.Load(path); err == nil {
			originalModel = model.(*pmx.PmxModel)

			if err := originalModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
				return
			}

			// 物理剛体の名前を変更して表示枠に追加
			uc.insertPhysicsBonePrefix(originalModel)
			uc.appendPhysicsBoneToDisplaySlots(originalModel)
		} else {
			errChan <- err
		}
	}()

	// 出力モーション読み込み
	wg.Add(1)
	go func() {
		defer wg.Done()
		rep := repository.NewPmxRepository(true)
		if model, err := rep.Load(path); err == nil {
			bakeModel = model.(*pmx.PmxModel)

			if err := bakeModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
				return
			}

			// 物理剛体の名前を変更して表示枠に追加
			uc.insertPhysicsBonePrefix(bakeModel)
			uc.appendPhysicsBoneToDisplaySlots(bakeModel)

			// 物理剛体を無効化
			uc.fixPhysicsRigidBodies(bakeModel)
		} else {
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

	bakeSet.OriginalModel = originalModel
	bakeSet.OriginalModelPath = path
	bakeSet.BakedModel = bakeModel

	return nil
}

// appendPhysicsBoneToDisplaySlots 物理ボーンを表示枠に追加
func (uc *LoadUsecase) appendPhysicsBoneToDisplaySlots(model *pmx.PmxModel) {
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
func (uc *LoadUsecase) insertPhysicsBonePrefix(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	digits := int(math.Log10(float64(model.Bones.Length()))) + 1

	// 物理ボーンの名前に接頭辞を追加
	model.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		if bone.HasDynamicPhysics() {
			// ボーンINDEXを0埋めして設定
			formattedBoneName := fmt.Sprintf("BB%0*d_%s", digits, boneIndex, bone.Name())

			// BoneNameEncodingServiceを使用
			bone.SetName(uc.encodeName(formattedBoneName, 14))
		}
		return true
	})

	model.Bones.UpdateNameIndexes()
}

// fixPhysicsRigidBodies 物理剛体を修正
func (uc *LoadUsecase) fixPhysicsRigidBodies(model *pmx.PmxModel) {
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

// encodeName ボーン名を指定されたバイト制限でエンコード
func (uc *LoadUsecase) encodeName(name string, limit int) string {
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

	// 末尾が文字化けしている場合は削除
	if len(decodedText) > 0 && (decodedText[len(decodedText)-1] < 0x20 || decodedText[len(decodedText)-1] > 0x7E) {
		decodedText = decodedText[:len(decodedText)-1]
	}

	return decodedText
}
