package domain

import (
	"encoding/json"

	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

type BakeSet struct {
	Index       int  // インデックス
	IsTerminate bool // 処理停止フラグ

	// Value Objectsを使用したファイルパス
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

	PhysicsTableModel *PhysicsTableModel `json:"physics_table"` // 物理ボーンツリー
	OutputTableModel  *OutputTableModel  `json:"output_table"`  // 出力定義テーブル

	// Helper依存注入（効率化）
	helper *BakeSetHelper `json:"-"` // ビジネスロジックヘルパー
}

func NewPhysicsSet(index int) *BakeSet {
	return &BakeSet{
		Index:             index,
		PhysicsTableModel: NewPhysicsTableModel(),
		OutputTableModel:  NewOutputTableModel(),
		helper:            NewBakeSetHelper(), // Helper注入
	}
}

func (ss *BakeSet) UnmarshalJSON(data []byte) error {
	type Alias BakeSet
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(ss),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// aux.PhysicsTableModel = NewPhysicsTableModel()
	// aux.OutputTableModel = NewOutputTableModel()
	aux.helper = NewBakeSetHelper()

	return nil
}

func (ss *BakeSet) MaxFrame() float32 {
	if ss.OriginalMotion == nil {
		return 0
	}

	return ss.OriginalMotion.MaxFrame()
}
func (ss *BakeSet) CreateOutputModelPath() string {
	return ss.helper.CreateOutputModelPath(ss.OriginalModel)
}

func (ss *BakeSet) CreateOutputMotionPath() string {
	return ss.helper.CreateOutputMotionPath(ss.OriginalMotion, ss.BakedModel)
}

func (ss *BakeSet) setMotion(originalMotion, outputMotion *vmd.VmdMotion) {
	if originalMotion == nil || outputMotion == nil {
		ss.OriginalMotionPath = ""
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.OutputMotionPath = ""
		ss.OutputMotion = vmd.NewVmdMotion("")

		return
	}

	ss.OriginalMotionPath = originalMotion.Path()
	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion
	ss.OutputMotion = outputMotion
}

func (ss *BakeSet) setModels(originalModel, physicsBakedModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.OriginalModelPath = ""
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.BakedModel = nil
		return
	}

	ss.OriginalModelPath = originalModel.Path()
	ss.OriginalModelName = originalModel.Name()
	ss.OriginalModel = originalModel
	ss.BakedModel = physicsBakedModel
}

// SetModels ドメインロジックでモデルを設定（公開メソッド）
func (ss *BakeSet) SetModels(originalModel, bakedModel *pmx.PmxModel) error {
	if originalModel == nil {
		ss.setModels(nil, nil)
		return nil
	}

	// ヘルパーを使用してビジネスロジックを実行
	if err := ss.helper.ProcessModelsForBakeSet(originalModel, bakedModel); err != nil {
		return err
	}

	ss.setModels(originalModel, bakedModel)
	ss.OutputModelPath = ss.helper.CreateOutputModelPath(originalModel)

	return nil
}

// ClearModels モデルをクリア（公開メソッド）
func (ss *BakeSet) ClearModels() {
	ss.setModels(nil, nil)
}

// SetMotions ドメインロジックでモーションを設定（公開メソッド）
func (ss *BakeSet) SetMotions(originalMotion, outputMotion *vmd.VmdMotion) error {
	ss.setMotion(originalMotion, outputMotion)
	ss.OutputMotionPath = ss.CreateOutputMotionPath()
	return nil
}

// ClearMotions モーションをクリア（公開メソッド）
func (ss *BakeSet) ClearMotions() {
	ss.setMotion(nil, nil)
}

func (ss *BakeSet) Delete() {
	ss.OriginalMotionPath = ""
	ss.OriginalModelPath = ""
	ss.OutputModelPath = ""
	ss.OutputMotionPath = ""

	ss.OriginalMotionName = ""
	ss.OriginalModelName = ""
	ss.OutputModelName = ""

	ss.OriginalMotion = nil
	ss.OriginalModel = nil
	ss.OutputMotion = nil
}

// GetOutputMotionOnlyChecked 物理ボーンだけ残す（ヘルパーに委譲）
func (ss *BakeSet) GetOutputMotionOnlyChecked(records []*OutputBoneRecord) ([]*vmd.VmdMotion, error) {
	return ss.helper.ProcessOutputMotion(
		ss.OriginalModel,
		ss.OriginalMotion,
		ss.OutputMotion,
		ss.OutputMotionPath,
		records,
	)
}
