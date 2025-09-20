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

// WindTableViewDialog 物理設定ダイアログのロジックを管理
type WindTableViewDialog struct {
	store                *WidgetStore
	doDelete             bool
	lastChangedTimestamp int64 // 最後に値が変更されたタイムスタンプ

	startFrameEdit     *walk.NumberEdit // 開始フレーム入力
	endFrameEdit       *walk.NumberEdit // 終了フレーム入力
	directionXEdit     *walk.NumberEdit // 風向きX入力
	directionYEdit     *walk.NumberEdit // 風向きY入力
	directionZEdit     *walk.NumberEdit // 風向きZ入力
	speedEdit          *walk.NumberEdit // 風速入力
	randomnessEdit     *walk.NumberEdit // 乱れ入力
	turbulenceFreqEdit *walk.NumberEdit // 乱流周波数入力
	dragCoeffEdit      *walk.NumberEdit // 抗力係数入力
	liftCoeffEdit      *walk.NumberEdit // 揚力係数入力
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
	var presetComboBox *walk.ComboBox

	builder := declarative.NewBuilder(p.store.Window())

	dialog := &declarative.Dialog{
		AssignTo:      &dlg,
		CancelButton:  &cancelBtn,
		DefaultButton: &okBtn,
		Title:         mi18n.T("風物理設定"),
		Layout:        declarative.VBox{},
		MinSize:       declarative.Size{Width: 250, Height: 250},
		MaxSize:       declarative.Size{Width: 250, Height: 250},
		DataBinder: declarative.DataBinder{
			AssignTo:   &db,
			DataSource: record,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:   declarative.Grid{Columns: 6},
				Children: p.createFormWidgets(&presetComboBox),
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

func (p *WindTableViewDialog) createFormWidgets(presetComboBox **walk.ComboBox) []declarative.Widget {

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
			AssignTo:           &p.startFrameEdit,
			ToolTipText:        mi18n.T("開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           float64(p.store.minFrame()),
			MaxValue:           float64(p.store.maxFrame() + 1),
			DefaultValue:       float64(p.store.minFrame()),
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
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
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
			Text:        mi18n.T("風向きX"),
			ToolTipText: mi18n.T("風向きX説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風向きX説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo:           &p.directionXEdit,
			Value:              declarative.Bind("WindConfig.Direction.X"), // 初期値
			MinValue:           -10.0,                                      // 最小値
			MaxValue:           10.0,                                       // 最大値
			DefaultValue:       0.0,                                        // 初期値
			Decimals:           2,                                          // 小数点以下の桁数
			Increment:          0.1,                                        // 増分
			SpinButtonsVisible: true,                                       // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.TextLabel{
			Text:        mi18n.T("風向きY"),
			ToolTipText: mi18n.T("風向きY説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風向きY説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo:           &p.directionYEdit,
			Value:              declarative.Bind("WindConfig.Direction.Y"), // 初期値
			MinValue:           -10.0,                                      // 最小値
			MaxValue:           10.0,                                       // 最大値
			DefaultValue:       0,                                          // 初期値
			Decimals:           2,                                          // 小数点以下の桁数
			Increment:          0.1,                                        // 増分
			SpinButtonsVisible: true,                                       // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.TextLabel{
			Text:        mi18n.T("風向きZ"),
			ToolTipText: mi18n.T("風向きZ説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風向きZ説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo:           &p.directionZEdit,
			Value:              declarative.Bind("WindConfig.Direction.Z"), // 初期値
			MinValue:           -10.0,                                      // 最小値
			MaxValue:           10.0,                                       // 最大値
			DefaultValue:       0,                                          // 初期値
			Decimals:           2,                                          // 小数点以下の桁数
			Increment:          0.1,                                        // 増分
			SpinButtonsVisible: true,                                       // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
		declarative.TextLabel{
			Text:        mi18n.T("風プリセット"),
			ToolTipText: mi18n.T("風プリセット説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("風プリセット説明"))
			},
			MinSize: declarative.Size{Width: 80, Height: 20},
			MaxSize: declarative.Size{Width: 80, Height: 20},
		},
		declarative.ComboBox{
			AssignTo: presetComboBox,
			Value:    0, // 初期値
			Model: []string{
				mi18n.T("そよ風"),
				mi18n.T("強風"),
				mi18n.T("台風"),
			},
			OnCurrentIndexChanged: func() {
				switch (*presetComboBox).CurrentIndex() {
				case 0: // そよ風
					p.directionXEdit.ChangeValue(3.5)
					p.directionYEdit.ChangeValue(0.0)
					p.directionZEdit.ChangeValue(0.3)
					p.speedEdit.ChangeValue(2.0)
					p.randomnessEdit.ChangeValue(1.0)
					p.turbulenceFreqEdit.ChangeValue(0.5)
					p.dragCoeffEdit.ChangeValue(0.2)
					p.liftCoeffEdit.ChangeValue(0.1)
				case 1: // 強風
					p.directionXEdit.ChangeValue(5.0)
					p.directionYEdit.ChangeValue(0.5)
					p.directionZEdit.ChangeValue(0.0)
					p.speedEdit.ChangeValue(20.0)
					p.randomnessEdit.ChangeValue(1.0)
					p.turbulenceFreqEdit.ChangeValue(1.5)
					p.dragCoeffEdit.ChangeValue(0.8)
					p.liftCoeffEdit.ChangeValue(1.5)
				case 2: // 台風
					p.directionXEdit.ChangeValue(-0.6)
					p.directionYEdit.ChangeValue(10.0)
					p.directionZEdit.ChangeValue(1.0)
					p.speedEdit.ChangeValue(100.0)
					p.randomnessEdit.ChangeValue(0.6)
					p.turbulenceFreqEdit.ChangeValue(3.0)
					p.dragCoeffEdit.ChangeValue(1.0)
					p.liftCoeffEdit.ChangeValue(10.0)
				}
				p.onChangeValue()
			},
		},
		declarative.HSpacer{
			ColumnSpan: 4,
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
			AssignTo:           &p.speedEdit,
			MinValue:           0.0,    // 最小値
			MaxValue:           1000.0, // 最大値
			DefaultValue:       0.0,    // 初期値
			Decimals:           2,      // 小数点以下の桁数
			Increment:          0.1,    // 増分
			SpinButtonsVisible: true,   // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.randomnessEdit,
			MinValue:           0.0,  // 最小値
			MaxValue:           1.0,  // 最大値
			DefaultValue:       0.0,  // 初期値
			Decimals:           2,    // 小数点以下の桁数
			Increment:          0.01, // 増分
			SpinButtonsVisible: true, // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.turbulenceFreqEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.5,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.1,   // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.dragCoeffEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.8,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.1,   // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
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
			AssignTo:           &p.liftCoeffEdit,
			MinValue:           0.0,   // 最小値
			MaxValue:           100.0, // 最大値
			DefaultValue:       0.2,   // 初期値
			Decimals:           2,     // 小数点以下の桁数
			Increment:          0.1,   // 増分
			SpinButtonsVisible: true,  // スピンボタンを表示
			MinSize:            declarative.Size{Width: 80, Height: 20},
			MaxSize:            declarative.Size{Width: 80, Height: 20},
			OnValueChanged: func() {
				p.onChangeValue()
			},
		},
	}
}

func (p *WindTableViewDialog) createButtonWidgets(
	okBtn, deleteBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder,
) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("物理設定登録説明"),
			OnClicked: func() {
				if !((p.startFrameEdit).Value() < (p.endFrameEdit).Value()) {
					mlog.E(mi18n.T("ワールド物理範囲設定エラー"), nil, "")
					return
				}

				if err := (*db).Submit(); err != nil {
					mlog.E(mi18n.T("焼き込み設定変更エラー"), err, "")
					return
				}
				(*dlg).Accept()
			},
			MinSize:    declarative.Size{Width: 80, Height: 20},
			MaxSize:    declarative.Size{Width: 80, Height: 20},
			ColumnSpan: 2,
		},
		declarative.PushButton{
			AssignTo:    deleteBtn,
			Text:        mi18n.T("削除"),
			ToolTipText: mi18n.T("物理設定削除説明"),
			OnClicked: func() {
				p.doDelete = true
				(*dlg).Accept()
			},
			MinSize:    declarative.Size{Width: 80, Height: 20},
			MaxSize:    declarative.Size{Width: 80, Height: 20},
			ColumnSpan: 2,
		},
		declarative.PushButton{
			AssignTo:    cancelBtn,
			Text:        mi18n.T("キャンセル"),
			ToolTipText: mi18n.T("物理設定キャンセル説明"),
			OnClicked: func() {
				(*dlg).Cancel()
			},
			MinSize:    declarative.Size{Width: 80, Height: 20},
			MaxSize:    declarative.Size{Width: 80, Height: 20},
			ColumnSpan: 2,
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

func (p *WindTableViewDialog) onChangeValue() {
	// 最後に値が変更されたタイムスタンプから0.5秒以内なら無視
	nowTimestamp := time.Now().UnixMilli()
	if p.lastChangedTimestamp+500 > nowTimestamp {
		return
	}
	p.store.setWidgetEnabled(false)

	record := entity.NewWindRecord(
		float32(p.startFrameEdit.Value()),
		float32(p.endFrameEdit.Value()),
	)
	record.WindConfig.Direction.X = p.directionXEdit.Value()
	record.WindConfig.Direction.Y = p.directionYEdit.Value()
	record.WindConfig.Direction.Z = p.directionZEdit.Value()
	record.WindConfig.Speed = float32(p.speedEdit.Value())
	record.WindConfig.Randomness = float32(p.randomnessEdit.Value())
	record.WindConfig.TurbulenceFreqHz = float32(p.turbulenceFreqEdit.Value())
	record.WindConfig.DragCoeff = float32(p.dragCoeffEdit.Value())
	record.WindConfig.LiftCoeff = float32(p.liftCoeffEdit.Value())

	windMotion := vmd.NewVmdMotion("")

	p.store.physicsUsecase.ApplyWindMotion(
		windMotion,
		[]*entity.WindRecord{record},
	)

	p.store.mWidgets.Window().StoreWindMotion(0, windMotion)
	p.store.mWidgets.Window().TriggerPhysicsReset()

	p.store.setWidgetEnabled(true)
	p.lastChangedTimestamp = nowTimestamp
}
