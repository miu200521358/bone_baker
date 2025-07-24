package usecase

import (
	"testing"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

// MockModelRepository モックリポジトリ
type MockModelRepository struct{}

func (m *MockModelRepository) LoadWithPhysics(path string, enablePhysics bool) (*pmx.PmxModel, error) {
	// テスト用のモックデータを返す
	return nil, nil
}

func (m *MockModelRepository) Save(path string, model *pmx.PmxModel) error {
	return nil
}

// MockMotionRepository モックリポジトリ
type MockMotionRepository struct{}

func (m *MockMotionRepository) Load(path string, clearPhysics bool) (*vmd.VmdMotion, error) {
	return nil, nil
}

func (m *MockMotionRepository) Save(path string, motion *vmd.VmdMotion) error {
	return nil
}

// MockBakeSetRepository モックリポジトリ
type MockBakeSetRepository struct{}

func (m *MockBakeSetRepository) Save(bakeSets []*domain.BakeSet, jsonPath string) error {
	return nil
}

func (m *MockBakeSetRepository) Load(jsonPath string) ([]*domain.BakeSet, error) {
	return []*domain.BakeSet{}, nil
}

func (m *MockBakeSetRepository) GetByID(id int) (*domain.BakeSet, error) {
	return domain.NewPhysicsSet(id), nil
}

func (m *MockBakeSetRepository) SaveSingle(bakeSet *domain.BakeSet) error {
	return nil
}

func TestBakeUsecase_CreatePhysicsTree(t *testing.T) {
	// モックリポジトリとサービスの作成
	modelRepo := &MockModelRepository{}
	motionRepo := &MockMotionRepository{}
	bakeSetRepo := &MockBakeSetRepository{}
	bakeSetService := domain.NewBakeSetService()

	// Usecaseの作成
	usecase := NewBakeUsecase(modelRepo, motionRepo, bakeSetRepo, bakeSetService)

	// テスト用のBakeSetを作成
	bakeSet := domain.NewPhysicsSet(0)

	// 物理ツリー作成のテスト
	err := usecase.CreatePhysicsTree(bakeSet)
	if err != nil {
		t.Errorf("CreatePhysicsTree failed: %v", err)
	}

	// 出力ツリー作成のテスト
	err = usecase.CreateOutputTree(bakeSet)
	if err != nil {
		t.Errorf("CreateOutputTree failed: %v", err)
	}
}

func TestPhysicsUsecase_CreatePhysicsTree(t *testing.T) {
	// モックリポジトリとサービスの作成
	bakeSetRepo := &MockBakeSetRepository{}
	modelRepo := &MockModelRepository{}
	bakeSetService := domain.NewBakeSetService()

	// PhysicsUsecaseの作成
	physicsUsecase := NewPhysicsUsecase(bakeSetRepo, modelRepo, bakeSetService)

	// テスト用のBakeSetを作成
	bakeSet := domain.NewPhysicsSet(0)

	// 物理ツリー作成のテスト
	err := physicsUsecase.CreatePhysicsTree(bakeSet)
	if err != nil {
		t.Errorf("CreatePhysicsTree failed: %v", err)
	}

	// 物理ツリー取得のテスト
	treeDTO, err := physicsUsecase.GetPhysicsTree(bakeSet)
	if err != nil {
		t.Errorf("GetPhysicsTree failed: %v", err)
	}

	if treeDTO == nil {
		t.Error("Expected non-nil tree DTO")
	}
}

func TestOutputUsecase_CreateOutputTree(t *testing.T) {
	// モックリポジトリとサービスの作成
	bakeSetRepo := &MockBakeSetRepository{}
	modelRepo := &MockModelRepository{}
	bakeSetService := domain.NewBakeSetService()

	// OutputUsecaseの作成
	outputUsecase := NewOutputUsecase(bakeSetRepo, modelRepo, bakeSetService)

	// テスト用のBakeSetを作成
	bakeSet := domain.NewPhysicsSet(0)

	// 出力ツリー作成のテスト
	err := outputUsecase.CreateOutputTree(bakeSet)
	if err != nil {
		t.Errorf("CreateOutputTree failed: %v", err)
	}

	// 出力ツリー取得のテスト
	treeDTO, err := outputUsecase.GetOutputTree(bakeSet)
	if err != nil {
		t.Errorf("GetOutputTree failed: %v", err)
	}

	if treeDTO == nil {
		t.Error("Expected non-nil tree DTO")
	}
}
