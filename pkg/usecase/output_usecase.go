package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/bone_baker/pkg/usecase/dto"
)

// OutputUsecase 出力関連のユースケース
type OutputUsecase interface {
	CreateOutputTree(bakeSet *domain.BakeSet) error
	GetOutputTree(bakeSet *domain.BakeSet) (*dto.OutputTreeDTO, error)
	SetChildrenChecked(bakeSet *domain.BakeSet, itemID string, checked bool) error
	SetIkChecked(bakeSet *domain.BakeSet, checked bool) error
	SetPhysicsChecked(bakeSet *domain.BakeSet, checked bool) error
}

type outputUsecase struct {
	bakeSetRepo    repository.BakeSetRepository
	modelRepo      repository.ModelRepository
	bakeSetService *domain.BakeSetService
}

// NewOutputUsecase OutputUsecaseのコンストラクタ
func NewOutputUsecase(
	bakeSetRepo repository.BakeSetRepository,
	modelRepo repository.ModelRepository,
	bakeSetService *domain.BakeSetService,
) OutputUsecase {
	return &outputUsecase{
		bakeSetRepo:    bakeSetRepo,
		modelRepo:      modelRepo,
		bakeSetService: bakeSetService,
	}
}

// CreateOutputTree 出力ツリーを作成
func (ou *outputUsecase) CreateOutputTree(bakeSet *domain.BakeSet) error {
	if bakeSet == nil || bakeSet.OriginalModel == nil {
		return nil
	}

	tree := domain.NewOutputModel()

	for _, boneIndex := range bakeSet.OriginalModel.Bones.LayerSortedIndexes {
		if bone, err := bakeSet.OriginalModel.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := domain.NewOutputItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*domain.OutputItem).AddChild(item)
			}
		}
	}

	bakeSet.OutputTree = tree
	return nil
}

// GetOutputTree 出力ツリーをDTOで取得
func (ou *outputUsecase) GetOutputTree(bakeSet *domain.BakeSet) (*dto.OutputTreeDTO, error) {
	if bakeSet == nil || bakeSet.OutputTree == nil {
		return &dto.OutputTreeDTO{Items: []dto.OutputItemDTO{}}, nil
	}

	return ou.convertOutputTreeToDTO(bakeSet.OutputTree), nil
}

// SetChildrenChecked 子要素のチェック状態を設定
func (ou *outputUsecase) SetChildrenChecked(bakeSet *domain.BakeSet, itemID string, checked bool) error {
	if bakeSet == nil || bakeSet.OutputTree == nil {
		return nil
	}

	if item := bakeSet.OutputTree.GetByID(itemID); item != nil {
		if outputItem, ok := item.(*domain.OutputItem); ok {
			ou.setChildrenCheckedRecursive(outputItem, checked)
		}
	}

	return nil
}

// SetIkChecked IKボーンのチェック状態を設定
func (ou *outputUsecase) SetIkChecked(bakeSet *domain.BakeSet, checked bool) error {
	if bakeSet == nil {
		return nil
	}

	return ou.bakeSetService.UpdateIkBoneChecked(bakeSet, checked)
}

// SetPhysicsChecked 物理ボーンのチェック状態を設定
func (ou *outputUsecase) SetPhysicsChecked(bakeSet *domain.BakeSet, checked bool) error {
	if bakeSet == nil {
		return nil
	}

	return ou.bakeSetService.UpdatePhysicsBoneChecked(bakeSet, checked)
}

// setChildrenCheckedRecursive 再帰的に子要素のチェック状態を設定
func (ou *outputUsecase) setChildrenCheckedRecursive(item *domain.OutputItem, checked bool) {
	item.SetChecked(checked)

	for _, child := range item.Children() {
		if childItem, ok := child.(*domain.OutputItem); ok {
			ou.setChildrenCheckedRecursive(childItem, checked)
		}
	}
}

// convertOutputTreeToDTO Domain → DTO変換
func (ou *outputUsecase) convertOutputTreeToDTO(tree *domain.OutputModel) *dto.OutputTreeDTO {
	items := make([]dto.OutputItemDTO, 0)

	for _, node := range tree.GetAllNodes() {
		items = append(items, ou.convertOutputItemToDTO(node))
	}

	return &dto.OutputTreeDTO{Items: items}
}

// convertOutputItemToDTO 出力アイテムをDTOに変換
func (ou *outputUsecase) convertOutputItemToDTO(item *domain.OutputItem) dto.OutputItemDTO {
	children := make([]dto.OutputItemDTO, 0)
	for _, child := range item.Children() {
		if childItem, ok := child.(*domain.OutputItem); ok {
			children = append(children, ou.convertOutputItemToDTO(childItem))
		}
	}

	return dto.OutputItemDTO{
		ID:        item.Text(),
		Text:      item.Text(),
		Checked:   item.Checked(),
		IsIK:      item.AsIk(),
		IsPhysics: item.AsPhysics(),
		Children:  children,
	}
}
