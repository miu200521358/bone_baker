package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// createOutputTableView テーブルビューを作成
func createOutputTableView(store *WidgetStore) declarative.TableView {
	return declarative.TableView{
		AssignTo:         &store.OutputTableView,
		Model:            newOutputTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 80},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("出力対象ボーン"), Width: 200},
		},
		OnItemClicked: createOutputTableViewDialog(store, false),
	}
}

func createOutputTableViewDialog(store *WidgetStore, isAdd bool) func() {
	return func() {
		var record *entity.OutputRecord
		recordIndex := -1
		switch isAdd {
		case true:
			if store.currentSet().OriginalMotion == nil || store.currentSet().OriginalModel == nil {
				record = entity.NewOutputRecord(0, 0, nil)
			} else {
				record = entity.NewOutputRecord(
					store.currentSet().OriginalMotion.MinFrame(),
					store.currentSet().OriginalMotion.MaxFrame(),
					store.currentSet().OriginalModel)
			}
		case false:
			record = store.currentSet().OutputRecords[store.OutputTableView.CurrentIndex()]
			recordIndex = store.OutputTableView.CurrentIndex()
		}
		dialog := newOutputTableViewDialog(store)
		dialog.show(record, recordIndex)
	}
}

type OutputTableModel struct {
	walk.TableModelBase
	Records  []*entity.OutputRecord // 出力ボーンレコード
	tv       *walk.TableView        // テーブルビュー
	TreeView *walk.TreeView         // 出力ボーンツリー
}

func newOutputTableModel() *OutputTableModel {
	m := new(OutputTableModel)
	m.Records = make([]*entity.OutputRecord, 0)
	return m
}

func newOutputTableModelWithRecords(records []*entity.OutputRecord) *OutputTableModel {
	m := new(OutputTableModel)
	m.Records = records
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
		return item.ItemNames()
	}

	panic("unexpected col")
}
