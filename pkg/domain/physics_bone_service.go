package domain

import (
	"fmt"
	"math"

	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

// PhysicsBoneService 物理ボーン処理に関するドメインサービス
// PhysicsBoneManagerインターフェースを実装
type PhysicsBoneService struct{}

// NewPhysicsBoneService コンストラクタ
func NewPhysicsBoneService() *PhysicsBoneService {
	return &PhysicsBoneService{}
}

// NewPhysicsBoneManager PhysicsBoneManagerインターフェースとしてのコンストラクタ
func NewPhysicsBoneManager() PhysicsBoneManager {
	return &PhysicsBoneService{}
}

// NewPhysicsBoneProcessor PhysicsBoneProcessorインターフェースとしてのコンストラクタ
func NewPhysicsBoneProcessor() PhysicsBoneProcessor {
	return &PhysicsBoneService{}
}

// NewPhysicsBoneDisplayManager PhysicsBoneDisplayManagerインターフェースとしてのコンストラクタ
func NewPhysicsBoneDisplayManager() PhysicsBoneDisplayManager {
	return &PhysicsBoneService{}
}

// NewPhysicsBoneNamer PhysicsBoneNamerインターフェースとしてのコンストラクタ
func NewPhysicsBoneNamer() PhysicsBoneNamer {
	return &PhysicsBoneService{}
}

// NewPhysicsRigidBodyFixer PhysicsRigidBodyFixerインターフェースとしてのコンストラクタ
func NewPhysicsRigidBodyFixer() PhysicsRigidBodyFixer {
	return &PhysicsBoneService{}
}

// ProcessPhysicsBones 物理ボーンの全処理を実行
func (s *PhysicsBoneService) ProcessPhysicsBones(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	// 物理ボーンの名前に接頭辞を追加
	s.InsertPhysicsBonePrefix(model)
	// 物理ボーンを表示枠に追加
	s.AppendPhysicsBoneToDisplaySlots(model)
}

// AppendPhysicsBoneToDisplaySlots 物理ボーンを表示枠に追加
func (s *PhysicsBoneService) AppendPhysicsBoneToDisplaySlots(model *pmx.PmxModel) {
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

// InsertPhysicsBonePrefix 物理ボーンの名前に接頭辞を追加
func (s *PhysicsBoneService) InsertPhysicsBonePrefix(model *pmx.PmxModel) {
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
			encodingService := NewBoneNameEncodingService()
			bone.SetName(encodingService.EncodeName(formattedBoneName, 15))
		}
		return true
	})

	model.Bones.UpdateNameIndexes()
}

// FixPhysicsRigidBodies 物理剛体を修正
func (s *PhysicsBoneService) FixPhysicsRigidBodies(model *pmx.PmxModel) {
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
