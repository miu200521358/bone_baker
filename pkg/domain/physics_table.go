package domain

import (
	"strings"

	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type PhysicsTableModel struct {
	walk.TableModelBase
	tv       *walk.TableView      `json:"-"`       // テーブルビュー
	Records  []*PhysicsBoneRecord `json:"records"` // 物理ボーンレコード
	TreeView *walk.TreeView       `json:"-"`       // 物理ボーンツリー
}

func NewPhysicsTableModel() *PhysicsTableModel {
	m := new(PhysicsTableModel)
	m.Records = make([]*PhysicsBoneRecord, 0)
	return m
}

func (m *PhysicsTableModel) RowCount() int {
	return len(m.Records)
}

func (m *PhysicsTableModel) SetParent(parent *walk.TableView) {
	m.tv = parent
}

func (m *PhysicsTableModel) Value(row, col int) any {
	if row < 0 || row >= len(m.Records) {
		return nil
	}

	item := m.Records[row]

	switch col {
	case 0:
		return row + 1 // 行番号
	case 1:
		return int(item.StartFrame)
	case 2:
		return int(item.EndFrame)
	case 3:
		return item.Gravity
	case 4:
		return item.MaxSubSteps
	case 5:
		return item.FixedTimeStep
	case 6:
		return item.IsStartDeform
	case 7:
		nodes := item.TreeModel.ModifiedNodes(nil)
		nodeNames := make([]string, 0, len(nodes))
		for _, n := range nodes {
			nodeNames = append(nodeNames, n.RigidBodyName)
		}
		return strings.Join(nodeNames, ", ")
	}

	panic("unexpected col")
}

func NewPhysicsBoneRecord(model *pmx.PmxModel, startFrame, endFrame float32) *PhysicsBoneRecord {
	return &PhysicsBoneRecord{
		StartFrame:    startFrame,
		EndFrame:      endFrame,
		MaxStartFrame: startFrame,
		MaxEndFrame:   endFrame,
		Gravity:       -9.8,  // 重力の初期値
		MaxSubSteps:   2,     // 最大演算回数の初期値
		FixedTimeStep: 60,    // 固定フレーム時間の初期値
		IsStartDeform: false, // 開始用整形の初期値
		TreeModel:     NewPhysicsRigidBodyTreeModel(model),
	}
}

func (m *PhysicsTableModel) AddRecord(model *pmx.PmxModel, startFrame, endFrame float32) {
	m.Records = append(m.Records, NewPhysicsBoneRecord(model, startFrame, endFrame))
}

func (m *PhysicsTableModel) RemoveRow(index int) {
	if index < 0 || index >= len(m.Records) {
		return
	}
	m.Records = append(m.Records[:index], m.Records[index+1:]...)
	if m.tv != nil {
		m.tv.SetModel(m) // モデルを更新
	}
}

type PhysicsBoneRecord struct {
	StartFrame     float32                    `json:"start_frame"`     // 区間開始フレーム
	EndFrame       float32                    `json:"end_frame"`       // 区間終了フレーム
	MaxStartFrame  float32                    `json:"max_start_frame"` // 最大値開始フレーム
	MaxEndFrame    float32                    `json:"max_end_frame"`   // 最大値終了フレーム
	Gravity        float64                    `json:"gravity"`         // 重力
	MaxSubSteps    int                        `json:"max_sub_steps"`   // 最大演算回数
	FixedTimeStep  float64                    `json:"fixed_time_step"` // 物理演算頻度
	IsStartDeform  bool                       `json:"is_start_deform"` // 開始用整形有無
	SizeRatio      *mmath.MVec3               `json:"size_ratio"`      // 大きさの比率
	MassRatio      float64                    `json:"mass_ratio"`      // 質量の比率
	TensionRatio   float64                    `json:"tension_ratio"`   // 張りの比率
	StiffnessRatio float64                    `json:"stiffness_ratio"` // 硬さの比率
	TreeModel      *PhysicsRigidBodyTreeModel `json:"tree"`            // 出力ボーンツリー
}
