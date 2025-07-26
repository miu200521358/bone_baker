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
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
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

	bakeState.OutputMotionPicker = widget.NewVmdSaveFilePicker(
		mi18n.T("焼き込み後モーション(Vmd)"),
		mi18n.T("焼き込み後モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
		},
	)

	bakeState.OutputModelPicker = widget.NewPmxSaveFilePicker(
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

	bakeState.OriginalMotionPicker = widget.NewVmdLoadFilePicker(
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

	bakeState.OriginalModelPicker = widget.NewPmxLoadFilePicker(
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

	bakeState.AddSetButton = widget.NewMPushButton()
	bakeState.AddSetButton.SetLabel(mi18n.T("セット追加"))
	bakeState.AddSetButton.SetTooltip(mi18n.T("セット追加説明"))
	bakeState.AddSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	bakeState.AddSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
		bakeState.BakeSets = append(bakeState.BakeSets,
			domain.NewPhysicsSet(len(bakeState.BakeSets)))
		bakeState.AddAction()
	})

	bakeState.ResetSetButton = widget.NewMPushButton()
	bakeState.ResetSetButton.SetLabel(mi18n.T("セット全削除"))
	bakeState.ResetSetButton.SetTooltip(mi18n.T("セット全削除説明"))
	bakeState.ResetSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	bakeState.ResetSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
		for n := range 2 {
			for m := range bakeState.NavToolBar.Actions().Len() {
				mWidgets.Window().StoreModel(n, m, nil)
				mWidgets.Window().StoreMotion(n, m, nil)
			}
		}

		bakeState.ResetSet()
	})

	bakeState.LoadSetButton = widget.NewMPushButton()
	bakeState.LoadSetButton.SetLabel(mi18n.T("セット設定読込"))
	bakeState.LoadSetButton.SetTooltip(mi18n.T("セット設定読込説明"))
	bakeState.LoadSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	bakeState.LoadSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
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

	bakeState.SaveSetButton = widget.NewMPushButton()
	bakeState.SaveSetButton.SetLabel(mi18n.T("セット設定保存"))
	bakeState.SaveSetButton.SetTooltip(mi18n.T("セット設定保存説明"))
	bakeState.SaveSetButton.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	bakeState.SaveSetButton.SetOnClicked(func(cw *controller.ControlWindow) {
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

	bakeState.SaveModelButton = widget.NewMPushButton()
	bakeState.SaveModelButton.SetLabel(mi18n.T("モデル保存"))
	bakeState.SaveModelButton.SetTooltip(mi18n.T("モデル保存説明"))
	bakeState.SaveModelButton.SetMinSize(declarative.Size{Width: 256, Height: 20})
	bakeState.SaveModelButton.SetStretchFactor(20)
	bakeState.SaveModelButton.SetOnClicked(func(cw *controller.ControlWindow) {
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

	bakeState.SaveMotionButton = widget.NewMPushButton()
	bakeState.SaveMotionButton.SetLabel(mi18n.T("モーション保存"))
	bakeState.SaveMotionButton.SetTooltip(mi18n.T("モーション保存説明"))
	bakeState.SaveMotionButton.SetMinSize(declarative.Size{Width: 256, Height: 20})
	bakeState.SaveMotionButton.SetStretchFactor(20)
	bakeState.SaveMotionButton.SetOnClicked(func(cw *controller.ControlWindow) {
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
					if err := rep.Save(motion.Path(), motion, false); err != nil {
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

	mWidgets.Widgets = append(mWidgets.Widgets, bakeState.Player, bakeState.OriginalMotionPicker,
		bakeState.OriginalModelPicker, bakeState.OutputMotionPicker,
		bakeState.OutputModelPicker, bakeState.AddSetButton, bakeState.ResetSetButton,
		bakeState.LoadSetButton, bakeState.SaveSetButton, bakeState.SaveMotionButton,
		bakeState.SaveModelButton)
	mWidgets.SetOnLoaded(func() {
		bakeState.BakeSets = append(bakeState.BakeSets, domain.NewPhysicsSet(len(bakeState.BakeSets)))
		bakeState.AddAction()
	})
	mWidgets.SetOnChangePlaying(func(playing bool) {
		mWidgets.Window().SetSaveDelta(0, playing)
		bakeState.SetWidgetEnabled(!playing)

		// // フレームドロップ無効
		// mWidgets.Window().SetCheckedFrameDropEnabled(!playing)

		// 情報表示
		mWidgets.Window().SetCheckedShowInfoEnabled(playing)

		if playing {
			// // FPS無制限
			// mWidgets.Window().TriggerUnLimitFps()

			bakeState.SetWidgetPlayingEnabled(true)

			// 焼き込み開始時にINDEX加算
			deltaIndex := mWidgets.Window().GetDeltaMotionCount(0, bakeState.CurrentIndex())
			if deltaIndex > 0 {
				// 既に焼き込みが1回以上行われている場合はインクリメント
				deltaIndex += 1
			}

			// for _, physicsSet := range bakeState.BakeSets {
			// 	if physicsSet.OriginalMotion != nil {
			// 		if copiedOriginalMotion, err := physicsSet.OriginalMotion.Copy(); err == nil {
			// 			mWidgets.Window().StoreDeltaMotion(0, physicsSet.Index, deltaIndex, copiedOriginalMotion)
			// 		}
			// 	}
			// }

			mWidgets.Window().SetSaveDeltaIndex(0, deltaIndex)
		} else {
			// 焼き込み完了時に範囲を更新
			deltaCnt := mWidgets.Window().GetDeltaMotionCount(0, bakeState.CurrentIndex())
			bakeState.OutputMotionIndexEdit.SetRange(1, float64(deltaCnt))
			bakeState.OutputMotionIndexEdit.SetValue(float64(deltaCnt))

			// 30FPS制限
			mWidgets.Window().TriggerFps30Limit()
		}
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
						Text: mi18n.T("物理設定オプション"),
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							mlog.ILT(mi18n.T("物理設定オプション"), mi18n.T("物理設定オプション説明"))
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
								MinSize: declarative.Size{Width: 80, Height: 20},
							},
							declarative.NumberEdit{
								AssignTo:           &bakeState.GravityEdit,
								Value:              -9.8,   // 初期値
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
								MinSize: declarative.Size{Width: 80, Height: 20},
							},
							declarative.NumberEdit{
								AssignTo:           &bakeState.MaxSubStepsEdit,
								Value:              2.0,   // 初期値
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
								MinSize: declarative.Size{Width: 80, Height: 20},
							},
							declarative.NumberEdit{
								AssignTo:           &bakeState.FixedTimeStepEdit,
								Value:              60.0,    // 初期値
								MinValue:           10.0,    // 最小値
								MaxValue:           48000.0, // 最大値
								Decimals:           0,       // 小数点以下の桁数
								Increment:          10.0,    // 増分
								SpinButtonsVisible: true,    // スピンボタンを表示
								StretchFactor:      20,
								MinSize:            declarative.Size{Width: 60, Height: 20},
								MaxSize:            declarative.Size{Width: 60, Height: 20},
							},
							declarative.TextLabel{
								Text:        mi18n.T("質量倍率"),
								ToolTipText: mi18n.T("質量倍率説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.IL("%s", mi18n.T("質量倍率説明"))
								},
								MinSize: declarative.Size{Width: 80, Height: 20},
							},
							declarative.NumberEdit{
								AssignTo: &bakeState.MassEdit,
								OnValueChanged: func() {
									if currentItem := bakeState.PhysicsTreeView.CurrentItem(); currentItem != nil {
										currentItem.(*domain.PhysicsItem).CalcMass(bakeState.MassEdit.Value())
										bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).PublishItemChanged(currentItem)
									}
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
								MinSize: declarative.Size{Width: 80, Height: 20},
							},
							declarative.NumberEdit{
								AssignTo: &bakeState.StiffnessEdit,
								OnValueChanged: func() {
									if currentItem := bakeState.PhysicsTreeView.CurrentItem(); currentItem != nil {
										// 選択されている物理ボーンの硬さを更新
										currentItem.(*domain.PhysicsItem).CalcStiffness(bakeState.StiffnessEdit.Value())
										bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).PublishItemChanged(currentItem)
									}
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
								MinSize: declarative.Size{Width: 80, Height: 20},
							},
							declarative.NumberEdit{
								AssignTo: &bakeState.TensionEdit,
								OnValueChanged: func() {
									if currentItem := bakeState.PhysicsTreeView.CurrentItem(); currentItem != nil {
										currentItem.(*domain.PhysicsItem).CalcTension(bakeState.TensionEdit.Value())
										bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).PublishItemChanged(currentItem)
									}
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
							declarative.PushButton{
								Text:        mi18n.T("物理設定変更"),
								ToolTipText: mi18n.T("物理設定変更説明"),
								ColumnSpan:  4,
								OnClicked: func() {
									bakeState.SetWidgetEnabled(false)

									gravity := mWidgets.Window().Gravity()
									gravity.Y = bakeState.GravityEdit.Value() // 重力のY成分を更新
									mWidgets.Window().SetGravity(gravity)

									mWidgets.Window().SetMaxSubSteps(int(bakeState.MaxSubStepsEdit.Value()))
									mWidgets.Window().SetFixedTimeStep(int(bakeState.FixedTimeStepEdit.Value()))

									model := bakeState.CurrentSet().OriginalModel
									model.RigidBodies.ForEach(func(rigidIndex int, rb *pmx.RigidBody) bool {
										physicsItem := bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).AtByBoneIndex(rb.BoneIndex)

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
											physicsItemA = bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).AtByBoneIndex(rigidBodyA.BoneIndex)
										}
										if physicsItemA == nil {
											physicsItemA = domain.NewPhysicsItem(nil, nil)
										}
										if rigidBodyB != nil && rigidBodyB.BoneIndex >= 0 {
											physicsItemB = bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).AtByBoneIndex(rigidBodyB.BoneIndex)
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

									bakeState.CurrentSet().OriginalModel = model
									mWidgets.Window().StoreModel(0, bakeState.CurrentIndex(), model)
									bakeState.OutputModelPicker.ChangePath(bakeState.CurrentSet().CreateOutputModelPath())
									mWidgets.Window().TriggerPhysicsReset()

									if mWidgets.Window().Playing() {
										// 再生中は、調整系だけ有効にする
										bakeState.SetWidgetPlayingEnabled(true)
									} else {
										bakeState.SetWidgetEnabled(true)
									}

									controller.Beep()
								},
							},
							declarative.PushButton{
								Text:        mi18n.T("物理リセット"),
								ToolTipText: mi18n.T("物理リセット説明"),
								ColumnSpan:  2,
								OnClicked: func() {
									bakeState.SetWidgetEnabled(false)

									// 物理ツリーをリセット
									bakeState.PhysicsTreeView.Model().(*domain.PhysicsBoneTreeModel).Reset()

									bakeState.MassEdit.SetValue(1.0)
									bakeState.StiffnessEdit.SetValue(1.0)
									bakeState.TensionEdit.SetValue(1.0)

									if err := bakeState.bakeUsecase.LoadModel(bakeState.CurrentSet(), bakeState.CurrentSet().OriginalModelPath); err == nil {
										mWidgets.Window().StoreModel(0, bakeState.CurrentIndex(), bakeState.CurrentSet().OriginalModel)
										bakeState.OutputModelPicker.ChangePath(bakeState.CurrentSet().CreateOutputModelPath())
										mWidgets.Window().TriggerPhysicsReset()
									}

									if mWidgets.Window().Playing() {
										// 再生中は、調整系だけ有効にする
										bakeState.SetWidgetPlayingEnabled(true)
									} else {
										bakeState.SetWidgetEnabled(true)
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
								AssignTo: &bakeState.PhysicsTreeView,
								Model:    domain.NewPhysicsBoneTreeModel(),
								MinSize:  declarative.Size{Width: 230, Height: 150},
								OnCurrentItemChanged: func() {
									// 物理ボーンツリーの選択が変更されたときの処理
									currentItem := bakeState.PhysicsTreeView.CurrentItem()
									if currentItem == nil {
										return
									}

									physicsItem := currentItem.(*domain.PhysicsItem)
									bakeState.MassEdit.SetValue(physicsItem.MassRatio())
									bakeState.StiffnessEdit.SetValue(physicsItem.StiffnessRatio())
									bakeState.TensionEdit.SetValue(physicsItem.TensionRatio())
								},
							},
						},
					},
					declarative.VSeparator{},
					bakeState.OutputModelPicker.Widgets(),
					bakeState.SaveModelButton.Widgets(),
					declarative.VSeparator{},
					bakeState.OutputMotionPicker.Widgets(),
					declarative.Composite{
						Layout: declarative.Grid{Columns: 6},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("焼き込み保存設定テーブル"),
								ToolTipText: mi18n.T("焼き込み保存設定テーブル説明"),
							},
							declarative.HSpacer{
								ColumnSpan: 3,
							},
							declarative.TextLabel{
								Text:        mi18n.T("焼き込み履歴INDEX"),
								ToolTipText: mi18n.T("焼き込み履歴INDEX説明"),
							},
							declarative.NumberEdit{
								SpinButtonsVisible: true,
								AssignTo:           &bakeState.OutputMotionIndexEdit,
								Decimals:           0,
								Increment:          1,
								MinValue:           1,
								MaxValue:           2,
								OnValueChanged: func() {
									// 出力モーションインデックスが変更されたときの処理
									currentSet := bakeState.CurrentSet()
									deltaIndex := int(bakeState.OutputMotionIndexEdit.Value() - 1)
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
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.TableView{
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
								OnItemClicked: func() {
									// アイテムがクリックされたら、入力ダイアログを表示する
									var dlg *walk.Dialog
									var cancelBtn *walk.PushButton
									var okBtn *walk.PushButton
									var db *walk.DataBinder
									var treeView *walk.TreeView
									var ikCheckBox *walk.CheckBox
									var physicsCheckBox *walk.CheckBox

									builder := declarative.NewBuilder(bakeTab)

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
								},
							},
						},
					},
					// declarative.Composite{
					// 	Layout: declarative.VBox{},
					// 	Children: []declarative.Widget{
					// 		declarative.TreeView{
					// 			AssignTo:  &bakeState.OutputTreeView,
					// 			Model:     domain.NewOutputModel(),
					// 			MinSize:   declarative.Size{Width: 230, Height: 200},
					// 			Checkable: true,
					// 			OnItemChecked: func(item walk.TreeItem) {
					// 				// 無限ループを防ぐためのフラグチェック
					// 				treeModel := bakeState.OutputTreeView.Model()
					// 				if treeModel == nil || item == nil || bakeState.IsOutputUpdatingChildren {
					// 					return
					// 				}

					// 				checked := bakeState.OutputTreeView.Checked(item)

					// 				// 子どもアイテムも同じチェック状態に設定
					// 				bakeState.SetOutputChildrenChecked(item, checked)
					// 			},
					// 		},
					// 	},
					// },
					bakeState.SaveMotionButton.Widgets(),
				},
			},
			bakeState.Player.Widgets(),
		},
	}
}
