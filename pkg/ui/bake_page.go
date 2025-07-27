package ui

import (
	"path/filepath"
	"time"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
	"github.com/miu200521358/mlib_go/pkg/config/mconfig"
	"github.com/miu200521358/mlib_go/pkg/config/merr"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

func NewBakePage(mWidgets *controller.MWidgets) declarative.TabPage {
	var bakeTab *walk.TabPage

	// Usecaseの依存性注入
	bakeUsecase := usecase.NewBakeUsecase()
	bakeState := NewBakeState(bakeUsecase)

	bakeState.Player = widget.NewMotionPlayer()
	bakeState.Player.SetLabelTexts(mi18n.T("焼き込み停止"), mi18n.T("焼き込み再生"))
	bakeState.Player.SetOnEnabledInPlaying(newEnabledPlaying(bakeState))
	bakeState.Player.SetOnChangePlayingPre(newOnChangePlayingPre(bakeState, mWidgets))

	bakeState.OutputMotionPicker = newOutputMotionFilePicker()
	bakeState.OutputModelPicker = newOutputModelFilePicker(bakeState)
	bakeState.OriginalMotionPicker = newOriginalMotionFilePicker(bakeState)
	bakeState.OriginalModelPicker = newOriginalModelFilePicker(bakeState)

	bakeState.AddSetButton = newAddSetButton(bakeState)
	bakeState.ResetSetButton = newResetSetButton(bakeState, mWidgets)
	bakeState.LoadSetButton = newLoadSetButton(bakeState, mWidgets)
	bakeState.SaveSetButton = newSaveSetButton(bakeState)

	bakeState.SaveModelButton = newSaveModelButton(bakeState)
	bakeState.SaveMotionButton = newSaveMotionButton(bakeState)

	mWidgets.Widgets = append(mWidgets.Widgets, bakeState.Player, bakeState.OriginalMotionPicker,
		bakeState.OriginalModelPicker, bakeState.OutputMotionPicker,
		bakeState.OutputModelPicker, bakeState.AddSetButton, bakeState.ResetSetButton,
		bakeState.LoadSetButton, bakeState.SaveSetButton, bakeState.SaveMotionButton,
		bakeState.SaveModelButton)
	mWidgets.SetOnLoaded(func() {
		bakeState.BakeSets = append(bakeState.BakeSets, domain.NewPhysicsSet(len(bakeState.BakeSets)))
		bakeState.AddAction()
	})

	return declarative.TabPage{
		Title:    mi18n.T("焼き込み"),
		AssignTo: &bakeTab,
		Layout:   declarative.VBox{},
		Background: declarative.SolidColorBrush{
			Color: controller.ColorTabBackground,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:  declarative.HBox{},
				MinSize: declarative.Size{Width: 200, Height: 40},
				MaxSize: declarative.Size{Width: 5120, Height: 40},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					bakeState.AddSetButton.Widgets(),
					bakeState.ResetSetButton.Widgets(),
					bakeState.LoadSetButton.Widgets(),
					bakeState.SaveSetButton.Widgets(),
				},
			},
			// セットスクロール
			declarative.ScrollView{
				Layout:        declarative.VBox{},
				MinSize:       declarative.Size{Width: 200, Height: 40},
				MaxSize:       declarative.Size{Width: 5120, Height: 40},
				VerticalFixed: true,
				Children: []declarative.Widget{
					// ナビゲーション用ツールバー
					declarative.ToolBar{
						AssignTo:           &bakeState.NavToolBar,
						MinSize:            declarative.Size{Width: 200, Height: 25},
						MaxSize:            declarative.Size{Width: 5120, Height: 25},
						DefaultButtonWidth: 200,
						Orientation:        walk.Horizontal,
						ButtonStyle:        declarative.ToolBarButtonTextOnly,
					},
				},
			},
			// セットごとの焼き込み内容
			declarative.ScrollView{
				Layout:  declarative.VBox{},
				MinSize: declarative.Size{Width: 126, Height: 350},
				MaxSize: declarative.Size{Width: 2560, Height: 5120},
				Children: []declarative.Widget{
					bakeState.OriginalModelPicker.Widgets(),
					bakeState.OriginalMotionPicker.Widgets(),
					declarative.VSeparator{},
					declarative.TextLabel{
						Text:        mi18n.T("物理設定オプションテーブル"),
						ToolTipText: mi18n.T("物理設定オプションテーブル説明"),
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							mlog.ILT(mi18n.T("物理設定オプションテーブル"), mi18n.T("物理設定オプションテーブル説明"))
						},
					},
					newPhysicsTableView(bakeState, mWidgets),
					declarative.VSeparator{},
					bakeState.OutputModelPicker.Widgets(),
					bakeState.SaveModelButton.Widgets(),
					declarative.VSeparator{},
					bakeState.OutputMotionPicker.Widgets(),
					declarative.Composite{
						Layout:   declarative.Grid{Columns: 6},
						Children: newBakedHistoryWidgets(bakeState, mWidgets),
					},
					declarative.TextLabel{
						Text:        mi18n.T("焼き込み保存設定テーブル"),
						ToolTipText: mi18n.T("焼き込み保存設定テーブル説明"),
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							mlog.ILT(mi18n.T("焼き込み保存設定テーブル"), mi18n.T("焼き込み保存設定テーブル説明"))
						},
						ColumnSpan: 6,
					},
					newOutputTableView(bakeState, mWidgets),
					bakeState.SaveMotionButton.Widgets(),
				},
			},
			bakeState.Player.Widgets(),
		},
	}
}

