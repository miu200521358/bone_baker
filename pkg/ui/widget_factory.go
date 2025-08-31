package ui

import (
	"fmt"
	"path/filepath"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
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

// WidgetFactory UIウィジェット作成のファクトリー
type WidgetFactory struct {
	bakeState      *BakeState
	mWidgets       *controller.MWidgets
	physicsUsecase *usecase.PhysicsUsecase
}

// NewWidgetFactory コンストラクタ
func NewWidgetFactory(bakeState *BakeState, mWidgets *controller.MWidgets) *WidgetFactory {
	return &WidgetFactory{
		bakeState:      bakeState,
		mWidgets:       mWidgets,
		physicsUsecase: usecase.NewPhysicsUsecase(),
	}
}

// CreateFilePickerWidgets ファイルピッカーウィジェット群を作成
func (wf *WidgetFactory) CreateFilePickerWidgets() {
	wf.bakeState.OutputMotionPicker = wf.createOutputMotionFilePicker()
	wf.bakeState.OutputModelPicker = wf.createOutputModelFilePicker()
	wf.bakeState.OriginalMotionPicker = wf.createOriginalMotionFilePicker()
	wf.bakeState.OriginalModelPicker = wf.createOriginalModelFilePicker()
}

// CreateButtonWidgets ボタンウィジェット群を作成
func (wf *WidgetFactory) CreateButtonWidgets() {
	wf.bakeState.AddSetButton = wf.createAddSetButton()
	wf.bakeState.ResetSetButton = wf.createResetSetButton()
	wf.bakeState.LoadSetButton = wf.createLoadSetButton()
	wf.bakeState.SaveSetButton = wf.createSaveSetButton()
	wf.bakeState.SaveModelButton = wf.createSaveModelButton()
	wf.bakeState.SaveMotionButton = wf.createSaveMotionButton()
	wf.bakeState.AddPhysicsButton = wf.createAddPhysicsButton()
	wf.bakeState.AddOutputButton = wf.createAddOutputButton()
	wf.bakeState.BakeHistoryClearButton = wf.createBakeHistoryClearButton()
}

// CreatePlayerWidget プレイヤーウィジェットを作成
func (wf *WidgetFactory) CreatePlayerWidget() {
	wf.bakeState.Player = widget.NewMotionPlayer()
	wf.bakeState.Player.SetLabelTexts(mi18n.T("焼き込み停止"), mi18n.T("焼き込み再生"))
	wf.bakeState.Player.SetOnEnabledInPlaying(wf.createEnabledPlaying())
	wf.bakeState.Player.SetOnChangePlayingPre(wf.createOnChangePlayingPre())
	wf.bakeState.Player.SetStartPlayingResetType(func() vmd.PhysicsResetType {
		return vmd.PHYSICS_RESET_TYPE_START_FIT_FRAME
	})
}

// CreateTableViews テーブルビューを作成
func (wf *WidgetFactory) CreatePhysicsTableView() declarative.TableView {
	return declarative.TableView{
		AssignTo:         &wf.bakeState.PhysicsTableView,
		Model:            domain.NewPhysicsTableModel(),
		AlternatingRowBG: true,
		MinSize:          declarative.Size{Width: 230, Height: 150},
		Columns: []declarative.TableViewColumn{
			{Title: "#", Width: 30},
			{Title: mi18n.T("開始F"), Width: 60},
			{Title: mi18n.T("最大開始F"), Width: 60},
			{Title: mi18n.T("最大終了F"), Width: 60},
			{Title: mi18n.T("終了F"), Width: 60},
			{Title: mi18n.T("重力"), Width: 60},
			{Title: mi18n.T("最大演算回数"), Width: 100},
			{Title: mi18n.T("物理演算頻度"), Width: 100},
			// {Title: mi18n.T("開始時用整形"), Width: 100},
			{Title: mi18n.T("調整剛体"), Width: 300},
		},
		OnItemClicked: wf.createPhysicsTableViewDialog(false),
	}
}

func (wf *WidgetFactory) CreateOutputTableView() declarative.TableView {
	return declarative.TableView{
		AssignTo:         &wf.bakeState.OutputTableView,
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
		OnItemClicked: wf.createOutputTableViewDialog(),
	}
}

// CreateBakedHistoryWidgets 焼き込み履歴ウィジェットを作成
func (wf *WidgetFactory) CreateBakedHistoryWidgets() []declarative.Widget {
	return []declarative.Widget{
		declarative.TextLabel{
			Text:        mi18n.T("焼き込み履歴INDEX"),
			ToolTipText: mi18n.T("焼き込み履歴INDEX説明"),
		},
		declarative.NumberEdit{
			SpinButtonsVisible: true,
			AssignTo:           &wf.bakeState.BakedHistoryIndexEdit,
			Decimals:           0,
			Increment:          1,
			MinValue:           1,
			MaxValue:           2,
			OnValueChanged:     wf.createHistoryIndexChangeHandler(),
		},
		wf.bakeState.BakeHistoryClearButton.Widgets(),
		declarative.HSpacer{
			ColumnSpan: 1,
		},
	}
}

// Private methods for widget creation

func (wf *WidgetFactory) createOutputMotionFilePicker() *widget.FilePicker {
	return widget.NewVmdSaveFilePicker(
		mi18n.T("焼き込み後モーション(Vmd)"),
		mi18n.T("焼き込み後モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
		},
	)
}

func (wf *WidgetFactory) createOutputModelFilePicker() *widget.FilePicker {
	return widget.NewPmxSaveFilePicker(
		mi18n.T("変更後モデル(Pmx)"),
		mi18n.T("変更後モデル説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			// 実際に保存するのは、物理有効な元モデル
			model := wf.bakeState.CurrentSet().OriginalModel
			if model == nil {
				return
			}

			if err := rep.Save(path, model, false); err != nil {
				mlog.ET(mi18n.T("保存失敗"), err, "")
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					wf.bakeState.SetWidgetEnabled(true)
				}
			}
		},
	)
}

