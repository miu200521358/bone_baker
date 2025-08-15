package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
)

type BakeUsecase struct {
	modelUsecase      *ModelUsecase
	motionUsecase     *MotionUsecase
	bakeSetRepository domain.BakeSetRepository
}

func NewBakeUsecase(bakeSetRepository domain.BakeSetRepository) *BakeUsecase {
	return &BakeUsecase{
		modelUsecase:      NewModelUsecase(),
		motionUsecase:     NewMotionUsecase(),
		bakeSetRepository: bakeSetRepository,
	}
}

// LoadModelForBakeSet BakeSet用モデル読み込みのビジネスロジック
func (uc *BakeUsecase) LoadModelForBakeSet(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		uc.modelUsecase.ClearModelsInBakeSet(bakeSet)
		return nil
	}

	// モデルペア読み込み
	originalModel, bakedModel, err := uc.modelUsecase.LoadModelPair(path)
	if err != nil {
		return err
	}

	// BakeSetにモデル設定
	return uc.modelUsecase.SetModelsInBakeSet(bakeSet, originalModel, bakedModel)
}

// LoadMotionForBakeSet BakeSet用モーション読み込みのビジネスロジック
func (uc *BakeUsecase) LoadMotionForBakeSet(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		uc.motionUsecase.ClearMotionsInBakeSet(bakeSet)
		return nil
	}

	// モーションペア読み込み
	originalMotion, outputMotion, err := uc.motionUsecase.LoadMotionPair(path)
	if err != nil {
		return err
	}

	// BakeSetにモーション設定
	return uc.motionUsecase.SetMotionsInBakeSet(bakeSet, originalMotion, outputMotion)
}

// SaveBakeSet セット保存のビジネスロジック
func (uc *BakeUsecase) SaveBakeSet(bakeSets []*domain.BakeSet, jsonPath string) error {
	return uc.bakeSetRepository.Save(bakeSets, jsonPath)
}

// LoadBakeSet セット読み込みのビジネスロジック
func (uc *BakeUsecase) LoadBakeSet(jsonPath string) ([]*domain.BakeSet, error) {
	return uc.bakeSetRepository.Load(jsonPath)
}