func newOutputMotionFilePicker() *widget.FilePicker {
	return widget.NewVmdSaveFilePicker(
		mi18n.T("焼き込み後モーション(Vmd)"),
		mi18n.T("焼き込み後モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
		},
	)
}

func newOutputModelFilePicker(bakeState *BakeState) *widget.FilePicker {
	return widget.NewPmxSaveFilePicker(
		mi18n.T("変更後モデル(Pmx)"),
		mi18n.T("変更後モデル説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			// 実際に保存するのは、物理有効な元モデル
			model := bakeState.CurrentSet().OriginalModel
			if model == nil {
				return
			}

			if err := rep.Save(path, model, false); err != nil {
				mlog.ET(mi18n.T("保存失敗"), err, "")
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					bakeState.SetWidgetEnabled(true)
				}
			}
		},
	)
}

func newOriginalMotionFilePicker(bakeState *BakeState) *widget.FilePicker {
	return widget.NewVmdLoadFilePicker(
		"vmd",
		mi18n.T("モーション(Vmd)"),
		mi18n.T("モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := bakeState.LoadMotion(cw, path, true); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					bakeState.SetWidgetEnabled(true)
				}
			}
		},
	)
}

func newOriginalModelFilePicker(bakeState *BakeState) *widget.FilePicker {
	return widget.NewPmxLoadFilePicker(
		"pmx",
		mi18n.T("モデル(Pmx)"),
		mi18n.T("モデル説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := bakeState.LoadModel(cw, path); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					bakeState.SetWidgetEnabled(true)
				}
			}
		},
	)
}

func newAddSetButton(bakeState *BakeState) *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("セット追加"))
	btn.SetTooltip(mi18n.T("セット追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		bakeState.BakeSets = append(bakeState.BakeSets,
			domain.NewPhysicsSet(len(bakeState.BakeSets)))
		bakeState.AddAction()
	})
	return btn
}

func newResetSetButton(bakeState *BakeState, mWidgets *controller.MWidgets) *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("セット全削除"))
	btn.SetTooltip(mi18n.T("セット全削除説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		for n := range 2 {
			for m := range bakeState.NavToolBar.Actions().Len() {
				mWidgets.Window().StoreModel(n, m, nil)
				mWidgets.Window().StoreMotion(n, m, nil)
			}
		}

		bakeState.ResetSet()
	})
	return btn
}