func (wf *WidgetFactory) createOriginalMotionFilePicker() *widget.FilePicker {
	return widget.NewVmdLoadFilePicker(
		"vmd",
		mi18n.T("モーション(Vmd)"),
		mi18n.T("モーション説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := wf.bakeState.LoadMotion(cw, path, true); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					wf.bakeState.SetWidgetEnabled(true)
				}
			}
		},
	)
}

func (wf *WidgetFactory) createOriginalModelFilePicker() *widget.FilePicker {
	return widget.NewPmxLoadFilePicker(
		"pmx",
		mi18n.T("モデル(Pmx)"),
		mi18n.T("モデル説明"),
		func(cw *controller.ControlWindow, rep repository.IRepository, path string) {
			if err := wf.bakeState.LoadModel(cw, path); err != nil {
				if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
					wf.bakeState.SetWidgetEnabled(true)
				}
			}
		},
	)
}

func (wf *WidgetFactory) createAddSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定追加"))
	btn.SetTooltip(mi18n.T("設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		wf.bakeState.BakeSets = append(wf.bakeState.BakeSets,
			domain.NewPhysicsSet(len(wf.bakeState.BakeSets)))
		wf.bakeState.AddAction()
	})
	return btn
}

func (wf *WidgetFactory) createResetSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定全削除"))
	btn.SetTooltip(mi18n.T("設定全削除説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		for n := range 2 {
			for m := range wf.bakeState.NavToolBar.Actions().Len() {
				wf.mWidgets.Window().StoreModel(n, m, nil)
				wf.mWidgets.Window().StoreMotion(n, m, nil)
			}
		}
		wf.bakeState.ResetSet()
	})
	return btn
}

