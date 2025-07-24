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

func TestBakeUsecase_UpdatePhysicsParameters(t *testing.T) {
	// モックリポジトリとサービスの作成
	modelRepo := &MockModelRepository{}
	motionRepo := &MockMotionRepository{}
	bakeSetRepo := &MockBakeSetRepository{}
	bakeSetService := domain.NewBakeSetService()

	// Usecaseの作成
	usecase := NewBakeUsecase(modelRepo, motionRepo, bakeSetRepo, bakeSetService)

	// テスト用のBakeSetを作成
	bakeSet := domain.NewPhysicsSet(0)

	// 物理パラメータ更新のテスト（物理ツリーがnilの場合のテスト）
	err := usecase.UpdatePhysicsStiffness(bakeSet, "testItem", 1.5)
	if err != nil {
		t.Errorf("UpdatePhysicsStiffness failed: %v", err)
	}

	err = usecase.UpdatePhysicsTension(bakeSet, "testItem", 2.0)
	if err != nil {
		t.Errorf("UpdatePhysicsTension failed: %v", err)
	}
}
