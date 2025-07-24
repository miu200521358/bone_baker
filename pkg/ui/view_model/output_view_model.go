package view_model

import "github.com/miu200521358/bone_baker/pkg/usecase/dto"

// OutputViewModel 出力ツリーのビューモデル
type OutputViewModel struct {
	currentBakeSetID   int
	outputTree         *dto.OutputTreeDTO
	isUpdatingChildren bool
	isUpdatingIk       bool
	isUpdatingPhysics  bool
	ikChecked          bool
	physicsChecked     bool
}

// NewOutputViewModel OutputViewModelのコンストラクタ
func NewOutputViewModel() *OutputViewModel {
	return &OutputViewModel{
		outputTree: &dto.OutputTreeDTO{Items: []dto.OutputItemDTO{}},
	}
}

// SetCurrentBakeSetID 現在のBakeSetIDを設定
func (ovm *OutputViewModel) SetCurrentBakeSetID(bakeSetID int) {
	ovm.currentBakeSetID = bakeSetID
}

// GetCurrentBakeSetID 現在のBakeSetIDを取得
func (ovm *OutputViewModel) GetCurrentBakeSetID() int {
	return ovm.currentBakeSetID
}

// SetOutputTree 出力ツリーを設定
func (ovm *OutputViewModel) SetOutputTree(tree *dto.OutputTreeDTO) {
	ovm.outputTree = tree
}

// GetOutputTree 出力ツリーを取得
func (ovm *OutputViewModel) GetOutputTree() *dto.OutputTreeDTO {
	return ovm.outputTree
}

// SetIsUpdatingChildren 子要素更新中フラグを設定
func (ovm *OutputViewModel) SetIsUpdatingChildren(updating bool) {
	ovm.isUpdatingChildren = updating
}

// IsUpdatingChildren 子要素更新中かどうか
func (ovm *OutputViewModel) IsUpdatingChildren() bool {
	return ovm.isUpdatingChildren
}

// SetIsUpdatingIk IK更新中フラグを設定
func (ovm *OutputViewModel) SetIsUpdatingIk(updating bool) {
	ovm.isUpdatingIk = updating
}

// IsUpdatingIk IK更新中かどうか
func (ovm *OutputViewModel) IsUpdatingIk() bool {
	return ovm.isUpdatingIk
}

// SetIsUpdatingPhysics 物理更新中フラグを設定
func (ovm *OutputViewModel) SetIsUpdatingPhysics(updating bool) {
	ovm.isUpdatingPhysics = updating
}

// IsUpdatingPhysics 物理更新中かどうか
func (ovm *OutputViewModel) IsUpdatingPhysics() bool {
	return ovm.isUpdatingPhysics
}

// SetIkChecked IKチェック状態を設定
func (ovm *OutputViewModel) SetIkChecked(checked bool) {
	ovm.ikChecked = checked
}

// IsIkChecked IKチェック状態を取得
func (ovm *OutputViewModel) IsIkChecked() bool {
	return ovm.ikChecked
}

// SetPhysicsChecked 物理チェック状態を設定
func (ovm *OutputViewModel) SetPhysicsChecked(checked bool) {
	ovm.physicsChecked = checked
}

// IsPhysicsChecked 物理チェック状態を取得
func (ovm *OutputViewModel) IsPhysicsChecked() bool {
	return ovm.physicsChecked
}
