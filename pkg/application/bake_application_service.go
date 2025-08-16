package application

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
)

// BakeApplicationService アプリケーションサービス
// UIとドメイン層の橋渡しを行い、複数のユースケースを組み合わせたビジネスフローを提供
// BakeApplicationServiceInterfaceを実装
type BakeApplicationService struct {
	bakeUsecase *usecase.BakeUsecase
}

// NewBakeApplicationService コンストラクタ
func NewBakeApplicationService(bakeUsecase *usecase.BakeUsecase) BakeApplicationServiceInterface {
	return &BakeApplicationService{
		bakeUsecase: bakeUsecase,
	}
}

// セット管理関連のメソッド

// CreateNewBakeSet 新しい焼き込みセットを作成
func (s *BakeApplicationService) CreateNewBakeSet(index int) *domain.BakeSet {
	return domain.NewPhysicsSet(index)
}

// LoadBakeSetFromFile ファイルから焼き込みセット設定を読み込み
func (s *BakeApplicationService) LoadBakeSetFromFile(jsonPath string) ([]*domain.BakeSet, error) {
	return s.bakeUsecase.LoadBakeSet(jsonPath)
}

// SaveBakeSetToFile 焼き込みセット設定をファイルに保存
func (s *BakeApplicationService) SaveBakeSetToFile(bakeSets []*domain.BakeSet, jsonPath string) error {
	return s.bakeUsecase.SaveBakeSet(bakeSets, jsonPath)
}

// ファイル操作関連のメソッド（インターフェース実装）

// LoadModelForBakeSet セット用のモデル読み込み処理（UI依存を排除）
func (s *BakeApplicationService) LoadModelForBakeSet(
	bakeSet *domain.BakeSet,
	path string,
) (*ModelLoadResult, error) {
	// BakeUsecase経由でモデル読み込み
	if err := s.bakeUsecase.LoadModelForBakeSet(bakeSet, path); err != nil {
		return &ModelLoadResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}, err
	}

	return &ModelLoadResult{
		OriginalModel: bakeSet,
		BakedModel:    bakeSet,
		Success:       true,
	}, nil
}

// LoadMotionForBakeSet セット用のモーション読み込み処理（UI依存を排除）
func (s *BakeApplicationService) LoadMotionForBakeSet(
	bakeSet *domain.BakeSet,
	path string,
) (*MotionLoadResult, error) {
	// BakeUsecase経由でモーション読み込み
	if err := s.bakeUsecase.LoadMotionForBakeSet(bakeSet, path); err != nil {
		return &MotionLoadResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}, err
	}

	return &MotionLoadResult{
		BakeSet: bakeSet,
		Success: true,
	}, nil
}

// LoadModelForBakeSetWithUI セット用のモデル読み込み処理（UI依存版）
func (s *BakeApplicationService) LoadModelForBakeSetWithUI(
	bakeSet *domain.BakeSet,
	path string,
	cw *controller.ControlWindow,
	setIndex int,
) error {
	// インターフェース版を呼び出し
	_, err := s.LoadModelForBakeSet(bakeSet, path)
	if err != nil {
		return err
	}

	// ウィンドウへの反映
	cw.StoreModel(0, setIndex, bakeSet.OriginalModel)
	cw.StoreModel(1, setIndex, bakeSet.BakedModel)

	return nil
}

// LoadMotionForBakeSetWithUI セット用のモーション読み込み処理（UI依存版）
func (s *BakeApplicationService) LoadMotionForBakeSetWithUI(
	bakeSet *domain.BakeSet,
	path string,
	cw *controller.ControlWindow,
	setIndex int,
) error {
	// インターフェース版を呼び出し
	_, err := s.LoadMotionForBakeSet(bakeSet, path)
	if err != nil {
		return err
	}

	// ウィンドウへの反映
	if bakeSet.OriginalMotion != nil {
		cw.StoreMotion(0, setIndex, bakeSet.OriginalMotion)
	}
	if bakeSet.OutputMotion != nil {
		cw.StoreMotion(1, setIndex, bakeSet.OutputMotion)
	}

	return nil
}

// 物理設定管理関連のメソッド

// InitializePhysicsTable 物理設定テーブルの初期化
func (s *BakeApplicationService) InitializePhysicsTable(bakeSet *domain.BakeSet) {
	if bakeSet.OriginalMotion != nil {
		bakeSet.PhysicsTableModel = domain.NewPhysicsTableModel()
		bakeSet.PhysicsTableModel.AddRecord(
			bakeSet.OriginalModel,
			0,
			bakeSet.OriginalMotion.MaxFrame())
	}
}

// InitializeOutputTable 出力設定テーブルの初期化
func (s *BakeApplicationService) InitializeOutputTable(bakeSet *domain.BakeSet) {
	if bakeSet.OriginalMotion != nil {
		bakeSet.OutputTableModel = domain.NewOutputTableModel()
		bakeSet.OutputTableModel.AddRecord(
			bakeSet.OriginalModel,
			0,
			bakeSet.OriginalMotion.MaxFrame())
	}
}

// 焼き込み処理制御関連のメソッド

// CalculateMaxFrame 全セットの最大フレーム数を計算
func (s *BakeApplicationService) CalculateMaxFrame(bakeSets []*domain.BakeSet) float32 {
	maxFrame := float32(0)
	for _, bakeSet := range bakeSets {
		if bakeSet.OriginalMotion != nil && maxFrame < bakeSet.OriginalMotion.MaxFrame() {
			maxFrame = bakeSet.OriginalMotion.MaxFrame()
		}
	}
	return maxFrame
}

// ClearDeltaMotions 焼き込み履歴をクリア
func (s *BakeApplicationService) ClearDeltaMotions(
	cw *controller.ControlWindow,
	bakeSets []*domain.BakeSet,
) {
	for n := range len(bakeSets) {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}
}

// PrepareForBaking 焼き込み準備処理（UI依存を排除）
func (s *BakeApplicationService) PrepareForBaking(bakeSet *domain.BakeSet) *BakingPreparationResult {
	if bakeSet == nil {
		return &BakingPreparationResult{
			Success:      false,
			ErrorMessage: "BakeSet is nil",
		}
	}

	preparedOriginalMotion := bakeSet.OriginalMotion != nil
	preparedOutputMotion := false

	if bakeSet.OriginalMotion != nil {
		// 出力モーション準備の検証
		if _, err := bakeSet.OriginalMotion.Copy(); err == nil {
			preparedOutputMotion = true
		}
	}

	return &BakingPreparationResult{
		BakeSet:                bakeSet,
		PreparedOriginalMotion: preparedOriginalMotion,
		PreparedOutputMotion:   preparedOutputMotion,
		Success:                preparedOriginalMotion && preparedOutputMotion,
	}
}

// PrepareForBakingWithUI 焼き込み準備処理（UI依存版）
func (s *BakeApplicationService) PrepareForBakingWithUI(
	cw *controller.ControlWindow,
	bakeSet *domain.BakeSet,
	setIndex int,
) {
	// インターフェース版を呼び出し
	result := s.PrepareForBaking(bakeSet)
	if !result.Success {
		return
	}

	// ウィンドウへの反映
	cw.StoreMotion(0, setIndex, bakeSet.OriginalMotion)
	if bakeSet.OriginalMotion != nil {
		if copiedMotion, err := bakeSet.OriginalMotion.Copy(); err == nil {
			cw.StoreMotion(1, setIndex, copiedMotion)
		}
	}
}
