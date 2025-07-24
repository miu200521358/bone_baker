package repository

import (
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	mlibRepo "github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

type modelRepositoryImpl struct{}

// NewModelRepository ModelRepositoryの実装を返す
func NewModelRepository() repository.ModelRepository {
	return &modelRepositoryImpl{}
}

func (r *modelRepositoryImpl) LoadWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error) {
	pmxRep := mlibRepo.NewPmxRepository(enablePhysics)
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

func (r *modelRepositoryImpl) Save(path string, model *pmx.PmxModel) error {
	pmxRep := mlibRepo.NewPmxRepository(true)
	return pmxRep.Save(path, model, false)
}
