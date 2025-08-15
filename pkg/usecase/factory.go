package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/infrastructure"
)

// ServiceFactory アプリケーション全体のサービス生成を担うメインファクトリ
type ServiceFactory struct {
	repositoryFactory      *infrastructure.RepositoryFactory
	domainServiceFactory   *domain.PhysicsBoneServiceFactory
	valueObjectFactory     *domain.ValueObjectFactory
	encodingServiceFactory *domain.BoneNameEncodingServiceFactory
}

// NewServiceFactory メインファクトリのコンストラクタ
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{
		repositoryFactory:      infrastructure.NewRepositoryFactory(),
		domainServiceFactory:   domain.NewPhysicsBoneServiceFactory(),
		valueObjectFactory:     domain.NewValueObjectFactory(),
		encodingServiceFactory: domain.NewBoneNameEncodingServiceFactory(),
	}
}

// CreateBakeUsecase BakeUsecaseを依存関係込みで作成
func (f *ServiceFactory) CreateBakeUsecase() *BakeUsecase {
	repository := f.repositoryFactory.CreateBakeSetRepository()
	return NewBakeUsecase(repository)
}

// CreateBakeUsecaseWithSeparatedInterfaces 分離されたインターフェースでBakeUsecaseを作成
func (f *ServiceFactory) CreateBakeUsecaseWithSeparatedInterfaces() *BakeUsecase {
	reader := f.repositoryFactory.CreateBakeSetReader()
	writer := f.repositoryFactory.CreateBakeSetWriter()
	return NewBakeUsecaseWithSeparatedInterfaces(reader, writer)
}

// CreatePhysicsBoneManager PhysicsBoneManagerを作成
func (f *ServiceFactory) CreatePhysicsBoneManager() domain.PhysicsBoneManager {
	return f.domainServiceFactory.CreatePhysicsBoneManager()
}

// CreatePhysicsBoneProcessor PhysicsBoneProcessorを作成
func (f *ServiceFactory) CreatePhysicsBoneProcessor() domain.PhysicsBoneProcessor {
	return f.domainServiceFactory.CreatePhysicsBoneProcessor()
}

// CreateBoneNameEncodingService BoneNameEncodingServiceを作成
func (f *ServiceFactory) CreateBoneNameEncodingService() *domain.BoneNameEncodingService {
	return f.encodingServiceFactory.CreateBoneNameEncodingService()
}

// CreateFilePath FilePathオブジェクトを作成
func (f *ServiceFactory) CreateFilePath(path string) *domain.FilePath {
	return f.valueObjectFactory.CreateFilePath(path)
}

// CreateFrameRange FrameRangeオブジェクトを作成
func (f *ServiceFactory) CreateFrameRange(startFrame, endFrame float32) (*domain.FrameRange, error) {
	return f.valueObjectFactory.CreateFrameRange(startFrame, endFrame)
}

// GetRepositoryFactory Repositoryファクトリを取得
func (f *ServiceFactory) GetRepositoryFactory() *infrastructure.RepositoryFactory {
	return f.repositoryFactory
}

// GetDomainServiceFactory Domainサービスファクトリを取得
func (f *ServiceFactory) GetDomainServiceFactory() *domain.PhysicsBoneServiceFactory {
	return f.domainServiceFactory
}

// GetValueObjectFactory ValueObjectファクトリを取得
func (f *ServiceFactory) GetValueObjectFactory() *domain.ValueObjectFactory {
	return f.valueObjectFactory
}
