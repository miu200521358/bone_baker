package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/bone_baker/pkg/usecase/dto"
)

// PhysicsUsecase 物理関連のユースケース
type PhysicsUsecase interface {
	CreatePhysicsTree(bakeSet *domain.BakeSet) error
	GetPhysicsTree(bakeSet *domain.BakeSet) (*dto.PhysicsTreeDTO, error)
	UpdateStiffness(bakeSet *domain.BakeSet, itemID string, stiffnessRatio float64) error
	UpdateTension(bakeSet *domain.BakeSet, itemID string, tensionRatio float64) error
	UpdateMass(bakeSet *domain.BakeSet, itemID string, massRatio float64) error
	GetCurrentPhysicsItem(bakeSet *domain.BakeSet, itemID string) (*dto.PhysicsItemDTO, error)
	ResetPhysicsTree(bakeSet *domain.BakeSet) error
}

type physicsUsecase struct {
	bakeSetRepo    repository.BakeSetRepository
	modelRepo      repository.ModelRepository
	bakeSetService *domain.BakeSetService
}

// NewPhysicsUsecase PhysicsUsecaseのコンストラクタ
func NewPhysicsUsecase(
	bakeSetRepo repository.BakeSetRepository,
	modelRepo repository.ModelRepository,
	bakeSetService *domain.BakeSetService,
) PhysicsUsecase {
	return &physicsUsecase{
		bakeSetRepo:    bakeSetRepo,
		modelRepo:      modelRepo,
		bakeSetService: bakeSetService,
	}
}

// CreatePhysicsTree 物理ツリーを作成
func (pu *physicsUsecase) CreatePhysicsTree(bakeSet *domain.BakeSet) error {
	if bakeSet == nil || bakeSet.OriginalModel == nil {
		return nil
	}

	tree := domain.NewPhysicsModel()

	for _, boneIndex := range bakeSet.OriginalModel.Bones.LayerSortedIndexes {
		if bone, err := bakeSet.OriginalModel.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := domain.NewPhysicsItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*domain.PhysicsItem).AddChild(item)
			}
		}
	}

	// 物理ボーンを持つアイテムのみを保存
	tree.SaveOnlyPhysicsItems()

	bakeSet.PhysicsTree = tree
	return nil
}

// GetPhysicsTree 物理ツリーをDTOで取得
func (pu *physicsUsecase) GetPhysicsTree(bakeSet *domain.BakeSet) (*dto.PhysicsTreeDTO, error) {
	if bakeSet == nil || bakeSet.PhysicsTree == nil {
		return &dto.PhysicsTreeDTO{Items: []dto.PhysicsItemDTO{}}, nil
	}

	return pu.convertPhysicsTreeToDTO(bakeSet.PhysicsTree), nil
}

// UpdateStiffness 硬さを更新
func (pu *physicsUsecase) UpdateStiffness(bakeSet *domain.BakeSet, itemID string, stiffnessRatio float64) error {
	if bakeSet == nil || bakeSet.PhysicsTree == nil {
		return nil
	}

	if item := bakeSet.PhysicsTree.GetByID(itemID); item != nil {
		if physicsItem, ok := item.(*domain.PhysicsItem); ok {
			physicsItem.CalcStiffness(stiffnessRatio)
		}
	}

	return nil
}

// UpdateTension 張りを更新
func (pu *physicsUsecase) UpdateTension(bakeSet *domain.BakeSet, itemID string, tensionRatio float64) error {
	if bakeSet == nil || bakeSet.PhysicsTree == nil {
		return nil
	}

	if item := bakeSet.PhysicsTree.GetByID(itemID); item != nil {
		if physicsItem, ok := item.(*domain.PhysicsItem); ok {
			physicsItem.CalcTension(tensionRatio)
		}
	}

	return nil
}

// UpdateMass 質量を更新
func (pu *physicsUsecase) UpdateMass(bakeSet *domain.BakeSet, itemID string, massRatio float64) error {
	if bakeSet == nil || bakeSet.PhysicsTree == nil {
		return nil
	}

	if item := bakeSet.PhysicsTree.GetByID(itemID); item != nil {
		if physicsItem, ok := item.(*domain.PhysicsItem); ok {
			physicsItem.CalcMass(massRatio)
		}
	}

	return nil
}

// GetCurrentPhysicsItem 現在の物理アイテム情報を取得
func (pu *physicsUsecase) GetCurrentPhysicsItem(bakeSet *domain.BakeSet, itemID string) (*dto.PhysicsItemDTO, error) {
	if bakeSet == nil || bakeSet.PhysicsTree == nil {
		return nil, nil
	}

	if item := bakeSet.PhysicsTree.GetByID(itemID); item != nil {
		if physicsItem, ok := item.(*domain.PhysicsItem); ok {
			return &dto.PhysicsItemDTO{
				ID:             physicsItem.Text(),
				Text:           physicsItem.Text(),
				MassRatio:      physicsItem.MassRatio(),
				StiffnessRatio: physicsItem.StiffnessRatio(),
				TensionRatio:   physicsItem.TensionRatio(),
				HasPhysics:     physicsItem.HasPhysicsChild(),
			}, nil
		}
	}

	return nil, nil
}

// ResetPhysicsTree 物理ツリーをリセット
func (pu *physicsUsecase) ResetPhysicsTree(bakeSet *domain.BakeSet) error {
	if bakeSet != nil && bakeSet.PhysicsTree != nil {
		bakeSet.PhysicsTree.Reset()
	}

	return nil
}

// convertPhysicsTreeToDTO Domain → DTO変換
func (pu *physicsUsecase) convertPhysicsTreeToDTO(tree *domain.PhysicsModel) *dto.PhysicsTreeDTO {
	items := make([]dto.PhysicsItemDTO, 0)

	for i := 0; i < tree.RootCount(); i++ {
		if node := tree.RootAt(i); node != nil {
			if physicsItem, ok := node.(*domain.PhysicsItem); ok {
				items = append(items, pu.convertPhysicsItemToDTO(physicsItem))
			}
		}
	}

	return &dto.PhysicsTreeDTO{Items: items}
}

// convertPhysicsItemToDTO 物理アイテムをDTOに変換
func (pu *physicsUsecase) convertPhysicsItemToDTO(item *domain.PhysicsItem) dto.PhysicsItemDTO {
	return dto.PhysicsItemDTO{
		ID:             item.Text(),
		Text:           item.Text(),
		MassRatio:      item.MassRatio(),
		StiffnessRatio: item.StiffnessRatio(),
		TensionRatio:   item.TensionRatio(),
		HasPhysics:     item.HasPhysicsChild(),
	}
}
