package ui

import (
	"fmt"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/usecase"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/walk"
)

type BakeState struct {
	AddSetButton             *widget.MPushButton  // セット追加ボタン
	ResetSetButton           *widget.MPushButton  // セットリセットボタン
	SaveSetButton            *widget.MPushButton  // セット保存ボタン
	LoadSetButton            *widget.MPushButton  // セット読込ボタン
	NavToolBar               *walk.ToolBar        // セットツールバー
	currentIndex             int                  // 現在のインデックス
	OriginalMotionPicker     *widget.FilePicker   // 元モーション
	OriginalModelPicker      *widget.FilePicker   // 物理焼き込み先モデル
	OutputMotionPicker       *widget.FilePicker   // 出力モーション
	OutputModelPicker        *widget.FilePicker   // 出力モデル
	OutputMotionIndexEdit    *walk.NumberEdit     // 出力モーションインデックスプルダウン
	SaveModelButton          *widget.MPushButton  // モデル保存ボタン
	StartFrameEdit           *walk.NumberEdit     // 開始フレーム
	EndFrameEdit             *walk.NumberEdit     // 終了フレーム
	SaveMotionButton         *widget.MPushButton  // モーション保存ボタン
	Player                   *widget.MotionPlayer // モーションプレイヤー
	GravityEdit              *walk.NumberEdit     // 重力値入力
	MassEdit                 *walk.NumberEdit     // 質量入力
	StiffnessEdit            *walk.NumberEdit     // 硬さ入力
	TensionEdit              *walk.NumberEdit     // 張り入力
	MaxSubStepsEdit          *walk.NumberEdit     // 最大最大演算回数
	FixedTimeStepEdit        *walk.NumberEdit     // 固定タイムステップ入力
	PhysicsTreeView          *walk.TreeView       // 物理ボーン表示ツリー
	OutputTreeView           *walk.TreeView       // 出力ボーン表示ツリー
	OutputTableView          *walk.TableView      // 出力定義テーブル
	IsOutputUpdatingChildren bool                 // 子どもアイテム更新中フラグ
	OutputIkCheckBox         *walk.CheckBox       // 出力IKチェックボックス
	IsOutputUpdatingIk       bool                 // 出力IK更新中フラグ
	OutputPhysicsCheckBox    *walk.CheckBox       // 出力物理チェックボックス
	IsOutputUpdatingPhysics  bool                 // 出力物理更新中フラグ
	BakeSets                 []*domain.BakeSet    `json:"bake_sets"` // ボーン焼き込みセット

	// Usecase（依存性注入）
	bakeUsecase *usecase.BakeUsecase
}

func NewBakeState(bakeUsecase *usecase.BakeUsecase) *BakeState {
	return &BakeState{
		bakeUsecase:  bakeUsecase,
		BakeSets:     make([]*domain.BakeSet, 0),
		currentIndex: -1,
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

	// 物理ツリーのモデル変更
	if ss.CurrentSet().PhysicsTree == nil {
		ss.CurrentSet().PhysicsTree = domain.NewPhysicsModel()
	}
	ss.PhysicsTreeView.SetModel(ss.CurrentSet().PhysicsTree)

	// 出力ツリーのモデル変更
	if ss.CurrentSet().OutputTree == nil {
		ss.CurrentSet().OutputTree = domain.NewOutputModel()
	}
	ss.OutputTreeView.SetModel(ss.CurrentSet().OutputTree)
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
	return ss.bakeUsecase.SaveBakeSet(ss.BakeSets, jsonPath)
}

// LoadSet セット情報を読み込む
func (ss *BakeState) LoadSet(jsonPath string) error {
	bakeSets, err := ss.bakeUsecase.LoadBakeSet(jsonPath)
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

	if err := bakeState.bakeUsecase.LoadModel(bakeState.CurrentSet(), path); err != nil {
		return err
	}

	cw.StoreModel(0, bakeState.CurrentIndex(), bakeState.CurrentSet().OriginalModel)
	cw.StoreModel(1, bakeState.CurrentIndex(), bakeState.CurrentSet().BakedModel)

	cw.StoreMotion(0, bakeState.CurrentIndex(), bakeState.CurrentSet().OriginalMotion)
	if bakeState.CurrentSet().OriginalMotion != nil {
		if copiedMotion, err := bakeState.CurrentSet().OriginalMotion.Copy(); err == nil {
			cw.StoreMotion(1, bakeState.CurrentIndex(), copiedMotion)
		}
	}

	// 物理ツリーモデ作成
	bakeState.createPhysicsTree()
	// 出力ツリーモデル作成
	bakeState.createOutputTree()

	for n := range bakeState.BakeSets {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}

	bakeState.OutputMotionIndexEdit.SetValue(1.0)
	bakeState.OutputMotionIndexEdit.SetRange(1.0, 2.0)

	bakeState.OutputModelPicker.ChangePath(bakeState.CurrentSet().OutputModelPath)
	bakeState.SetWidgetEnabled(true)

	return nil
}

func (bakeState *BakeState) createOutputTree() {
	// 出力ツリーのモデル変更
	tree := domain.NewOutputModel()

	for _, boneIndex := range bakeState.CurrentSet().OriginalModel.Bones.LayerSortedIndexes {
		if bone, err := bakeState.CurrentSet().OriginalModel.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := domain.NewOutputItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*domain.OutputItem).AddChild(item)
			}
		}
	}

	bakeState.CurrentSet().OutputTree = tree

	if err := bakeState.OutputTreeView.SetModel(tree); err != nil {
		mlog.E(mi18n.T("出力ボーンツリー設定失敗エラー"), err, "")
	}
}

