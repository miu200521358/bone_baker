package ui

import (
	"fmt"
	"path/filepath"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mconfig"
	"github.com/miu200521358/mlib_go/pkg/config/merr"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

const (
	jsonPathKey = "json_path"
)

func (s *WidgetStore) createPlayerWidget() {
	s.Player = widget.NewMotionPlayer()
	s.Player.SetLabelTexts(mi18n.T("焼き込み停止"), mi18n.T("焼き込み再生"))
	s.Player.SetOnEnabledInPlaying(s.createEnabledPlaying())
	s.Player.SetOnChangePlayingPre(s.createOnChangePlayingPre())
	s.Player.SetStartPlayingResetType(func() vmd.PhysicsResetType {
		return vmd.PHYSICS_RESET_TYPE_START_FIT_FRAME
	})
}

func (s *WidgetStore) createEnabledPlaying() func(playing bool) {
	return func(playing bool) {
		s.setWidgetEnabled(!playing)
		if playing {
			// 再生中も操作可ウィジェットを有効化
			s.setWidgetPlayingEnabled(true)
		}
	}
}

// setWidgetEnabled 物理焼き込み有効無効設定
func (s *WidgetStore) setWidgetEnabled(enabled bool) {
	s.AddSetButton.SetEnabled(enabled)
	s.ResetSetButton.SetEnabled(enabled)
	s.SaveSetButton.SetEnabled(enabled)
	s.LoadSetButton.SetEnabled(enabled)

	s.OriginalMotionPicker.SetEnabled(enabled)
	s.OriginalModelPicker.SetEnabled(enabled)
	s.OutputMotionPicker.SetEnabled(enabled)
	s.OutputModelPicker.SetEnabled(enabled)

	s.AddOutputButton.SetEnabled(enabled)
	s.OutputTableView.SetEnabled(enabled)

	s.BakedHistoryIndexEdit.SetEnabled(enabled)
	s.BakeHistoryClearButton.SetEnabled(enabled)

	s.SaveModelButton.SetEnabled(enabled)
	s.SaveMotionButton.SetEnabled(enabled)

	s.setWidgetPlayingEnabled(enabled)
}

func (s *WidgetStore) setWidgetPlayingEnabled(enabled bool) {
	s.Player.SetEnabled(enabled)

	s.AddPhysicsButton.SetEnabled(enabled)
	s.AddRigidBodyButton.SetEnabled(enabled)

	s.PhysicsTableView.SetEnabled(enabled)
	s.RigidBodyTableWidget.SetEnabled(enabled)
}

func (s *WidgetStore) createOnChangePlayingPre() func(playing bool) {
	return func(playing bool) {
		s.Window().SetSaveDelta(0, playing)

		// 情報表示
		s.Window().SetCheckedShowInfoEnabled(playing)
		// フレームドロップOFF
		s.Window().SetFrameDropEnabled(false)

		if playing {
			// 焼き込み開始時にINDEX加算
			deltaIndex := s.Window().GetDeltaMotionCount(0, s.CurrentIndex)
			if deltaIndex > 0 {
				// 既に焼き込みが1回以上行われている場合はインクリメント
				deltaIndex += 1
			}
			s.Window().SetSaveDeltaIndex(0, deltaIndex+1)

			// 再生フレーム
			mlog.IL(mi18n.T("焼き込み再生開始: 焼き込み履歴INDEX[%d]"), deltaIndex+1)
		} else {
			deltaIndex := s.Window().GetDeltaMotionCount(0, s.CurrentIndex)
			s.BakedHistoryIndexEdit.SetRange(1.0, float64(deltaIndex))
			s.BakedHistoryIndexEdit.SetValue(float64(deltaIndex))

			// 焼き込み完了時に出力モーションを取得
			s.createHistoryIndexChangeHandler()()
		}
	}
}

// createFilePickerWidgets ファイルピッカーウィジェット群を作成
func (s *WidgetStore) createFilePickerWidgets() {
	s.OriginalModelPicker = s.createOriginalModelFilePicker()
	s.OriginalMotionPicker = s.createOriginalMotionFilePicker()
	s.OutputModelPicker = s.createOutputModelFilePicker()
	s.OutputMotionPicker = s.createOutputMotionFilePicker()
}

// createButtonWidgets ボタンウィジェット群を作成
func (s *WidgetStore) createButtonWidgets() {
	s.AddSetButton = s.createAddSetButton()
	s.ResetSetButton = s.createResetSetButton()
	s.LoadSetButton = s.createLoadSetButton()
	s.SaveSetButton = s.createSaveSetButton()
	s.SaveModelButton = s.createSaveModelButton()
	s.SaveMotionButton = s.createSaveMotionButton()
	s.AddPhysicsButton = s.createAddPhysicsButton()
	s.AddRigidBodyButton = s.createAddRigidBodyButton()
	s.AddOutputButton = s.createAddOutputButton()
	s.BakeHistoryClearButton = s.createBakeHistoryClearButton()
}

func (s *WidgetStore) createOutputMotionFilePicker() *widget.FilePicker {
	return widget.NewVmdSaveFilePicker(
		mi18n.T("焼き込み後モーション(Vmd)"),
		mi18n.T("焼き込み後モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
		},
	)
}

func (s *WidgetStore) createOutputModelFilePicker() *widget.FilePicker {
	return widget.NewPmxSaveFilePicker(
		mi18n.T("変更後モデル(Pmx)"),
		mi18n.T("変更後モデル説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			// 実際に保存するのは、物理有効な元モデル
			model := s.currentSet().OriginalModel
			if model == nil {
				return
			}

			if err := rep.Save(path, model, false); err != nil {
				mlog.ET(mi18n.T("保存失敗"), err, "")
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					s.setWidgetEnabled(true)
				}
			}
		},
	)
}

