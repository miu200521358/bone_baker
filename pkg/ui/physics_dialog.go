package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
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
		Title:         mi18n.T("物理設定"),
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
			AssignTo:           gravityEdit,
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
			AssignTo:           maxSubStepsEdit,
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
			AssignTo:           fixedTimeStepEdit,
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
		declarative.Label{
			Text:        mi18n.T("設定最大開始フレーム"),
			ToolTipText: mi18n.T("設定最大開始フレーム説明"),
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxStartFrame"),
			ToolTipText:        mi18n.T("設定最大開始フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.bakeState.CurrentSet().MaxFrame() + 1),
		},
		declarative.Label{
			Text:        mi18n.T("設定最大終了フレーム"),
			ToolTipText: mi18n.T("設定最大終了フレーム説明"),
		},
		declarative.NumberEdit{
			Value:              declarative.Bind("MaxEndFrame"),
			ToolTipText:        mi18n.T("設定最大終了フレーム説明"),
			SpinButtonsVisible: true,
			Decimals:           0,
			Increment:          1,
			MinValue:           0,
			MaxValue:           float64(p.bakeState.CurrentSet().MaxFrame() + 1),
		},
		declarative.VSeparator{
			ColumnSpan: 2,
		},
	}

	// 物理編集ウィジェットを追加
	widgets = append(widgets, p.createPhysicsEditWidgets(sizeXEdit, sizeYEdit, sizeZEdit, massEdit, stiffnessEdit, tensionEdit, treeView)...)

	// 物理アイテムツリー
	widgets = append(widgets, declarative.TreeView{
		AssignTo:   treeView,
		Model:      p.bakeState.CurrentSet().PhysicsTableModel.Records[p.bakeState.PhysicsTableView.CurrentIndex()].TreeModel,
		MinSize:    declarative.Size{Width: 230, Height: 200},
		ColumnSpan: 6,
		OnCurrentItemChanged: func() {
			p.updateEditValues(*treeView, *sizeXEdit, *sizeYEdit, *sizeZEdit, *massEdit, *stiffnessEdit, *tensionEdit)
		},
	})

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
			Text:        mi18n.T("大きさY倍率"),
			ToolTipText: mi18n.T("大きさY倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("大きさY倍率説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: sizeYEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *domain.PhysicsItem) {
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
			MinSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: sizeZEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *domain.PhysicsItem) {
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
		declarative.TextLabel{
			Text:        mi18n.T("硬さ倍率"),
			ToolTipText: mi18n.T("硬さ倍率説明"),
			OnMouseDown: func(x, y int, button walk.MouseButton) {
				mlog.IL("%s", mi18n.T("硬さ倍率説明"))
			},
			MinSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: stiffnessEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *domain.PhysicsItem) {
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
			MinSize: declarative.Size{Width: 100, Height: 20},
		},
		declarative.NumberEdit{
			AssignTo: tensionEdit,
			OnValueChanged: func() {
				p.updateItemProperty(*treeView, func(item *domain.PhysicsItem) {
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
	}
}

// createButtonWidgets ボタンウィジェットを作成
func (p *PhysicsTableViewDialog) createButtonWidgets(okBtn, cancelBtn **walk.PushButton, dlg **walk.Dialog, db **walk.DataBinder) []declarative.Widget {
	return []declarative.Widget{
		declarative.PushButton{
			AssignTo:    okBtn,
			Text:        mi18n.T("登録"),
			ToolTipText: mi18n.T("物理設定登録説明"),
			OnClicked: func() {
				if err := (*db).Submit(); err != nil {
					mlog.ET(mi18n.T("焼き込み設定変更エラー"), err, "")
					return
				}
				(*dlg).Accept()
			},
		},
		declarative.PushButton{
			AssignTo:    cancelBtn,
			Text:        mi18n.T("削除"),
			ToolTipText: mi18n.T("物理設定削除説明"),
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
			AssignTo:    cancelBtn,
			Text:        mi18n.T("キャンセル"),
			ToolTipText: mi18n.T("物理設定キャンセル説明"),
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
	sizeXEdit.ChangeValue(currentItem.SizeRatio.X)
	sizeYEdit.ChangeValue(currentItem.SizeRatio.Y)
	sizeZEdit.ChangeValue(currentItem.SizeRatio.Z)
	massEdit.ChangeValue(currentItem.MassRatio)
	stiffnessEdit.ChangeValue(currentItem.StiffnessRatio)
	tensionEdit.ChangeValue(currentItem.TensionRatio)
}

// handleDialogOK ダイアログOK処理
func (p *PhysicsTableViewDialog) handleDialogOK() {
	p.bakeState.SetWidgetEnabled(false)

	p.bakeState.CurrentSet().PhysicsTableModel.AddRecord(
		p.bakeState.CurrentSet().OriginalModel,
		0,
		p.bakeState.CurrentSet().MaxFrame())
	p.bakeState.PhysicsTableView.SetModel(p.bakeState.CurrentSet().PhysicsTableModel)
	p.bakeState.PhysicsTableView.SetCurrentIndex(len(p.bakeState.CurrentSet().PhysicsTableModel.Records) - 1)

	physicsWorldMotion := p.mWidgets.Window().LoadPhysicsWorldMotion(0)
	physicsModelMotion := p.mWidgets.Window().LoadPhysicsModelMotion(0, p.bakeState.CurrentIndex())

	for _, record := range p.bakeState.PhysicsTableView.Model().(*domain.PhysicsTableModel).Records {
		for f := record.StartFrame; f <= record.EndFrame; f++ {
			physicsWorldMotion.AppendGravityFrame(vmd.NewGravityFrameByValue(f, &mmath.MVec3{
				X: 0,
				Y: float64(record.Gravity),
				Z: 0,
			}))
			physicsWorldMotion.AppendMaxSubStepsFrame(vmd.NewMaxSubStepsFrameByValue(f, record.MaxSubSteps))
			physicsWorldMotion.AppendFixedTimeStepFrame(vmd.NewFixedTimeStepFrameByValue(f, record.FixedTimeStep))

			if f == record.StartFrame {
				if record.IsStartDeform {
					// 開始時用整形をON
					physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_START_FIT_FRAME))
				} else {
					// 前フレームから継続して物理演算を行う
					physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
				}
			} else {
				// 開始と終了以外はリセットしない
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_NONE))
			}

			// 剛体・ジョイントパラは台形の線形補間で変形させる
			frameRatio := float32(0.0)
			if f < record.MaxStartFrame && f > record.StartFrame &&
				record.MaxStartFrame > record.StartFrame {
				// StartFrame から MaxStartFrame の間：0倍から指定倍率まで線形補間
				frameRatio = (f - record.StartFrame) / (record.MaxStartFrame - record.StartFrame)
				// 変動中はリセッし続ける
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			} else if f > record.MaxEndFrame && f < record.EndFrame &&
				record.MaxEndFrame < record.EndFrame {
				// MaxEndFrame から EndFrame の間：指定倍率から0倍まで線形補間
				frameRatio = (record.EndFrame - f) / (record.EndFrame - record.MaxEndFrame)
				// 変動中はリセッし続ける
				physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(f, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
			} else if f >= record.MaxStartFrame && f <= record.MaxEndFrame {
				// MAXの間はそのまま最大倍率
				frameRatio = 1.0
			} else {
				// StartFrame以前とEndFrame以後は元の値（倍率なし）
				frameRatio = 0.0
			}
			frameRatio64 := float64(frameRatio)

			// 剛体
			model := p.bakeState.CurrentSet().OriginalModel
			model.RigidBodies.ForEach(func(rigidIndex int, rb *pmx.RigidBody) bool {
				rigidBodyItem := record.TreeModel.AtByRigidBodyIndex(rb.Index())

				if rigidBodyItem == nil || !rigidBodyItem.(*domain.PhysicsItem).Modified {
					physicsModelMotion.AppendRigidBodyFrame(rb.Name(),
						vmd.NewRigidBodyFrameByValues(
							f,
							rb.Size,
							rb.RigidBodyParam.Mass,
						))

					return true
				}

				// 質量の計算：元の質量 + (元の質量 * (massRatio - 1.0) * frameRatio)
				sizeRatio := rigidBodyItem.(*domain.PhysicsItem).SizeRatio
				calculatedSize := rb.Size.Added(rb.Size.Muled(sizeRatio.SubedScalar(1.0).MuledScalar(frameRatio64)))

				massRatio := rigidBodyItem.(*domain.PhysicsItem).MassRatio
				calculatedMass := rb.RigidBodyParam.Mass + (rb.RigidBodyParam.Mass * (massRatio - 1.0) * frameRatio64)

				physicsModelMotion.AppendRigidBodyFrame(rb.Name(),
					vmd.NewRigidBodyFrameByValues(
						f,
						calculatedSize,
						calculatedMass,
					))

				return true
			})

			// ジョイント
			model.Joints.ForEach(func(jointIndex int, joint *pmx.Joint) bool {
				rigidBodyItemA := record.TreeModel.AtByRigidBodyIndex(joint.RigidbodyIndexA)
				rigidBodyItemB := record.TreeModel.AtByRigidBodyIndex(joint.RigidbodyIndexB)

				if rigidBodyItemA == nil && rigidBodyItemB == nil {
					// ジョイントの両端が未設定の場合はスキップ
					physicsModelMotion.AppendJointFrame(joint.Name(),
						vmd.NewJointFrameByValues(
							f,
							joint.JointParam.TranslationLimitMin,
							joint.JointParam.TranslationLimitMax,
							joint.JointParam.RotationLimitMin,
							joint.JointParam.RotationLimitMax,
							joint.JointParam.SpringConstantTranslation,
							joint.JointParam.SpringConstantRotation,
						))

					return true
				}

				if ((rigidBodyItemA != nil && !rigidBodyItemA.(*domain.PhysicsItem).Modified) || rigidBodyItemA == nil) &&
					((rigidBodyItemB != nil && !rigidBodyItemB.(*domain.PhysicsItem).Modified) || rigidBodyItemB == nil) {
					// 両方の剛体が未変更の場合はスキップ
					physicsModelMotion.AppendJointFrame(joint.Name(),
						vmd.NewJointFrameByValues(
							f,
							joint.JointParam.TranslationLimitMin,
							joint.JointParam.TranslationLimitMax,
							joint.JointParam.RotationLimitMin,
							joint.JointParam.RotationLimitMax,
							joint.JointParam.SpringConstantTranslation,
							joint.JointParam.SpringConstantRotation,
						))

					return true
				}

				// ジョイントのパラメータを台形の線形補間で変形させる
				var stiffnessRatioA, stiffnessRatioB float64
				var tensionRatioA, tensionRatioB float64
				if rigidBodyItemA != nil && rigidBodyItemA.(*domain.PhysicsItem).Modified {
					stiffnessRatioA = rigidBodyItemA.(*domain.PhysicsItem).StiffnessRatio
					tensionRatioA = rigidBodyItemA.(*domain.PhysicsItem).TensionRatio
				} else {
					stiffnessRatioA = 1.0
					tensionRatioA = 1.0
				}
				if rigidBodyItemB != nil && rigidBodyItemB.(*domain.PhysicsItem).Modified {
					stiffnessRatioB = rigidBodyItemB.(*domain.PhysicsItem).StiffnessRatio
					tensionRatioB = rigidBodyItemB.(*domain.PhysicsItem).TensionRatio
				} else {
					stiffnessRatioB = 1.0
					tensionRatioB = 1.0
				}

				// 両剛体の平均倍率を計算
				avgStiffnessRatio := mmath.Mean([]float64{stiffnessRatioA, stiffnessRatioB})
				avgTensionRatio := mmath.Mean([]float64{tensionRatioA, tensionRatioB})

				// 台形状の変化を適用
				calculatedRotationLimitMin := joint.JointParam.RotationLimitMin.Added(
					joint.JointParam.RotationLimitMin.DivedScalar((avgStiffnessRatio - 1.0) * frameRatio64))
				calculatedRotationLimitMax := joint.JointParam.RotationLimitMax.Added(
					joint.JointParam.RotationLimitMax.DivedScalar((avgStiffnessRatio - 1.0) * frameRatio64))
				calculatedSpringConstantRotation := joint.JointParam.SpringConstantRotation.Added(
					joint.JointParam.SpringConstantRotation.MuledScalar((avgTensionRatio - 1.0) * frameRatio64))

				physicsModelMotion.AppendJointFrame(joint.Name(),
					vmd.NewJointFrameByValues(
						f,
						joint.JointParam.TranslationLimitMin,
						joint.JointParam.TranslationLimitMax,
						calculatedRotationLimitMin,
						calculatedRotationLimitMax,
						joint.JointParam.SpringConstantTranslation,
						calculatedSpringConstantRotation,
					))

				return true
			})
		}

		// 最初フレームの前には物理リセットしない（次キーフレを呼んでしまうので）
		if record.StartFrame > 0 {
			physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.StartFrame-1, vmd.PHYSICS_RESET_TYPE_NONE))
		}
		// 最後のフレームの後に物理リセットする
		physicsWorldMotion.AppendPhysicsResetFrame(vmd.NewPhysicsResetFrameByValue(record.EndFrame+1, vmd.PHYSICS_RESET_TYPE_CONTINUE_FRAME))
	}
	p.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	p.mWidgets.Window().StorePhysicsModelMotion(0, p.bakeState.CurrentIndex(), physicsModelMotion)
	p.mWidgets.Window().TriggerPhysicsReset()

	p.bakeState.SetWidgetEnabled(true)
	controller.Beep()

	// 更新
	p.bakeState.PhysicsTableView.SetModel(p.bakeState.CurrentSet().PhysicsTableModel)

}
