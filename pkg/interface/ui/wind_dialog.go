package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// WindTableViewDialog 物理設定ダイアログのロジックを管理
type WindTableViewDialog struct {
	store    *WidgetStore
	doDelete bool
}

// newWindTableViewDialog コンストラクタ
func newWindTableViewDialog(store *WidgetStore) *WindTableViewDialog {
	return &WindTableViewDialog{
		store: store,
	}
}

// show 物理設定ダイアログを表示
func (p *WindTableViewDialog) show(record *entity.WindRecord, recordIndex int) {
	// アイテムがクリックされたら、入力ダイアログを表示する
	var dlg *walk.Dialog
	var okBtn *walk.PushButton
	var deleteBtn *walk.PushButton
	var cancelBtn *walk.PushButton
	var db *walk.DataBinder
	var startFrameEdit *walk.NumberEdit     // 開始フレーム入力
	var endFrameEdit *walk.NumberEdit       // 終了フレーム入力
	var directionXEdit *walk.NumberEdit     // 風向きX入力
	var directionYEdit *walk.NumberEdit     // 風向きY入力
	var directionZEdit *walk.NumberEdit     // 風向きZ入力
	var speedEdit *walk.NumberEdit          // 風速入力
	var randomnessEdit *walk.NumberEdit     // 乱れ入力
	var turbulenceFreqEdit *walk.NumberEdit // 乱流周波数入力
	var dragCoeffEdit *walk.NumberEdit      // 抗力係数入力
	var liftCoeffEdit *walk.NumberEdit      // 揚力係数入力

	builder := declarative.NewBuilder(p.store.Window())

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("物理設定"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 250, Height: 250},
		MaxSize:       declarative.Size{Width: 250, Height: 250},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: record,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.Grid{Columns: 6},
				Children: p.createFormWidgets(&startFrameEdit, &endFrameEdit,
					&directionXEdit, &directionYEdit, &directionZEdit,
					&speedEdit, &randomnessEdit, &turbulenceFreqEdit,
					&dragCoeffEdit, &liftCoeffEdit),
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Alignment: declarative.AlignHFarVCenter,
				},
				Children: p.createButtonWidgets(&startFrameEdit, &endFrameEdit, &okBtn, &deleteBtn, &cancelBtn, &dlg, &db),
			},
		},
	}

	if cmd, err := dialog.Run(builder.Parent().Form()); err == nil && cmd == walk.DlgCmdOK {
		p.handleDialogOK(record, recordIndex)
	}
}