func (s *WidgetStore) createOriginalMotionFilePicker() *widget.FilePicker {
	return widget.NewVmdLoadFilePicker(
		"vmd",
		mi18n.T("モーション(Vmd)"),
		mi18n.T("モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := s.loadMotion(cw, path); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					s.setWidgetEnabled(true)
				}
			}
		},
	)
}

func (s *WidgetStore) createOriginalModelFilePicker() *widget.FilePicker {
	return widget.NewPmxLoadFilePicker(
		"pmx",
		mi18n.T("モデル(Pmx)"),
		mi18n.T("モデル説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := s.loadModel(cw, path); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					s.setWidgetEnabled(true)
				}
			}
		},
	)
}

func (s *WidgetStore) createAddSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定追加"))
	btn.SetTooltip(mi18n.T("設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		s.BakeSets = append(s.BakeSets, entity.NewBakeSet(len(s.BakeSets)))
		s.AddAction()
	})
	return btn
}

func (s *WidgetStore) createResetSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定全削除"))
	btn.SetTooltip(mi18n.T("設定全削除説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		for n := range 2 {
			for m := range s.NavToolBar.Actions().Len() {
				s.Window().StoreModel(n, m, nil)
				s.Window().StoreMotion(n, m, nil)
			}
		}
		// s.ResetSet()
	})
	return btn
}

func (s *WidgetStore) createLoadSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定設定読込"))
	btn.SetTooltip(mi18n.T("設定設定読込説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		choices := mconfig.LoadUserConfig(jsonPathKey)
		var initialDirPath string
		if len(choices) > 0 {
			initialDirPath = filepath.Dir(choices[0])
		}

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
			s.loadBakeSets(dlg.FilePath)
			mconfig.SaveUserConfig(jsonPathKey, dlg.FilePath, 1)
		}
	})
	return btn
}

func (s *WidgetStore) createSaveSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定設定保存"))
	btn.SetTooltip(mi18n.T("設定設定保存説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		initialDirPath := filepath.Dir(s.currentSet().OriginalMotionPath)
		filePath := ""
		if s.currentSet().OriginalModel != nil {
			// モーション側にモデルファイル名でJSONデフォルト名を入れる
			_, name, _ := mfile.SplitPath(s.currentSet().OriginalModel.Path())
			filePath = fmt.Sprintf("BoneBaker_%s.json", name)
		}

		dlg := walk.FileDialog{
			Title: mi18n.T(
				"ファイル選択ダイアログタイトル",
				map[string]any{"Title": "Json"}),
			Filter:         "Json files (*.json)|*.json",
			FilterIndex:    1,
			InitialDirPath: initialDirPath,
			FilePath:       filePath,
		}
		if ok, err := dlg.ShowSave(nil); err != nil {
			walk.MsgBox(nil, mi18n.T("ファイル選択ダイアログ選択エラー"), err.Error(), walk.MsgBoxIconError)
		} else if ok {
			s.saveBakeSets(dlg.FilePath)
			mconfig.SaveUserConfig(jsonPathKey, dlg.FilePath, 1)
		}
	})
	return btn
}

func (s *WidgetStore) createSaveModelButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モデル保存"))
	btn.SetTooltip(mi18n.T("モデル保存説明"))
	btn.SetMinSize(declarative.Size{Width: 256, Height: 20})
	btn.SetStretchFactor(20)
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		s.setWidgetEnabled(false)

		for _, physicsSet := range s.BakeSets {
			if physicsSet.OutputModelPath != "" && physicsSet.OriginalModel != nil {
				rep := repository.NewPmxRepository(true)
				if err := rep.Save(physicsSet.OutputModelPath, physicsSet.OriginalModel, false); err != nil {
					mlog.ET(mi18n.T("モデル保存失敗"), err, "")
					if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
						s.setWidgetEnabled(true)
						return
					}
				}
			}
		}

		s.setWidgetEnabled(true)
		controller.Beep()
	})
	return btn
}

