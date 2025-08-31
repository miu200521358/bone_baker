package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

type PhysicsUsecase struct {
}

func NewPhysicsUsecase() *PhysicsUsecase {
	return &PhysicsUsecase{}
}

func (u *PhysicsUsecase) ApplyPhysicsMotion(
	physicsWorldMotion, physicsModelMotion *vmd.VmdMotion,
	records []*domain.PhysicsBoneRecord,
	model *pmx.PmxModel,
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
				if record.IsStartDeform {
					// 開始時用整形をON
					physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_START_FIT_FRAME))
				} else {
					// 前フレームから継続して物理演算を行う
					physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
				}
			} else {
				// 開始と終了以外はリセットしない
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_NONE))
			}

			// 剛体・ジョイントパラは台形の線形補間で変形させる
			frameRatio := float32(0.0)
			if f < record.MaxStartFrame && f > record.StartFrame &&
				record.MaxStartFrame > record.StartFrame {
				// StartFrame から MaxStartFrame の間：0倍から指定倍率まで線形補間
				frameRatio = (f - record.StartFrame) / (record.MaxStartFrame - record.StartFrame)
				// 変動中はリセッし続ける
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			} else if f > record.MaxEndFrame && f < record.EndFrame &&
				record.MaxEndFrame < record.EndFrame {
				// MaxEndFrame から EndFrame の間：指定倍率から0倍まで線形補間
				frameRatio = (record.EndFrame - f) / (record.EndFrame - record.MaxEndFrame)
				// 変動中はリセッし続ける
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			} else if f >= record.MaxStartFrame && f <= record.MaxEndFrame {
				// MAXの間はそのまま最大倍率
				frameRatio = 1.0
			} else {
				// StartFrame以前とEndFrame以後は元の値（倍率なし）
				frameRatio = 0.0
			}
			frameRatio64 := float64(frameRatio)

			// 剛体
			model.RigidBodies.ForEach(func(rigidIndex int, rb *pmx.RigidBody) bool {
				rigidBodyItem := record.TreeModel.AtByRigidBodyIndex(rb.Index())

				if rigidBodyItem == nil || !rigidBodyItem.(*domain.PhysicsItem).Modified {
					return true
				}

				// 質量の計算：元の質量 + (元の質量 * (massRatio - 1.0) * frameRatio)
				sizeRatio := rigidBodyItem.(*domain.PhysicsItem).SizeRatio
				calculatedSize := rb.Size.Added(rb.Size.Muled(sizeRatio.SubedScalar(1.0).MuledScalar(frameRatio64)))

				massRatio := rigidBodyItem.(*domain.PhysicsItem).MassRatio
				calculatedMass := rb.RigidBodyParam.Mass + (rb.RigidBodyParam.Mass * (massRatio - 1.0) * frameRatio64)

				physicsModelMotion.AppendRigidBodyFrame(rb.Name(),
					vmd.NewRigidBodyFrameByValues(
						f,
						calculatedSize,
						calculatedMass,
					))

				return true
			})

			// ジョイント
			model.Joints.ForEach(func(jointIndex int, joint *pmx.Joint) bool {
				rigidBodyItemA := record.TreeModel.AtByRigidBodyIndex(joint.RigidbodyIndexA)
				rigidBodyItemB := record.TreeModel.AtByRigidBodyIndex(joint.RigidbodyIndexB)

				if rigidBodyItemA == nil && rigidBodyItemB == nil {
					// ジョイントの両端が未設定の場合はスキップ
					return true
				}

				if ((rigidBodyItemA != nil && !rigidBodyItemA.(*domain.PhysicsItem).Modified) || rigidBodyItemA == nil) &&
					((rigidBodyItemB != nil && !rigidBodyItemB.(*domain.PhysicsItem).Modified) || rigidBodyItemB == nil) {
					// 両方の剛体が未変更の場合はスキップ
					return true
				}

				// ジョイントのパラメータを台形の線形補間で変形させる
				var stiffnessRatioA, stiffnessRatioB float64
				var tensionRatioA, tensionRatioB float64
				if rigidBodyItemA != nil && rigidBodyItemA.(*domain.PhysicsItem).Modified {
					stiffnessRatioA = rigidBodyItemA.(*domain.PhysicsItem).StiffnessRatio
					tensionRatioA = rigidBodyItemA.(*domain.PhysicsItem).TensionRatio
				} else {
					stiffnessRatioA = 1.0
					tensionRatioA = 1.0
				}
				if rigidBodyItemB != nil && rigidBodyItemB.(*domain.PhysicsItem).Modified {
					stiffnessRatioB = rigidBodyItemB.(*domain.PhysicsItem).StiffnessRatio
					tensionRatioB = rigidBodyItemB.(*domain.PhysicsItem).TensionRatio
				} else {
					stiffnessRatioB = 1.0
					tensionRatioB = 1.0
				}

				// 両剛体の平均倍率を計算
				avgStiffnessRatio := mmath.Mean([]float64{stiffnessRatioA, stiffnessRatioB})
				avgTensionRatio := mmath.Mean([]float64{tensionRatioA, tensionRatioB})

				// 台形状の変化を適用
				calculatedRotationLimitMin := joint.JointParam.RotationLimitMin.Added(
					joint.JointParam.RotationLimitMin.DivedScalar((avgStiffnessRatio - 1.0) * frameRatio64))
				calculatedRotationLimitMax := joint.JointParam.RotationLimitMax.Added(
					joint.JointParam.RotationLimitMax.DivedScalar((avgStiffnessRatio - 1.0) * frameRatio64))
				calculatedSpringConstantRotation := joint.JointParam.SpringConstantRotation.Added(
					joint.JointParam.SpringConstantRotation.MuledScalar((avgTensionRatio - 1.0) * frameRatio64))

				physicsModelMotion.AppendJointFrame(joint.Name(),
					vmd.NewJointFrameByValues(
						f,
						joint.JointParam.TranslationLimitMin,
						joint.JointParam.TranslationLimitMax,
						calculatedRotationLimitMin,
						calculatedRotationLimitMax,
						joint.JointParam.SpringConstantTranslation,
						calculatedSpringConstantRotation,
					))

				return true
			})
		}

		// 最初フレームの前には物理リセットしない（次キーフレを呼んでしまうので）
		if record.StartFrame > 0 {
			physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.StartFrame-1, vmd.PHYSICS_RESET_TYPE_NONE))
		}
		// 最後のフレームの後に物理リセットする
		physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.EndFrame+1, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
	}

}
