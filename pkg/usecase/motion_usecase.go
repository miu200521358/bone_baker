package usecase

import (
	"sync"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

// MotionUsecase モーション操作専用のユースケース
type MotionUsecase struct {
	vmdRepo repository.VmdRepository
}

// NewMotionUsecase コンストラクタ
func NewMotionUsecase() *MotionUsecase {
	return &MotionUsecase{}
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

		vmdRep := repository.NewVmdVpdRepository(false)
		if data, err := vmdRep.Load(path); err == nil {
			originalMotion = data.(*vmd.VmdMotion)
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	// 出力モーション読み込み（物理有効）
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
