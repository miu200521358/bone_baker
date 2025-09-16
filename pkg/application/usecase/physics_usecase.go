package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

type PhysicsUsecase struct {
}

func NewPhysicsUsecase() *PhysicsUsecase {
	return &PhysicsUsecase{}
}

func (u *PhysicsUsecase) ApplyPhysicsWorldMotion(
	physicsWorldMotion *vmd.VmdMotion,
	records []*entity.PhysicsRecord,
) {
	for _, record := range records {
		for f := record.StartFrame; f <= record.EndFrame; f++ {
			physicsWorldMotion.AppendGravityFrame(vmd.NewGravityFrameByValue(f, &mmath.MVec3{
				X: 0,
				Y: float64(record.Gravity),
				Z: 0,
			}))
			physicsWorldMotion.AppendMaxSubStepsFrame(vmd.NewMaxSubStepsFrameByValue(f, record.MaxSubSteps))
			physicsWorldMotion.AppendFixedTimeStepFrame(vmd.NewFixedTimeStepFrameByValue(f, record.FixedTimeStep))

			if f == record.StartFrame {
				// 前フレームから継続して物理演算を行う
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			} else {
				// 開始と終了以外はリセットしない
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_NONE))
			}
		}

		// 最初フレームの前には物理リセットしない（次キーフレを呼んでしまうので）
		if record.StartFrame > 0 {
			physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.StartFrame-1, vmd.PHYSICS_RESET_TYPE_NONE))
		}
		// 最後のフレームの後に物理更新停止する
		physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.EndFrame+1, vmd.PHYSICS_RESET_TYPE_NONE))
	}
}

func (u *PhysicsUsecase) ApplyPhysicsModelMotion(
	physicsWorldMotion, physicsModelMotion *vmd.VmdMotion,
	records []*entity.RigidBodyRecord,
	model *pmx.PmxModel,
) {
	for _, record := range records {
		for _, f := range []float32{max(0, record.StartFrame-1), record.StartFrame, record.EndFrame, record.EndFrame + 1} {
			// 最初と最後に初期化キーを入れる

			// 前フレームから継続して物理演算を行う
			physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))

			// 剛体
			model.RigidBodies.ForEach(func(rigidIndex int, rb *pmx.RigidBody) bool {
				rigidBodyItem := record.Tree.AtByRigidBodyIndex(rb.Index())

				if rigidBodyItem == nil || !rigidBodyItem.Modified {
					return true
				}

				physicsModelMotion.AppendRigidBodyFrame(rb.Name(),
					vmd.NewRigidBodyFrameByValues(
						f,
						rb.Position.Copy(),
						rb.Size.Copy(),
						rb.RigidBodyParam.Mass,
					))

				return true
			})

			// ジョイント
			model.Joints.ForEach(func(jointIndex int, joint *pmx.Joint) bool {
				rigidBodyItemA := record.Tree.AtByRigidBodyIndex(joint.RigidBodyIndexA)
				rigidBodyItemB := record.Tree.AtByRigidBodyIndex(joint.RigidBodyIndexB)

				if rigidBodyItemA == nil && rigidBodyItemB == nil {
					// ジョイントの両端が未設定の場合はスキップ
					return true
				}

				if ((rigidBodyItemA != nil && !rigidBodyItemA.Modified) || rigidBodyItemA == nil) &&
					((rigidBodyItemB != nil && !rigidBodyItemB.Modified) || rigidBodyItemB == nil) {
					// 両方の剛体が未変更の場合はスキップ
					return true
				}

				physicsModelMotion.AppendJointFrame(joint.Name(),
					vmd.NewJointFrameByValues(
						f,
						joint.JointParam.TranslationLimitMin.Copy(),
						joint.JointParam.TranslationLimitMax.Copy(),
						joint.JointParam.RotationLimitMin.Copy(),
						joint.JointParam.RotationLimitMax.Copy(),
						joint.JointParam.SpringConstantTranslation.Copy(),
						joint.JointParam.SpringConstantRotation.Copy(),
					))

				return true
			})
		}

		// 台形の線形補間で変形させる
		for _, f := range []float32{record.MaxStartFrame, record.MaxEndFrame} {
			// 最初と最後に最大キーを入れる

			// 前フレームから継続して物理演算を行う
			physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))

			// 剛体
			model.RigidBodies.ForEach(func(rigidIndex int, rb *pmx.RigidBody) bool {
				rigidBodyItem := record.Tree.AtByRigidBodyIndex(rb.Index())

				if rigidBodyItem == nil || !rigidBodyItem.Modified {
					return true
				}

				physicsModelMotion.AppendRigidBodyFrame(rb.Name(),
					vmd.NewRigidBodyFrameByValues(
						f,
						rb.Position.Added(rigidBodyItem.Position),
						rb.Size.Muled(rigidBodyItem.SizeRatio),
						rb.RigidBodyParam.Mass*rigidBodyItem.MassRatio,
					))

				return true
			})

			// ジョイント
			model.Joints.ForEach(func(jointIndex int, joint *pmx.Joint) bool {
				rigidBodyItemA := record.Tree.AtByRigidBodyIndex(joint.RigidBodyIndexA)
				rigidBodyItemB := record.Tree.AtByRigidBodyIndex(joint.RigidBodyIndexB)

				if rigidBodyItemA == nil || rigidBodyItemB == nil {
					// ジョイントが繋がっている剛体のいずれかが未設定の場合はスキップ
					return true
				}

				if !rigidBodyItemA.Modified && !rigidBodyItemB.Modified {
					// 両方の剛体が未変更の場合はスキップ
					return true
				}

				// 両剛体の平均倍率を計算
				avgStiffnessRatio := mmath.Mean([]float64{rigidBodyItemA.StiffnessRatio, rigidBodyItemB.StiffnessRatio})
				avgTensionRatio := mmath.Mean([]float64{rigidBodyItemA.TensionRatio, rigidBodyItemB.TensionRatio})

				physicsModelMotion.AppendJointFrame(joint.Name(),
					vmd.NewJointFrameByValues(
						f,
						joint.JointParam.TranslationLimitMin.Copy(),
						joint.JointParam.TranslationLimitMax.Copy(),
						joint.JointParam.RotationLimitMin.DivedScalar(avgStiffnessRatio),
						joint.JointParam.RotationLimitMax.DivedScalar(avgStiffnessRatio),
						joint.JointParam.SpringConstantTranslation.MuledScalar(avgStiffnessRatio),
						joint.JointParam.SpringConstantRotation.MuledScalar(avgTensionRatio),
					))

				return true
			})
		}

		// 最初フレームの前には物理リセットしない（次キーフレを呼んでしまうので）
		if record.StartFrame > 0 {
			physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.StartFrame-1, vmd.PHYSICS_RESET_TYPE_NONE))
		}
		// 最後のフレームの後に物理更新停止する
		physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.EndFrame+1, vmd.PHYSICS_RESET_TYPE_NONE))
	}
}

