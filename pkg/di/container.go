package di

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	domainRepository "github.com/miu200521358/bone_baker/pkg/domain/repository"
	infrastructureRepository "github.com/miu200521358/bone_baker/pkg/infrastructure/repository"
	"github.com/miu200521358/bone_baker/pkg/interface/controller"
	bakeController "github.com/miu200521358/bone_baker/pkg/interface/controller"
	"github.com/miu200521358/bone_baker/pkg/usecase"
)

// Container 依存性注入コンテナ
type Container struct {
	// Repositories
	BakeSetRepo domainRepository.BakeSetRepository
	ModelRepo   domainRepository.ModelRepository
	MotionRepo  domainRepository.MotionRepository

	// Services
	BakeSetService *domain.BakeSetService

	// Usecases
	PhysicsUsecase usecase.PhysicsUsecase
	OutputUsecase  usecase.OutputUsecase
	BakeUsecase    *usecase.BakeUsecase

	// Controllers
	PhysicsController *controller.PhysicsController
	OutputController  *controller.OutputController
	BakeController    *bakeController.BakeController
}

// NewContainer 依存性注入コンテナのコンストラクタ
func NewContainer() *Container {
	// Repositories
	bakeSetRepo := infrastructureRepository.NewBakeSetRepository()
	modelRepo := infrastructureRepository.NewModelRepository()
	motionRepo := infrastructureRepository.NewMotionRepository()

	// Services
	bakeSetService := domain.NewBakeSetService()

	// Usecases
	physicsUsecase := usecase.NewPhysicsUsecase(bakeSetRepo, modelRepo, bakeSetService)
	outputUsecase := usecase.NewOutputUsecase(bakeSetRepo, modelRepo, bakeSetService)
	bakeUsecase := usecase.NewBakeUsecase(modelRepo, motionRepo, bakeSetRepo, bakeSetService)

	// Controllers
	physicsController := controller.NewPhysicsController(physicsUsecase)
	outputController := controller.NewOutputController(outputUsecase)
	bakeController := bakeController.NewBakeController(bakeUsecase)

	return &Container{
		BakeSetRepo:       bakeSetRepo,
		ModelRepo:         modelRepo,
		MotionRepo:        motionRepo,
		BakeSetService:    bakeSetService,
		PhysicsUsecase:    physicsUsecase,
		OutputUsecase:     outputUsecase,
		BakeUsecase:       bakeUsecase,
		PhysicsController: physicsController,
		OutputController:  outputController,
		BakeController:    bakeController,
	}
}
