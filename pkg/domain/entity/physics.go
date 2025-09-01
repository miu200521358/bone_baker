package entity

// 全体構成用物理定義
type PhysicsRecord struct {
	StartFrame    float32 `json:"start_frame"`     // 区間開始フレーム
	EndFrame      float32 `json:"end_frame"`       // 区間終了フレーム
	Gravity       float64 `json:"gravity"`         // 重力
	MaxSubSteps   int     `json:"max_sub_steps"`   // 最大演算回数
	FixedTimeStep float64 `json:"fixed_time_step"` // 物理演算頻度
}

func NewPhysicsRecord(startFrame, endFrame float32) *PhysicsRecord {
	return &PhysicsRecord{
		StartFrame:    startFrame,
		EndFrame:      endFrame,
		Gravity:       -9.8, // 重力の初期値
		MaxSubSteps:   2,    // 最大演算回数の初期値
		FixedTimeStep: 60,   // 固定フレーム時間の初期値
	}
}
