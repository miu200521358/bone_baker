package entity

import (
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/physics"
)

// 風用物理定義
type WindRecord struct {
	StartFrame float32             `json:"start_frame"` // 区間開始フレーム
	EndFrame   float32             `json:"end_frame"`   // 区間終了フレーム
	WindConfig *physics.WindConfig `json:"wind_config"` // 風の設定
}

func NewWindRecord(startFrame, endFrame float32) *WindRecord {
	return &WindRecord{
		StartFrame: startFrame,
		EndFrame:   endFrame,
		WindConfig: &physics.WindConfig{
			Enabled:          true,
			Direction:        mmath.NewMVec3(), // X方向
			Speed:            0.0,              // 基本風速 [unit/s]
			Randomness:       0.0,              // 乱れの強さ(0..1)
			TurbulenceFreqHz: 0.5,              // 乱流の周波数[Hz]
			DragCoeff:        1.0,              // 抵抗係数（0.5*rho*Cd*A を吸収）
			LiftCoeff:        0.2,              // 揚力係数（0.5*rho*Cl*A を吸収）
		},
	}
}
