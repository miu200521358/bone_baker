package usecase

import (
	"sync"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

type BakeUsecase struct {
	modelRepo      repository.ModelRepository
	motionRepo     repository.MotionRepository
	bakeSetRepo    repository.BakeSetRepository
	bakeSetService *domain.BakeSetService
}

func NewBakeUsecase(
	modelRepo repository.ModelRepository,
	motionRepo repository.MotionRepository,
	bakeSetRepo repository.BakeSetRepository,
	bakeSetService *domain.BakeSetService,
) *BakeUsecase {
	return &BakeUsecase{
		modelRepo:      modelRepo,
		motionRepo:     motionRepo,
		bakeSetRepo:    bakeSetRepo,
		bakeSetService: bakeSetService,
	}
}

// LoadModel モデル読み込みのビジネスロジック
func (uc *BakeUsecase) LoadModel(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearModels()
		return nil
	}

	// 元モデル読み込み（物理有効）
	originalModel, err := uc.modelRepo.LoadWithPhysics(path, true)
	if err != nil {
		return err
	}

	// 焼き込み用モデル読み込み（物理無効）
	bakedModel, err := uc.modelRepo.LoadWithPhysics(path, false)
	if err != nil {
		return err
	}

	// ドメインサービスを使ってビジネスロジックを実行
	if err := uc.bakeSetService.ProcessPhysicsModel(originalModel, bakedModel); err != nil {
		return err
	}

	// モデルを設定
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

		if motion, err := uc.motionRepo.Load(path, false); err == nil {
			originalMotion = motion
		} else {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if motion, err := uc.motionRepo.Load(path, true); err == nil {
			outputMotion = motion
		} else {
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
	return uc.bakeSetRepo.Save(bakeSets, jsonPath)
}

// LoadBakeSet セット読み込みのビジネスロジック
func (uc *BakeUsecase) LoadBakeSet(jsonPath string) ([]*domain.BakeSet, error) {
	return uc.bakeSetRepo.Load(jsonPath)
}

// ExportMotions モーション出力のビジネスロジック
func (uc *BakeUsecase) ExportMotions(bakeSet *domain.BakeSet, startFrame, endFrame float64) error {
	motions, err := bakeSet.GetOutputMotionOnlyChecked(startFrame, endFrame)
	if err != nil {
		return err
	}

	for _, motion := range motions {
		if err := uc.motionRepo.Save(motion.Path(), motion); err != nil {
			return err
		}
	}

	return nil
}
