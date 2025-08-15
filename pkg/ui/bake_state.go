package ui

import (
	"fmt"

	"github.com/miu200521358/bone_baker/pkg/application"
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/walk"
)

type BakeState struct {
	AddSetButton          *widget.MPushButton  // セット追加ボタン
	ResetSetButton        *widget.MPushButton  // セットリセットボタン
	SaveSetButton         *widget.MPushButton  // セット保存ボタン
	LoadSetButton         *widget.MPushButton  // セット読込ボタン
	NavToolBar            *walk.ToolBar        // セットツールバー
	currentIndex          int                  // 現在のインデックス
	OriginalMotionPicker  *widget.FilePicker   // 元モーション
	OriginalModelPicker   *widget.FilePicker   // 物理焼き込み先モデル
	OutputMotionPicker    *widget.FilePicker   // 出力モーション
	OutputModelPicker     *widget.FilePicker   // 出力モデル
	BakedHistoryIndexEdit *walk.NumberEdit     // 出力モーションインデックスプルダウン
	SaveModelButton       *widget.MPushButton  // モデル保存ボタン
	SaveMotionButton      *widget.MPushButton  // モーション保存ボタン
	Player                *widget.MotionPlayer // モーションプレイヤー
	PhysicsTableView      *walk.TableView      // 物理ボーン表示テーブル
	OutputTableView       *walk.TableView      // 出力定義テーブル
	BakeSets              []*domain.BakeSet    `json:"bake_sets"` // ボーン焼き込みセット

	// Application Service（依存性注入）
	applicationService *application.BakeApplicationService
}

func NewBakeState(applicationService *application.BakeApplicationService) *BakeState {
	return &BakeState{
		applicationService: applicationService,
		BakeSets:           make([]*domain.BakeSet, 0),
		currentIndex:       -1,
	}
}

func (ss *BakeState) AddAction() {
	index := ss.NavToolBar.Actions().Len()

	action := ss.newAction(index)
	ss.NavToolBar.Actions().Add(action)
	ss.ChangeCurrentAction(index)
}

func (ss *BakeState) newAction(index int) *walk.Action {
	action := walk.NewAction()
	action.SetCheckable(true)
	action.SetExclusive(true)
	action.SetText(fmt.Sprintf(" No. %d ", index+1))

	action.Triggered().Attach(func() {
		ss.ChangeCurrentAction(index)
	})

	return action
}

func (ss *BakeState) ResetSet() {
	// 一旦全部削除
	for range ss.NavToolBar.Actions().Len() {
		index := ss.NavToolBar.Actions().Len() - 1
		ss.BakeSets[index].Delete()
		ss.NavToolBar.Actions().RemoveAt(index)
	}

	ss.BakeSets = make([]*domain.BakeSet, 0)
	ss.currentIndex = -1

	// 1セット追加
	ss.BakeSets = append(ss.BakeSets, domain.NewPhysicsSet(len(ss.BakeSets)))
	ss.AddAction()
}

func (ss *BakeState) ChangeCurrentAction(index int) {
	// 一旦すべてのチェックを外す
	for i := range ss.NavToolBar.Actions().Len() {
		ss.NavToolBar.Actions().At(i).SetChecked(false)
	}

	// 該当INDEXのみチェックON
	ss.currentIndex = index
	ss.NavToolBar.Actions().At(index).SetChecked(true)

	// 物理焼き込みセットの情報を表示
	ss.OriginalModelPicker.ChangePath(ss.CurrentSet().OriginalModelPath)
	ss.OriginalMotionPicker.ChangePath(ss.CurrentSet().OriginalMotionPath)
	ss.OutputModelPicker.ChangePath(ss.CurrentSet().OutputModelPath)
	ss.OutputMotionPicker.ChangePath(ss.CurrentSet().OutputMotionPath)

	// // 物理ツリーのモデル変更
	// if ss.CurrentSet().PhysicsBoneTreeModel == nil {
	// 	ss.CurrentSet().PhysicsBoneTreeModel = domain.NewPhysicsBoneTreeModel()
	// }
	// ss.PhysicsTreeView.SetModel(ss.CurrentSet().PhysicsBoneTreeModel)
}

func (ss *BakeState) ClearOptions() {
	ss.Player.Reset(ss.MaxFrame())
}

func (ss *BakeState) MaxFrame() float32 {
	maxFrame := float32(0)
	for _, physicsSet := range ss.BakeSets {
		if physicsSet.OriginalMotion != nil && maxFrame < physicsSet.OriginalMotion.MaxFrame() {
			maxFrame = physicsSet.OriginalMotion.MaxFrame()
		}
	}

	return maxFrame
}

func (ss *BakeState) SetCurrentIndex(index int) {
	ss.currentIndex = index
}

func (ss *BakeState) CurrentIndex() int {
	return ss.currentIndex
}

