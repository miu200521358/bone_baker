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

// PhysicsTableViewDialog 物理設定ダイアログのロジックを管理
type PhysicsTableViewDialog struct {
	store                *WidgetStore
	doDelete             bool
	lastChangedTimestamp int64 // 最後に値が変更されたタイムスタンプ

	startFrameEdit    *walk.NumberEdit // 開始フレーム入力
	endFrameEdit      *walk.NumberEdit // 終了フレーム入力
	gravityEdit       *walk.NumberEdit // 重力値入力
	maxSubStepsEdit   *walk.NumberEdit // 最大最大演算回数
	fixedTimeStepEdit *walk.NumberEdit // 固定タイムステップ入力
}

// newPhysicsTableViewDialog コンストラクタ
func newPhysicsTableViewDialog(store *WidgetStore) *PhysicsTableViewDialog {
	return &PhysicsTableViewDialog{
		store: store,
	}
}

// show 物理設定ダイアログを表示
func (p *PhysicsTableViewDialog) show(record *entity.PhysicsRecord, recordIndex int) {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var okBtn *walk.PushButton
	var deleteBtn *walk.PushButton
	var cancelBtn *walk.PushButton
	var db *walk.DataBinder

	builder := declarative.NewBuilder(p.store.Window())

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("ワールド物理設定"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 250, Height: 250},
		MaxSize:       declarative.Size{Width: 250, Height: 250},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: record,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:   declarative.Grid{Columns: 2},
				Children: p.createFormWidgets(),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: p.createButtonWidgets(&okBtn, &deleteBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if _, err := dialog.Run(builder.Parent().Form()); err == nil {
		// どのボタンでも
		p.handleDialogOK(record, recordIndex)
	}
}

func (p *PhysicsTableViewDialog) createFormWidgets() []declarative.Widget {

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
			AssignTo:           &p.startFrameEdit,
			ToolTipText:        mi18n.T("開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.minFrame()),
			MaxValue:           float64(p.store.maxFrame() + 1),
			DefaultValue:       float64(p.store.minFrame()),
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
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
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			AssignTo:           &p.endFrameEdit,
			ToolTipText:        mi18n.T("終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.minFrame()),
			MaxValue:           float64(p.store.maxFrame() + 1),
			DefaultValue:       float64(p.store.maxFrame()),
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.TextLabel{
			Text:        mi18n.T("重力"),
			ToolTipText: mi18n.T("重力説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("重力説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("Gravity"),
			AssignTo:           &p.gravityEdit,
			MinValue:           -100.0, // 最小値
			MaxValue:           100.0,  // 最大値
			DefaultValue:       -9.8,
			Decimals:           1,    // 小数点以下の桁数
			Increment:          0.1,  // 増分
			SpinButtonsVisible: true, // スピンボタンを表示
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.TextLabel{
			Text:        mi18n.T("最大演算回数"),
			ToolTipText: mi18n.T("最大演算回数説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("最大演算回数説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxSubSteps"),
			AssignTo:           &p.maxSubStepsEdit,
			MinValue:           1.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       2,
			Decimals:           0,    // 小数点以下の桁数
			Increment:          1.0,  // 増分
			SpinButtonsVisible: true, // スピンボタンを表示
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.TextLabel{
			Text:        mi18n.T("物理演算頻度"),
			ToolTipText: mi18n.T("物理演算頻度説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("物理演算頻度説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
			MaxSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("FixedTimeStep"),
			AssignTo:           &p.fixedTimeStepEdit,
			MinValue:           10.0,    // 最小値
			MaxValue:           48000.0, // 最大値
			DefaultValue:       60,
			Decimals:           0,    // 小数点以下の桁数
			Increment:          10.0, // 増分
			SpinButtonsVisible: true, // スピンボタンを表示
			StretchFactor:      20,
			MinSize:            declarative.Size{Width: 100, Height: 20},
			MaxSize:            declarative.Size{Width: 100, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
	}
}

func (p *PhysicsTableViewDialog) createButtonWidgets(
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("物理設定登録説明"),
			OnClicked: func() {
				if !(p.startFrameEdit.Value() < p.endFrameEdit.Value()) {
					mlog.E(mi18n.T("ワールド物理範囲設定エラー"), nil, "")
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
			ToolTipText: mi18n.T("物理設定削除説明"),
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
			ToolTipText: mi18n.T("物理設定キャンセル説明"),
			OnClicked: func() {
				(*dlg).Cancel()
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
	}
}

func (p *PhysicsTableViewDialog) handleDialogOK(record *entity.PhysicsRecord, recordIndex int) {
	p.store.setWidgetEnabled(false)

	if recordIndex == -1 {
		p.store.PhysicsRecords =
			append(p.store.PhysicsRecords, record)
		p.store.PhysicsTableView.SetCurrentIndex(len(p.store.PhysicsRecords) - 1)
	} else {
		p.store.PhysicsRecords[recordIndex] = record
		p.store.PhysicsTableView.SetCurrentIndex(recordIndex)
	}

	physicsWorldMotion := vmd.NewVmdMotion("")

	p.store.physicsUsecase.ApplyPhysicsWorldMotion(
		physicsWorldMotion,
		p.store.PhysicsRecords,
	)

	p.store.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	p.store.mWidgets.Window().TriggerPhysicsReset()

	p.store.setWidgetEnabled(true)

	// 更新
	p.store.PhysicsTableView.SetModel(newPhysicsTableModelWithRecords(p.store.PhysicsRecords))
}

func (p *PhysicsTableViewDialog) onChangeValue() {
	// 最後に値が変更されたタイムスタンプから0.5秒以内なら無視
	nowTimestamp := time.Now().UnixMilli()
	if p.lastChangedTimestamp+500 > nowTimestamp {
		return
	}
	p.store.setWidgetEnabled(false)

	record := entity.NewPhysicsRecord(
		float32(p.startFrameEdit.Value()),
		float32(p.endFrameEdit.Value()),
	)
	record.Gravity = p.gravityEdit.Value()
	record.MaxSubSteps = int(p.maxSubStepsEdit.Value())
	record.FixedTimeStep = p.fixedTimeStepEdit.Value()

	physicsWorldMotion := vmd.NewVmdMotion("")

	p.store.physicsUsecase.ApplyPhysicsWorldMotion(
		physicsWorldMotion,
		[]*entity.PhysicsRecord{record},
	)

	p.store.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	p.store.mWidgets.Window().TriggerPhysicsReset()

	p.store.setWidgetEnabled(true)
	p.lastChangedTimestamp = nowTimestamp
}
