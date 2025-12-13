package usecase

import (
	"fmt"
	"runtime"
	"slices"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/merr"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/miter"
)

type OutputUsecase struct {
}

func NewOutputUsecase() *OutputUsecase {
	return &OutputUsecase{}
}

func (uc *OutputUsecase) GetBakedBoneFlags(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	records []*entity.OutputRecord,
) (outputBoneFlags [][]entity.OutputBoneFlag, isContainsReduce bool) {
	minFrame := originalMotion.MinFrame()
	maxFrame := originalMotion.MaxFrame()
	// 焼き込み対象ボーン一覧を取得
	for _, record := range records {
		if minFrame == 0 || record.StartFrame < minFrame {
			minFrame = record.StartFrame
		}
		if maxFrame == 0 || record.EndFrame > maxFrame {
			maxFrame = record.EndFrame
		}
	}
	frameCount := int(maxFrame + 1)

	// ボーン毎のフレーム焼き込み有無を設定
	outputBoneFlags = make([][]entity.OutputBoneFlag, originalModel.Bones.Length())
	originalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		outputBoneFlags[boneIndex] = make([]entity.OutputBoneFlag, frameCount)

		for f := float32(0); f <= maxFrame; f++ {
			if originalMotion.BoneFrames.Contains(bone.Name()) && originalMotion.BoneFrames.Get(bone.Name()).Contains(f) {
				// 元モーションに登録されている場合、焼き込み対象
				outputBoneFlags[boneIndex][int(f)] = entity.OutputBoneFlagOriginal
			}

			for _, record := range records {
				if slices.Contains(record.ItemBoneNames(), bone.Name()) && f >= record.StartFrame && f <= record.EndFrame {
					// 出力対象レコードに登録されている場合、焼き込み対象
					if record.Reduce {
						outputBoneFlags[boneIndex][int(f)] = entity.OutputBoneFlagReduce
						isContainsReduce = true
					} else {
						outputBoneFlags[boneIndex][int(f)] = entity.OutputBoneFlagBake
					}
				}
			}
		}

		return true
	})

	return outputBoneFlags, isContainsReduce
}

// ProcessOutputMotion 出力モーション処理のビジネスロジック
func (uc *OutputUsecase) ProcessOutputMotions(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	outputMotion *vmd.VmdMotion,
	outputMotionPath string,
	records []*entity.OutputRecord,
	outputBoneFlags [][]entity.OutputBoneFlag,
	isContainsReduce bool,
	incrementCompletedCount func(),
	isTerminate func() bool,
) ([]*vmd.VmdMotion, error) {
	// 焼き込みモーションを生成
	bakedMotion, err := uc.bakeMotion(originalModel, originalMotion, outputMotion, outputBoneFlags, incrementCompletedCount, isTerminate)
	if err != nil {
		return nil, err
	}

	var reducedMotion *vmd.VmdMotion

	if isContainsReduce {
		// 間引き後のキーフレームを生成
		reducedBoneFrames, err := uc.generateReducedBoneFrames(originalModel, bakedMotion, incrementCompletedCount, isTerminate)
		if err != nil {
			return nil, err
		}

		// 間引きモーションを生成
		reducedMotion, err = uc.reduceMotion(originalModel, originalMotion, outputMotion, outputBoneFlags, reducedBoneFrames, incrementCompletedCount, isTerminate)
		if err != nil {
			return nil, err
		}
	} else {
		reducedMotion = bakedMotion
	}

	// 最大件数で分割
	return uc.splitMotion(originalModel, originalMotion, outputMotionPath, reducedMotion, incrementCompletedCount, isTerminate)
}

