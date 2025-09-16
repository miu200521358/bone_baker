package ui

import (
	"time"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// RigidBodyTableViewDialog モデル物理設定ダイアログのロジックを管理
type RigidBodyTableViewDialog struct {
	store                *WidgetStore
	doDelete             bool
	lastChangedTimestamp int64 // 最後に値が変更されたタイムスタンプ

	startFrameEdit    *walk.NumberEdit // 開始フレーム入力
	endFrameEdit      *walk.NumberEdit // 終了フレーム入力
	maxStartFrameEdit *walk.NumberEdit // 最大開始フレーム入力
	maxEndFrameEdit   *walk.NumberEdit // 最大終了フレーム入力
	sizeXEdit         *walk.NumberEdit // 大きさX入力
	sizeYEdit         *walk.NumberEdit // 大きさY入力
	sizeZEdit         *walk.NumberEdit // 大きさZ入力
	positionXEdit     *walk.NumberEdit // 位置X入力
	positionYEdit     *walk.NumberEdit // 位置Y入力
	positionZEdit     *walk.NumberEdit // 位置Z入力
	massEdit          *walk.NumberEdit // 質量入力
	stiffnessEdit     *walk.NumberEdit // 硬さ入力
	tensionEdit       *walk.NumberEdit // 張り入力
	treeView          *walk.TreeView   // 剛体ツリービュー
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
				Layout:   declarative.Grid{Columns: 6},
				Children: p.createFormWidgets(&p.treeView, treeModel),
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

func (p *RigidBodyTableViewDialog) createFormWidgets(treeView **walk.TreeView, treeModel *RigidBodyTreeModel) []declarative.Widget {

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
			AssignTo:           &p.startFrameEdit,
			ToolTipText:        mi18n.T("開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			DefaultValue:       float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.maxStartFrameEdit,
			ToolTipText:        mi18n.T("最大開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			DefaultValue:       float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.maxEndFrameEdit,
			ToolTipText:        mi18n.T("最大終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			DefaultValue:       float64(p.store.currentSet().OriginalMotion.MaxFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.endFrameEdit,
			ToolTipText:        mi18n.T("終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.currentSet().OriginalMotion.MinFrame()),
			MaxValue:           float64(p.store.currentSet().OriginalMotion.MaxFrame() + 1),
			DefaultValue:       float64(p.store.currentSet().OriginalMotion.MaxFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.HSpacer{
			ColumnSpan: 2,
		},
		declarative.TextLabel{
			Text:        mi18n.T("位置X"),
			ToolTipText: mi18n.T("位置X説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("位置X説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: &p.positionXEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              0,     // 初期値
			MinValue:           0.00,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0,     // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("位置Y"),
			ToolTipText: mi18n.T("位置Y説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("位置Y説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: &p.positionYEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              0,     // 初期値
			MinValue:           0.00,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0,     // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("位置Z"),
			ToolTipText: mi18n.T("位置Z説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("位置Z説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: &p.positionZEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              0,     // 初期値
			MinValue:           0.00,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0,     // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
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
			AssignTo: &p.sizeXEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1,     // 初期値
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
			AssignTo: &p.sizeYEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1,     // 初期値
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
			AssignTo: &p.sizeZEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1,     // 初期値
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
			AssignTo: &p.massEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1,     // 初期値
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
			AssignTo: &p.stiffnessEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1,     // 初期値
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
			AssignTo: &p.tensionEdit,
			OnValueChanged: func() {
				p.onChangeValue()
			},
			Value:              1,     // 初期値
			MinValue:           0.01,  // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1,     // 初期値
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
				p.updateEditValues(*treeView)
			},
		},
	}
}

// updateEditValues 編集値を更新
func (p *RigidBodyTableViewDialog) updateEditValues(treeView *walk.TreeView) {
	if treeView.CurrentItem() == nil {
		return
	}

	// 選択されたアイテムの情報を更新
	currentItem := treeView.CurrentItem().(*RigidBodyTreeItem)
	p.positionXEdit.ChangeValue(currentItem.item.Position.X)
	p.positionYEdit.ChangeValue(currentItem.item.Position.Y)
	p.positionZEdit.ChangeValue(currentItem.item.Position.Z)
	p.sizeXEdit.ChangeValue(currentItem.item.SizeRatio.X)
	p.sizeYEdit.ChangeValue(currentItem.item.SizeRatio.Y)
	p.sizeZEdit.ChangeValue(currentItem.item.SizeRatio.Z)
	p.massEdit.ChangeValue(currentItem.item.MassRatio)
	p.stiffnessEdit.ChangeValue(currentItem.item.StiffnessRatio)
	p.tensionEdit.ChangeValue(currentItem.item.TensionRatio)
}

// updateItemProperty アイテムプロパティを更新
func (p *RigidBodyTableViewDialog) updateItemProperty(updateFunc func(*RigidBodyTreeItem)) {
	if p.treeView.CurrentItem() == nil {
		return
	}
	updateFunc(p.treeView.CurrentItem().(*RigidBodyTreeItem))
	// モデルの更新
	p.treeView.Model().(*RigidBodyTreeModel).PublishItemChanged(p.treeView.CurrentItem())
}

func (p *RigidBodyTableViewDialog) createButtonWidgets(
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("モデル物理設定登録説明"),
			OnClicked: func() {
				if !((p.startFrameEdit).Value() <= (p.maxStartFrameEdit).Value() &&
					(p.maxStartFrameEdit).Value() < (p.maxEndFrameEdit).Value() &&
					(p.maxEndFrameEdit).Value() <= (p.endFrameEdit).Value()) {
					mlog.E(mi18n.T("モデル物理範囲設定エラー"), nil, "")
					return
				}

				if err := (*db).Submit(); err != nil {
					mlog.E(mi18n.T("焼き込み設定変更エラー"), err, "")
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
	physicsModelMotion := vmd.NewVmdMotion("")

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

	// 削除フラグをリセット
	p.doDelete = false
}

func (p *RigidBodyTableViewDialog) onChangeValue() {
	// 最後に値が変更されたタイムスタンプから0.5秒以内なら無視
	nowTimestamp := time.Now().UnixMilli()
	if p.lastChangedTimestamp+500 > nowTimestamp {
		return
	}

	physicsWorldMotion := p.store.mWidgets.Window().LoadPhysicsWorldMotion(0)
	physicsModelMotion := vmd.NewVmdMotion("")

	record := entity.NewRigidBodyRecord(float32(p.startFrameEdit.Value()), float32(p.endFrameEdit.Value()), p.store.currentSet().OriginalModel)
	record.MaxStartFrame = float32(p.maxStartFrameEdit.Value())
	record.MaxEndFrame = float32(p.maxEndFrameEdit.Value())

	p.updateItemProperty(func(item *RigidBodyTreeItem) {
		item.CalcPositionX((p.positionXEdit).Value())
		item.CalcPositionY((p.positionYEdit).Value())
		item.CalcPositionZ((p.positionZEdit).Value())
		item.CalcSizeX((p.sizeXEdit).Value())
		item.CalcSizeY((p.sizeYEdit).Value())
		item.CalcSizeZ((p.sizeZEdit).Value())
		item.CalcMass((p.massEdit).Value())
		item.CalcTension((p.tensionEdit).Value())
		item.CalcStiffness((p.stiffnessEdit).Value())
	})

	p.store.physicsUsecase.ApplyPhysicsModelMotion(
		physicsWorldMotion,
		physicsModelMotion,
		[]*entity.RigidBodyRecord{record},
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
	p.lastChangedTimestamp = nowTimestamp
}
