package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
)

type BakeSet struct {
	Index       int  // インデックス
	IsTerminate bool // 処理停止フラグ

	// Value Objectsを使用したファイルパス
	originalMotionPath *FilePath `json:"-"` // 元モーションパス（Value Object）
	originalModelPath  *FilePath `json:"-"` // 元モデルパス（Value Object）
	outputMotionPath   *FilePath `json:"-"` // 出力モーションパス（Value Object）
	outputModelPath    *FilePath `json:"-"` // 出力モデルパス（Value Object）

	// JSONシリアライズ用の文字列フィールド（後方互換性）
	OriginalMotionPathStr string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPathStr  string `json:"original_model_path"`  // 元モデルパス

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
		Index:              index,
		PhysicsTableModel:  NewPhysicsTableModel(),
		OutputTableModel:   NewOutputTableModel(),
		originalMotionPath: NewFilePath(""),
		originalModelPath:  NewFilePath(""),
		outputMotionPath:   NewFilePath(""),
		outputModelPath:    NewFilePath(""),
		helper:             NewBakeSetHelper(), // Helper注入
	}
}

// Getter methods for Value Objects
func (ss *BakeSet) OriginalMotionPath() string {
	if ss.originalMotionPath == nil {
		return ""
	}
	return ss.originalMotionPath.Value()
}

func (ss *BakeSet) OriginalModelPath() string {
	if ss.originalModelPath == nil {
		return ""
	}
	return ss.originalModelPath.Value()
}

func (ss *BakeSet) OutputMotionPath() string {
	if ss.outputMotionPath == nil {
		return ""
	}
	return ss.outputMotionPath.Value()
}

func (ss *BakeSet) OutputModelPath() string {
	if ss.outputModelPath == nil {
		return ""
	}
	return ss.outputModelPath.Value()
}

// Setter methods for Value Objects
func (ss *BakeSet) SetOriginalMotionPath(path string) {
	ss.originalMotionPath = NewFilePath(path)
	ss.OriginalMotionPathStr = path
}

func (ss *BakeSet) SetOriginalModelPath(path string) {
	ss.originalModelPath = NewFilePath(path)
	ss.OriginalModelPathStr = path
}

func (ss *BakeSet) SetOutputMotionPath(path string) {
	ss.outputMotionPath = NewFilePath(path)
}

func (ss *BakeSet) SetOutputModelPath(path string) {
	ss.outputModelPath = NewFilePath(path)
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
		ss.SetOriginalMotionPath("")
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.SetOutputMotionPath("")
		ss.OutputMotion = vmd.NewVmdMotion("")

		return
	}

	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion
	ss.OutputMotion = outputMotion
}

func (ss *BakeSet) setModels(originalModel, physicsBakedModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.SetOriginalModelPath("")
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.BakedModel = nil
		return
	}

	ss.SetOriginalModelPath(originalModel.Path())
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
	ss.SetOutputModelPath(ss.helper.CreateOutputModelPath(originalModel))

	return nil
}

// ClearModels モデルをクリア（公開メソッド）
func (ss *BakeSet) ClearModels() {
	ss.setModels(nil, nil)
}

// SetMotions ドメインロジックでモーションを設定（公開メソッド）
func (ss *BakeSet) SetMotions(originalMotion, outputMotion *vmd.VmdMotion) error {
	ss.setMotion(originalMotion, outputMotion)
	ss.SetOutputMotionPath(ss.CreateOutputMotionPath())
	return nil
}

// ClearMotions モーションをクリア（公開メソッド）
func (ss *BakeSet) ClearMotions() {
	ss.setMotion(nil, nil)
}

func (ss *BakeSet) Delete() {
	ss.SetOriginalMotionPath("")
	ss.SetOriginalModelPath("")
	ss.SetOutputMotionPath("")
	ss.SetOutputModelPath("")

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
		ss.OutputMotionPath(),
		records,
	)
}
