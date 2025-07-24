package repository

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
)

// BakeSetRepository BakeSetのリポジトリインターフェース
type BakeSetRepository interface {
	Save(bakeSets []*domain.BakeSet, jsonPath string) error
	Load(jsonPath string) ([]*domain.BakeSet, error)
	GetByID(id int) (*domain.BakeSet, error)
	SaveSingle(bakeSet *domain.BakeSet) error
}
