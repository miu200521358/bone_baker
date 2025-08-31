package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

// ModelRepository モデル永続化の抽象インターフェース
type ModelRepository interface {
	// Load モデル読み込み
	Load(path string, enablePhysics bool) (*pmx.PmxModel, error)
}
