package entity

import (
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
)

// 1モデル分の焼き込み設定
type BakeSet struct {
	Index int // インデックス

	OriginalMotionPath string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPath  string `json:"original_model_path"`  // 元モデルパス
	OutputMotionPath   string `json:"-"`                    // 出力モーションパス
	OutputModelPath    string `json:"-"`                    // 出力モデルパス

	OriginalMotion *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel  *pmx.PmxModel  `json:"-"` // 元モデル
	BakedModel     *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	OutputMotion   *vmd.VmdMotion `json:"-"` // 出力結果モーション

	RigidBodyRecords []*RigidBodyRecord `json:"rigid_body_records"` // モデル物理設定レコード
	OutputRecords    []*OutputRecord    `json:"output_records"`     // 出力設定レコード
}

func NewBakeSet(index int) *BakeSet {
	return &BakeSet{
		Index: index,
	}
}

func (s *BakeSet) Clear() {
	s.ClearModel()
	s.ClearMotion()

	s.RigidBodyRecords = make([]*RigidBodyRecord, 0)
	s.OutputRecords = make([]*OutputRecord, 0)
}

func (s *BakeSet) ClearModel() {
	s.OriginalModel = nil
	s.BakedModel = nil
	s.OriginalModelPath = ""
	s.OutputModelPath = ""
}

func (s *BakeSet) ClearMotion() {
	s.OriginalMotion = nil
	s.OutputMotion = nil
	s.OutputMotion = nil
	s.OriginalMotionPath = ""
	s.OutputMotionPath = ""
}

func (s *BakeSet) OriginalMotionName() string {
	if s.OriginalMotion == nil {
		return ""
	}
	return s.OriginalMotion.Name()
}

func (s *BakeSet) OriginalModelName() string {
	if s.OriginalModel == nil {
		return ""
	}
	return s.OriginalModel.Name()
}

// CreateOutputMotionPath 出力モーションパスを生成
func (s *BakeSet) CreateOutputMotionPath() string {
	_, fileName, _ := mfile.SplitPath(s.BakedModel.Path())
	return mfile.CreateOutputPath(s.OriginalMotion.Path(), fmt.Sprintf("BB_%s", fileName))
}
