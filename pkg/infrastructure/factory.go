package infrastructure

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
)

// RepositoryFactory Infrastructureレイヤーのリポジトリ生成を担うファクトリ
type RepositoryFactory struct{}

// NewRepositoryFactory ファクトリのコンストラクタ
func NewRepositoryFactory() *RepositoryFactory {
	return &RepositoryFactory{}
}

// CreateBakeSetRepository BakeSetRepositoryの実装を作成
func (f *RepositoryFactory) CreateBakeSetRepository() domain.BakeSetRepository {
	return NewFileBakeSetRepository()
}

// CreateBakeSetReader BakeSetReader専用の実装を作成
func (f *RepositoryFactory) CreateBakeSetReader() domain.BakeSetReader {
	return NewFileBakeSetRepository()
}

// CreateBakeSetWriter BakeSetWriter専用の実装を作成
func (f *RepositoryFactory) CreateBakeSetWriter() domain.BakeSetWriter {
	return NewFileBakeSetRepository()
}

// CreateBakeSetPersistenceManager BakeSetPersistenceManagerの実装を作成
func (f *RepositoryFactory) CreateBakeSetPersistenceManager() domain.BakeSetPersistenceManager {
	return NewFileBakeSetRepository()
}
