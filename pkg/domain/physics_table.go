package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type PhysicsTableModel struct {
	walk.TableModelBase
	tv       *walk.TableView
	Records  []*PhysicsBoneRecord
	TreeView *walk.TreeView // 物理ボーンツリー
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
	}

	panic("unexpected col")
}

func (m *PhysicsTableModel) AddRecord(model *pmx.PmxModel, startFrame, endFrame float32) {
	item := &PhysicsBoneRecord{
		StartFrame:    startFrame,
		EndFrame:      endFrame,
		Gravity:       -9.8,  // 重力の初期値
		MaxSubSteps:   2,     // 最大演算回数の初期値
		FixedTimeStep: 60,    // 固定フレーム時間の初期値
		IsStartDeform: false, // 開始用整形の初期値
		TreeModel:     NewPhysicsRigidBodyTreeModel(model),
	}
	m.Records = append(m.Records, item)
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
	StartFrame     float32
	EndFrame       float32
	Gravity        float64
	MaxSubSteps    int
	FixedTimeStep  float64
	IsStartDeform  bool                       // 開始用整形有無
	SizeRatio      *mmath.MVec3               // 大きさの比率
	MassRatio      float64                    // 質量の比率
	TensionRatio   float64                    // 張りの比率
	StiffnessRatio float64                    // 硬さの比率
	TreeModel      *PhysicsRigidBodyTreeModel // 出力ボーンツリー
}