func (bakeState *BakeState) createPhysicsTree() {
	// 物理ツリーのモデル変更
	tree := domain.NewPhysicsModel()

	for _, boneIndex := range bakeState.CurrentSet().OriginalModel.Bones.LayerSortedIndexes {
		if bone, err := bakeState.CurrentSet().OriginalModel.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := domain.NewPhysicsItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*domain.PhysicsItem).AddChild(item)
			}
		}
	}

	// 物理ボーンを持つアイテムのみを保存
	tree.SaveOnlyPhysicsItems()

	bakeState.CurrentSet().PhysicsTree = tree

	if err := bakeState.PhysicsTreeView.SetModel(tree); err != nil {
		mlog.E(mi18n.T("物理ボーンツリー設定失敗エラー"), err, "")
	}
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

	if err := bakeState.bakeUsecase.LoadMotion(bakeState.CurrentSet(), path); err != nil {
		return err
	}

	if bakeState.CurrentSet().OriginalMotion != nil {
		cw.StoreMotion(0, bakeState.CurrentIndex(), bakeState.CurrentSet().OriginalMotion)
	}

	if bakeState.CurrentSet().OutputMotion != nil {
		cw.StoreMotion(1, bakeState.CurrentIndex(), bakeState.CurrentSet().OutputMotion)
	}

	for n := range bakeState.BakeSets {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}

	bakeState.OutputMotionIndexEdit.SetValue(1.0)
	bakeState.OutputMotionIndexEdit.SetRange(1.0, 2.0)

	// モーションプレイヤーのリセット
	if bakeState.CurrentSet().OriginalMotion != nil {
		bakeState.Player.Reset(bakeState.CurrentSet().OriginalMotion.MaxFrame())
		bakeState.StartFrameEdit.SetRange(0, float64(bakeState.CurrentSet().OriginalMotion.MaxFrame()))
		bakeState.EndFrameEdit.SetRange(0, float64(bakeState.CurrentSet().OriginalMotion.MaxFrame()))
		bakeState.EndFrameEdit.SetValue(float64(bakeState.CurrentSet().OriginalMotion.MaxFrame()))
	}

	bakeState.OutputMotionPicker.SetPath(bakeState.CurrentSet().OutputMotionPath)
	bakeState.SetWidgetEnabled(true)

	return nil
}

// SetWidgetEnabled 物理焼き込み有効無効設定
func (bakeState *BakeState) SetWidgetEnabled(enabled bool) {
	bakeState.StartFrameEdit.SetEnabled(enabled)
	bakeState.EndFrameEdit.SetEnabled(enabled)
	bakeState.OutputMotionIndexEdit.SetEnabled(enabled)

	bakeState.AddSetButton.SetEnabled(enabled)
	bakeState.ResetSetButton.SetEnabled(enabled)
	bakeState.SaveSetButton.SetEnabled(enabled)
	bakeState.LoadSetButton.SetEnabled(enabled)

	bakeState.OriginalMotionPicker.SetEnabled(enabled)
	bakeState.OriginalModelPicker.SetEnabled(enabled)
	bakeState.OutputMotionPicker.SetEnabled(enabled)
	bakeState.OutputModelPicker.SetEnabled(enabled)

	bakeState.OutputIkCheckBox.SetEnabled(enabled)
	bakeState.OutputPhysicsCheckBox.SetEnabled(enabled)
	bakeState.OutputTreeView.SetEnabled(enabled)

	bakeState.SaveModelButton.SetEnabled(enabled)
	bakeState.SaveMotionButton.SetEnabled(enabled)

	bakeState.SetWidgetPlayingEnabled(enabled)
}

