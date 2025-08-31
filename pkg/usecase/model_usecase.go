package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

// ModelUsecase モデル操作専用のユースケース
type ModelUsecase struct {
	modelRepository domain.ModelRepository
}

// NewModelUsecase コンストラクタ
func NewModelUsecase(modelRepository domain.ModelRepository) *ModelUsecase {
	return &ModelUsecase{
		modelRepository: modelRepository,
	}
}

// LoadModelPair 元モデルと焼き込み用モデルのペアを読み込み
func (uc *ModelUsecase) LoadModelPair(path string) (*pmx.PmxModel, *pmx.PmxModel, error) {
	if path == "" {
		return nil, nil, nil
	}

	// 元モデル読み込み（物理有効）
	originalModel, err := uc.loadModelWithPhysics(path, true)
	if err != nil {
		return nil, nil, err
	}

	// 焼き込み用モデル読み込み（物理無効）
	bakedModel, err := uc.loadModelWithPhysics(path, false)
	if err != nil {
		return nil, nil, err
	}

	return originalModel, bakedModel, nil
}

// SetModelsInBakeSet BakeSetにモデルを設定
func (uc *ModelUsecase) SetModelsInBakeSet(bakeSet *domain.BakeSet, originalModel, bakedModel *pmx.PmxModel) error {
	return bakeSet.SetModels(originalModel, bakedModel)
}

// ClearModelsInBakeSet BakeSetのモデルをクリア
func (uc *ModelUsecase) ClearModelsInBakeSet(bakeSet *domain.BakeSet) {
	bakeSet.ClearModels()
}

// loadModelWithPhysics モデル読み込み（内部メソッド）
func (uc *ModelUsecase) loadModelWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error) {
	return uc.modelRepository.Load(path, enablePhysics)
}
