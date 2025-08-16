package application

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
)

// BakeApplicationServiceInterface アプリケーションサービスのインターフェース
// UI層との結合度を下げ、テスタビリティを向上させる
type BakeApplicationServiceInterface interface {
	// セット管理関連
	CreateNewBakeSet(index int) *domain.BakeSet
	LoadBakeSetFromFile(jsonPath string) ([]*domain.BakeSet, error)
	SaveBakeSetToFile(bakeSets []*domain.BakeSet, jsonPath string) error

	// ファイル操作関連（UI依存を排除）
	LoadModelForBakeSet(bakeSet *domain.BakeSet, path string) (*ModelLoadResult, error)
	LoadMotionForBakeSet(bakeSet *domain.BakeSet, path string) (*MotionLoadResult, error)

	// 物理設定管理関連
	InitializePhysicsTable(bakeSet *domain.BakeSet)
	InitializeOutputTable(bakeSet *domain.BakeSet)

	// 焼き込み処理制御関連
	CalculateMaxFrame(bakeSets []*domain.BakeSet) float32
	PrepareForBaking(bakeSet *domain.BakeSet) *BakingPreparationResult
}

// ModelLoadResult モデル読み込み結果
// UI層への通知用データを含む
type ModelLoadResult struct {
	OriginalModel *domain.BakeSet
	BakedModel    *domain.BakeSet
	SetIndex      int
	Success       bool
	ErrorMessage  string
}

// MotionLoadResult モーション読み込み結果
// UI層への通知用データを含む
type MotionLoadResult struct {
	BakeSet      *domain.BakeSet
	SetIndex     int
	Success      bool
	ErrorMessage string
}

// BakingPreparationResult 焼き込み準備結果
// UI層への通知用データを含む
type BakingPreparationResult struct {
	BakeSet                *domain.BakeSet
	SetIndex               int
	PreparedOriginalMotion bool
	PreparedOutputMotion   bool
	Success                bool
	ErrorMessage           string
}
