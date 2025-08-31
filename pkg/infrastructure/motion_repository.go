package infrastructure

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

// VmdMotionRepository VmdRepositoryのアダプター実装
type VmdMotionRepository struct{}

// NewVmdMotionRepository コンストラクタ
func NewVmdMotionRepository() domain.MotionRepository {
	return &VmdMotionRepository{}
}

// Load モーション読み込み
func (r *VmdMotionRepository) Load(path string, isLog bool) (*vmd.VmdMotion, error) {
	vmdRep := repository.NewVmdVpdRepository(isLog)
	data, err := vmdRep.Load(path)
	if err != nil {
		mlog.ET(mi18n.T("読み込み失敗"), err, "")
		return nil, err
	}

	motion := data.(*vmd.VmdMotion)
	return motion, nil
}
