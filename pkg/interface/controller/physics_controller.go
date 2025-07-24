package controller

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
	"github.com/miu200521358/bone_baker/pkg/usecase/dto"
)

// PhysicsController 物理関連のコントローラー
type PhysicsController struct {
	physicsUsecase usecase.PhysicsUsecase
}

// NewPhysicsController PhysicsControllerのコンストラクタ
func NewPhysicsController(physicsUsecase usecase.PhysicsUsecase) *PhysicsController {
	return &PhysicsController{
		physicsUsecase: physicsUsecase,
	}
}

// CreatePhysicsTree 物理ツリーを作成
func (pc *PhysicsController) CreatePhysicsTree(bakeSet *domain.BakeSet) error {
	return pc.physicsUsecase.CreatePhysicsTree(bakeSet)
}

// GetPhysicsTree 物理ツリーを取得
func (pc *PhysicsController) GetPhysicsTree(bakeSet *domain.BakeSet) (*dto.PhysicsTreeDTO, error) {
	return pc.physicsUsecase.GetPhysicsTree(bakeSet)
}

// UpdateStiffness 硬さを更新
func (pc *PhysicsController) UpdateStiffness(bakeSet *domain.BakeSet, itemID string, stiffnessRatio float64) error {
	return pc.physicsUsecase.UpdateStiffness(bakeSet, itemID, stiffnessRatio)
}

// UpdateTension 張りを更新
func (pc *PhysicsController) UpdateTension(bakeSet *domain.BakeSet, itemID string, tensionRatio float64) error {
	return pc.physicsUsecase.UpdateTension(bakeSet, itemID, tensionRatio)
}

// UpdateMass 質量を更新
func (pc *PhysicsController) UpdateMass(bakeSet *domain.BakeSet, itemID string, massRatio float64) error {
	return pc.physicsUsecase.UpdateMass(bakeSet, itemID, massRatio)
}

// GetCurrentPhysicsItem 現在の物理アイテム情報を取得
func (pc *PhysicsController) GetCurrentPhysicsItem(bakeSet *domain.BakeSet, itemID string) (*dto.PhysicsItemDTO, error) {
	return pc.physicsUsecase.GetCurrentPhysicsItem(bakeSet, itemID)
}

// ResetPhysicsTree 物理ツリーをリセット
func (pc *PhysicsController) ResetPhysicsTree(bakeSet *domain.BakeSet) error {
	return pc.physicsUsecase.ResetPhysicsTree(bakeSet)
}