func (uc *OutputUsecase) bakeMotion(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	outputMotion *vmd.VmdMotion,
	outputBoneFlags [][]entity.OutputBoneFlag,
	incrementCompletedCount func(),
	isTerminate func() bool,
) (bakedMotion *vmd.VmdMotion, err error) {
	bakedMotion, err = originalMotion.Copy()
	if err != nil {
		return nil, err
	}

	logBlockSize := runtime.NumCPU() * 100
	blockSize, _ := miter.GetBlockSize(len(originalModel.Bones.Names()))

	// 焼き込み処理
	err = miter.IterParallelByList(originalModel.Bones.Names(), blockSize, logBlockSize,
		func(boneIndex int, boneName string) error {
			for f, outputFlag := range outputBoneFlags[boneIndex] {
				if isTerminate() {
					return merr.NewTerminateError("manual terminate")
				}

				incrementCompletedCount()

				switch outputFlag {
				case entity.OutputBoneFlagBake, entity.OutputBoneFlagReduce:
					// 焼き込み出力対象の場合、出力モーションから取得
					bakedBf := outputMotion.BoneFrames.Get(boneName).Get(float32(f))

					bf := vmd.NewBoneFrame(float32(f))
					bf.Position = bakedBf.FilledPosition().Copy()     // 位置を保存
					bf.Rotation = bakedBf.FilledUnitRotation().Copy() // (モーフ・付与親含む)トータル回転を保存

					bakedMotion.InsertBoneFrame(boneName, bf)
				}
			}
			mlog.I(fmt.Sprintf(mi18n.T("--- [%07d/%07d] キーフレーム焼き込み処理中 [%s] ..."), originalModel.Bones.Length(), originalModel.Bones.Length(), boneName))

			return nil
		},
		func(iterIndex, allCount int) {
			mlog.I(fmt.Sprintf(mi18n.T("--- [%07d/%07d] キーフレーム焼き込み処理中 ..."), iterIndex, allCount))
		})
	if err != nil {
		return nil, err
	}

	return bakedMotion, nil
}

func (uc *OutputUsecase) generateReducedBoneFrames(
	originalModel *pmx.PmxModel,
	bakedMotion *vmd.VmdMotion,
	incrementCompletedCount func(),
	isTerminate func() bool,
) (reducedBoneFrames []*vmd.BoneNameFrames, err error) {
	blockSize, _ := miter.GetBlockSize(len(originalModel.Bones.Names()))

	reducedBoneFrames = make([]*vmd.BoneNameFrames, len(originalModel.Bones.Names()))

	// 間引き処理
	err = miter.IterParallelByList(originalModel.Bones.Names(), blockSize, 1,
		func(boneIndex int, boneName string) error {
			if isTerminate() {
				return merr.NewTerminateError("manual terminate")
			}

			reducedBoneFrames[boneIndex] = bakedMotion.BoneFrames.Get(boneName).Reduce()

			incrementCompletedCount()

			return nil
		},
		func(iterIndex, allCount int) {
			mlog.I(fmt.Sprintf(mi18n.T("--- [%03d/%03d] キーフレーム間引き処理中 [%s] ..."), iterIndex, allCount, originalModel.Bones.Names()[iterIndex]))
		})
	if err != nil {
		return nil, err
	}

	return reducedBoneFrames, nil
}

