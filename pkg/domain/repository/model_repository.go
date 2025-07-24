package repository

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

// ModelRepository モデルのリポジトリインターフェース
type ModelRepository interface {
	LoadWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error)
	Save(path string, model *pmx.PmxModel) error
}
