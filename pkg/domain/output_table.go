package domain

import (
	"strings"

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
	item := m.Records[row]

	switch col {
	case 0:
		return row + 1 // 行番号
	case 1:
		return int(item.StartFrame)
	case 2:
		return int(item.EndFrame)
	case 3:
		return item.ResetFrame
	case 4:
		return len(item.TargetBoneNames)
	case 5:
		return strings.Join(item.TargetBoneNames, ", ")
	}

	panic("unexpected col")
}

func (m *OutputTableModel) AddRecord(startFrame, endFrame float32) {
	item := &OutputBoneRecord{
		Checked:    false,
		StartFrame: startFrame,
		EndFrame:   endFrame,
		ResetFrame: -5, // 初期値として5F前のリセットを入れる
	}
	m.Records = append(m.Records, item)
}

func (m *OutputTableModel) Checked(row int) bool {
	return m.Records[row].Checked
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
	Checked         bool
	StartFrame      float32
	EndFrame        float32
	ResetFrame      int
	TargetBoneNames []string
}
