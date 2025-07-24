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
