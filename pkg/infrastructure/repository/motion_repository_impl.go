package repository

import (
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	mlibRepo "github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

type motionRepositoryImpl struct{}

// NewMotionRepository MotionRepositoryの実装を返す
func NewMotionRepository() repository.MotionRepository {
	return &motionRepositoryImpl{}
}

func (r *motionRepositoryImpl) Load(path string, enableOverride bool) (*vmd.VmdMotion, error) {
	vmdRep := mlibRepo.NewVmdVpdRepository(enableOverride)
	data, err := vmdRep.Load(path)
	if err != nil {
		mlog.ET(mi18n.T("読み込み失敗"), err, "")
		return nil, err
	}

	return data.(*vmd.VmdMotion), nil
}

func (r *motionRepositoryImpl) Save(path string, motion *vmd.VmdMotion) error {
	vmdRep := mlibRepo.NewVmdVpdRepository(false)
	return vmdRep.Save(path, motion, false)
}
