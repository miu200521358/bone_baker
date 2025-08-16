package domain

import (
	"errors"
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
)

// BakeSetHelper BakeSetの複雑なビジネスロジックを処理するヘルパー
type BakeSetHelper struct {
	physicsBoneService *PhysicsBoneService
}

// NewBakeSetHelper コンストラクタ
func NewBakeSetHelper() *BakeSetHelper {
	return &BakeSetHelper{
		physicsBoneService: NewPhysicsBoneService(),
	}
}

// ProcessModelsForBakeSet BakeSet用のモデル処理（ビジネスロジック）
func (h *BakeSetHelper) ProcessModelsForBakeSet(originalModel, bakedModel *pmx.PmxModel) error {
	if originalModel == nil {
		return nil
	}

	// ドメインルールの適用
	h.physicsBoneService.ProcessPhysicsBones(originalModel)
	h.physicsBoneService.ProcessPhysicsBones(bakedModel)

	if bakedModel != nil {
		h.physicsBoneService.FixPhysicsRigidBodies(bakedModel)
	}

	return nil
}

// CreateOutputModelPath 出力モデルパスを生成
func (h *BakeSetHelper) CreateOutputModelPath(originalModel *pmx.PmxModel) string {
	if originalModel == nil {
		return ""
	}
	return mfile.CreateOutputPath(originalModel.Path(), "BB")
}

// CreateOutputMotionPath 出力モーションパスを生成
func (h *BakeSetHelper) CreateOutputMotionPath(originalMotion *vmd.VmdMotion, bakedModel *pmx.PmxModel) string {
	if originalMotion == nil || bakedModel == nil {
		return ""
	}

	_, fileName, _ := mfile.SplitPath(bakedModel.Path())
	return mfile.CreateOutputPath(
		originalMotion.Path(), fmt.Sprintf("BB_%s", fileName))
}

// ProcessOutputMotion 出力モーション処理のビジネスロジック
func (h *BakeSetHelper) ProcessOutputMotion(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	outputMotion *vmd.VmdMotion,
	outputMotionPath string,
	records []*OutputBoneRecord,
) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)

	if originalModel == nil || outputMotion == nil || len(records) == 0 {
		return motions, errors.New(mi18n.T("物理焼き込みセットの元モデルまたは出力モーションが設定されていません"))
	}

	// 既存モーションに焼き込みボーンを追加挿入
	bakedMotion, err := originalMotion.Copy()
	if err != nil {
		return motions, fmt.Errorf(mi18n.T("元モーションのコピーに失敗しました: %w"), err)
	}
	bakedMotion.SetPath(outputMotionPath)

	keyCounts := make([]int, int(originalMotion.MaxFrame()+1+1))
	for _, record := range records {
		if record == nil || record.OutputBoneTreeModel == nil {
			continue
		}

		for f := record.StartFrame; f <= record.EndFrame; f++ {
			for _, boneName := range record.TargetBoneNames {
				bf := outputMotion.BoneFrames.Get(boneName).Get(f)
				if bf == nil {
					continue
				}

				if bone, err := originalModel.Bones.GetByName(boneName); err == nil {
					if bone.HasPhysics() {
						bf.DisablePhysics = true
					}
					bakedMotion.AppendBoneFrame(boneName, bf)
					keyCounts[int(f)]++

					// 次のキーフレ物理有効で登録
					if bone.HasPhysics() {
						bf := outputMotion.BoneFrames.Get(boneName).Get(f + 1)
						if bf == nil {
							continue
						}
						bf.DisablePhysics = false
						bakedMotion.AppendBoneFrame(boneName, bf)
						keyCounts[int(f)]++
					}
				}
			}
		}
	}

	if mmath.Sum(keyCounts) == 0 {
		return motions, errors.New(mi18n.T("焼き込み対象キーフレームなし"))
	}

	// モーション分割処理
	return h.splitMotions(bakedMotion, keyCounts, outputMotionPath, originalModel, originalMotion)
}

// splitMotions モーション分割処理（内部メソッド）
func (h *BakeSetHelper) splitMotions(
	bakedMotion *vmd.VmdMotion,
	keyCounts []int,
	outputMotionPath string,
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)
	dirPath, fileName, ext := mfile.SplitPath(outputMotionPath)

	motion := vmd.NewVmdMotion("")
	motion.SetPath(fmt.Sprintf("%s%s_%04d%s", dirPath, fileName, 0, ext))
	motion.MorphFrames, _ = originalMotion.MorphFrames.Copy()

	frameCount := 0
	for f := 0; f <= len(keyCounts); f++ {
		if f < len(keyCounts)-1 && frameCount+keyCounts[int(f+1)] > vmd.MAX_BONE_FRAMES {
			// キーフレーム数が上限を超える場合は切り替える
			motions = append(motions, motion)

			motion = vmd.NewVmdMotion(fmt.Sprintf("%s%s_%04d%s", dirPath, fileName, f, ext))
			motion.MorphFrames, _ = originalMotion.MorphFrames.Copy()
			frameCount = 0
		}

		originalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
			if bakedMotion.BoneFrames.Get(bone.Name()).Contains(float32(f)) {
				motion.AppendBoneFrame(bone.Name(), bakedMotion.BoneFrames.Get(bone.Name()).Get(float32(f)))
			}
			return true
		})

		if f < len(keyCounts)-1 {
			frameCount += keyCounts[int(f)]
		}
	}

	motions = append(motions, motion)
	return motions, nil
}