func newLoadSetButton(bakeState *BakeState, mWidgets *controller.MWidgets) *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("セット設定読込"))
	btn.SetTooltip(mi18n.T("セット設定読込説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		choices := mconfig.LoadUserConfig("physics_set_path")
		var initialDirPath string
		if len(choices) > 0 {
			// ファイルパスからディレクトリパスを取得
			initialDirPath = filepath.Dir(choices[0])
		}

		// ファイル選択ダイアログを開く
		dlg := walk.FileDialog{
			Title: mi18n.T(
				"ファイル選択ダイアログタイトル",
				map[string]any{"Title": "Json"}),
			Filter:         "Json files (*.json)|*.json",
			FilterIndex:    1,
			InitialDirPath: initialDirPath,
		}
		if ok, err := dlg.ShowOpen(nil); err != nil {
			walk.MsgBox(nil, mi18n.T("ファイル選択ダイアログ選択エラー"), err.Error(), walk.MsgBoxIconError)
		} else if ok {
			bakeState.SetWidgetEnabled(false)
			mconfig.SaveUserConfig("physics_set_path", dlg.FilePath, 1)

			for n := range 2 {
				for m := range bakeState.NavToolBar.Actions().Len() {
					mWidgets.Window().StoreModel(n, m, nil)
					mWidgets.Window().StoreMotion(n, m, nil)
				}
			}

			bakeState.ResetSet()
			bakeState.LoadSet(dlg.FilePath)

			for range len(bakeState.BakeSets) - 1 {
				bakeState.AddAction()
			}

			for index := range bakeState.BakeSets {
				bakeState.ChangeCurrentAction(index)
				bakeState.OriginalMotionPicker.SetForcePath(bakeState.BakeSets[index].OriginalMotionPath)
				bakeState.OutputModelPicker.SetForcePath(bakeState.BakeSets[index].OutputModelPath)
			}

			bakeState.SetCurrentIndex(0)
			bakeState.SetWidgetEnabled(true)
		}
	})
	return btn
}

func newSaveSetButton(bakeState *BakeState) *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("セット設定保存"))
	btn.SetTooltip(mi18n.T("セット設定保存説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		// 焼き込み元モーションパスを初期パスとする
		initialDirPath := filepath.Dir(bakeState.CurrentSet().OriginalMotionPath)

		// ファイル選択ダイアログを開く
		dlg := walk.FileDialog{
			Title: mi18n.T(
				"ファイル選択ダイアログタイトル",
				map[string]any{"Title": "Json"}),
			Filter:         "Json files (*.json)|*.json",
			FilterIndex:    1,
			InitialDirPath: initialDirPath,
		}
		if ok, err := dlg.ShowSave(nil); err != nil {
			walk.MsgBox(nil, mi18n.T("ファイル選択ダイアログ選択エラー"), err.Error(), walk.MsgBoxIconError)
		} else if ok {
			bakeState.SaveSet(dlg.FilePath)
			mconfig.SaveUserConfig("physics_set_path", dlg.FilePath, 1)
		}
	})
	return btn
}

func newSaveModelButton(bakeState *BakeState) *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モデル保存"))
	btn.SetTooltip(mi18n.T("モデル保存説明"))
	btn.SetMinSize(declarative.Size{Width: 256, Height: 20})
	btn.SetStretchFactor(20)
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		bakeState.SetWidgetEnabled(false)

		for _, physicsSet := range bakeState.BakeSets {
			if physicsSet.OutputModelPath != "" && physicsSet.OriginalModel != nil {
				// 保存するのは物理が有効になっている元モデル
				rep := repository.NewPmxRepository(true)
				if err := rep.Save(physicsSet.OutputModelPath, physicsSet.OriginalModel, false); err != nil {
					mlog.ET(mi18n.T("モデル保存失敗"), err, "")
					if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
						bakeState.SetWidgetEnabled(true)
					}
				}
			}
		}

		bakeState.SetWidgetEnabled(true)
		controller.Beep()
	})

	return btn
}

