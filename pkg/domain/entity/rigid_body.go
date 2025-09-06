package entity

import (
	"strings"

	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

type RigidBodyRecord struct {
	StartFrame     float32          `json:"start_frame"`     // 区間開始フレーム
	EndFrame       float32          `json:"end_frame"`       // 区間終了フレーム
	MaxStartFrame  float32          `json:"max_start_frame"` // 最大値開始フレーム
	MaxEndFrame    float32          `json:"max_end_frame"`   // 最大値終了フレーム
	SizeRatio      *mmath.MVec3     `json:"size_ratio"`      // 大きさの比率
	MassRatio      float64          `json:"mass_ratio"`      // 質量の比率
	TensionRatio   float64          `json:"tension_ratio"`   // 張りの比率
	StiffnessRatio float64          `json:"stiffness_ratio"` // 硬さの比率
	Items          []*RigidBodyItem `json:"items"`           // 剛体アイテム一覧
}

func (r *RigidBodyRecord) ItemNames() string {
	var names []string
	for _, item := range r.Items {
		names = append(names, item.RigidBody.Name())
	}
	return strings.Join(names, ", ")
}

type RigidBodyItem struct {
	Bone           *pmx.Bone      // 剛体に紐付くボーン情報
	RigidBody      *pmx.RigidBody // 剛体情報
	SizeRatio      *mmath.MVec3   `json:"size_ratio"`       // 大きさ比率
	MassRatio      float64        `json:"mass_ratio"`       // 質量比率
	StiffnessRatio float64        `json:"stiffness_ratio"`  // 硬さ比率
	TensionRatio   float64        `json:"tension_ratio"`    // 張り比率
	Modified       bool           `json:"modified"`         // 変更されたかどうか
	RigidBodyIndex int            `json:"rigid_body_index"` // 剛体インデックス
	RigidBodyName  string         `json:"rigid_body_name"`  // 剛体名
}

func NewRigidBodyRecord(startFrame, endFrame float32) *RigidBodyRecord {
	return &RigidBodyRecord{
		StartFrame: startFrame,
		EndFrame:   endFrame,
	}
}
