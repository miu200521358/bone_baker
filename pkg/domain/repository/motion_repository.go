package repository

import (
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

// MotionRepository モーションのリポジトリインターフェース
type MotionRepository interface {
	Load(path string, enableOverride bool) (*vmd.VmdMotion, error)
	Save(path string, motion *vmd.VmdMotion) error
}