func newSaveMotionButton(bakeState *BakeState) *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モーション保存"))
	btn.SetTooltip(mi18n.T("モーション保存説明"))
	btn.SetMinSize(declarative.Size{Width: 256, Height: 20})
	btn.SetStretchFactor(20)
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		bakeState.SetWidgetEnabled(false)

		for _, physicsSet := range bakeState.BakeSets {
			if physicsSet.OutputMotionPath != "" && physicsSet.OutputMotion != nil {
				// チェックボーンのみ残す
				motions, err := physicsSet.GetOutputMotionOnlyChecked(
					bakeState.OutputTableView.Model().(*domain.OutputTableModel).Records,
				)
				if err != nil {
					mlog.ET(mi18n.T("モーション保存失敗"), err, "")
					return
				}

				for _, motion := range motions {
					rep := repository.NewVmdRepository(true)
					if err := rep.Save("", motion, false); err != nil {
						mlog.ET(mi18n.T("モーション保存失敗"), err, "")
						if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
							bakeState.SetWidgetEnabled(true)
						}
					}
				}
			}
		}

		bakeState.SetWidgetEnabled(true)
		controller.Beep()
	})

	return btn
}

func newEnabledPlaying(bakeState *BakeState) func(playing bool) {
	return func(playing bool) {
		bakeState.SetWidgetEnabled(!playing)
		if playing {
			// 再生中も操作可ウィジェットを有効化
			bakeState.SetWidgetPlayingEnabled(true)
		}
	}
}

func newOnChangePlayingPre(bakeState *BakeState, mWidgets *controller.MWidgets) func(playing bool) {
	return func(playing bool) {
		mWidgets.Window().SetSaveDelta(0, playing)

		// 情報表示
		mWidgets.Window().SetCheckedShowInfoEnabled(playing)

		if playing {
			// 焼き込み開始時にINDEX加算
			deltaIndex := mWidgets.Window().GetDeltaMotionCount(0, bakeState.CurrentIndex())
			if deltaIndex > 0 {
				// 既に焼き込みが1回以上行われている場合はインクリメント
				deltaIndex += 1
			}
			mWidgets.Window().SetSaveDeltaIndex(0, deltaIndex)

			// 再生フレーム
			mlog.IL(mi18n.T("焼き込み再生開始: 焼き込み履歴INDEX[%d]"), deltaIndex+1)
			mWidgets.Window().SetFrame(mWidgets.Window().Frame() - 2)
			mWidgets.Window().StorePhysicsReset(vmd.PHYSICS_RESET_TYPE_START_FIT_FRAME)
		} else {
			// 焼き込み完了時に範囲を更新
			deltaCnt := mWidgets.Window().GetDeltaMotionCount(0, bakeState.CurrentIndex())
			bakeState.BakedHistoryIndexEdit.SetRange(1, float64(deltaCnt))
			bakeState.BakedHistoryIndexEdit.SetValue(float64(deltaCnt))
		}
	}
}

func newPhysicsTableView(bakeState *BakeState, mWidgets *controller.MWidgets) declarative.TableView {
	return declarative.TableView{
		AssignTo:         &bakeState.PhysicsTableView,
		Model:            domain.NewPhysicsTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 150},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("重力"), Width: 60},
			{Title: mi18n.T("最大演算回数"), Width: 100},
			{Title: mi18n.T("物理演算頻度"), Width: 100},
			{Title: mi18n.T("開始時用整形"), Width: 100},
		},
		OnItemClicked: newPhysicsTableViewDialog(bakeState, mWidgets),
	}
}