// ApplyWindMotion 風設定をVMDモーションに適用する
func (u *PhysicsUsecase) ApplyWindMotion(
	windMotion *vmd.VmdMotion,
	records []*entity.WindRecord,
) {
	for _, record := range records {
		for f := record.StartFrame; f <= record.EndFrame; f++ {
			windMotion.AppendWindEnabledFrame(vmd.NewWindEnabledFrameByValue(f, record.WindConfig.Enabled))
			windMotion.AppendWindDirectionFrame(vmd.NewWindDirectionFrameByValue(f, record.WindConfig.Direction))
			windMotion.AppendWindDragCoeffFrame(vmd.NewWindDragCoeffFrameByValue(f, record.WindConfig.DragCoeff))
			windMotion.AppendWindLiftCoeffFrame(vmd.NewWindLiftCoeffFrameByValue(f, record.WindConfig.LiftCoeff))
			windMotion.AppendWindRandomnessFrame(vmd.NewWindRandomnessFrameByValue(f, record.WindConfig.Randomness))
			windMotion.AppendWindSpeedFrame(vmd.NewWindSpeedFrameByValue(f, record.WindConfig.Speed))
			windMotion.AppendWindTurbulenceFreqHzFrame(vmd.NewWindTurbulenceFreqHzFrameByValue(f, record.WindConfig.TurbulenceFreqHz))

			if f == record.StartFrame {
				// 前フレームから継続して物理演算を行う
				windMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			} else {
				// 開始と終了以外はリセットしない
				windMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_NONE))
			}
		}

		// 最初フレームの前には物理リセットしない（次キーフレを呼んでしまうので）
		if record.StartFrame > 0 {
			windMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.StartFrame-1, vmd.PHYSICS_RESET_TYPE_NONE))
		}
		// 最後のフレームの後に物理更新停止する
		windMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.EndFrame+1, vmd.PHYSICS_RESET_TYPE_NONE))
	}
}
