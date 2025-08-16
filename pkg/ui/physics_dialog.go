package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// PhysicsTableViewDialog 物理設定ダイアログのロジックを管理
type PhysicsTableViewDialog struct {
	bakeState *BakeState
	mWidgets  *controller.MWidgets
}

// NewPhysicsTableViewDialog コンストラクタ
func NewPhysicsTableViewDialog(bakeState *BakeState, mWidgets *controller.MWidgets) *PhysicsTableViewDialog {
	return &PhysicsTableViewDialog{
		bakeState: bakeState,
		mWidgets:  mWidgets,
	}
}

// Show 物理設定ダイアログを表示
func (p *PhysicsTableViewDialog) Show() {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var cancelBtn *walk.PushButton
	var okBtn *walk.PushButton
	var db *walk.DataBinder
	var treeView *walk.TreeView
	var gravityEdit *walk.NumberEdit       // 重力値入力
	var sizeXEdit *walk.NumberEdit         // 大きさX入力
	var sizeYEdit *walk.NumberEdit         // 大きさY入力
	var sizeZEdit *walk.NumberEdit         // 大きさZ入力
	var massEdit *walk.NumberEdit          // 質量入力
	var stiffnessEdit *walk.NumberEdit     // 硬さ入力
	var tensionEdit *walk.NumberEdit       // 張り入力
	var maxSubStepsEdit *walk.NumberEdit   // 最大最大演算回数
	var fixedTimeStepEdit *walk.NumberEdit // 固定タイムステップ入力

	builder := declarative.NewBuilder(p.mWidgets.Window())

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("物理設定変更"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 600, Height: 300},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: p.bakeState.CurrentSet().PhysicsTableModel.Records[p.bakeState.PhysicsTableView.CurrentIndex()],
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:   declarative.Grid{Columns: 6},
				Children: p.createFormWidgets(&gravityEdit, &sizeXEdit, &sizeYEdit, &sizeZEdit, &massEdit, &stiffnessEdit, &tensionEdit, &maxSubStepsEdit, &fixedTimeStepEdit, &treeView),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: p.createButtonWidgets(&okBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if cmd, err := dialog.Run(builder.Parent().Form()); err == nil && cmd == walk.DlgCmdOK {
		p.handleDialogOK()
	}
}

// createFormWidgets フォームウィジェットを作成
func (p *PhysicsTableViewDialog) createFormWidgets(gravityEdit, sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit, maxSubStepsEdit, fixedTimeStepEdit **walk.NumberEdit, treeView **walk.TreeView) []declarative.Widget {
	widgets := []declarative.Widget{
		declarative.Label{
			Text:        mi18n.T("設定開始フレーム"),
			ToolTipText: mi18n.T("設定開始フレーム説明"),
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("StartFrame"),
			ToolTipText:        mi18n.T("設定開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.bakeState.CurrentSet().MaxFrame() + 1),
		},
		declarative.Label{
			Text:        mi18n.T("設定終了フレーム"),
			ToolTipText: mi18n.T("設定終了フレーム説明"),
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			ToolTipText:        mi18n.T("設定終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.bakeState.CurrentSet().MaxFrame() + 1),
		},
		declarative.Label{
			Text:        mi18n.T("開始時用整形"),
			ToolTipText: mi18n.T("開始時用整形説明"),
		},
		declarative.CheckBox{
			Checked:     declarative.Bind("IsStartDeform"),
			ToolTipText: mi18n.T("開始時用整形説明"),
		},
		declarative.VSeparator{
			ColumnSpan: 6,
		},
		declarative.TreeView{
			AssignTo:   treeView,
			Model:      p.bakeState.CurrentSet().PhysicsTableModel.Records[p.bakeState.PhysicsTableView.CurrentIndex()].TreeModel,
			MinSize:    declarative.Size{Width: 230, Height: 200},
			ColumnSpan: 6,
			OnCurrentItemChanged: func() {
				p.updateEditValues(*treeView, *sizeXEdit, *sizeYEdit, *sizeZEdit, *massEdit, *stiffnessEdit, *tensionEdit)
			},
		},
	}

	// 物理編集ウィジェットを追加
	widgets = append(widgets, p.createPhysicsEditWidgets(sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit, treeView)...)

	return widgets
}

// createPhysicsEditWidgets 物理編集ウィジェットを作成
func (p *PhysicsTableViewDialog) createPhysicsEditWidgets(sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit **walk.NumberEdit, treeView **walk.TreeView) []declarative.Widget {
	return []declarative.Widget{
		declarative.TextLabel{
			Text:        mi18n.T("大きさX倍率"),
			ToolTipText: mi18n.T("大きさX倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("大きさX倍率説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: sizeXEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *domain.PhysicsItem) {
					item.CalcSizeX((*sizeXEdit).Value())
				})
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("質量倍率"),
			ToolTipText: mi18n.T("質量倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("質量倍率説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: massEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *domain.PhysicsItem) {
					item.CalcMass((*massEdit).Value())
				})
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.HSpacer{
			ColumnSpan: 2,
		},
	}
}

// createButtonWidgets ボタンウィジェットを作成
func (p *PhysicsTableViewDialog) createButtonWidgets(okBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder) []declarative.Widget {
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
				p.bakeState.CurrentSet().PhysicsTableModel.RemoveRow(p.bakeState.PhysicsTableView.CurrentIndex())
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

// updateItemProperty アイテムプロパティを更新
func (p *PhysicsTableViewDialog) updateItemProperty(treeView *walk.TreeView, updateFunc func(*domain.PhysicsItem)) {
	if treeView.CurrentItem() == nil {
		return
	}
	updateFunc(treeView.CurrentItem().(*domain.PhysicsItem))
	// モデルの更新
	treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
}

// updateEditValues 編集値を更新
func (p *PhysicsTableViewDialog) updateEditValues(treeView *walk.TreeView, sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit *walk.NumberEdit) {
	if treeView.CurrentItem() == nil {
		return
	}

	// 選択されたアイテムの情報を更新
	currentItem := treeView.CurrentItem().(*domain.PhysicsItem)
	sizeXEdit.ChangeValue(currentItem.SizeRatio().X)
	massEdit.ChangeValue(currentItem.MassRatio())
}

// handleDialogOK ダイアログOK処理
func (p *PhysicsTableViewDialog) handleDialogOK() {
	p.bakeState.SetWidgetEnabled(false)

	// 簡略化した物理処理（元の複雑なロジックは後でDomainサービスに移動）
	physicsWorldMotion := p.mWidgets.Window().LoadPhysicsWorldMotion(0)
	p.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	p.mWidgets.Window().TriggerPhysicsReset()

	p.bakeState.SetWidgetEnabled(true)
	controller.Beep()

	// 次の作業用の行を追加
	currentIndex := p.bakeState.PhysicsTableView.CurrentIndex()
	if currentIndex == len(p.bakeState.CurrentSet().PhysicsTableModel.Records)-1 {
		p.bakeState.CurrentSet().PhysicsTableModel.AddRecord(
			p.bakeState.CurrentSet().OriginalModel,
			0,
			p.bakeState.CurrentSet().MaxFrame())
	}
	p.bakeState.PhysicsTableView.SetModel(p.bakeState.CurrentSet().PhysicsTableModel)
}
