package usecase

import (
	"sync"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

// MotionUsecase モーション操作専用のユースケース
type MotionUsecase struct {
	motionRepository domain.MotionRepository
}

// NewMotionUsecase コンストラクタ
func NewMotionUsecase(motionRepository domain.MotionRepository) *MotionUsecase {
	return &MotionUsecase{
		motionRepository: motionRepository,
	}
}

// LoadMotionPair 元モーションと出力モーションのペアを読み込み
func (uc *MotionUsecase) LoadMotionPair(path string) (*vmd.VmdMotion, *vmd.VmdMotion, error) {
	if path == "" {
		return nil, nil, nil
	}

	var wg sync.WaitGroup
	var originalMotion, outputMotion *vmd.VmdMotion
	errChan := make(chan error, 2)

	// 元モーション読み込み（物理無効）
	wg.Add(1)
	go func() {
		defer wg.Done()
		if motion, err := uc.motionRepository.Load(path, false); err == nil {
			originalMotion = motion
		} else {
			errChan <- err
		}
	}()

	// 出力モーション読み込み（物理有効）
	wg.Add(1)
	go func() {
		defer wg.Done()
		if motion, err := uc.motionRepository.Load(path, true); err == nil {
			outputMotion = motion
		} else {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, nil, err
		}
	}

	return originalMotion, outputMotion, nil
}

// SetMotionsInBakeSet BakeSetにモーションを設定
func (uc *MotionUsecase) SetMotionsInBakeSet(bakeSet *domain.BakeSet, originalMotion, outputMotion *vmd.VmdMotion) error {
	return bakeSet.SetMotions(originalMotion, outputMotion)
}

// ClearMotionsInBakeSet BakeSetのモーションをクリア
func (uc *MotionUsecase) ClearMotionsInBakeSet(bakeSet *domain.BakeSet) {
	bakeSet.ClearMotions()
}
