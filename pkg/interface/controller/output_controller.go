package controller

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
	"github.com/miu200521358/bone_baker/pkg/usecase/dto"
)

// OutputController 出力関連のコントローラー
type OutputController struct {
	outputUsecase usecase.OutputUsecase
}

// NewOutputController OutputControllerのコンストラクタ
func NewOutputController(outputUsecase usecase.OutputUsecase) *OutputController {
	return &OutputController{
		outputUsecase: outputUsecase,
	}
}

// CreateOutputTree 出力ツリーを作成
func (oc *OutputController) CreateOutputTree(bakeSet *domain.BakeSet) error {
	return oc.outputUsecase.CreateOutputTree(bakeSet)
}

// GetOutputTree 出力ツリーを取得
func (oc *OutputController) GetOutputTree(bakeSet *domain.BakeSet) (*dto.OutputTreeDTO, error) {
	return oc.outputUsecase.GetOutputTree(bakeSet)
}

// SetChildrenChecked 子要素のチェック状態を設定
func (oc *OutputController) SetChildrenChecked(bakeSet *domain.BakeSet, itemID string, checked bool) error {
	return oc.outputUsecase.SetChildrenChecked(bakeSet, itemID, checked)
}

// SetIkChecked IKボーンのチェック状態を設定
func (oc *OutputController) SetIkChecked(bakeSet *domain.BakeSet, checked bool) error {
	return oc.outputUsecase.SetIkChecked(bakeSet, checked)
}

// SetPhysicsChecked 物理ボーンのチェック状態を設定
func (oc *OutputController) SetPhysicsChecked(bakeSet *domain.BakeSet, checked bool) error {
	return oc.outputUsecase.SetPhysicsChecked(bakeSet, checked)
}
