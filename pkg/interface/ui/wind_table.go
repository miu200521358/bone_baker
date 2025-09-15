package ui

import (
	"fmt"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// createWindTableView テーブルビューを作成
func createWindTableView(store *WidgetStore) declarative.TableView {
	return declarative.TableView{
		AssignTo:         &store.WindTableView,
		Model:            newWindTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 80},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("風向き"), Width: 120},
			{Title: mi18n.T("風速"), Width: 100},
			{Title: mi18n.T("ランダム"), Width: 100},
			{Title: mi18n.T("乱流周波数"), Width: 100},
			{Title: mi18n.T("抗力係数"), Width: 100},
			{Title: mi18n.T("揚力係数"), Width: 100},
		},
		OnItemClicked: createWindTableViewDialog(store, false),
	}
}
func createWindTableViewDialog(store *WidgetStore, isAdd bool) func() {
	return func() {
		var record *entity.WindRecord
		recordIndex := -1
		switch isAdd {
		case true:
			if store.currentSet().OriginalMotion == nil {
				record = entity.NewWindRecord(0, 0)
			} else {
				record = entity.NewWindRecord(store.minFrame(), store.maxFrame())
			}
		case false:
			record = store.WindRecords[store.WindTableView.CurrentIndex()]
			recordIndex = store.WindTableView.CurrentIndex()
		}
		dialog := newWindTableViewDialog(store)
		dialog.show(record, recordIndex)
	}
}

type WindTableModel struct {
	walk.TableModelBase
	Records  []*entity.WindRecord // 物理ボーンレコード
	tv       *walk.TableView      // テーブルビュー
	TreeView *walk.TreeView       // 物理ボーンツリー
}

func newWindTableModel() *WindTableModel {
	m := new(WindTableModel)
	m.Records = make([]*entity.WindRecord, 0)
	return m
}

func newWindTableModelWithRecords(records []*entity.WindRecord) *WindTableModel {
	m := new(WindTableModel)
	m.Records = records
	return m
}

func (m *WindTableModel) RowCount() int {
	return len(m.Records)
}

func (m *WindTableModel) SetParent(parent *walk.TableView) {
	m.tv = parent
}

func (m *WindTableModel) Value(row, col int) any {
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
		return fmt.Sprintf("X:%.2f Y:%.2f Z:%.2f",
			item.WindConfig.Direction.X,
			item.WindConfig.Direction.Y,
			item.WindConfig.Direction.Z)
	case 4:
		return item.WindConfig.Speed
	case 5:
		return item.WindConfig.Randomness
	case 6:
		return item.WindConfig.TurbulenceFreqHz
	case 7:
		return item.WindConfig.DragCoeff
	case 8:
		return item.WindConfig.LiftCoeff
	}

	panic("unexpected col")
}