func (wf *WidgetFactory) createLoadSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定設定読込"))
	btn.SetTooltip(mi18n.T("設定設定読込説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		choices := mconfig.LoadUserConfig("physics_set_path")
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
			wf.handleLoadSet(dlg.FilePath)
		}
	})
	return btn
}

func (wf *WidgetFactory) createSaveSetButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("設定設定保存"))
	btn.SetTooltip(mi18n.T("設定設定保存説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		initialDirPath := filepath.Dir(wf.bakeState.CurrentSet().OriginalMotionPath)
		filePath := ""
		if wf.bakeState.CurrentSet().OriginalModel != nil {
			// モーション側にモデルファイル名でJSONデフォルト名を入れる
			_, name, _ := mfile.SplitPath(wf.bakeState.CurrentSet().OriginalModel.Path())
			filePath = fmt.Sprintf("%s.json", name)
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
			wf.bakeState.SaveSet(dlg.FilePath)
			mconfig.SaveUserConfig("physics_set_path", dlg.FilePath, 1)
		}
	})
	return btn
}

func (wf *WidgetFactory) createSaveModelButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モデル保存"))
	btn.SetTooltip(mi18n.T("モデル保存説明"))
	btn.SetMinSize(declarative.Size{Width: 256, Height: 20})
	btn.SetStretchFactor(20)
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		wf.bakeState.SetWidgetEnabled(false)

		for _, physicsSet := range wf.bakeState.BakeSets {
			if physicsSet.OutputModelPath != "" && physicsSet.OriginalModel != nil {
				rep := repository.NewPmxRepository(true)
				if err := rep.Save(physicsSet.OutputModelPath, physicsSet.OriginalModel, false); err != nil {
					mlog.ET(mi18n.T("モデル保存失敗"), err, "")
					if ok := merr.ShowErrorDialog(cw.AppConfig(), err); ok {
						wf.bakeState.SetWidgetEnabled(true)
					}
				}
			}
		}

		wf.bakeState.SetWidgetEnabled(true)
		controller.Beep()
	})
	return btn
}

func (wf *WidgetFactory) createSaveMotionButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("モーション保存"))
	btn.SetTooltip(mi18n.T("モーション保存説明"))
	btn.SetMinSize(declarative.Size{Width: 256, Height: 20})
	btn.SetStretchFactor(20)
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		wf.bakeState.SetWidgetEnabled(false)

		for _, physicsSet := range wf.bakeState.BakeSets {
			if physicsSet.OutputMotionPath != "" && physicsSet.OutputMotion != nil {
				// チェックボーンのみ残す
				motions, err := physicsSet.GetOutputMotionOnlyChecked(
					wf.bakeState.OutputTableView.Model().(*domain.OutputTableModel).Records,
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
							wf.bakeState.SetWidgetEnabled(true)
						}
					}
				}
			}
		}

		wf.bakeState.SetWidgetEnabled(true)
		controller.Beep()
	})
	return btn
}

func (wf *WidgetFactory) createAddPhysicsButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("物理設定追加"))
	btn.SetTooltip(mi18n.T("物理設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		wf.createPhysicsTableViewDialog(true)() // ダイアログを表示
	})
	return btn
}

func (wf *WidgetFactory) createAddOutputButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("出力設定追加"))
	btn.SetTooltip(mi18n.T("出力設定追加説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		wf.bakeState.CurrentSet().OutputTableModel.AddRecord(
			wf.bakeState.CurrentSet().OriginalModel,
			0,
			wf.bakeState.CurrentSet().MaxFrame())
		wf.bakeState.OutputTableView.SetModel(wf.bakeState.CurrentSet().OutputTableModel)
		wf.bakeState.OutputTableView.SetCurrentIndex(len(wf.bakeState.CurrentSet().OutputTableModel.Records) - 1)
		wf.createOutputTableViewDialog()() // ダイアログを表示
	})
	return btn
}