func (s *WidgetStore) createSaveMotionButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モーション保存"))
	btn.SetTooltip(mi18n.T("モーション保存説明"))
	btn.SetMinSize(declarative.Size{Width: 256, Height: 20})
	btn.SetStretchFactor(20)
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		s.setWidgetEnabled(false)

		for _, bakeSet := range s.BakeSets {
			if bakeSet.OutputMotionPath != "" && bakeSet.OutputMotion != nil {
				motions, err := s.outputUsecase.ProcessOutputMotions(
					bakeSet.OriginalModel,
					bakeSet.OriginalMotion,
					bakeSet.OutputMotion,
					bakeSet.OutputMotionPath,
					bakeSet.OutputRecords,
				)

				if err != nil {
					mlog.ET(mi18n.T("モーション保存失敗"), err, "")
					return
				}

				for _, motion := range motions {
					rep := repository.NewVmdRepository(true)
					mlog.IL(fmt.Sprintf(mi18n.T("モーション保存開始: [%.0f-%.0f]"), motion.MinFrame(), motion.MaxFrame()))
					if err := rep.Save("", motion, false); err != nil {
						mlog.ET(fmt.Sprintf(mi18n.T("モーション保存失敗"), motion.Path()), err, "")
						if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
							s.setWidgetEnabled(true)
							return
						}
					}
				}
			}
		}

		s.setWidgetEnabled(true)
		controller.Beep()
	})
	return btn
}

func (s *WidgetStore) createAddPhysicsButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("ワールド物理設定追加"))
	btn.SetTooltip(mi18n.T("ワールド物理設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 150, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		createPhysicsTableViewDialog(s, true)() // ダイアログを表示
	})
	return btn
}

func (s *WidgetStore) createAddRigidBodyButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モデル物理設定追加"))
	btn.SetTooltip(mi18n.T("モデル物理設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 150, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		createRigidBodyTableViewDialog(s, true)() // ダイアログを表示
	})
	return btn
}

func (s *WidgetStore) createAddOutputButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("出力設定追加"))
	btn.SetTooltip(mi18n.T("出力設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 150, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		createOutputTableViewDialog(s, true)() // ダイアログを表示
	})
	return btn
}

// createBakedHistoryWidgets 焼き込み履歴ウィジェットを作成
func (s *WidgetStore) createBakedHistoryWidgets() []declarative.Widget {
	return []declarative.Widget{
		declarative.TextLabel{
			Text:        mi18n.T("焼き込み履歴INDEX"),
			ToolTipText: mi18n.T("焼き込み履歴INDEX説明"),
		},
		declarative.NumberEdit{
			SpinButtonsVisible: true,
			AssignTo:           &s.BakedHistoryIndexEdit,
			Decimals:           0,
			Increment:          1,
			MinValue:           1,
			MaxValue:           2,
			OnValueChanged:     s.createHistoryIndexChangeHandler(),
		},
		s.BakeHistoryClearButton.Widgets(),
		declarative.HSpacer{
			ColumnSpan: 1,
		},
	}
}

func (s *WidgetStore) createHistoryIndexChangeHandler() func() {
	return func() {
		// 出力モーションインデックスが変更されたときの処理
		currentSet := s.currentSet()
		deltaIndex := int(s.BakedHistoryIndexEdit.Value() - 1)
		if deltaIndex < 0 ||
			deltaIndex >= s.mWidgets.Window().GetDeltaMotionCount(0, currentSet.Index) {
			deltaIndex = 0
		}

		// 物理ありのモーションを取得
		outputMotion := s.mWidgets.Window().LoadDeltaMotion(0, currentSet.Index, deltaIndex)
		// 物理確認用として設定
		s.mWidgets.Window().StoreMotion(1, currentSet.Index, outputMotion)
		s.mWidgets.Window().TriggerPhysicsReset()

		// 出力モーションを更新
		currentSet.OutputMotion = outputMotion
		currentSet.OutputMotionPath = currentSet.CreateOutputMotionPath()
		s.OutputMotionPicker.ChangePath(currentSet.OutputMotionPath)
	}
}

func (s *WidgetStore) createBakeHistoryClearButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("焼き込み履歴クリア"))
	btn.SetTooltip(mi18n.T("焼き込み履歴クリア説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		s.Window().ClearDeltaMotion(0, s.CurrentIndex)
		s.Window().ClearDeltaMotion(1, s.CurrentIndex)
		s.Window().SetSaveDeltaIndex(0, 0)
		s.Window().SetSaveDeltaIndex(1, 0)

		s.BakedHistoryIndexEdit.SetValue(1.0)
		s.BakedHistoryIndexEdit.SetRange(1.0, 2.0)
	})
	return btn
}
