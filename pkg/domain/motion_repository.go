package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

// MotionRepository モーション永続化の抽象インターフェース
type MotionRepository interface {
	// LoadWithPhysics 物理設定を考慮したモーション読み込み
	LoadWithPhysics(path string, enablePhysics bool) (*vmd.VmdMotion, error)
}
