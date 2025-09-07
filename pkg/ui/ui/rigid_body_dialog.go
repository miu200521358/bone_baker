package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// RigidBodyTableViewDialog モデル物理設定ダイアログのロジックを管理
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

// Show モデル物理設定ダイアログを表示
func (p *RigidBodyTableViewDialog) Show(record *entity.RigidBodyRecord, recordIndex int) {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var okBtn *walk.PushButton
	var deleteBtn *walk.PushButton
	var cancelBtn *walk.PushButton
	var db *walk.DataBinder
	var treeView *walk.TreeView
	var startFrameEdit *walk.NumberEdit    // 開始フレーム入力
	var endFrameEdit *walk.NumberEdit      // 終了フレーム入力
	var maxStartFrameEdit *walk.NumberEdit // 最大開始フレーム入力
	var maxEndFrameEdit *walk.NumberEdit   // 最大終了フレーム入力
	var sizeXEdit *walk.NumberEdit         // 大きさX入力
	var sizeYEdit *walk.NumberEdit         // 大きさY入力
	var sizeZEdit *walk.NumberEdit         // 大きさZ入力
	var massEdit *walk.NumberEdit          // 質量入力
	var stiffnessEdit *walk.NumberEdit     // 硬さ入力
	var tensionEdit *walk.NumberEdit       // 張り入力

	builder := declarative.NewBuilder(p.store.Window())
	treeModel := newRigidBodyTreeModel(record)

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("モデル物理設定"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 500, Height: 400},
		MaxSize:       declarative.Size{Width: 500, Height: 400},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: record,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.Grid{Columns: 6},
				Children: p.createFormWidgets(&startFrameEdit, &endFrameEdit,
					&maxStartFrameEdit, &maxEndFrameEdit, &sizeXEdit, &sizeYEdit, &sizeZEdit, &massEdit, &stiffnessEdit, &tensionEdit, &treeView, treeModel),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: p.createButtonWidgets(startFrameEdit, endFrameEdit,
					maxStartFrameEdit, maxEndFrameEdit, &okBtn, &deleteBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if cmd, err := dialog.Run(builder.Parent().Form()); err == nil && cmd == walk.DlgCmdOK {
		p.handleDialogOK(record, recordIndex)
	}
}

func (p *RigidBodyTableViewDialog) createFormWidgets(startFrameEdit, endFrameEdit,
	maxStartFrameEdit, maxEndFrameEdit, sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit **walk.NumberEdit, treeView **walk.TreeView, treeModel *RigidBodyTreeModel) []declarative.Widget {

	return []declarative.Widget{
		declarative.Label{
			Text:        mi18n.T("開始フレーム"),
			ToolTipText: mi18n.T("開始フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("開始フレーム説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
			MaxSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("StartFrame"),
			AssignTo:           startFrameEdit,
			ToolTipText:        mi18n.T("開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("最大開始フレーム"),
			ToolTipText: mi18n.T("最大開始フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("最大開始フレーム説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
			MaxSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxStartFrame"),
			AssignTo:           maxStartFrameEdit,
			ToolTipText:        mi18n.T("最大開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.HSpacer{
			ColumnSpan: 2,
		},
		declarative.Label{
			Text:        mi18n.T("最大終了フレーム"),
			ToolTipText: mi18n.T("最大終了フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("最大終了フレーム説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
			MaxSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxEndFrame"),
			AssignTo:           maxEndFrameEdit,
			ToolTipText:        mi18n.T("最大終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("終了フレーム"),
			ToolTipText: mi18n.T("終了フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("終了フレーム説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
			MaxSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			AssignTo:           endFrameEdit,
			ToolTipText:        mi18n.T("終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.HSpacer{
			ColumnSpan: 2,
		},
		declarative.TextLabel{
			Text:        mi18n.T("大きさX倍率"),
			ToolTipText: mi18n.T("大きさX倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("大きさX倍率説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: sizeXEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *RigidBodyTreeItem) {
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
			Text:        mi18n.T("大きさY倍率"),
			ToolTipText: mi18n.T("大きさY倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("大きさY倍率説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: sizeYEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *RigidBodyTreeItem) {
					item.CalcSizeY((*sizeYEdit).Value())
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
			Text:        mi18n.T("大きさZ倍率"),
			ToolTipText: mi18n.T("大きさZ倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("大きさZ倍率説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: sizeZEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *RigidBodyTreeItem) {
					item.CalcSizeZ((*sizeZEdit).Value())
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
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: massEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *RigidBodyTreeItem) {
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
		declarative.TextLabel{
			Text:        mi18n.T("硬さ倍率"),
			ToolTipText: mi18n.T("硬さ倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("硬さ倍率説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: stiffnessEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *RigidBodyTreeItem) {
					item.CalcStiffness((*stiffnessEdit).Value())
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
			Text:        mi18n.T("張り倍率"),
			ToolTipText: mi18n.T("張り倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("張り倍率説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: tensionEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *RigidBodyTreeItem) {
					item.CalcTension((*tensionEdit).Value())
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
		declarative.TreeView{
			AssignTo:   treeView,
			Model:      treeModel,
			MinSize:    declarative.Size{Width: 450, Height: 200},
			ColumnSpan: 6,
			OnCurrentItemChanged: func() {
				p.updateEditValues(*treeView, *sizeXEdit, *sizeYEdit, *sizeZEdit, *massEdit, *stiffnessEdit, *tensionEdit)
			},
		},
	}
}

// updateEditValues 編集値を更新
func (p *RigidBodyTableViewDialog) updateEditValues(treeView *walk.TreeView, sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit *walk.NumberEdit) {
	if treeView.CurrentItem() == nil {
		return
	}

	// 選択されたアイテムの情報を更新
	currentItem := treeView.CurrentItem().(*RigidBodyTreeItem)
	sizeXEdit.ChangeValue(currentItem.item.SizeRatio.X)
	sizeYEdit.ChangeValue(currentItem.item.SizeRatio.Y)
	sizeZEdit.ChangeValue(currentItem.item.SizeRatio.Z)
	massEdit.ChangeValue(currentItem.item.MassRatio)
	stiffnessEdit.ChangeValue(currentItem.item.StiffnessRatio)
	tensionEdit.ChangeValue(currentItem.item.TensionRatio)
}

// updateItemProperty アイテムプロパティを更新
func (p *RigidBodyTableViewDialog) updateItemProperty(treeView *walk.TreeView, updateFunc func(*RigidBodyTreeItem)) {
	if treeView.CurrentItem() == nil {
		return
	}
	updateFunc(treeView.CurrentItem().(*RigidBodyTreeItem))
	// モデルの更新
	treeView.Model().(*RigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
}

func (p *RigidBodyTableViewDialog) createButtonWidgets(
	startFrameEdit, endFrameEdit, maxStartFrameEdit, maxEndFrameEdit *walk.NumberEdit,
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("モデル物理設定登録説明"),
			OnClicked: func() {
				if !(startFrameEdit.Value() < maxStartFrameEdit.Value() &&
					maxStartFrameEdit.Value() < maxEndFrameEdit.Value() &&
					maxEndFrameEdit.Value() < endFrameEdit.Value()) {
					mlog.ET(mi18n.T("モデル物理範囲設定エラー"), nil, "")
					return
				}

				if err := (*db).Submit(); err != nil {
					mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
					return
				}
				(*dlg).Accept()
			},
		},
		declarative.PushButton{
			AssignTo:    deleteBtn,
			Text:        mi18n.T("削除"),
			ToolTipText: mi18n.T("モデル物理設定削除説明"),
			OnClicked: func() {
				p.doDelete = true
				(*dlg).Accept()
			},
		},
		declarative.PushButton{
			AssignTo:    cancelBtn,
			Text:        mi18n.T("キャンセル"),
			ToolTipText: mi18n.T("モデル物理設定キャンセル説明"),
			OnClicked: func() {
				(*dlg).Cancel()
			},
		},
	}
}

func (p *RigidBodyTableViewDialog) handleDialogOK(record *entity.RigidBodyRecord, recordIndex int) {
	p.store.setWidgetEnabled(false)

	if p.doDelete {
		// 削除処理
		if recordIndex >= 0 && recordIndex < len(p.store.currentSet().RigidBodyRecords) {
			// 指定インデックスのレコードを削除
			records := p.store.currentSet().RigidBodyRecords
			p.store.currentSet().RigidBodyRecords = append(records[:recordIndex], records[recordIndex+1:]...)
		}
	} else {
		// 追加・更新処理
		if recordIndex == -1 {
			// 新規追加
			p.store.currentSet().RigidBodyRecords = append(p.store.currentSet().RigidBodyRecords, record)
		} else {
			// 更新
			if recordIndex >= 0 && recordIndex < len(p.store.currentSet().RigidBodyRecords) {
				p.store.currentSet().RigidBodyRecords[recordIndex] = record
			}
		}
	}

	physicsWorldMotion := p.store.mWidgets.Window().LoadPhysicsWorldMotion(0)
	physicsModelMotion := p.store.mWidgets.Window().LoadPhysicsModelMotion(0, p.store.CurrentIndex)

	p.store.physicsUsecase.ApplyPhysicsModelMotion(
		physicsWorldMotion,
		physicsModelMotion,
		p.store.currentSet().RigidBodyRecords,
		p.store.currentSet().OriginalModel,
	)

	p.store.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	p.store.mWidgets.Window().StorePhysicsModelMotion(0, p.store.CurrentIndex, physicsModelMotion)
	p.store.mWidgets.Window().TriggerPhysicsReset()

	// 台形テーブルの再描画を強制
	if p.store.RigidBodyTableWidget != nil {
		p.store.RigidBodyTableWidget.Invalidate()
	}

	p.store.setWidgetEnabled(true)

	controller.Beep()

	// 削除フラグをリセット
	p.doDelete = false
}