func newPhysicsTableViewDialog(bakeState *BakeState, mWidgets *controller.MWidgets) func() {
	return func() {
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

		builder := declarative.NewBuilder(mWidgets.Window())

		dialog := &declarative.Dialog{
			AssignTo:      &dlg,
			CancelButton:  &cancelBtn,
			DefaultButton: &okBtn,
			Title:         mi18n.T("物理設定変更"),
			Layout:        declarative.VBox{},
			MinSize:       declarative.Size{Width: 600, Height: 200},
			DataBinder: declarative.DataBinder{
				AssignTo:   &db,
				DataSource: bakeState.CurrentSet().PhysicsTableModel.Records[bakeState.PhysicsTableView.CurrentIndex()],
			},
			Children: []declarative.Widget{
				declarative.Composite{
					Layout: declarative.Grid{Columns: 6},
					Children: []declarative.Widget{
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
							MaxValue:           float64(bakeState.CurrentSet().MaxFrame() + 1),
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
							MaxValue:           float64(bakeState.CurrentSet().MaxFrame() + 1),
						},
						declarative.Label{
							Text:        mi18n.T("開始時用整形"),
							ToolTipText: mi18n.T("開始時用整形説明"),
						},
						declarative.CheckBox{
							Checked:     declarative.Bind("IsStartDeform"),
							ToolTipText: mi18n.T("開始時用整形説明"),
						},
						declarative.TextLabel{
							Text:        mi18n.T("重力"),
							ToolTipText: mi18n.T("重力説明"),
							OnMouseDown: func(x, y int, button walk.MouseButton) {
								mlog.IL("%s", mi18n.T("重力説明"))
							},
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							Value:              declarative.Bind("Gravity"),
							AssignTo:           &gravityEdit,
							MinValue:           -100.0, // 最小値
							MaxValue:           100.0,  // 最大値
							Decimals:           1,      // 小数点以下の桁数
							Increment:          0.1,    // 増分
							SpinButtonsVisible: true,   // スピンボタンを表示
							MinSize:            declarative.Size{Width: 60, Height: 20},
							MaxSize:            declarative.Size{Width: 60, Height: 20},
						},
						declarative.TextLabel{
							Text:        mi18n.T("最大演算回数"),
							ToolTipText: mi18n.T("最大演算回数説明"),
							OnMouseDown: func(x, y int, button walk.MouseButton) {
								mlog.IL("%s", mi18n.T("最大演算回数説明"))
							},
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							Value:              declarative.Bind("MaxSubSteps"),
							AssignTo:           &maxSubStepsEdit,
							MinValue:           1.0,   // 最小値
							MaxValue:           100.0, // 最大値
							Decimals:           0,     // 小数点以下の桁数
							Increment:          1.0,   // 増分
							SpinButtonsVisible: true,  // スピンボタンを表示
							MinSize:            declarative.Size{Width: 60, Height: 20},
							MaxSize:            declarative.Size{Width: 60, Height: 20},
						},
						declarative.TextLabel{
							Text:        mi18n.T("物理演算頻度"),
							ToolTipText: mi18n.T("物理演算頻度説明"),
							OnMouseDown: func(x, y int, button walk.MouseButton) {
								mlog.IL("%s", mi18n.T("物理演算頻度説明"))
							},
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							Value:              declarative.Bind("FixedTimeStep"),
							AssignTo:           &fixedTimeStepEdit,
							MinValue:           10.0,    // 最小値
							MaxValue:           48000.0, // 最大値
							Decimals:           0,       // 小数点以下の桁数
							Increment:          10.0,    // 増分
							SpinButtonsVisible: true,    // スピンボタンを表示
							StretchFactor:      20,
							MinSize:            declarative.Size{Width: 60, Height: 20},
							MaxSize:            declarative.Size{Width: 60, Height: 20},
						},
						declarative.VSeparator{
							ColumnSpan: 6,
						},
						declarative.TextLabel{
							Text:        mi18n.T("大きさX倍率"),
							ToolTipText: mi18n.T("大きさX倍率説明"),
							OnMouseDown: func(x, y int, button walk.MouseButton) {
								mlog.IL("%s", mi18n.T("大きさX倍率説明"))
							},
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							AssignTo: &sizeXEdit,
							OnValueChanged: func() {
								// 選択されているアイテムの大きさXを更新
								treeView.CurrentItem().(*domain.PhysicsItem).CalcSizeX(sizeXEdit.Value())
								// モデルの更新
								treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
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
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							AssignTo: &sizeYEdit,
							OnValueChanged: func() {
								// 選択されているアイテムの大きさYを更新
								treeView.CurrentItem().(*domain.PhysicsItem).CalcSizeY(sizeYEdit.Value())
								// モデルの更新
								treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
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
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							AssignTo: &sizeZEdit,
							OnValueChanged: func() {
								// 選択されているアイテムの大きさZを更新
								treeView.CurrentItem().(*domain.PhysicsItem).CalcSizeZ(sizeZEdit.Value())
								// モデルの更新
								treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
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
							AssignTo: &massEdit,
							OnValueChanged: func() {
								// 選択されているアイテムの質量を更新
								treeView.CurrentItem().(*domain.PhysicsItem).CalcMass(massEdit.Value())
								// モデルの更新
								treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
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
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							AssignTo: &stiffnessEdit,
							OnValueChanged: func() {
								// 選択されているアイテムの硬さを更新
								treeView.CurrentItem().(*domain.PhysicsItem).CalcStiffness(stiffnessEdit.Value())
								// モデルの更新
								treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())
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
							MinSize: declarative.Size{Width: 100, Height: 20},
						},
						declarative.NumberEdit{
							AssignTo: &tensionEdit,
							OnValueChanged: func() {
								// 選択されているアイテムの張りを更新
								treeView.CurrentItem().(*domain.PhysicsItem).CalcTension(tensionEdit.Value())
								// モデルの更新
								treeView.Model().(*domain.PhysicsRigidBodyTreeModel).PublishItemChanged(treeView.CurrentItem())

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
							AssignTo:   &treeView,
							Model:      bakeState.CurrentSet().PhysicsTableModel.Records[bakeState.PhysicsTableView.CurrentIndex()].TreeModel,
							MinSize:    declarative.Size{Width: 230, Height: 200},
							ColumnSpan: 6,
							OnCurrentItemChanged: func() {
								if treeView.CurrentItem() != nil {
									// 選択されたアイテムの情報を更新
									currentItem := treeView.CurrentItem().(*domain.PhysicsItem)
									sizeXEdit.ChangeValue(currentItem.SizeRatio().X)
									sizeYEdit.ChangeValue(currentItem.SizeRatio().Y)
									sizeZEdit.ChangeValue(currentItem.SizeRatio().Z)
									massEdit.ChangeValue(currentItem.MassRatio())
									stiffnessEdit.ChangeValue(currentItem.StiffnessRatio())
									tensionEdit.ChangeValue(currentItem.TensionRatio())
								}
							},
						},
					},
				},
				declarative.Composite{
					Layout: declarative.HBox{
						Alignment: declarative.AlignHFarVCenter,
					},
					Children: []declarative.Widget{
						declarative.PushButton{
							AssignTo: &okBtn,
							Text:     mi18n.T("登録"),
							OnClicked: func() {
								if err := db.Submit(); err != nil {
									mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
									return
								}
								dlg.Accept()
							},
						},
						declarative.PushButton{
							AssignTo: &cancelBtn,
							Text:     mi18n.T("削除"),
							OnClicked: func() {
								// 削除処理
								bakeState.CurrentSet().PhysicsTableModel.RemoveRow(bakeState.PhysicsTableView.CurrentIndex())
								if err := db.Submit(); err != nil {
									mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
									return
								}
								dlg.Accept()
							},
						},
						declarative.PushButton{
							AssignTo: &cancelBtn,
							Text:     mi18n.T("キャンセル"),
							OnClicked: func() {
								dlg.Cancel()
							},
						},
					},
				},
			},
		}

		if cmd, err := dialog.Run(builder.Parent().Form()); err == nil && cmd == walk.DlgCmdOK {
			bakeState.SetWidgetEnabled(false)

			physicsMotion := mWidgets.Window().LoadPhysicsMotion(0)
			for _, record := range bakeState.PhysicsTableView.Model().(*domain.PhysicsTableModel).Records {
				for f := record.StartFrame; f <= record.EndFrame; f++ {
					physicsMotion.AppendGravityFrame(vmd.NewGravityFrameByValue(f, &mmath.MVec3{
						X: 0,
						Y: float64(record.Gravity),
						Z: 0,
					}))
					physicsMotion.AppendMaxSubStepsFrame(vmd.NewMaxSubStepsFrameByValue(f, record.MaxSubSteps))
					physicsMotion.AppendFixedTimeStepFrame(vmd.NewFixedTimeStepFrameByValue(f, record.FixedTimeStep))
					if f == record.StartFrame {
						if record.IsStartDeform {
							// 開始時用整形をON
							physicsMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_START_FIT_FRAME))
						} else {
							// 前フレームから継続して物理演算を行う
							physicsMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
						}
					} else {
						// 開始と終了以外はリセットしない
						physicsMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_NONE))
					}
				}

				// 最後のフレームの後に物理リセットする
				physicsMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.EndFrame+1, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			}
			mWidgets.Window().StorePhysicsMotion(0, physicsMotion)
			mWidgets.Window().TriggerPhysicsReset()

			bakeState.SetWidgetEnabled(true)
			controller.Beep()

			// 次の作業用の行を追加して、更新
			currentIndex := bakeState.PhysicsTableView.CurrentIndex()
			if currentIndex == len(bakeState.CurrentSet().PhysicsTableModel.Records)-1 {
				// 最後の行が選択されている場合は、新しい行を追加
				bakeState.CurrentSet().PhysicsTableModel.AddRecord(
					bakeState.CurrentSet().OriginalModel,
					0,
					bakeState.CurrentSet().MaxFrame())
			}
			bakeState.PhysicsTableView.SetModel(bakeState.CurrentSet().PhysicsTableModel)
		}
	}
}

