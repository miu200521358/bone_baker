package domain

import (
	"strings"

	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type OutputTableModel struct {
	walk.TableModelBase
	tv      *walk.TableView
	Records []*OutputBoneRecord
}

func NewOutputTableModel() *OutputTableModel {
	m := new(OutputTableModel)
	m.Records = make([]*OutputBoneRecord, 0)
	return m
}

func (m *OutputTableModel) RowCount() int {
	return len(m.Records)
}

func (m *OutputTableModel) SetParent(parent *walk.TableView) {
	m.tv = parent
}

func (m *OutputTableModel) Value(row, col int) any {
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
		return len(item.TargetBoneNames)
	case 4:
		return strings.Join(item.TargetBoneNames, ", ")
	}

	panic("unexpected col")
}

func (m *OutputTableModel) AddRecord(model *pmx.PmxModel, startFrame, endFrame float32) {
	item := &OutputBoneRecord{
		StartFrame:          startFrame,
		EndFrame:            endFrame,
		OutputBoneTreeModel: NewOutputBoneTreeModel(model),
	}
	m.Records = append(m.Records, item)
}

func (m *OutputTableModel) RemoveRow(index int) {
	if index < 0 || index >= len(m.Records) {
		return
	}
	m.Records = append(m.Records[:index], m.Records[index+1:]...)
	if m.tv != nil {
		m.tv.SetModel(m) // モデルを更新
	}
}

type OutputBoneRecord struct {
	StartFrame          float32
	EndFrame            float32
	TargetBoneNames     []string
	OutputBoneTreeModel *OutputBoneTreeModel // 出力ボーンツリー
}
