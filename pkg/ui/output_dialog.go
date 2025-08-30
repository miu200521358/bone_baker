package ui

import (
	"time"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// OutputTableViewDialog 出力設定ダイアログのロジックを管理
type OutputTableViewDialog struct {
	bakeState *BakeState
	mWidgets  *controller.MWidgets
}

// NewOutputTableViewDialog コンストラクタ
func NewOutputTableViewDialog(bakeState *BakeState, mWidgets *controller.MWidgets) *OutputTableViewDialog {
	return &OutputTableViewDialog{
		bakeState: bakeState,
		mWidgets:  mWidgets,
	}
}

// Show 出力設定ダイアログを表示
func (o *OutputTableViewDialog) Show() {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var cancelBtn *walk.PushButton
	var okBtn *walk.PushButton
	var db *walk.DataBinder
	var treeView *walk.TreeView
	var ikCheckBox *walk.CheckBox
	var physicsCheckBox *walk.CheckBox
	var standardCheckBox *walk.CheckBox
	var fingerCheckBox *walk.CheckBox

	builder := declarative.NewBuilder(o.mWidgets.Window())

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("焼き込み設定変更"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 600, Height: 200},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: o.bakeState.CurrentSet().OutputTableModel.Records[o.bakeState.OutputTableView.CurrentIndex()],
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:   declarative.Grid{Columns: 5},
				Children: o.createFormWidgets(&treeView, &ikCheckBox, &physicsCheckBox, &standardCheckBox, &fingerCheckBox),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: o.createButtonWidgets(&okBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if cmd, err := dialog.RunWithFunc(builder.Parent().Form(), func(dialog *walk.Dialog) {
		// ダイアログが完全に表示された後に実行
		go func() {
			// 少し待ってからチェック状態を適用
			for range 5 {
				time.Sleep(10 * time.Millisecond)
				treeView.Synchronize(func() {
					treeView.ApplyRootCheckStates()
				})
			}
		}()
	}); err == nil && cmd == walk.DlgCmdOK {
		o.handleDialogOK()
	}
}

// createFormWidgets フォームウィジェットを作成
func (o *OutputTableViewDialog) createFormWidgets(treeView **walk.TreeView,
	ikCheckBox, physicsCheckBox, standardCheckBox, fingerCheckBox **walk.CheckBox) []declarative.Widget {
	return []declarative.Widget{
		declarative.Label{
			Text: mi18n.T("出力開始フレーム"),
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("StartFrame"),
			ToolTipText:        mi18n.T("出力開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(o.bakeState.CurrentSet().MaxFrame() + 1),
		},
		declarative.Label{
			Text: mi18n.T("出力終了フレーム"),
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			ToolTipText:        mi18n.T("出力終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(o.bakeState.CurrentSet().MaxFrame() + 1),
		},
		declarative.HSpacer{
			ColumnSpan: 1,
		},
		declarative.Label{
			Text: mi18n.T("焼き込み対象ボーン"),
		},
		declarative.CheckBox{
			AssignTo: physicsCheckBox,
			Text:     mi18n.T("物理焼き込み対象"),
			OnClicked: func() {
				(*treeView).Model().(*domain.OutputBoneTreeModel).SetOutputPhysicsChecked(*treeView, nil, (*physicsCheckBox).Checked())
			},
		},
		declarative.CheckBox{
			AssignTo: ikCheckBox,
			Text:     mi18n.T("IK焼き込み対象"),
			OnClicked: func() {
				(*treeView).Model().(*domain.OutputBoneTreeModel).SetOutputIkChecked(*treeView, nil, (*ikCheckBox).Checked())
			},
		},
		declarative.CheckBox{
			AssignTo: standardCheckBox,
			Text:     mi18n.T("準標準焼き込み対象"),
			OnClicked: func() {
				(*treeView).Model().(*domain.OutputBoneTreeModel).SetOutputStandardChecked(*treeView, nil, (*standardCheckBox).Checked())
			},
		},
		declarative.CheckBox{
			AssignTo: fingerCheckBox,
			Text:     mi18n.T("指焼き込み対象"),
			OnClicked: func() {
				(*treeView).Model().(*domain.OutputBoneTreeModel).SetOutputFingerChecked(*treeView, nil, (*fingerCheckBox).Checked())
			},
		},
		declarative.TreeView{
			AssignTo:   treeView,
			Model:      o.bakeState.CurrentSet().OutputTableModel.Records[o.bakeState.OutputTableView.CurrentIndex()].OutputBoneTreeModel,
			MinSize:    declarative.Size{Width: 230, Height: 200},
			Checkable:  true,
			ColumnSpan: 6,
		},
	}
}

// createButtonWidgets ボタンウィジェットを作成
func (o *OutputTableViewDialog) createButtonWidgets(okBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo: okBtn,
			Text:     mi18n.T("登録"),
			OnClicked: func() {
				if err := (*db).Submit(); err != nil {
					mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
					return
				}
				(*dlg).Accept()
			},
		},
		declarative.PushButton{
			AssignTo: cancelBtn,
			Text:     mi18n.T("削除"),
			OnClicked: func() {
				// 削除処理
				o.bakeState.CurrentSet().OutputTableModel.RemoveRow(o.bakeState.OutputTableView.CurrentIndex())
				if err := (*db).Submit(); err != nil {
					mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
					return
				}
				(*dlg).Accept()
			},
		},
		declarative.PushButton{
			AssignTo: cancelBtn,
			Text:     mi18n.T("キャンセル"),
			OnClicked: func() {
				(*dlg).Cancel()
			},
		},
	}
}

// handleDialogOK ダイアログOK処理
func (o *OutputTableViewDialog) handleDialogOK() {
	// チェックされたボーン名一覧を取得
	currentIndex := o.bakeState.OutputTableView.CurrentIndex()
	o.bakeState.CurrentSet().OutputTableModel.Records[currentIndex].TargetBoneNames =
		o.bakeState.CurrentSet().OutputTableModel.Records[currentIndex].OutputBoneTreeModel.GetCheckedBoneNames()

	// 更新
	o.bakeState.OutputTableView.SetModel(o.bakeState.CurrentSet().OutputTableModel)
}
