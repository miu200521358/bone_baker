package usecase

import (
	"fmt"
	"slices"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
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
	copiedMotion, err := originalMotion.Copy()
	if err != nil {
		return motions, err
	}

	logInterval := 10000

	for rIndex, record := range records {
		dirPath, fileName, ext := mfile.SplitPath(outputMotionPath)

		targetBoneNames := record.ItemBoneNames()
		if len(targetBoneNames) == 0 {
			mlog.W(fmt.Sprintf(mi18n.T("出力対象ボーンが見つからなかったため、出力設定をスキップします [No.%02d]"), rIndex+1))
			continue
		}

		motion := vmd.NewVmdMotion("")
		motion.SetName(fmt.Sprintf("%s_baked", originalModel.Name()))
		motion.SetPath(fmt.Sprintf("%s%s_%02d_%04d%s", dirPath, fileName, rIndex, int(record.StartFrame), ext))

		recordFrameCount := 0
		frameCount := 0
		logFrameCount := 0

		mlog.I(fmt.Sprintf(mi18n.T("焼き込み開始 [No.%02d][%04d-%04d]"), rIndex+1, int(record.StartFrame), int(record.EndFrame)))

		for f := record.StartFrame; f <= record.EndFrame; f++ {
			if recordFrameCount+frameCount*2 > vmd.MAX_BONE_FRAMES {
				// 次のフレームが上限を超える場合は切り替える
				mlog.I(fmt.Sprintf(mi18n.T("分割開始 [%04d] - [%d][%d]"), int(f), recordFrameCount, frameCount*2))

				motions = append(motions, motion)
				motion = vmd.NewVmdMotion("")
				motion.SetName(fmt.Sprintf("%s_baked", originalModel.Name()))
				motion.SetPath(fmt.Sprintf("%s%s_%02d_%04d%s", dirPath, fileName, rIndex, int(f), ext))
				recordFrameCount = 0
			} else {
				recordFrameCount += frameCount
			}

			frameCount = 0

			for _, boneName := range outputMotion.BoneFrames.Names() {
				if !slices.Contains(targetBoneNames, boneName) {
					// 出力対象ボーン以外はスキップ
					if originalMotion.BoneFrames.Get(boneName).Contains(f) {
						bone, err := originalModel.Bones.GetByName(boneName)
						if err != nil {
							continue
						}

						if bone.IsEffectorRotation() || bone.IsEffectorTranslation() {
							effectBone, err := originalModel.Bones.Get(bone.EffectIndex)
							if err != nil {
								continue
							}

							if slices.Contains(targetBoneNames, effectBone.Name()) {
								// 付与親ボーンは出力レコードに付与元ボーンが登録されている場合はスキップ
								continue
							}
						}

						// 元モーションの登録ボーンの場合、キーフレ登録
						originalBf := originalMotion.BoneFrames.Get(boneName).Get(f)
						bf := vmd.NewBoneFrame(f)
						bf.Position = originalBf.FilledPosition().Copy() // 位置を保存
						bf.Rotation = originalBf.FilledRotation().Copy() // 回転を保存
						if originalBf.Curves != nil {
							bf.Curves = originalBf.Curves.Copy()
						}

						motion.InsertBoneFrame(boneName, bf)
						frameCount++
					}

					continue
				}

				bone, err := originalModel.Bones.GetByName(boneName)
				if err != nil {
					continue
				}

				bf := vmd.NewBoneFrame(f)
				if slices.Contains(targetBoneNames, boneName) {
					// 焼き込み出力対象の場合、出力モーションから取得
					bakedBf := outputMotion.BoneFrames.Get(boneName).Get(f)
					bf.Position = bakedBf.FilledPosition().Copy()     // 位置を保存
					bf.Rotation = bakedBf.FilledUnitRotation().Copy() // (モーフ・付与親含む)トータル回転を保存
					if bakedBf.Curves != nil {
						bf.Curves = bakedBf.Curves.Copy()
					}

					// 物理ボーンの場合、物理無効で登録
					if bone.HasPhysics() {
						bf.DisablePhysics = true
					}
				} else {
					// 元モーションの登録ボーンの場合、元モーションから取得
					bf = copiedMotion.BoneFrames.Get(boneName).Get(f)
				}

				motion.InsertBoneFrame(boneName, bf)
				frameCount++

				if copiedMotion.BoneFrames.Contains(boneName) {
					// 元モーションの登録ボーンの場合、次のキーフレ登録

					// まずはキーを分割したとして元モーションに登録する
					nowBf := outputMotion.BoneFrames.Get(boneName).Get(f)
					copiedMotion.InsertBoneFrame(boneName, nowBf)

					// 補間曲線分割済みの次のキーフレ取得して、出力モーションに追加
					nextFrame := copiedMotion.BoneFrames.Get(boneName).NextFrame(f + 1)
					nextBf := copiedMotion.BoneFrames.Get(boneName).Get(nextFrame)
					motion.InsertBoneFrame(boneName, nextBf)
				} else if bone.HasPhysics() {
					// 出力対象レコードで物理ボーン場合、次のキーフレ物理有効で登録
					nextFrame := outputMotion.BoneFrames.Get(boneName).NextFrame(f + 1)
					nextBf := outputMotion.BoneFrames.Get(boneName).Get(nextFrame)
					nextBf.DisablePhysics = false
					motion.InsertBoneFrame(boneName, nextBf)
				}
			}

			if recordFrameCount/logInterval > logFrameCount/logInterval && recordFrameCount > 0 {
				mlog.I(fmt.Sprintf(mi18n.T("--- キーフレーム焼き込み処理中 [%s][%04d][%d] ..."), fileName, int(f), recordFrameCount))
				logFrameCount = recordFrameCount
			}
		}

		// 最後のモーションを保持
		motions = append(motions, motion)
	}

	return motions, nil
}