func (bakeState *BakeState) SetWidgetPlayingEnabled(enabled bool) {
	bakeState.Player.SetEnabled(enabled)

	bakeState.GravityEdit.SetEnabled(enabled)
	bakeState.MaxSubStepsEdit.SetEnabled(enabled)
	bakeState.FixedTimeStepEdit.SetEnabled(enabled)
	bakeState.MassEdit.SetEnabled(enabled)
	bakeState.StiffnessEdit.SetEnabled(enabled)
	bakeState.TensionEdit.SetEnabled(enabled)
	bakeState.PhysicsTreeView.SetEnabled(enabled)
}

// SetOutputChildrenChecked は指定されたアイテムの子どもを再帰的にチェック状態を設定する
func (bakeState *BakeState) SetOutputChildrenChecked(item walk.TreeItem, checked bool) {
	if item == nil || bakeState.IsOutputUpdatingChildren ||
		bakeState.IsOutputUpdatingPhysics || bakeState.IsOutputUpdatingIk {
		return
	}

	// 無限ループを防ぐためのフラグ
	bakeState.IsOutputUpdatingChildren = true
	defer func() {
		bakeState.IsOutputUpdatingChildren = false
	}()

	item.(*domain.OutputItem).SetChecked(checked)

	// 子どもの数を取得
	childCount := item.ChildCount()
	for i := range childCount {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 子どものチェック状態を設定
		if outputItem, ok := child.(*domain.OutputItem); ok {
			outputItem.SetChecked(checked)
			bakeState.OutputTreeView.SetChecked(outputItem, checked)
		}

		// 再帰的に孫も処理（フラグを一時的にクリアして再帰呼び出し）
		bakeState.IsOutputUpdatingChildren = false
		bakeState.SetOutputChildrenChecked(child, checked)
		bakeState.IsOutputUpdatingChildren = true
	}
}

// SetOutputPhysicsChecked は物理関連ボーンのチェック状態を設定する
func (bakeState *BakeState) SetOutputPhysicsChecked(item walk.TreeItem, checked bool) {
	// if bakeState.IsOutputUpdatingPhysics {
	// 	return
	// }

	// 無限ループを防ぐためのフラグ
	bakeState.IsOutputUpdatingPhysics = true
	defer func() {
		bakeState.IsOutputUpdatingPhysics = false
	}()

	if item == nil {
		for i := range bakeState.OutputTreeView.Model().RootCount() {
			item := bakeState.OutputTreeView.Model().RootAt(i)
			bakeState.SetOutputPhysicsChecked(item, checked)
		}
		return
	}

	// 子どもの数を取得
	for i := range item.ChildCount() {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 出力IKボーンのチェック状態を設定
		if outputItem, ok := child.(*domain.OutputItem); ok {
			if outputItem.AsPhysics() {
				outputItem.SetChecked(checked)
				bakeState.OutputTreeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		bakeState.SetOutputPhysicsChecked(child, checked)
	}
}

// SetOutputIkChecked はIK関連ボーンのチェック状態を設定する
func (bakeState *BakeState) SetOutputIkChecked(item walk.TreeItem, checked bool) {
	// if bakeState.IsOutputUpdatingIk {
	// 	return
	// }

	if item == nil {
		for i := range bakeState.OutputTreeView.Model().RootCount() {
			item := bakeState.OutputTreeView.Model().RootAt(i)
			bakeState.SetOutputIkChecked(item, checked)
		}
		return
	}

	// 無限ループを防ぐためのフラグ
	bakeState.IsOutputUpdatingIk = true
	defer func() {
		bakeState.IsOutputUpdatingIk = false
	}()

	// 子どもの数を取得
	for i := range item.ChildCount() {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 出力IKボーンのチェック状態を設定
		if outputItem, ok := child.(*domain.OutputItem); ok {
			if outputItem.AsIk() {
				outputItem.SetChecked(checked)
				bakeState.OutputTreeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		bakeState.SetOutputIkChecked(child, checked)
	}
}
