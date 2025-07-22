package ui

import (
	"path/filepath"

	"github.com/miu200521358/physics_fixer/pkg/domain"

	"github.com/miu200521358/mlib_go/pkg/config/mconfig"
	"github.com/miu200521358/mlib_go/pkg/config/merr"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
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
	physicsState.Player.SetLabelTexts(mi18n.T("焼き込み停止"), mi18n.T("焼き込み再生"))

	physicsState.OutputMotionPicker = widget.NewVmdSaveFilePicker(
		mi18n.T("物理焼き込み後モーション(Vmd)"),
		mi18n.T("物理焼き込み後モーションツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
		},
	)

	physicsState.OutputModelPicker = widget.NewPmxSaveFilePicker(
		mi18n.T("物理変更後モデル(Pmx)"),
		mi18n.T("物理変更後モデルツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			// 実際に保存するのは、物理有効な元モデル
			model := physicsState.CurrentSet().OriginalModel
			if model == nil {
				return
			}

			if err := rep.Save(path, model, false); err != nil {
				mlog.ET(mi18n.T("保存失敗"), err, "")
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetWidgetEnabled(true)
				}
			}
		},
	)

	physicsState.OriginalMotionPicker = widget.NewVmdVpdLoadFilePicker(
		"vmd",
		mi18n.T("モーション(Vmd/Vpd)"),
		mi18n.T("モーションツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := physicsState.LoadMotion(cw, path, true); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetWidgetEnabled(true)
				}
			}
		},
	)

	physicsState.OriginalModelPicker = widget.NewPmxLoadFilePicker(
		"pmx",
		mi18n.T("モデル(Pmx)"),
		mi18n.T("モデルツールチップ"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := physicsState.LoadModel(cw, path); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					physicsState.SetWidgetEnabled(true)
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
			physicsState.SetWidgetEnabled(false)
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
				physicsState.OutputModelPicker.SetForcePath(physicsState.PhysicsSets[index].OutputModelPath)
			}

			physicsState.SetCurrentIndex(0)
			physicsState.SetWidgetEnabled(true)
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

	physicsState.SaveModelButton = widget.NewMPushButton()
	physicsState.SaveModelButton.SetLabel(mi18n.T("モデル保存"))
	physicsState.SaveModelButton.SetTooltip(mi18n.T("モデル保存説明"))
	physicsState.SaveModelButton.SetMinSize(declarative.Size{Width: 256, Height: 20})
	physicsState.SaveModelButton.SetStretchFactor(20)
	physicsState.SaveModelButton.SetOnClicked(func(cw *controller.ControlWindow) {
		physicsState.SetWidgetEnabled(false)

		for _, physicsSet := range physicsState.PhysicsSets {
			if physicsSet.OutputModelPath != "" && physicsSet.OriginalModel != nil {
				// 保存するのは物理が有効になっている元モデル
				rep := repository.NewPmxRepository(true)
				if err := rep.Save(physicsSet.OutputModelPath, physicsSet.OriginalModel, false); err != nil {
					mlog.ET(mi18n.T("保存失敗"), err, "")
					if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
						physicsState.SetWidgetEnabled(true)
					}
				}
			}
		}

		physicsState.SetWidgetEnabled(true)
		controller.Beep()
	})

	physicsState.SaveMotionButton = widget.NewMPushButton()
	physicsState.SaveMotionButton.SetLabel(mi18n.T("モーション保存"))
	physicsState.SaveMotionButton.SetTooltip(mi18n.T("モーション保存説明"))
	physicsState.SaveMotionButton.SetMinSize(declarative.Size{Width: 256, Height: 20})
	physicsState.SaveMotionButton.SetStretchFactor(20)
	physicsState.SaveMotionButton.SetOnClicked(func(cw *controller.ControlWindow) {
		physicsState.SetWidgetEnabled(false)

		for _, physicsSet := range physicsState.PhysicsSets {
			if physicsSet.OutputMotionPath != "" && physicsSet.OutputMotion != nil {
				// 物理ボーンのみ残す
				motion, err := physicsSet.GetOutputMotionOnlyPhysics(
					physicsState.StartFrameEdit.Value(),
					physicsState.EndFrameEdit.Value(),
				)
				if err != nil {
					mlog.ET(mi18n.T("保存失敗"), err, "")
					return
				}

				rep := repository.NewVmdRepository(true)
				if err := rep.Save(physicsSet.OutputMotionPath, motion, false); err != nil {
					mlog.ET(mi18n.T("保存失敗"), err, "")
					if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
						physicsState.SetWidgetEnabled(true)
					}
				}
			}
		}

		physicsState.SetWidgetEnabled(true)
		controller.Beep()
	})

	mWidgets.Widgets = append(mWidgets.Widgets, physicsState.Player, physicsState.OriginalMotionPicker,
		physicsState.OriginalModelPicker, physicsState.OutputMotionPicker,
		physicsState.OutputModelPicker, physicsState.AddSetButton, physicsState.ResetSetButton,
		physicsState.LoadSetButton, physicsState.SaveSetButton, physicsState.SaveMotionButton,
		physicsState.SaveModelButton)
	mWidgets.SetOnLoaded(func() {
		physicsState.PhysicsSets = append(physicsState.PhysicsSets, domain.NewPhysicsSet(len(physicsState.PhysicsSets)))
		physicsState.AddAction()
	})
	mWidgets.SetOnChangePlaying(func(playing bool) {
		mWidgets.Window().SetSaveDelta(0, playing)
		physicsState.SetWidgetEnabled(!playing)

		// フレームドロップ無効
		mWidgets.Window().SetFrameDropEnabled(false)

		if playing {
			physicsState.SetWidgetPlayingEnabled(true)

			// 焼き込み開始時にINDEX加算
			deltaIndex := mWidgets.Window().GetDeltaMotionCount(0, physicsState.CurrentIndex())
			deltaIndex += 1

			for _, physicsSet := range physicsState.PhysicsSets {
				if physicsSet.OriginalMotion != nil {
					if copiedOriginalMotion, err := physicsSet.OriginalMotion.Copy(); err == nil {
						mWidgets.Window().StoreDeltaMotion(0, physicsSet.Index, deltaIndex, copiedOriginalMotion)
					}
				}
			}

			mWidgets.Window().SetSaveDeltaIndex(0, deltaIndex)
		} else {
			// 焼き込み完了時に範囲を更新
			deltaCnt := mWidgets.Window().GetDeltaMotionCount(0, physicsState.CurrentIndex())
			physicsState.OutputMotionIndexEdit.SetRange(1, float64(deltaCnt))
			physicsState.OutputMotionIndexEdit.SetValue(float64(deltaCnt))
		}
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
					physicsState.OriginalModelPicker.Widgets(),
					physicsState.OriginalMotionPicker.Widgets(),
					declarative.VSeparator{},
					declarative.TextLabel{
						Text: mi18n.T("物理焼き込みオプション"),
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							mlog.ILT(mi18n.T("物理焼き込みオプション"), mi18n.T("物理焼き込みオプション説明"))
						},
					},
					declarative.Composite{
						Layout: declarative.Grid{Columns: 6},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("重力"),
								ToolTipText: mi18n.T("重力説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("重力説明"))
								},
								StretchFactor: 1,
							},
							declarative.NumberEdit{
								AssignTo:           &physicsState.GravityEdit,
								Value:              -9.8,   // 初期値
								MinValue:           -100.0, // 最小値
								MaxValue:           100.0,  // 最大値
								Decimals:           1,      // 小数点以下の桁数
								Increment:          0.1,    // 増分
								SpinButtonsVisible: true,   // スピンボタンを表示
								StretchFactor:      20,
							},
							declarative.TextLabel{
								Text:        mi18n.T("サブステップ数"),
								ToolTipText: mi18n.T("サブステップ数説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("サブステップ数説明"))
								},
								StretchFactor: 1,
							},
							declarative.NumberEdit{
								AssignTo:           &physicsState.MaxSubStepsEdit,
								Value:              2.0,   // 初期値
								MinValue:           1.0,   // 最小値
								MaxValue:           100.0, // 最大値
								Decimals:           0,     // 小数点以下の桁数
								Increment:          1.0,   // 増分
								SpinButtonsVisible: true,  // スピンボタンを表示
								StretchFactor:      20,
							},
							declarative.TextLabel{
								Text:        mi18n.T("演算精度"),
								ToolTipText: mi18n.T("演算精度説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("演算精度説明"))
								},
								StretchFactor: 1,
							},
							declarative.NumberEdit{
								AssignTo:           &physicsState.FixedTimeStepEdit,
								Value:              60.0,   // 初期値
								MinValue:           10.0,   // 最小値
								MaxValue:           4800.0, // 最大値
								Decimals:           0,      // 小数点以下の桁数
								Increment:          10.0,   // 増分
								SpinButtonsVisible: true,   // スピンボタンを表示
								StretchFactor:      20,
							},
							declarative.TextLabel{
								Text:        mi18n.T("質量"),
								ToolTipText: mi18n.T("質量説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("質量説明"))
								},
								StretchFactor: 1,
							},
							declarative.NumberEdit{
								AssignTo: &physicsState.MassEdit,
								OnValueChanged: func() {
									physicsState.PhysicsTreeView.CurrentItem().(*domain.PhysicsItem).CalcMass(
										physicsState.MassEdit.Value(),
									)
									physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).PublishItemChanged(physicsState.PhysicsTreeView.CurrentItem())
								},
								Value:              1,    // 初期値
								MinValue:           0.0,  // 最小値
								MaxValue:           10.0, // 最大値
								Decimals:           2,    // 小数点以下の桁数
								Increment:          0.01, // 増分
								SpinButtonsVisible: true, // スピンボタンを表示
								StretchFactor:      20,
							},
							declarative.TextLabel{
								Text:        mi18n.T("硬さ"),
								ToolTipText: mi18n.T("硬さ説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("硬さ説明"))
								},
								StretchFactor: 1,
							},
							declarative.NumberEdit{
								AssignTo: &physicsState.StiffnessEdit,
								OnValueChanged: func() {
									physicsState.PhysicsTreeView.CurrentItem().(*domain.PhysicsItem).CalcStiffness(
										physicsState.StiffnessEdit.Value(),
									)
									physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).PublishItemChanged(physicsState.PhysicsTreeView.CurrentItem())
								},
								Value:              1,    // 初期値
								MinValue:           0.0,  // 最小値
								MaxValue:           10.0, // 最大値
								Decimals:           2,    // 小数点以下の桁数
								Increment:          0.01, // 増分
								SpinButtonsVisible: true, // スピンボタンを表示
								StretchFactor:      20,
							},
							declarative.TextLabel{
								Text:        mi18n.T("張り"),
								ToolTipText: mi18n.T("張り説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("張り説明"))
								},
								StretchFactor: 1,
							},
							declarative.NumberEdit{
								AssignTo: &physicsState.TensionEdit,
								OnValueChanged: func() {
									physicsState.PhysicsTreeView.CurrentItem().(*domain.PhysicsItem).CalcTension(
										physicsState.TensionEdit.Value(),
									)
									physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).PublishItemChanged(physicsState.PhysicsTreeView.CurrentItem())
								},
								Value:              1,    // 初期値
								MinValue:           0.0,  // 最小値
								MaxValue:           10.0, // 最大値
								Decimals:           2,    // 小数点以下の桁数
								Increment:          0.01, // 増分
								SpinButtonsVisible: true, // スピンボタンを表示
								StretchFactor:      20,
							},
							declarative.PushButton{
								Text:          mi18n.T("物理設定変更"),
								ToolTipText:   mi18n.T("物理設定変更説明"),
								ColumnSpan:    4,
								StretchFactor: 30,
								OnClicked: func() {
									physicsState.SetWidgetEnabled(false)

									gravity := mWidgets.Window().Gravity()
									gravity.Y = physicsState.GravityEdit.Value() // 重力のY成分を更新
									mWidgets.Window().SetGravity(gravity)

									mWidgets.Window().SetMaxSubSteps(int(physicsState.MaxSubStepsEdit.Value()))
									mWidgets.Window().SetFixedTimeStep(int(physicsState.FixedTimeStepEdit.Value()))

									model := physicsState.CurrentSet().OriginalModel
									model.RigidBodies.ForEach(func(rigidIndex int, rb *pmx.RigidBody) bool {
										physicsItem := physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).AtByBoneIndex(rb.BoneIndex)

										if physicsItem == nil {
											return true
										}

										// 質量、硬さ、張りを設定
										rb.RigidBodyParam.Mass *= physicsItem.(*domain.PhysicsItem).MassRatio()

										return true
									})
									model.Joints.ForEach(func(jointIndex int, joint *pmx.Joint) bool {
										rigidBodyA, _ := model.RigidBodies.Get(joint.RigidbodyIndexA)
										rigidBodyB, _ := model.RigidBodies.Get(joint.RigidbodyIndexB)

										var physicsItemA, physicsItemB walk.TreeItem
										if rigidBodyA != nil && rigidBodyA.BoneIndex >= 0 {
											physicsItemA = physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).AtByBoneIndex(rigidBodyA.BoneIndex)
										}
										if physicsItemA == nil {
											physicsItemA = domain.NewPhysicsItem(nil, nil)
										}
										if rigidBodyB != nil && rigidBodyB.BoneIndex >= 0 {
											physicsItemB = physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).AtByBoneIndex(rigidBodyB.BoneIndex)
										}
										if physicsItemB == nil {
											physicsItemB = domain.NewPhysicsItem(nil, nil)
										}

										// 質量、硬さ、張りを設定
										joint.JointParam.TranslationLimitMin.MulScalar(max(
											1/physicsItemA.(*domain.PhysicsItem).StiffnessRatio(),
											1/physicsItemB.(*domain.PhysicsItem).StiffnessRatio()))
										joint.JointParam.TranslationLimitMax.MulScalar(max(
											1/physicsItemA.(*domain.PhysicsItem).StiffnessRatio(),
											1/physicsItemB.(*domain.PhysicsItem).StiffnessRatio()))

										joint.JointParam.RotationLimitMin.MulScalar(max(
											1/physicsItemA.(*domain.PhysicsItem).StiffnessRatio(),
											1/physicsItemB.(*domain.PhysicsItem).StiffnessRatio()))
										joint.JointParam.RotationLimitMax.MulScalar(max(
											1/physicsItemA.(*domain.PhysicsItem).StiffnessRatio(),
											1/physicsItemB.(*domain.PhysicsItem).StiffnessRatio()))

										joint.JointParam.SpringConstantTranslation.MulScalar(max(
											physicsItemA.(*domain.PhysicsItem).StiffnessRatio(),
											physicsItemB.(*domain.PhysicsItem).StiffnessRatio()))
										joint.JointParam.SpringConstantRotation.MulScalar(max(
											physicsItemA.(*domain.PhysicsItem).TensionRatio(),
											physicsItemB.(*domain.PhysicsItem).TensionRatio()))

										return true
									})

									physicsState.CurrentSet().OriginalModel = model
									mWidgets.Window().StoreModel(0, physicsState.CurrentIndex(), model)
									physicsState.OutputModelPicker.ChangePath(physicsState.CurrentSet().CreateOutputModelPath())
									mWidgets.Window().TriggerPhysicsReset()

									if mWidgets.Window().Playing() {
										// 再生中は、調整系だけ有効にする
										physicsState.SetWidgetPlayingEnabled(true)
									} else {
										physicsState.SetWidgetEnabled(true)
									}

									controller.Beep()
								},
							},
							declarative.PushButton{
								Text:          mi18n.T("物理リセット"),
								ToolTipText:   mi18n.T("物理リセット説明"),
								ColumnSpan:    2,
								StretchFactor: 30,
								OnClicked: func() {
									physicsState.SetWidgetEnabled(false)

									// 物理ツリーをリセット
									physicsState.PhysicsTreeView.Model().(*domain.PhysicsModel).Reset()

									physicsState.MassEdit.SetValue(1.0)
									physicsState.StiffnessEdit.SetValue(1.0)
									physicsState.TensionEdit.SetValue(1.0)

									if err := physicsState.CurrentSet().LoadModel(physicsState.CurrentSet().OriginalModelPath); err == nil {
										mWidgets.Window().StoreModel(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalModel)
										physicsState.OutputModelPicker.ChangePath(physicsState.CurrentSet().CreateOutputModelPath())
										mWidgets.Window().TriggerPhysicsReset()
									}

									if mWidgets.Window().Playing() {
										// 再生中は、調整系だけ有効にする
										physicsState.SetWidgetPlayingEnabled(true)
									} else {
										physicsState.SetWidgetEnabled(true)
									}

									controller.Beep()
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.TreeView{
								AssignTo: &physicsState.PhysicsTreeView,
								Model:    domain.NewPhysicsModel(),
								MinSize:  declarative.Size{Width: 230, Height: 200},
								OnCurrentItemChanged: func() {
									// 物理ボーンツリーの選択が変更されたときの処理
									currentItem := physicsState.PhysicsTreeView.CurrentItem()
									if currentItem == nil {
										return
									}

									physicsItem := currentItem.(*domain.PhysicsItem)
									physicsState.MassEdit.SetValue(physicsItem.MassRatio())
									physicsState.StiffnessEdit.SetValue(physicsItem.StiffnessRatio())
									physicsState.TensionEdit.SetValue(physicsItem.TensionRatio())
								},
							},
						},
					},
					physicsState.OutputModelPicker.Widgets(),
					physicsState.OutputMotionPicker.Widgets(),
					declarative.VSeparator{},
					physicsState.SaveModelButton.Widgets(),
					declarative.Composite{
						Layout: declarative.HBox{},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("焼き込みIndex"),
								ToolTipText: mi18n.T("焼き込みIndex説明"),
							},
							declarative.NumberEdit{
								SpinButtonsVisible: true,
								AssignTo:           &physicsState.OutputMotionIndexEdit,
								Decimals:           0,
								Increment:          1,
								MinValue:           1,
								MaxValue:           2,
								OnValueChanged: func() {
									// 出力モーションインデックスが変更されたときの処理
									currentSet := physicsState.CurrentSet()
									deltaIndex := int(physicsState.OutputMotionIndexEdit.Value() - 1)
									if deltaIndex < 0 ||
										deltaIndex >= mWidgets.Window().GetDeltaMotionCount(0, currentSet.Index) {
										// インデックスが範囲外の場合は、0に戻す
										deltaIndex = 0
									}

									// 物理ありのモーションを取得
									outputMotion := mWidgets.Window().LoadDeltaMotion(0, currentSet.Index, deltaIndex)
									mlog.I("変形情報呼び出し: [motion(%d)] %p", deltaIndex, outputMotion)
									// 物理確認用として設定
									mWidgets.Window().StoreMotion(1, currentSet.Index, outputMotion)
									mWidgets.Window().TriggerPhysicsReset()

									// 出力モーションを更新
									currentSet.OutputMotion = outputMotion
									currentSet.OutputMotionPath = currentSet.CreateOutputMotionPath()
									physicsState.OutputMotionPicker.ChangePath(currentSet.OutputMotionPath)
								},
							},
							declarative.TextLabel{
								Text:        mi18n.T("開始"),
								ToolTipText: mi18n.T("開始フレーム説明"),
							},
							declarative.NumberEdit{
								ToolTipText:        mi18n.T("開始フレーム説明"),
								SpinButtonsVisible: true,
								AssignTo:           &physicsState.StartFrameEdit,
								Decimals:           0,
								Increment:          1,
								MinValue:           0,
								MaxValue:           1,
							},
							declarative.TextLabel{
								Text:        mi18n.T("終了"),
								ToolTipText: mi18n.T("終了フレーム説明"),
							},
							declarative.NumberEdit{
								ToolTipText:        mi18n.T("終了フレーム説明"),
								SpinButtonsVisible: true,
								AssignTo:           &physicsState.EndFrameEdit,
								Decimals:           0,
								Increment:          1,
								MinValue:           0,
								MaxValue:           1,
							},
						},
					},
					physicsState.SaveMotionButton.Widgets(),
				},
			},
			physicsState.Player.Widgets(),
		},
	}
}
