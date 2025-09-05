package entity

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

// 1モデル分の焼き込み設定
type BakeSet struct {
	Index int // インデックス

	OriginalMotionPath string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPath  string `json:"original_model_path"`  // 元モデルパス
	OutputMotionPath   string `json:"-"`                    // 出力モーションパス
	OutputModelPath    string `json:"-"`                    // 出力モデルパス

	OriginalMotionName string `json:"-"` // 元モーション名
	OriginalModelName  string `json:"-"` // 元モーション名
	OutputModelName    string `json:"-"` // 物理焼き込み先モデル名

	OriginalMotion *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel  *pmx.PmxModel  `json:"-"` // 元モデル
	BakedModel     *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	OutputMotion   *vmd.VmdMotion `json:"-"` // 出力結果モーション

	PhysicsRecords   []*PhysicsRecord   `json:"physics_records"`    // 物理設定レコード
	RigidBodyRecords []*RigidBodyRecord `json:"rigid_body_records"` // 剛体設定レコード
}

func NewBakeSet(index int) *BakeSet {
	return &BakeSet{
		Index: index,
	}
}

func (s *BakeSet) Clear() {
	s.ClearModel()
	s.ClearMotion()

	s.PhysicsRecords = make([]*PhysicsRecord, 0)
	s.RigidBodyRecords = make([]*RigidBodyRecord, 0)
}

func (s *BakeSet) ClearModel() {
	s.OriginalModel = nil
	s.BakedModel = nil
	s.OriginalModelName = ""
	s.OriginalModelPath = ""
	s.OutputModelName = ""
	s.OutputModelPath = ""
}

func (s *BakeSet) ClearMotion() {
	s.OriginalMotion = nil
	s.OutputMotion = nil
	s.OutputMotion = nil
	s.OriginalMotionName = ""
	s.OriginalMotionPath = ""
	s.OutputMotionPath = ""
}