func (wf *WidgetFactory) createBakeHistoryClearButton() *widget.MPushButton {
	btn := widget.NewMPushButton()
	btn.SetLabel(mi18n.T("焼き込み履歴クリア"))
	btn.SetTooltip(mi18n.T("焼き込み履歴クリア説明"))
	btn.SetMaxSize(declarative.Size{Width: 100, Height: 20})
	btn.SetOnClicked(func(cw *controller.ControlWindow) {
		wf.mWidgets.Window().ClearDeltaMotion(0, wf.bakeState.CurrentIndex())
		wf.mWidgets.Window().ClearDeltaMotion(1, wf.bakeState.CurrentIndex())
		wf.mWidgets.Window().SetSaveDeltaIndex(0, 0)
		wf.mWidgets.Window().SetSaveDeltaIndex(1, 0)

		wf.bakeState.BakedHistoryIndexEdit.SetValue(1.0)
		wf.bakeState.BakedHistoryIndexEdit.SetRange(1.0, 2.0)
	})
	return btn
}

// Event handler methods

func (wf *WidgetFactory) createEnabledPlaying() func(playing bool) {
	return func(playing bool) {
		wf.bakeState.SetWidgetEnabled(!playing)
		if playing {
			// 再生中も操作可ウィジェットを有効化
			wf.bakeState.SetWidgetPlayingEnabled(true)
		}
	}
}

func (wf *WidgetFactory) createOnChangePlayingPre() func(playing bool) {
	return func(playing bool) {
		wf.mWidgets.Window().SetSaveDelta(0, playing)

		// 情報表示
		wf.mWidgets.Window().SetCheckedShowInfoEnabled(playing)

		if playing {
			// 焼き込み開始時にINDEX加算
			deltaIndex := wf.mWidgets.Window().GetDeltaMotionCount(0, wf.bakeState.CurrentIndex())
			if deltaIndex > 0 {
				// 既に焼き込みが1回以上行われている場合はインクリメント
				deltaIndex += 1
			}
			wf.mWidgets.Window().SetSaveDeltaIndex(0, deltaIndex)

			// 再生フレーム
			mlog.IL(mi18n.T("焼き込み再生開始: 焼き込み履歴INDEX[%d]"), deltaIndex+1)
		} else {
			// 焼き込み完了時に出力モーションを取得
			wf.createHistoryIndexChangeHandler()()
		}
	}
}

func (wf *WidgetFactory) createPhysicsTableViewDialog(isAdd bool) func() {
	return func() {
		var record *domain.PhysicsBoneRecord
		recordIndex := -1
		switch isAdd {
		case true:
			if wf.bakeState.CurrentSet().OriginalMotion == nil {
				record = domain.NewPhysicsBoneRecord(wf.bakeState.CurrentSet().OriginalModel,
					0,
					0)
			} else {
				record = domain.NewPhysicsBoneRecord(wf.bakeState.CurrentSet().OriginalModel,
					wf.bakeState.CurrentSet().OriginalMotion.MinFrame(),
					wf.bakeState.CurrentSet().OriginalMotion.MaxFrame())
			}
		case false:
			record = wf.bakeState.CurrentSet().PhysicsTableModel.Records[wf.bakeState.PhysicsTableView.CurrentIndex()]
			recordIndex = wf.bakeState.PhysicsTableView.CurrentIndex()
		}
		dialog := NewPhysicsTableViewDialog(wf.bakeState, wf.mWidgets)
		dialog.Show(record, recordIndex)
	}
}

func (wf *WidgetFactory) createOutputTableViewDialog() func() {
	return func() {
		dialog := NewOutputTableViewDialog(wf.bakeState, wf.mWidgets)
		dialog.Show()
	}
}

