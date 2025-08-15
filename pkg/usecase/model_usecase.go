package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

// ModelUsecase モデル操作専用のユースケース
type ModelUsecase struct {
	pmxRepo repository.PmxRepository
}

// NewModelUsecase コンストラクタ
func NewModelUsecase() *ModelUsecase {
	return &ModelUsecase{}
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

// loadModelWithPhysics 物理設定を考慮したモデル読み込み（内部メソッド）
func (uc *ModelUsecase) loadModelWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error) {
	pmxRep := repository.NewPmxRepository(enablePhysics)
	data, err := pmxRep.Load(path)
	if err != nil {
		mlog.ET(mi18n.T("読み込み失敗"), err, "")
		return nil, err
	}

	model := data.(*pmx.PmxModel)
	if err := model.Bones.InsertShortageOverrideBones(); err != nil {
		mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
		return nil, err
	}

	return model, nil
}
