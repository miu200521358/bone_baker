package ui

import (
	"path/filepath"

	"github.com/miu200521358/physics_fixer/pkg/domain"

	"github.com/miu200521358/mlib_go/pkg/config/mconfig"
	"github.com/miu200521358/mlib_go/pkg/config/merr"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

func NewPhysicsPage(mWidgets *controller.MWidgets) declarative.TabPage {
	var physicsTab *walk.TabPage
	physicsState := new(PhysicsState)

	physicsState.Player = widget.NewMotionPlayer()
	physicsState.GravitySlider = widget.NewTextSlider(
		mi18n.T("重力"),
		mi18n.T("重力説明"),
		float32(-10.0), // sliderMin
		float32(20.0),  // sliderMax
		float32(9.8),   // initialValue
		8,              // gridColumns
		5,              // sliderColumns
		func(cw *controller.ControlWindow) {
		})

	physicsState.OutputMotionPicker = widget.NewVmdSaveFilePicker(
		mi18n.T("物理焼き込み後モーション(Vmd)"),
		mi18n.T("物理焼き込み後モーションツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			motion := cw.LoadMotion(0, physicsState.CurrentIndex())
			if motion == nil {
				return
			}

			if err := rep.Save(path, motion, false); err != nil {
				mlog.ET(mi18n.T("保存失敗"), err, "")
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetPhysicsEnabled(true)
				}
			}
		},
	)

	physicsState.OutputModelPicker = widget.NewPmxSaveFilePicker(
		mi18n.T("物理変更後モデル(Pmx)"),
		mi18n.T("物理変更後モデルツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			model := cw.LoadModel(0, physicsState.CurrentIndex())
			if model == nil {
				return
			}

			if err := rep.Save(path, model, false); err != nil {
				mlog.ET(mi18n.T("保存失敗"), err, "")
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetPhysicsEnabled(true)
				}
			}
		},
	)

	physicsState.OriginalMotionPicker = widget.NewVmdVpdLoadFilePicker(
		"vmd",
		mi18n.T("モーション(Vmd/Vpd)"),
		mi18n.T("モーションツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := physicsState.LoadPhysicsMotion(cw, path, true); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetPhysicsEnabled(true)
				}
			}
		},
	)

	physicsState.PhysicsModelPicker = widget.NewPmxLoadFilePicker(
		"pmx",
		mi18n.T("モデル(Pmx)"),
		mi18n.T("モデルツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := physicsState.LoadPhysicsModel(cw, path); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetPhysicsEnabled(true)
				}
			}
		},
	)

	physicsState.AddSetButton = widget.NewMPushButton()
	physicsState.AddSetButton.SetLabel(mi18n.T("セット追加"))
	physicsState.AddSetButton.SetTooltip(mi18n.T("セット追加説明"))
	physicsState.AddSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	physicsState.AddSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
		physicsState.PhysicsSets = append(physicsState.PhysicsSets,
			domain.NewPhysicsSet(len(physicsState.PhysicsSets)))
		physicsState.AddAction()
	})

	physicsState.ResetSetButton = widget.NewMPushButton()
	physicsState.ResetSetButton.SetLabel(mi18n.T("セット全削除"))
	physicsState.ResetSetButton.SetTooltip(mi18n.T("セット全削除説明"))
	physicsState.ResetSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	physicsState.ResetSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
		for n := range 2 {
			for m := range physicsState.NavToolBar.Actions().Len() {
				mWidgets.Window().StoreModel(n, m, nil)
				mWidgets.Window().StoreMotion(n, m, nil)
			}
		}

		physicsState.ResetSet()
	})

	physicsState.LoadSetButton = widget.NewMPushButton()
	physicsState.LoadSetButton.SetLabel(mi18n.T("セット設定読込"))
	physicsState.LoadSetButton.SetTooltip(mi18n.T("セット設定読込説明"))
	physicsState.LoadSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	physicsState.LoadSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
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
			physicsState.SetPhysicsEnabled(false)
			mconfig.SaveUserConfig("physics_set_path", dlg.FilePath, 1)

			for n := range 2 {
				for m := range physicsState.NavToolBar.Actions().Len() {
					mWidgets.Window().StoreModel(n, m, nil)
					mWidgets.Window().StoreMotion(n, m, nil)
				}
			}

			physicsState.ResetSet()
			physicsState.LoadSet(dlg.FilePath)

			for range len(physicsState.PhysicsSets) - 1 {
				physicsState.AddAction()
			}

			for index := range physicsState.PhysicsSets {
				physicsState.ChangeCurrentAction(index)
				physicsState.OriginalMotionPicker.SetForcePath(physicsState.PhysicsSets[index].OriginalMotionPath)
				physicsState.PhysicsModelPicker.SetForcePath(physicsState.PhysicsSets[index].PhysicsModelPath)
			}

			physicsState.SetCurrentIndex(0)
			physicsState.SetPhysicsEnabled(true)
		}
	})

	physicsState.SaveSetButton = widget.NewMPushButton()
	physicsState.SaveSetButton.SetLabel(mi18n.T("セット設定保存"))
	physicsState.SaveSetButton.SetTooltip(mi18n.T("セット設定保存説明"))
	physicsState.SaveSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	physicsState.SaveSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
		// 物理焼き込み元モーションパスを初期パスとする
		initialDirPath := filepath.Dir(physicsState.CurrentSet().OriginalMotionPath)

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
			physicsState.SaveSet(dlg.FilePath)
			mconfig.SaveUserConfig("physics_set_path", dlg.FilePath, 1)
		}
	})

	physicsState.SaveButton = widget.NewMPushButton()
	physicsState.SaveButton.SetLabel(mi18n.T("モーション保存"))
	physicsState.SaveButton.SetTooltip(mi18n.T("モーション保存説明"))
	physicsState.SaveButton.SetMinSize(declarative.Size{Width: 256, Height: 20})
	physicsState.SaveButton.SetStretchFactor(20)
	physicsState.SaveButton.SetOnClicked(func(cw *controller.ControlWindow) {
		physicsState.SetPhysicsEnabled(false)

		for _, physicsSet := range physicsState.PhysicsSets {
			if physicsSet.OutputMotionPath != "" && physicsSet.OutputMotion != nil {
				rep := repository.NewVmdRepository(true)
				if err := rep.Save(physicsSet.OutputMotionPath, physicsSet.OutputMotion, false); err != nil {
					mlog.ET(mi18n.T("保存失敗"), err, "")
					if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
						physicsState.SetPhysicsEnabled(true)
					}
				}
			}
		}

		physicsState.SetPhysicsEnabled(true)
		controller.Beep()
	})

	mWidgets.Widgets = append(mWidgets.Widgets, physicsState.Player, physicsState.OriginalMotionPicker,
		physicsState.PhysicsModelPicker, physicsState.OutputMotionPicker,
		physicsState.OutputModelPicker, physicsState.AddSetButton, physicsState.ResetSetButton,
		physicsState.LoadSetButton, physicsState.SaveSetButton, physicsState.SaveButton,
		physicsState.GravitySlider)
	mWidgets.SetOnLoaded(func() {
		physicsState.PhysicsSets = append(physicsState.PhysicsSets, domain.NewPhysicsSet(len(physicsState.PhysicsSets)))
		physicsState.AddAction()
	})
	mWidgets.SetOnChangePlaying(func(playing bool) {
		physicsState.SetPhysicsOptionEnabled(!playing)
	})

	return declarative.TabPage{
		Title:    mi18n.T("物理焼き込み"),
		AssignTo: &physicsTab,
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
					physicsState.AddSetButton.Widgets(),
					physicsState.ResetSetButton.Widgets(),
					physicsState.LoadSetButton.Widgets(),
					physicsState.SaveSetButton.Widgets(),
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
						AssignTo:           &physicsState.NavToolBar,
						MinSize:            declarative.Size{Width: 200, Height: 25},
						MaxSize:            declarative.Size{Width: 5120, Height: 25},
						DefaultButtonWidth: 200,
						Orientation:        walk.Horizontal,
						ButtonStyle:        declarative.ToolBarButtonTextOnly,
					},
				},
			},

			// セットごとの物理焼き込み内容
			declarative.ScrollView{
				Layout:  declarative.VBox{},
				MinSize: declarative.Size{Width: 126, Height: 350},
				MaxSize: declarative.Size{Width: 2560, Height: 5120},
				Children: []declarative.Widget{
					physicsState.OriginalMotionPicker.Widgets(),
					physicsState.PhysicsModelPicker.Widgets(),
					declarative.VSeparator{},
					declarative.TextLabel{
						Text: mi18n.T("物理焼き込みオプション"),
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							mlog.ILT(mi18n.T("物理焼き込みオプション"), mi18n.T("物理焼き込みオプション説明"))
						},
					},
					physicsState.GravitySlider.Widgets(),
					// declarative.VSeparator{},
					// declarative.Composite{
					// 	Layout: declarative.Grid{Columns: 3},
					// 	Children: []declarative.Widget{
					// 		declarative.CheckBox{
					// 			AssignTo:    &physicsState.AdoptPhysicsCheck,
					// 			Text:        mi18n.T("即時反映"),
					// 			ToolTipText: mi18n.T("即時反映説明"),
					// 			Checked:     true,
					// 		},
					// 		declarative.CheckBox{
					// 			AssignTo:    &physicsState.AdoptAllCheck,
					// 			Text:        mi18n.T("全セット反映"),
					// 			ToolTipText: mi18n.T("全セット反映説明"),
					// 			Checked:     true,
					// 		},
					// 		physicsState.TerminateButton.Widgets(),
					// 	},
					// },
					declarative.VSeparator{},
					physicsState.OutputMotionPicker.Widgets(),
					physicsState.OutputModelPicker.Widgets(),
				},
			},
			physicsState.SaveButton.Widgets(),
			physicsState.Player.Widgets(),
		},
	}
}