func (wf *WidgetFactory) createHistoryIndexChangeHandler() func() {
	return func() {
		// 出力モーションインデックスが変更されたときの処理
		currentSet := wf.bakeState.CurrentSet()
		deltaIndex := int(wf.bakeState.BakedHistoryIndexEdit.Value() - 1)
		if deltaIndex < 0 ||
			deltaIndex >= wf.mWidgets.Window().GetDeltaMotionCount(0, currentSet.Index) {
			deltaIndex = 0
		}

		// 物理ありのモーションを取得
		outputMotion := wf.mWidgets.Window().LoadDeltaMotion(0, currentSet.Index, deltaIndex)
		// 物理確認用として設定
		wf.mWidgets.Window().StoreMotion(1, currentSet.Index, outputMotion)
		wf.mWidgets.Window().TriggerPhysicsReset()

		// 出力モーションを更新
		currentSet.OutputMotion = outputMotion
		currentSet.OutputMotionPath = currentSet.CreateOutputMotionPath()
		wf.bakeState.OutputMotionPicker.ChangePath(currentSet.OutputMotionPath)
	}
}

// Helper methods

func (wf *WidgetFactory) handleLoadSet(filePath string) {
	wf.bakeState.SetWidgetEnabled(false)
	mconfig.SaveUserConfig("physics_set_path", filePath, 1)

	for n := range 2 {
		for m := range wf.bakeState.NavToolBar.Actions().Len() {
			wf.mWidgets.Window().StoreModel(n, m, nil)
			wf.mWidgets.Window().StoreMotion(n, m, nil)
		}
	}

	wf.bakeState.ResetSet()
	wf.bakeState.LoadSet(filePath)

	for range len(wf.bakeState.BakeSets) - 1 {
		wf.bakeState.AddAction()
	}

	physicsWorldMotion := wf.mWidgets.Window().LoadPhysicsWorldMotion(0)

	for index := range wf.bakeState.BakeSets {
		wf.bakeState.ChangeCurrentAction(index)
		wf.bakeState.OriginalModelPicker.SetForcePath(wf.bakeState.BakeSets[index].OriginalModelPath)
		wf.bakeState.OriginalMotionPicker.SetForcePath(wf.bakeState.BakeSets[index].OriginalMotionPath)

		physicsTable := wf.bakeState.BakeSets[index].PhysicsTableModel
		newPhysicsTable := domain.NewPhysicsTableModel()
		for _, record := range physicsTable.Records {
			newPhysicsTable.AddRecord(
				wf.bakeState.BakeSets[index].OriginalModel,
				record.StartFrame,
				record.EndFrame,
			)
			latestRecord := newPhysicsTable.Records[len(newPhysicsTable.Records)-1]
			latestRecord.Gravity = record.Gravity
			latestRecord.MaxStartFrame = record.MaxStartFrame
			latestRecord.MaxEndFrame = record.MaxEndFrame
			latestRecord.MaxSubSteps = record.MaxSubSteps
			latestRecord.FixedTimeStep = record.FixedTimeStep
			latestRecord.IsStartDeform = record.IsStartDeform
			latestRecord.SizeRatio = record.SizeRatio
			latestRecord.MassRatio = record.MassRatio
			latestRecord.TensionRatio = record.TensionRatio
			latestRecord.StiffnessRatio = record.StiffnessRatio

			// JSONから復元したツリー情報を設定
			latestRecord.TreeModel.UpdateModifiedNodes(nil, record.TreeModel.Nodes)
		}

		// 物理設定
		wf.bakeState.BakeSets[index].PhysicsTableModel = newPhysicsTable
		wf.bakeState.PhysicsTableView.SetModel(newPhysicsTable)

		physicsModelMotion := wf.mWidgets.Window().LoadPhysicsModelMotion(0, index)

		wf.physicsUsecase.ApplyPhysicsMotion(
			physicsWorldMotion, physicsModelMotion,
			newPhysicsTable.Records,
			wf.bakeState.BakeSets[index].OriginalModel,
		)

		wf.mWidgets.Window().StorePhysicsModelMotion(0, index, physicsModelMotion)
	}

	wf.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	wf.mWidgets.Window().TriggerPhysicsReset()

	wf.bakeState.SetCurrentIndex(0)
	wf.bakeState.SetWidgetEnabled(true)
}