func (ss *BakeState) CurrentSet() *domain.BakeSet {
	if ss.currentIndex < 0 || ss.currentIndex >= len(ss.BakeSets) {
		return nil
	}

	return ss.BakeSets[ss.currentIndex]
}

// SaveSet セット情報を保存
func (ss *BakeState) SaveSet(jsonPath string) error {
	return ss.applicationService.SaveBakeSetToFile(ss.BakeSets, jsonPath)
}

// LoadSet セット情報を読み込む
func (ss *BakeState) LoadSet(jsonPath string) error {
	bakeSets, err := ss.applicationService.LoadBakeSetFromFile(jsonPath)
	if err != nil {
		return err
	}

	ss.BakeSets = bakeSets
	return nil
}

// LoadModel 元モデルを読み込む
func (bakeState *BakeState) LoadModel(
	cw *controller.ControlWindow, path string,
) error {
	bakeState.SetWidgetEnabled(false)

	// オプションクリア
	bakeState.ClearOptions()

	if err := bakeState.applicationService.LoadModelForBakeSet(
		bakeState.CurrentSet(), path, cw, bakeState.CurrentIndex()); err != nil {
		return err
	}

	bakeState.applicationService.PrepareForBaking(cw, bakeState.CurrentSet(), bakeState.CurrentIndex())
	bakeState.applicationService.ClearDeltaMotions(cw, bakeState.BakeSets)

	bakeState.BakedHistoryIndexEdit.SetValue(1.0)
	bakeState.BakedHistoryIndexEdit.SetRange(1.0, 2.0)

	bakeState.OutputModelPicker.ChangePath(bakeState.CurrentSet().OutputModelPath)
	bakeState.SetWidgetEnabled(true)

	return nil
}

// LoadMotion 物理焼き込みモーションを読み込む
func (bakeState *BakeState) LoadMotion(
	cw *controller.ControlWindow, path string, isClear bool,
) error {
	bakeState.SetWidgetEnabled(false)

	// オプションクリア
	if isClear {
		bakeState.ClearOptions()
	}

	if err := bakeState.applicationService.LoadMotionForBakeSet(
		bakeState.CurrentSet(), path, cw, bakeState.CurrentIndex()); err != nil {
		return err
	}

	bakeState.applicationService.ClearDeltaMotions(cw, bakeState.BakeSets)

	bakeState.BakedHistoryIndexEdit.SetValue(1.0)
	bakeState.BakedHistoryIndexEdit.SetRange(1.0, 2.0)

	if bakeState.CurrentSet().OriginalMotion != nil {
		// テーブル初期化処理をApplication Serviceに委譲
		bakeState.applicationService.InitializeOutputTable(bakeState.CurrentSet())
		bakeState.OutputTableView.SetModel(bakeState.CurrentSet().OutputTableModel)

		bakeState.applicationService.InitializePhysicsTable(bakeState.CurrentSet())
		bakeState.PhysicsTableView.SetModel(bakeState.CurrentSet().PhysicsTableModel)
	}

	// モーションプレイヤーのリセット
	bakeState.Player.Reset(bakeState.applicationService.CalculateMaxFrame(bakeState.BakeSets))

	bakeState.OutputMotionPicker.SetPath(bakeState.CurrentSet().OutputMotionPath)
	bakeState.SetWidgetEnabled(true)

	return nil
}

// SetWidgetEnabled 物理焼き込み有効無効設定
func (bakeState *BakeState) SetWidgetEnabled(enabled bool) {
	bakeState.OutputTableView.SetEnabled(enabled)

	bakeState.AddSetButton.SetEnabled(enabled)
	bakeState.ResetSetButton.SetEnabled(enabled)
	bakeState.SaveSetButton.SetEnabled(enabled)
	bakeState.LoadSetButton.SetEnabled(enabled)

	bakeState.OriginalMotionPicker.SetEnabled(enabled)
	bakeState.OriginalModelPicker.SetEnabled(enabled)
	bakeState.OutputMotionPicker.SetEnabled(enabled)
	bakeState.OutputModelPicker.SetEnabled(enabled)

	bakeState.SaveModelButton.SetEnabled(enabled)
	bakeState.SaveMotionButton.SetEnabled(enabled)

	bakeState.SetWidgetPlayingEnabled(enabled)
}

func (bakeState *BakeState) SetWidgetPlayingEnabled(enabled bool) {
	bakeState.Player.SetEnabled(enabled)

	bakeState.PhysicsTableView.SetEnabled(enabled)

	// bakeState.GravityEdit.SetEnabled(enabled)
	// bakeState.MaxSubStepsEdit.SetEnabled(enabled)
	// bakeState.FixedTimeStepEdit.SetEnabled(enabled)
	// bakeState.MassEdit.SetEnabled(enabled)
	// bakeState.StiffnessEdit.SetEnabled(enabled)
	// bakeState.TensionEdit.SetEnabled(enabled)
	// bakeState.PhysicsTreeView.SetEnabled(enabled)
}
