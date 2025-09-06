package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
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
