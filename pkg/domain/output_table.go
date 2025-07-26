package domain

import (
	"strings"

	"github.com/miu200521358/walk/pkg/walk"
)

type OutputTableModel struct {
	walk.TableModelBase
	tv      *walk.TableView
	Records []*OutputRecord
}

func NewOutputTableModel() *OutputTableModel {
	m := new(OutputTableModel)
	m.Records = make([]*OutputRecord, 0)
	m.AddRecord() // 初期行を追加
	return m
}

func (m *OutputTableModel) RowCount() int {
	return len(m.Records)
}

func (m *OutputTableModel) SetParent(parent *walk.TableView) {
	m.tv = parent
}

func (m *OutputTableModel) Value(row, col int) any {
	item := m.Records[row]

	switch col {
	case 0:
		return row + 1 // 行番号
	case 1:
		return item.StartFrame
	case 2:
		return item.EndFrame
	case 3:
		return item.ResetFrame
	case 4:
		return len(item.TargetBoneNames)
	case 5:
		return strings.Join(item.TargetBoneNames, ", ")
	}

	panic("unexpected col")
}

func (m *OutputTableModel) AddRecord() {
	item := &OutputRecord{}
	m.Records = append(m.Records, item)
}

type OutputRecord struct {
	StartFrame      int
	EndFrame        int
	ResetFrame      int
	TargetBoneNames []string
}
