package controller

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
)

// BakeController BakeのController
type BakeController struct {
	bakeUsecase *usecase.BakeUsecase
}

// NewBakeController コンストラクタ
func NewBakeController(bakeUsecase *usecase.BakeUsecase) *BakeController {
	return &BakeController{
		bakeUsecase: bakeUsecase,
	}
}

// LoadModel モデル読み込み
func (c *BakeController) LoadModel(bakeSet *domain.BakeSet, path string) error {
	return c.bakeUsecase.LoadModel(bakeSet, path)
}

// LoadMotion モーション読み込み
func (c *BakeController) LoadMotion(bakeSet *domain.BakeSet, path string) error {
	return c.bakeUsecase.LoadMotion(bakeSet, path)
}

// SaveBakeSet BakeSet保存
func (c *BakeController) SaveBakeSet(bakeSets []*domain.BakeSet, jsonPath string) error {
	return c.bakeUsecase.SaveBakeSet(bakeSets, jsonPath)
}

// LoadBakeSet BakeSet読み込み
func (c *BakeController) LoadBakeSet(jsonPath string) ([]*domain.BakeSet, error) {
	return c.bakeUsecase.LoadBakeSet(jsonPath)
}

// ExportMotions モーション出力
func (c *BakeController) ExportMotions(bakeSet *domain.BakeSet, startFrame, endFrame float64) error {
	return c.bakeUsecase.ExportMotions(bakeSet, startFrame, endFrame)
}

// CreatePhysicsTree 物理ツリー作成
func (c *BakeController) CreatePhysicsTree(bakeSet *domain.BakeSet) error {
	return c.bakeUsecase.CreatePhysicsTree(bakeSet)
}

// CreateOutputTree 出力ツリー作成
func (c *BakeController) CreateOutputTree(bakeSet *domain.BakeSet) error {
	return c.bakeUsecase.CreateOutputTree(bakeSet)
}

// UpdatePhysicsStiffness 物理パラメータ更新（硬さ）
func (c *BakeController) UpdatePhysicsStiffness(bakeSet *domain.BakeSet, itemID string, stiffnessRatio float64) error {
	return c.bakeUsecase.UpdatePhysicsStiffness(bakeSet, itemID, stiffnessRatio)
}

// UpdatePhysicsTension 物理パラメータ更新（張り）
func (c *BakeController) UpdatePhysicsTension(bakeSet *domain.BakeSet, itemID string, tensionRatio float64) error {
	return c.bakeUsecase.UpdatePhysicsTension(bakeSet, itemID, tensionRatio)
}

// SetOutputChildrenChecked 出力ツリーの子要素チェック状態更新
func (c *BakeController) SetOutputChildrenChecked(bakeSet *domain.BakeSet, itemID string, checked bool) error {
	return c.bakeUsecase.SetOutputChildrenChecked(bakeSet, itemID, checked)
}

// SetOutputIkChecked 出力ツリーのIKチェック状態更新
func (c *BakeController) SetOutputIkChecked(bakeSet *domain.BakeSet, checked bool) error {
	return c.bakeUsecase.SetOutputIkChecked(bakeSet, checked)
}

// SetOutputPhysicsChecked 出力ツリーの物理チェック状態更新
func (c *BakeController) SetOutputPhysicsChecked(bakeSet *domain.BakeSet, checked bool) error {
	return c.bakeUsecase.SetOutputPhysicsChecked(bakeSet, checked)
}
