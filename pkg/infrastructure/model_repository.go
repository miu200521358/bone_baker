package infrastructure

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

// PmxModelRepository PmxRepositoryのアダプター実装
type PmxModelRepository struct{}

// NewPmxModelRepository コンストラクタ
func NewPmxModelRepository() domain.ModelRepository {
	return &PmxModelRepository{}
}

// LoadWithPhysics 物理設定を考慮したモデル読み込み
func (r *PmxModelRepository) LoadWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error) {
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
