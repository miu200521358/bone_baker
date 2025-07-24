package usecase

import (
	"sync"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

type BakeUsecase struct {
	modelRepo      repository.ModelRepository
	motionRepo     repository.MotionRepository
	bakeSetRepo    repository.BakeSetRepository
	bakeSetService *domain.BakeSetService
}

func NewBakeUsecase(
	modelRepo repository.ModelRepository,
	motionRepo repository.MotionRepository,
	bakeSetRepo repository.BakeSetRepository,
	bakeSetService *domain.BakeSetService,
) *BakeUsecase {
	return &BakeUsecase{
		modelRepo:      modelRepo,
		motionRepo:     motionRepo,
		bakeSetRepo:    bakeSetRepo,
		bakeSetService: bakeSetService,
	}
}

// LoadModel モデル読み込みのビジネスロジック
func (uc *BakeUsecase) LoadModel(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearModels()
		return nil
	}

	// 元モデル読み込み（物理有効）
	originalModel, err := uc.modelRepo.LoadWithPhysics(path, true)
	if err != nil {
		return err
	}

	// 焼き込み用モデル読み込み（物理無効）
	bakedModel, err := uc.modelRepo.LoadWithPhysics(path, false)
	if err != nil {
		return err
	}

	// ドメインサービスを使ってビジネスロジックを実行
	if err := uc.bakeSetService.ProcessPhysicsModel(originalModel, bakedModel); err != nil {
		return err
	}

	// モデルを設定
	return bakeSet.SetModels(originalModel, bakedModel)
}

// LoadMotion モーション読み込みのビジネスロジック
func (uc *BakeUsecase) LoadMotion(bakeSet *domain.BakeSet, path string) error {
	if path == "" {
		bakeSet.ClearMotions()
		return nil
	}

	var wg sync.WaitGroup
	var originalMotion, outputMotion *vmd.VmdMotion
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if motion, err := uc.motionRepo.Load(path, false); err == nil {
			originalMotion = motion
		} else {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if motion, err := uc.motionRepo.Load(path, true); err == nil {
			outputMotion = motion
		} else {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return bakeSet.SetMotions(originalMotion, outputMotion)
}

// SaveBakeSet セット保存のビジネスロジック
func (uc *BakeUsecase) SaveBakeSet(bakeSets []*domain.BakeSet, jsonPath string) error {
	return uc.bakeSetRepo.Save(bakeSets, jsonPath)
}

// LoadBakeSet セット読み込みのビジネスロジック
func (uc *BakeUsecase) LoadBakeSet(jsonPath string) ([]*domain.BakeSet, error) {
	return uc.bakeSetRepo.Load(jsonPath)
}

// ExportMotions モーション出力のビジネスロジック
func (uc *BakeUsecase) ExportMotions(bakeSet *domain.BakeSet, startFrame, endFrame float64) error {
	motions, err := bakeSet.GetOutputMotionOnlyChecked(startFrame, endFrame)
	if err != nil {
		return err
	}

	for _, motion := range motions {
		if err := uc.motionRepo.Save(motion.Path(), motion); err != nil {
			return err
		}
	}

	return nil
}

// CreatePhysicsTree 物理ツリー作成のビジネスロジック
func (uc *BakeUsecase) CreatePhysicsTree(bakeSet *domain.BakeSet) error {
	if bakeSet.OriginalModel == nil {
		return nil
	}

	tree := domain.NewPhysicsModel()

	for _, boneIndex := range bakeSet.OriginalModel.Bones.LayerSortedIndexes {
		if bone, err := bakeSet.OriginalModel.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := domain.NewPhysicsItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*domain.PhysicsItem).AddChild(item)
			}
		}
	}

	bakeSet.PhysicsTree = tree
	return nil
}

// CreateOutputTree 出力ツリー作成のビジネスロジック
func (uc *BakeUsecase) CreateOutputTree(bakeSet *domain.BakeSet) error {
	if bakeSet.OriginalModel == nil {
		return nil
	}

	tree := domain.NewOutputModel()

	for _, boneIndex := range bakeSet.OriginalModel.Bones.LayerSortedIndexes {
		if bone, err := bakeSet.OriginalModel.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := domain.NewOutputItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*domain.OutputItem).AddChild(item)
			}
		}
	}

	bakeSet.OutputTree = tree
	return nil
}

// UpdatePhysicsStiffness 物理パラメータ更新（硬さ）
func (uc *BakeUsecase) UpdatePhysicsStiffness(bakeSet *domain.BakeSet, itemID string, stiffnessRatio float64) error {
	if bakeSet.PhysicsTree == nil {
		return nil
	}

	if item := bakeSet.PhysicsTree.GetByID(itemID); item != nil {
		if physicsItem, ok := item.(*domain.PhysicsItem); ok {
			physicsItem.CalcStiffness(stiffnessRatio)
		}
	}

	return nil
}

// UpdatePhysicsTension 物理パラメータ更新（張り）
func (uc *BakeUsecase) UpdatePhysicsTension(bakeSet *domain.BakeSet, itemID string, tensionRatio float64) error {
	if bakeSet.PhysicsTree == nil {
		return nil
	}

	if item := bakeSet.PhysicsTree.GetByID(itemID); item != nil {
		if physicsItem, ok := item.(*domain.PhysicsItem); ok {
			physicsItem.CalcTension(tensionRatio)
		}
	}

	return nil
}

// SetOutputChildrenChecked 出力ツリーの子要素チェック状態更新
func (uc *BakeUsecase) SetOutputChildrenChecked(bakeSet *domain.BakeSet, itemID string, checked bool) error {
	if bakeSet.OutputTree == nil {
		return nil
	}

	if item := bakeSet.OutputTree.GetByID(itemID); item != nil {
		if outputItem, ok := item.(*domain.OutputItem); ok {
			uc.setChildrenCheckedRecursive(outputItem, checked)
		}
	}

	return nil
}

// setChildrenCheckedRecursive 再帰的に子要素のチェック状態を更新
func (uc *BakeUsecase) setChildrenCheckedRecursive(item *domain.OutputItem, checked bool) {
	item.SetChecked(checked)

	for _, child := range item.Children() {
		if childItem, ok := child.(*domain.OutputItem); ok {
			uc.setChildrenCheckedRecursive(childItem, checked)
		}
	}
}

// SetOutputIkChecked 出力ツリーのIKチェック状態更新
func (uc *BakeUsecase) SetOutputIkChecked(bakeSet *domain.BakeSet, checked bool) error {
	if bakeSet.OutputTree == nil {
		return nil
	}

	// ドメインロジックでIKボーンのチェック状態を更新
	return uc.bakeSetService.UpdateIkBoneChecked(bakeSet, checked)
}

// SetOutputPhysicsChecked 出力ツリーの物理チェック状態更新
func (uc *BakeUsecase) SetOutputPhysicsChecked(bakeSet *domain.BakeSet, checked bool) error {
	if bakeSet.OutputTree == nil {
		return nil
	}

	// ドメインロジックで物理ボーンのチェック状態を更新
	return uc.bakeSetService.UpdatePhysicsBoneChecked(bakeSet, checked)
}
