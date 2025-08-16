package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

// ModelRepository モデル永続化の抽象インターフェース
type ModelRepository interface {
	// LoadWithPhysics 物理設定を考慮したモデル読み込み
	LoadWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error)
}
