package ui

import (
	"time"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// OutputTableViewDialog 出力設定ダイアログのロジックを管理
type OutputTableViewDialog struct {
	store    *WidgetStore
	doDelete bool
}

// newOutputTableViewDialog コンストラクタ
func newOutputTableViewDialog(store *WidgetStore) *OutputTableViewDialog {
	return &OutputTableViewDialog{
		store: store,
	}
}

// show 出力設定ダイアログを表示
func (p *OutputTableViewDialog) show(record *entity.OutputRecord, recordIndex int) {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var okBtn *walk.PushButton
	var deleteBtn *walk.PushButton
	var cancelBtn *walk.PushButton
	var db *walk.DataBinder
	var startFrameEdit *walk.NumberEdit // 開始フレーム入力
	var endFrameEdit *walk.NumberEdit   // 終了フレーム入力
	var treeView *walk.TreeView
	var ikCheckBox *walk.CheckBox
	var physicsCheckBox *walk.CheckBox
	var standardCheckBox *walk.CheckBox
	var fingerCheckBox *walk.CheckBox

	builder := declarative.NewBuilder(p.store.Window())
	treeModel := newOutputTreeModel(record)

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("出力設定"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 500, Height: 400},
		MaxSize:       declarative.Size{Width: 500, Height: 400},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: record,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:   declarative.Grid{Columns: 4},
				Children: p.createFormWidgets(&startFrameEdit, &endFrameEdit, &ikCheckBox, &physicsCheckBox, &standardCheckBox, &fingerCheckBox, &treeView, treeModel),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: p.createButtonWidgets(&startFrameEdit, &endFrameEdit, &okBtn, &deleteBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if cmd, err := dialog.RunWithFunc(builder.Parent().Form(), func(dialog *walk.Dialog) {
		// ダイアログが完全に表示された後に実行
		go func() {
			// 少し待ってからチェック状態を適用
			time.Sleep(10 * time.Millisecond)
			treeView.Synchronize(func() {
				treeView.ApplyRootCheckStates()
				treeView.ExpandAll()
			})
		}()
	}); err == nil && cmd == walk.DlgCmdOK {
		p.handleDialogOK(record, recordIndex)
	}
}

func (p *OutputTableViewDialog) createFormWidgets(startFrameEdit, endFrameEdit **walk.NumberEdit,
	ikCheckBox, physicsCheckBox, standardCheckBox, fingerCheckBox **walk.CheckBox, treeView **walk.TreeView, treeModel *OutputTreeModel) []declarative.Widget {

	return []declarative.Widget{
		declarative.Label{
			Text:        mi18n.T("開始フレーム"),
			ToolTipText: mi18n.T("開始フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("開始フレーム説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("StartFrame"),
			AssignTo:           startFrameEdit,
			ToolTipText:        mi18n.T("開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame()),
			DefaultValue:       float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("終了フレーム"),
			ToolTipText: mi18n.T("終了フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("終了フレーム説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			AssignTo:           endFrameEdit,
			ToolTipText:        mi18n.T("終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			DefaultValue:       float64(p.store.currentSet().OriginalMotion.MaxFrame()),
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
		},
		declarative.Label{
			Text: mi18n.T("出力対象ボーン"),
		},
		declarative.HSpacer{
			ColumnSpan: 3,
		},
		declarative.CheckBox{
			AssignTo:    physicsCheckBox,
			Text:        mi18n.T("物理焼き込み対象"),
			ToolTipText: mi18n.T("物理焼き込み対象説明"),
			OnClicked: func() {
				(*treeView).Model().(*OutputTreeModel).SetOutputPhysicsChecked(*treeView, nil, (*physicsCheckBox).Checked())
			},
		},
		declarative.CheckBox{
			AssignTo:    ikCheckBox,
			Text:        mi18n.T("IK焼き込み対象"),
			ToolTipText: mi18n.T("IK焼き込み対象説明"),
			OnClicked: func() {
				(*treeView).Model().(*OutputTreeModel).SetOutputIkChecked(*treeView, nil, (*ikCheckBox).Checked())
			},
		},
		declarative.CheckBox{
			AssignTo:    standardCheckBox,
			Text:        mi18n.T("準標準焼き込み対象"),
			ToolTipText: mi18n.T("準標準焼き込み対象説明"),
			OnClicked: func() {
				(*treeView).Model().(*OutputTreeModel).SetOutputStandardChecked(*treeView, nil, (*standardCheckBox).Checked())
			},
		},
		declarative.CheckBox{
			AssignTo:    fingerCheckBox,
			Text:        mi18n.T("指焼き込み対象"),
			ToolTipText: mi18n.T("指焼き込み対象説明"),
			OnClicked: func() {
				(*treeView).Model().(*OutputTreeModel).SetOutputFingerChecked(*treeView, nil, (*fingerCheckBox).Checked())
			},
		},
		declarative.TreeView{
			AssignTo:   treeView,
			Model:      treeModel,
			MinSize:    declarative.Size{Width: 450, Height: 200},
			Checkable:  true,
			ColumnSpan: 4,
		},
	}
}

func (p *OutputTableViewDialog) createButtonWidgets(
	startFrameEdit, endFrameEdit **walk.NumberEdit,
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("出力設定登録説明"),
			OnClicked: func() {
				if !((*startFrameEdit).Value() < (*endFrameEdit).Value()) {
					mlog.E(mi18n.T("出力範囲設定エラー"), nil, "")
					return
				}

				if err := (*db).Submit(); err != nil {
					mlog.E(mi18n.T("焼き込み設定変更エラー"), err, "")
					return
				}
				(*dlg).Accept()
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.PushButton{
			AssignTo:    deleteBtn,
			Text:        mi18n.T("削除"),
			ToolTipText: mi18n.T("出力設定削除説明"),
			OnClicked: func() {
				p.doDelete = true
				(*dlg).Cancel()
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.PushButton{
			AssignTo:    cancelBtn,
			Text:        mi18n.T("キャンセル"),
			ToolTipText: mi18n.T("出力設定キャンセル説明"),
			OnClicked: func() {
				(*dlg).Cancel()
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
	}
}

func (p *OutputTableViewDialog) handleDialogOK(record *entity.OutputRecord, recordIndex int) {
	p.store.setWidgetEnabled(false)

	if p.doDelete {
		p.store.currentSet().OutputRecords = append(p.store.currentSet().OutputRecords[:recordIndex], p.store.currentSet().OutputRecords[recordIndex+1:]...)
	} else {
		// 追加・更新処理
		if recordIndex == -1 {
			// 新規追加
			p.store.currentSet().OutputRecords = append(p.store.currentSet().OutputRecords, record)
		} else {
			// 更新
			if recordIndex >= 0 && recordIndex < len(p.store.currentSet().OutputRecords) {
				p.store.currentSet().OutputRecords[recordIndex] = record
			}
		}
	}

	p.store.setWidgetEnabled(true)

	// 削除フラグをリセット
	p.doDelete = false

	// 更新
	p.store.OutputTableView.SetModel(newOutputTableModelWithRecords(p.store.currentSet().OutputRecords))
}
