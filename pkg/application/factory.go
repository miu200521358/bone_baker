package application

import (
	"github.com/miu200521358/bone_baker/pkg/usecase"
)

// ApplicationServiceFactory Application Serviceの生成を担うファクトリ
type ApplicationServiceFactory struct{}

// NewApplicationServiceFactory ファクトリのコンストラクタ
func NewApplicationServiceFactory() *ApplicationServiceFactory {
	return &ApplicationServiceFactory{}
}

// CreateBakeApplicationService BakeApplicationServiceインターフェースの実装を作成
func (f *ApplicationServiceFactory) CreateBakeApplicationService(
	bakeUsecase *usecase.BakeUsecase,
) BakeApplicationServiceInterface {
	return NewBakeApplicationService(bakeUsecase)
}

// CreateBakeApplicationServiceConcrete 具象型のBakeApplicationServiceを作成
// UI層で具象型が必要な場合に使用
func (f *ApplicationServiceFactory) CreateBakeApplicationServiceConcrete(
	bakeUsecase *usecase.BakeUsecase,
) *BakeApplicationService {
	return &BakeApplicationService{bakeUsecase: bakeUsecase}
}
