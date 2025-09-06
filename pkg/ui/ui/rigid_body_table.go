package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// createRigidBodyTableView テーブルビューを作成
func createRigidBodyTableView(store *WidgetStore) declarative.TableView {
	return declarative.TableView{
		AssignTo:         &store.RigidBodyTableView,
		Model:            newRigidBodyTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 150},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("最大開始F"), Width: 60},
			{Title: mi18n.T("最大終了F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("対象剛体名"), Width: 300},
		},
		OnItemClicked: createRigidBodyTableViewDialog(store, false),
	}
}

func createRigidBodyTableViewDialog(store *WidgetStore, isAdd bool) func() {
	return func() {
		var record *entity.RigidBodyRecord
		recordIndex := -1
		switch isAdd {
		case true:
			if store.currentSet().OriginalMotion == nil {
				record = entity.NewRigidBodyRecord(0, 0)
			} else {
				record = entity.NewRigidBodyRecord(
					store.currentSet().OriginalMotion.MinFrame(),
					store.currentSet().OriginalMotion.MaxFrame())
			}
		case false:
			record = store.currentSet().RigidBodyRecords[store.RigidBodyTableView.CurrentIndex()]
			recordIndex = store.RigidBodyTableView.CurrentIndex()
		}
		dialog := NewRigidBodyTableViewDialog(store)
		dialog.Show(record, recordIndex)
	}
}

type RigidBodyTableModel struct {
	walk.TableModelBase
	Records  []*entity.RigidBodyRecord // 物理ボーンレコード
	tv       *walk.TableView           // テーブルビュー
	TreeView *walk.TreeView            // 物理ボーンツリー
}

func newRigidBodyTableModel() *RigidBodyTableModel {
	m := new(RigidBodyTableModel)
	m.Records = make([]*entity.RigidBodyRecord, 0)
	return m
}

func newRigidBodyTableModelWithRecords(records []*entity.RigidBodyRecord) *RigidBodyTableModel {
	m := new(RigidBodyTableModel)
	m.Records = records
	return m
}

func (m *RigidBodyTableModel) RowCount() int {
	return len(m.Records)
}

func (m *RigidBodyTableModel) SetParent(parent *walk.TableView) {
	m.tv = parent
}

func (m *RigidBodyTableModel) Value(row, col int) any {
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
		return int(item.MaxStartFrame)
	case 3:
		return int(item.MaxEndFrame)
	case 4:
		return int(item.EndFrame)
	case 5:
		return item.ItemNames()
	}

	panic("unexpected col")
}