func (uc *OutputUsecase) reduceMotion(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	outputMotion *vmd.VmdMotion,
	outputBoneFlags [][]entity.OutputBoneFlag,
	reducedBoneFrames []*vmd.BoneNameFrames,
	incrementCompletedCount func(),
	isTerminate func() bool,
) (reducedMotion *vmd.VmdMotion, err error) {
	reducedMotion, err = originalMotion.Copy()
	if err != nil {
		return nil, err
	}

	logInterval := 100000
	frameCount := 0
	maxFrameCount := len(outputBoneFlags[0]) / logInterval

	for boneIndex, boneName := range originalModel.Bones.Names() {
		for f, outputFlag := range outputBoneFlags[boneIndex] {
			if isTerminate() {
				return nil, merr.NewTerminateError("manual terminate")
			}

			if (outputFlag == entity.OutputBoneFlagReduce && reducedBoneFrames[boneIndex].Contains(float32(f))) ||
				outputFlag == entity.OutputBoneFlagBake {
				// 間引き出力対象で間引き後のフレームに含まれる場合、または焼き込み出力対象の場合、処理継続
				outputBf := outputMotion.BoneFrames.Get(boneName).Get(float32(f))

				bf := vmd.NewBoneFrame(float32(f))
				bf.Position = outputBf.FilledPosition().Copy() // 位置を保存
				bf.Rotation = outputBf.FilledRotation().Copy() // 回転を保存
				bf.Curves = nil

				bone, err := originalModel.Bones.GetByName(boneName)
				if err != nil {
					continue
				}

				// 物理ボーンの場合、物理無効で登録
				if bone.HasDynamicPhysics() {
					bf.DisablePhysics = true
				}

				reducedMotion.InsertBoneFrame(boneName, bf)
			}

			frameCount++
			if frameCount%logInterval == 0 {
				mlog.I(fmt.Sprintf(mi18n.T("--- [%03d/%03d] キーフレーム焼き込み処理中 ..."), frameCount/logInterval, maxFrameCount))
			}

			incrementCompletedCount()
		}
	}

	return reducedMotion, nil
}

func (uc *OutputUsecase) splitMotion(
	originalModel *pmx.PmxModel,
	originalMotion *vmd.VmdMotion,
	outputMotionPath string,
	reducedMotion *vmd.VmdMotion,
	incrementCompletedCount func(),
	isTerminate func() bool,
) (motions []*vmd.VmdMotion, err error) {
	motions = make([]*vmd.VmdMotion, 0)
	var motion *vmd.VmdMotion

	dirPath, fileName, ext := mfile.SplitPath(outputMotionPath)

	logInterval := 100000
	frameCount := 0
	prevFrameTotalCount := 0
	maxFrameCount := int(reducedMotion.MaxFrame()) / logInterval

	for f := float32(0); f < originalMotion.MaxFrame(); f++ {
		if isTerminate() {
			return nil, merr.NewTerminateError("manual terminate")
		}

		if len(motions) == 0 || prevFrameTotalCount+frameCount > vmd.MAX_BONE_FRAMES {
			// 最大登録数を超える場合、新規モーションを作成

			motion = vmd.NewVmdMotion("")
			motion.SetName(fmt.Sprintf("%s_baked", originalModel.Name()))
			motion.SetPath(fmt.Sprintf("%s%s_%02d_%04d%s", dirPath, fileName, len(motions)+1, int(f), ext))
			motions = append(motions, motion)

			prevFrameTotalCount = 0
		} else {
			prevFrameTotalCount += frameCount
		}

		frameCount = 0
		for _, boneName := range originalModel.Bones.Names() {
			incrementCompletedCount()

			if reducedMotion.BoneFrames.Get(boneName).Contains(f) {
				bone, err := originalModel.Bones.GetByName(boneName)
				if err != nil {
					continue
				}

				bf := reducedMotion.BoneFrames.Get(boneName).Get(f)

				if bone.HasDynamicPhysics() {
					// 物理ボーンの場合、物理無効で登録
					bf.DisablePhysics = true
				}

				motion.AppendBoneFrame(boneName, bf)

				// 補間曲線分割済みの次のキーフレ取得して、出力モーションに追加
				nextFrame := reducedMotion.BoneFrames.Get(boneName).NextFrame(f + 1)
				nextBf := reducedMotion.BoneFrames.Get(boneName).Get(nextFrame)

				if bone.HasDynamicPhysics() {
					// 物理ボーンの場合、物理有効で登録
					nextBf.DisablePhysics = false
				}

				motion.AppendBoneFrame(boneName, nextBf)
			}

			frameCount += 2
			if prevFrameTotalCount+frameCount%logInterval == 0 {
				mlog.I(fmt.Sprintf(mi18n.T("--- [%02d][%06d] キーフレーム焼き込み処理中 ..."), frameCount/logInterval, maxFrameCount))
			}
		}
	}

	return motions, nil
}
