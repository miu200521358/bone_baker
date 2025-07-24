package view_model

import "github.com/miu200521358/bone_baker/pkg/usecase/dto"

// PhysicsViewModel 物理ツリーのビューモデル
type PhysicsViewModel struct {
	currentBakeSetID int
	currentItemID    string
	stiffnessRatio   float64
	tensionRatio     float64
	massRatio        float64
	physicsTree      *dto.PhysicsTreeDTO
}

// NewPhysicsViewModel PhysicsViewModelのコンストラクタ
func NewPhysicsViewModel() *PhysicsViewModel {
	return &PhysicsViewModel{
		stiffnessRatio: 1.0,
		tensionRatio:   1.0,
		massRatio:      1.0,
		physicsTree:    &dto.PhysicsTreeDTO{Items: []dto.PhysicsItemDTO{}},
	}
}

// SetCurrentItem 現在選択中のアイテムを設定
func (pvm *PhysicsViewModel) SetCurrentItem(bakeSetID int, itemID string) {
	pvm.currentBakeSetID = bakeSetID
	pvm.currentItemID = itemID
}

// GetCurrentBakeSetID 現在のBakeSetIDを取得
func (pvm *PhysicsViewModel) GetCurrentBakeSetID() int {
	return pvm.currentBakeSetID
}

// GetCurrentItemID 現在のアイテムIDを取得
func (pvm *PhysicsViewModel) GetCurrentItemID() string {
	return pvm.currentItemID
}

// SetStiffnessRatio 硬さ比率を設定
func (pvm *PhysicsViewModel) SetStiffnessRatio(ratio float64) {
	pvm.stiffnessRatio = ratio
}

// GetStiffnessRatio 硬さ比率を取得
func (pvm *PhysicsViewModel) GetStiffnessRatio() float64 {
	return pvm.stiffnessRatio
}

// SetTensionRatio 張り比率を設定
func (pvm *PhysicsViewModel) SetTensionRatio(ratio float64) {
	pvm.tensionRatio = ratio
}

// GetTensionRatio 張り比率を取得
func (pvm *PhysicsViewModel) GetTensionRatio() float64 {
	return pvm.tensionRatio
}

// SetMassRatio 質量比率を設定
func (pvm *PhysicsViewModel) SetMassRatio(ratio float64) {
	pvm.massRatio = ratio
}

// GetMassRatio 質量比率を取得
func (pvm *PhysicsViewModel) GetMassRatio() float64 {
	return pvm.massRatio
}

// SetPhysicsTree 物理ツリーを設定
func (pvm *PhysicsViewModel) SetPhysicsTree(tree *dto.PhysicsTreeDTO) {
	pvm.physicsTree = tree
}

// GetPhysicsTree 物理ツリーを取得
func (pvm *PhysicsViewModel) GetPhysicsTree() *dto.PhysicsTreeDTO {
	return pvm.physicsTree
}