func newOutputTableView(bakeState *BakeState, mWidgets *controller.MWidgets) declarative.TableView {
	return declarative.TableView{
		AssignTo:         &bakeState.OutputTableView,
		Model:            domain.NewOutputTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 150},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("ボーン数"), Width: 60},
			{Title: mi18n.T("焼き込み対象ボーン名"), Width: 300},
		},
		OnItemClicked: newOutputTableViewDialog(bakeState, mWidgets),
	}
}

func newOutputTableViewDialog(bakeState *BakeState, mWidgets *controller.MWidgets) func() {
	return func() {
		// アイテムがクリックされたら、入力ダイアログを表示する
		var dlg *walk.Dialog
		var cancelBtn *walk.PushButton
		var okBtn *walk.PushButton
		var db *walk.DataBinder
		var treeView *walk.TreeView
		var ikCheckBox *walk.CheckBox
		var physicsCheckBox *walk.CheckBox

		builder := declarative.NewBuilder(mWidgets.Window())

		dialog := &declarative.Dialog{
			AssignTo:      &dlg,
			CancelButton:  &cancelBtn,
			DefaultButton: &okBtn,
			Title:         mi18n.T("焼き込み設定変更"),
			Layout:        declarative.VBox{},
			MinSize:       declarative.Size{Width: 600, Height: 200},
			DataBinder: declarative.DataBinder{
				AssignTo:   &db,
				DataSource: bakeState.CurrentSet().OutputTableModel.Records[bakeState.OutputTableView.CurrentIndex()],
			},
			Children: []declarative.Widget{
				declarative.Composite{
					Layout: declarative.Grid{Columns: 6},
					Children: []declarative.Widget{
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
							MaxValue:           float64(bakeState.CurrentSet().MaxFrame() + 1),
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
							MaxValue:           float64(bakeState.CurrentSet().MaxFrame() + 1),
						},
						declarative.Label{
							Text: mi18n.T("焼き込み対象ボーン"),
						},
						declarative.HSpacer{
							ColumnSpan: 1,
						},
						declarative.CheckBox{
							AssignTo: &ikCheckBox,
							Text:     mi18n.T("IK焼き込み対象"),
							OnClicked: func() {
								treeView.Model().(*domain.OutputBoneTreeModel).SetOutputIkChecked(treeView, nil, ikCheckBox.Checked())
							},
							ColumnSpan: 2,
						},
						declarative.CheckBox{
							AssignTo: &physicsCheckBox,
							Text:     mi18n.T("物理焼き込み対象"),
							OnClicked: func() {
								treeView.Model().(*domain.OutputBoneTreeModel).SetOutputPhysicsChecked(treeView, nil, physicsCheckBox.Checked())
							},
							ColumnSpan: 2,
						},
						declarative.TreeView{
							AssignTo:   &treeView,
							Model:      bakeState.CurrentSet().OutputTableModel.Records[bakeState.OutputTableView.CurrentIndex()].OutputBoneTreeModel,
							MinSize:    declarative.Size{Width: 230, Height: 200},
							Checkable:  true,
							ColumnSpan: 6,
						},
					},
				},
				declarative.Composite{
					Layout: declarative.HBox{
						Alignment: declarative.AlignHFarVCenter,
					},
					Children: []declarative.Widget{
						declarative.PushButton{
							AssignTo: &okBtn,
							Text:     mi18n.T("登録"),
							OnClicked: func() {
								if err := db.Submit(); err != nil {
									mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
									return
								}
								dlg.Accept()
							},
						},
						declarative.PushButton{
							AssignTo: &cancelBtn,
							Text:     mi18n.T("削除"),
							OnClicked: func() {
								// 削除処理
								bakeState.CurrentSet().OutputTableModel.RemoveRow(bakeState.OutputTableView.CurrentIndex())
								if err := db.Submit(); err != nil {
									mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
									return
								}
								dlg.Accept()
							},
						},
						declarative.PushButton{
							AssignTo: &cancelBtn,
							Text:     mi18n.T("キャンセル"),
							OnClicked: func() {
								dlg.Cancel()
							},
						},
					},
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
			// 次の作業用の行を追加して、更新
			currentIndex := bakeState.OutputTableView.CurrentIndex()
			bakeState.CurrentSet().OutputTableModel.Records[currentIndex].TargetBoneNames = bakeState.CurrentSet().OutputTableModel.Records[currentIndex].OutputBoneTreeModel.GetCheckedBoneNames()
			if currentIndex == len(bakeState.CurrentSet().OutputTableModel.Records)-1 {
				// 最後の行が選択されている場合は、新しい行を追加
				bakeState.CurrentSet().OutputTableModel.AddRecord(
					bakeState.CurrentSet().OriginalModel,
					0,
					bakeState.CurrentSet().MaxFrame())
			}
			bakeState.OutputTableView.SetModel(bakeState.CurrentSet().OutputTableModel)
		}
	}
}

func newBakedHistoryWidgets(bakeState *BakeState, mWidgets *controller.MWidgets) []declarative.Widget {
	return []declarative.Widget{
		declarative.HSpacer{
			ColumnSpan: 3,
		},
		declarative.TextLabel{
			Text:        mi18n.T("焼き込み履歴INDEX"),
			ToolTipText: mi18n.T("焼き込み履歴INDEX説明"),
		},
		declarative.NumberEdit{
			SpinButtonsVisible: true,
			AssignTo:           &bakeState.BakedHistoryIndexEdit,
			Decimals:           0,
			Increment:          1,
			MinValue:           1,
			MaxValue:           2,
			OnValueChanged: func() {
				// 出力モーションインデックスが変更されたときの処理
				currentSet := bakeState.CurrentSet()
				deltaIndex := int(bakeState.BakedHistoryIndexEdit.Value() - 1)
				if deltaIndex < 0 ||
					deltaIndex >= mWidgets.Window().GetDeltaMotionCount(0, currentSet.Index) {
					// インデックスが範囲外の場合は、0に戻す
					deltaIndex = 0
				}

				// 物理ありのモーションを取得
				outputMotion := mWidgets.Window().LoadDeltaMotion(0, currentSet.Index, deltaIndex)
				// mlog.I("変形情報呼び出し: [motion(%d)] %p", deltaIndex, outputMotion)
				// 物理確認用として設定
				mWidgets.Window().StoreMotion(1, currentSet.Index, outputMotion)
				mWidgets.Window().TriggerPhysicsReset()

				// 出力モーションを更新
				currentSet.OutputMotion = outputMotion
				currentSet.OutputMotionPath = currentSet.CreateOutputMotionPath()
				bakeState.OutputMotionPicker.ChangePath(currentSet.OutputMotionPath)
			},
		},
		declarative.PushButton{
			Text:        mi18n.T("焼き込み履歴クリア"),
			ToolTipText: mi18n.T("焼き込み履歴クリア説明"),
			OnClicked: func() {
				mWidgets.Window().ClearDeltaMotion(0, bakeState.CurrentIndex())
				mWidgets.Window().ClearDeltaMotion(1, bakeState.CurrentIndex())
				mWidgets.Window().SetSaveDeltaIndex(0, 0)
				mWidgets.Window().SetSaveDeltaIndex(1, 0)

				bakeState.BakedHistoryIndexEdit.SetValue(1.0)
				bakeState.BakedHistoryIndexEdit.SetRange(1.0, 2.0)
			},
		},
	}
}
