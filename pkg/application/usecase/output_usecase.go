package usecase

import (
	"errors"
	"fmt"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
)

type OutputUsecase struct {
}

func NewOutputUsecase() *OutputUsecase {
	return &OutputUsecase{}
}

// ProcessOutputMotion 出力モーション処理のビジネスロジック
func (uc *OutputUsecase) ProcessOutputMotions(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	outputMotion *vmd.VmdMotion,
	outputMotionPath string,
	records []*entity.OutputRecord,
) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)

	if originalModel == nil || outputMotion == nil || len(records) == 0 {
		return motions, errors.New(mi18n.T("物理焼き込みセットの元モデルまたは出力モーションが設定されていません"))
	}

	// 新規モーションを作成（焼き込み用のみとする）
	bakedMotion := vmd.NewVmdMotion(outputMotionPath)

	keyCounts := make([]int, int(originalMotion.MaxFrame()*2))
	for _, record := range records {
		if record == nil || record.Tree == nil || len(record.Tree.Items) == 0 {
			continue
		}

		for f := record.StartFrame; f <= record.EndFrame; f++ {
			for _, boneName := range record.ItemBoneNames() {
				bf := outputMotion.BoneFrames.Get(boneName).Get(f)
				if bf == nil {
					continue
				}
				bakedBf := vmd.NewBoneFrame(f)
				bakedBf.Position = bf.FilledPosition().Copy()
				bakedBf.Rotation = bf.FilledUnitRotation().Copy() // (モーフ・付与親含む)トータル回転を保存

				if bone, err := originalModel.Bones.GetByName(boneName); err == nil {
					if bone.HasPhysics() {
						bakedBf.DisablePhysics = true
					}
					bakedMotion.AppendBoneFrame(boneName, bakedBf)
					keyCounts[int(f)]++

					// 次のキーフレ物理有効で登録
					if bone.HasPhysics() {
						bf := outputMotion.BoneFrames.Get(boneName).Get(f + 1)
						if bf == nil {
							continue
						}
						bf.DisablePhysics = false
						bakedMotion.AppendBoneFrame(boneName, bf)
						keyCounts[int(f+1)]++
					}
				}
			}
		}
	}

	// TODO: 焼き込み結果を挿入する場合の分割処理

	if mmath.Sum(keyCounts) == 0 {
		return motions, errors.New(mi18n.T("焼き込み対象キーフレームなし"))
	}

	// モーション分割処理
	return uc.splitMotions(bakedMotion, keyCounts, outputMotionPath, originalModel, originalMotion)
}

// splitMotions モーション分割処理
func (uc *OutputUsecase) splitMotions(
	bakedMotion *vmd.VmdMotion,
	keyCounts []int,
	outputMotionPath string,
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)
	dirPath, fileName, ext := mfile.SplitPath(outputMotionPath)

	motion := vmd.NewVmdMotion("")
	motion.SetName(fmt.Sprintf("%s_baked", originalModel.Name()))
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