func (p *WindTableViewDialog) createFormWidgets(startFrameEdit, endFrameEdit,
	directionXEdit, directionYEdit, directionZEdit, speedEdit, randomnessEdit,
	turbulenceFreqEdit, dragCoeffEdit, liftCoeffEdit **walk.NumberEdit) []declarative.Widget {

	return []declarative.Widget{
		declarative.Label{
			Text:        mi18n.T("開始フレーム"),
			ToolTipText: mi18n.T("開始フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("開始フレーム説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("StartFrame"),
			AssignTo:           startFrameEdit,
			ToolTipText:        mi18n.T("開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.minFrame()),
			MaxValue:           float64(p.store.maxFrame() + 1),
			DefaultValue:       float64(p.store.minFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.Label{
			Text:        mi18n.T("終了フレーム"),
			ToolTipText: mi18n.T("終了フレーム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("終了フレーム説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("EndFrame"),
			AssignTo:           endFrameEdit,
			ToolTipText:        mi18n.T("終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.minFrame()),
			MaxValue:           float64(p.store.maxFrame() + 1),
			DefaultValue:       float64(p.store.maxFrame()),
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.HSpacer{
			ColumnSpan: 2,
		},
		declarative.TextLabel{
			Text:        mi18n.T("風向きX"),
			ToolTipText: mi18n.T("風向きX説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風向きX説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo:           directionXEdit,
			Value:              declarative.Bind("WindConfig.Direction.X"), // 初期値
			MinValue:           0.00,                                       // 最小値
			MaxValue:           100.0,                                      // 最大値
			DefaultValue:       0.0,                                        // 初期値
			Decimals:           2,                                          // 小数点以下の桁数
			Increment:          0.01,                                       // 増分
			SpinButtonsVisible: true,                                       // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("風向きY"),
			ToolTipText: mi18n.T("風向きY説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風向きY説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo:           directionYEdit,
			Value:              declarative.Bind("WindConfig.Direction.Y"), // 初期値
			MinValue:           0.00,                                       // 最小値
			MaxValue:           100.0,                                      // 最大値
			DefaultValue:       0,                                          // 初期値
			Decimals:           2,                                          // 小数点以下の桁数
			Increment:          0.01,                                       // 増分
			SpinButtonsVisible: true,                                       // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("風向きZ"),
			ToolTipText: mi18n.T("風向きZ説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風向きZ説明"))
			},
			MinSize: declarative.Size{Width: 150, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo:           directionZEdit,
			Value:              declarative.Bind("WindConfig.Direction.Z"), // 初期値
			MinValue:           0.00,                                       // 最小値
			MaxValue:           100.0,                                      // 最大値
			DefaultValue:       0,                                          // 初期値
			Decimals:           2,                                          // 小数点以下の桁数
			Increment:          0.01,                                       // 増分
			SpinButtonsVisible: true,                                       // スピンボタンを表示
			MinSize:            declarative.Size{Width: 60, Height: 20},
			MaxSize:            declarative.Size{Width: 60, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("風速"),
			ToolTipText: mi18n.T("風速説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風速説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("WindConfig.Speed"),
			AssignTo:           speedEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.0,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("ランダム"),
			ToolTipText: mi18n.T("ランダム説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("ランダム説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("WindConfig.Randomness"),
			AssignTo:           randomnessEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.0,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("乱流周波数"),
			ToolTipText: mi18n.T("乱流周波数説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("乱流周波数説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("WindConfig.TurbulenceFreqHz"),
			AssignTo:           turbulenceFreqEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.5,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("抗力係数"),
			ToolTipText: mi18n.T("抗力係数説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("抗力係数説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("WindConfig.DragCoeff"),
			AssignTo:           dragCoeffEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       1.0,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
		declarative.TextLabel{
			Text:        mi18n.T("揚力係数"),
			ToolTipText: mi18n.T("揚力係数説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("揚力係数説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("WindConfig.LiftCoeff"),
			AssignTo:           liftCoeffEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.2,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.01,  // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
		},
	}
}

func (p *WindTableViewDialog) createButtonWidgets(
	startFrameEdit, endFrameEdit **walk.NumberEdit,
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("物理設定登録説明"),
			OnClicked: func() {
				if !((*startFrameEdit).Value() < (*endFrameEdit).Value()) {
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
				(*dlg).Accept()
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

func (p *WindTableViewDialog) handleDialogOK(record *entity.WindRecord, recordIndex int) {
	p.store.setWidgetEnabled(false)

	if recordIndex == -1 {
		p.store.WindRecords =
			append(p.store.WindRecords, record)
		p.store.WindTableView.SetCurrentIndex(len(p.store.WindRecords) - 1)
	} else {
		p.store.WindRecords[recordIndex] = record
		p.store.WindTableView.SetCurrentIndex(recordIndex)
	}

	windMotion := vmd.NewVmdMotion("")

	p.store.physicsUsecase.ApplyWindMotion(
		windMotion,
		p.store.WindRecords,
	)

	p.store.mWidgets.Window().StoreWindMotion(0, windMotion)
	p.store.mWidgets.Window().TriggerPhysicsReset()

	p.store.setWidgetEnabled(true)

	// 更新
	p.store.WindTableView.SetModel(newWindTableModelWithRecords(p.store.WindRecords))
}
