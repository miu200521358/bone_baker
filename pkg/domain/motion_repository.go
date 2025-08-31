package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

// MotionRepository モーション永続化の抽象インターフェース
type MotionRepository interface {
	// Load モーション読み込み
	Load(path string, enablePhysics bool) (*vmd.VmdMotion, error)
}
