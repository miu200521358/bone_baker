package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// RigidBodyTableViewDialog 剛体設定ダイアログのロジックを管理
type RigidBodyTableViewDialog struct {
	store    *WidgetStore
	doDelete bool
}

// NewRigidBodyTableViewDialog コンストラクタ
func NewRigidBodyTableViewDialog(store *WidgetStore) *RigidBodyTableViewDialog {
	return &RigidBodyTableViewDialog{
		store: store,
	}
}

// Show 剛体設定ダイアログを表示
func (p *RigidBodyTableViewDialog) Show(record *entity.RigidBodyRecord, recordIndex int) {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var okBtn *walk.PushButton
	var deleteBtn *walk.PushButton
	var cancelBtn *walk.PushButton
	var db *walk.DataBinder
	var startFrameEdit *walk.NumberEdit    // 開始フレーム入力
	var endFrameEdit *walk.NumberEdit      // 終了フレーム入力
	var gravityEdit *walk.NumberEdit       // 重力値入力
	var maxSubStepsEdit *walk.NumberEdit   // 最大最大演算回数
	var fixedTimeStepEdit *walk.NumberEdit // 固定タイムステップ入力

	builder := declarative.NewBuilder(p.store.Window())

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("剛体設定"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 250, Height: 250},
		MaxSize:       declarative.Size{Width: 250, Height: 250},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: record,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.Grid{Columns: 2},
				Children: p.createFormWidgets(&startFrameEdit, &endFrameEdit,
					&gravityEdit, &maxSubStepsEdit, &fixedTimeStepEdit),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: p.createButtonWidgets(&okBtn, &deleteBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if cmd, err := dialog.Run(builder.Parent().Form()); err == nil && cmd == walk.DlgCmdOK {
		p.handleDialogOK(record, recordIndex)
	}
}

func (p *RigidBodyTableViewDialog) createFormWidgets(startFrameEdit, endFrameEdit,
	gravityEdit, maxSubStepsEdit, fixedTimeStepEdit **walk.NumberEdit) []declarative.Widget {

	return []declarative.Widget{
		declarative.Label{
			Text:        mi18n.T("設定開始フレーム"),
			ToolTipText: mi18n.T("設定開始フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("設定開始フレーム説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("StartFrame"),
			AssignTo:           startFrameEdit,
			ToolTipText:        mi18n.T("設定開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("設定最大開始フレーム"),
			ToolTipText: mi18n.T("設定最大開始フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("設定最大開始フレーム説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxStartFrame"),
			AssignTo:           startFrameEdit,
			ToolTipText:        mi18n.T("設定最大開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("設定最大終了フレーム"),
			ToolTipText: mi18n.T("設定最大終了フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("設定最大終了フレーム説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxEndFrame"),
			AssignTo:           endFrameEdit,
			ToolTipText:        mi18n.T("設定最大終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("設定終了フレーム"),
			ToolTipText: mi18n.T("設定終了フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("設定終了フレーム説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			AssignTo:           endFrameEdit,
			ToolTipText:        mi18n.T("設定終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
	}
}

func (p *RigidBodyTableViewDialog) createButtonWidgets(
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("剛体設定登録説明"),
			OnClicked: func() {
				if err := (*db).Submit(); err != nil {
					mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
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
			ToolTipText: mi18n.T("剛体設定削除説明"),
			OnClicked: func() {
				p.doDelete = true
				(*dlg).Accept()
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.PushButton{
			AssignTo:    cancelBtn,
			Text:        mi18n.T("キャンセル"),
			ToolTipText: mi18n.T("剛体設定キャンセル説明"),
			OnClicked: func() {
				(*dlg).Cancel()
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
	}
}

func (p *RigidBodyTableViewDialog) handleDialogOK(record *entity.RigidBodyRecord, recordIndex int) {
	p.store.setWidgetEnabled(false)

	p.store.setWidgetEnabled(true)
	controller.Beep()
}
