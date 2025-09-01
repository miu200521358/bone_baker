package physics

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// CreateTableViews テーブルビューを作成
func CreatePhysicsTableView(bakeSet *entity.BakeSet, physicsTableView *walk.TableView, mWidgets *controller.MWidgets) declarative.TableView {
	return declarative.TableView{
		AssignTo:         &physicsTableView,
		Model:            NewPhysicsTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 150},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("重力"), Width: 60},
			{Title: mi18n.T("最大演算回数"), Width: 100},
			{Title: mi18n.T("物理演算頻度"), Width: 100},
		},
		OnItemClicked: createPhysicsTableViewDialog(bakeSet, physicsTableView, mWidgets, false),
	}
}

func createPhysicsTableViewDialog(bakeSet *entity.BakeSet, physicsTableView *walk.TableView, mWidgets *controller.MWidgets, isAdd bool) func() {
	return func() {
		var record *entity.PhysicsRecord
		recordIndex := -1
		switch isAdd {
		case true:
			if bakeSet.OriginalMotion == nil {
				record = entity.NewPhysicsRecord(0, 0)
			} else {
				record = entity.NewPhysicsRecord(
					bakeSet.OriginalMotion.MinFrame(),
					bakeSet.OriginalMotion.MaxFrame())
			}
		case false:
			record = bakeSet.PhysicsRecords[physicsTableView.CurrentIndex()]
			recordIndex = physicsTableView.CurrentIndex()
		}
		dialog := newPhysicsTableViewDialog(bakeSet, mWidgets.Window())
		dialog.Show(record, recordIndex)
	}
}

type PhysicsTableModel struct {
	walk.TableModelBase
	Records  []*entity.PhysicsRecord // 物理ボーンレコード
	tv       *walk.TableView         // テーブルビュー
	TreeView *walk.TreeView          // 物理ボーンツリー
}

func NewPhysicsTableModel() *PhysicsTableModel {
	m := new(PhysicsTableModel)
	m.Records = make([]*entity.PhysicsRecord, 0)
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
	}

	panic("unexpected col")
}
