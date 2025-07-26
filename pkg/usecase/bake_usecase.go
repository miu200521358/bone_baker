package usecase

import (
	"sync"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

type BakeUsecase struct {
	pmxRepo repository.PmxRepository
	vmdRepo repository.VmdRepository
}

func NewBakeUsecase() *BakeUsecase {
	return &BakeUsecase{}
}

// LoadModel モデル読み込みのビジネスロジック
func (uc *BakeUsecase) LoadModel(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearModels()
		return nil
	}

	// 元モデル読み込み（物理有効）
	originalModel, err := uc.loadModelWithPhysics(path, true)
	if err != nil {
		return err
	}

	// 焼き込み用モデル読み込み（物理無効）
	bakedModel, err := uc.loadModelWithPhysics(path, false)
	if err != nil {
		return err
	}

	// ドメインロジックを使ってモデルを設定
	return bakeSet.SetModels(originalModel, bakedModel)
}

// LoadMotion モーション読み込みのビジネスロジック
func (uc *BakeUsecase) LoadMotion(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearMotions()
		return nil
	}

	var wg sync.WaitGroup
	var originalMotion, outputMotion *vmd.VmdMotion
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		vmdRep := repository.NewVmdVpdRepository(false)
		if data, err := vmdRep.Load(path); err == nil {
			originalMotion = data.(*vmd.VmdMotion)
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		vmdRep := repository.NewVmdVpdRepository(true)
		if data, err := vmdRep.Load(path); err == nil {
			outputMotion = data.(*vmd.VmdMotion)
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return bakeSet.SetMotions(originalMotion, outputMotion)
}

// SaveBakeSet セット保存のビジネスロジック
func (uc *BakeUsecase) SaveBakeSet(bakeSets []*domain.BakeSet, jsonPath string) error {
	return domain.SaveBakeSets(bakeSets, jsonPath)
}

// LoadBakeSet セット読み込みのビジネスロジック
func (uc *BakeUsecase) LoadBakeSet(jsonPath string) ([]*domain.BakeSet, error) {
	return domain.LoadBakeSets(jsonPath)
}

func (uc *BakeUsecase) loadModelWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error) {
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
