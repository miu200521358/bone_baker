package domain

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// BakeSetService BakeSetのドメインサービス
type BakeSetService struct{}

// NewBakeSetService コンストラクタ
func NewBakeSetService() *BakeSetService {
	return &BakeSetService{}
}

// ProcessPhysicsModel 物理モデルの処理
func (s *BakeSetService) ProcessPhysicsModel(originalModel, bakedModel *pmx.PmxModel) error {
	if originalModel == nil {
		return nil
	}

	s.processPhysicsBones(originalModel)
	s.processPhysicsBones(bakedModel)

	if bakedModel != nil {
		s.fixPhysicsRigidBodies(bakedModel)
	}

	return nil
}

// CreateOutputPath 出力パスの作成
func (s *BakeSetService) CreateOutputPath(basePath, suffix string) string {
	return mfile.CreateOutputPath(basePath, suffix)
}

// processPhysicsBones 物理ボーンの処理（ドメインロジック）
func (s *BakeSetService) processPhysicsBones(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	// 物理ボーンの名前に接頭辞を追加
	s.insertPhysicsBonePrefix(model)
	// 物理ボーンを表示枠に追加
	s.appendPhysicsBoneToDisplaySlots(model)
}

// appendPhysicsBoneToDisplaySlots 物理ボーンを表示枠に追加
func (s *BakeSetService) appendPhysicsBoneToDisplaySlots(model *pmx.PmxModel) {
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
func (s *BakeSetService) insertPhysicsBonePrefix(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	digits := int(math.Log10(float64(model.Bones.Length()))) + 1

	// 物理ボーンの名前に接頭辞を追加
	model.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		if bone.HasPhysics() {
			// ボーンINDEXを0埋めして設定
			formattedBoneName := fmt.Sprintf("PF%0*d_%s", digits, boneIndex, bone.Name())
			bone.SetName(s.encodeName(formattedBoneName, 15))
		}
		return true
	})
}

func (s *BakeSetService) encodeName(name string, limit int) string {
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

func (s *BakeSetService) fixPhysicsRigidBodies(model *pmx.PmxModel) {
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
