package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/miu200521358/physics_fixer/pkg/domain"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/walk"
)

type PhysicsState struct {
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
	OutputMotionIndexEdit *walk.NumberEdit     // 出力モーションインデックスプルダウン
	SaveModelButton       *widget.MPushButton  // モデル保存ボタン
	StartFrameEdit        *walk.NumberEdit     // 開始フレーム
	EndFrameEdit          *walk.NumberEdit     // 終了フレーム
	SaveMotionButton      *widget.MPushButton  // モーション保存ボタン
	Player                *widget.MotionPlayer // モーションプレイヤー
	GravityEdit           *walk.NumberEdit     // 重力値入力
	MaxSubStepsEdit       *walk.NumberEdit     // 最大サブステップ数
	PhysicsSets           []*domain.PhysicsSet `json:"physics_sets"` // 物理焼き込みセット
}

func (ss *PhysicsState) AddAction() {
	index := ss.NavToolBar.Actions().Len()

	action := ss.newAction(index)
	ss.NavToolBar.Actions().Add(action)
	ss.ChangeCurrentAction(index)
}

func (ss *PhysicsState) newAction(index int) *walk.Action {
	action := walk.NewAction()
	action.SetCheckable(true)
	action.SetExclusive(true)
	action.SetText(fmt.Sprintf(" No. %d ", index+1))

	action.Triggered().Attach(func() {
		ss.ChangeCurrentAction(index)
	})

	return action
}

func (ss *PhysicsState) ResetSet() {
	// 一旦全部削除
	for range ss.NavToolBar.Actions().Len() {
		index := ss.NavToolBar.Actions().Len() - 1
		ss.PhysicsSets[index].Delete()
		ss.NavToolBar.Actions().RemoveAt(index)
	}

	ss.PhysicsSets = make([]*domain.PhysicsSet, 0)
	ss.currentIndex = -1

	// 1セット追加
	ss.PhysicsSets = append(ss.PhysicsSets, domain.NewPhysicsSet(len(ss.PhysicsSets)))
	ss.AddAction()
}

func (ss *PhysicsState) ChangeCurrentAction(index int) {
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
}

func (ss *PhysicsState) ClearOptions() {
	ss.Player.Reset(ss.MaxFrame())
}

func (ss *PhysicsState) MaxFrame() float32 {
	maxFrame := float32(0)
	for _, physicsSet := range ss.PhysicsSets {
		if physicsSet.OriginalMotion != nil && maxFrame < physicsSet.OriginalMotion.MaxFrame() {
			maxFrame = physicsSet.OriginalMotion.MaxFrame()
		}
	}

	return maxFrame
}

func (ss *PhysicsState) SetCurrentIndex(index int) {
	ss.currentIndex = index
}

func (ss *PhysicsState) CurrentIndex() int {
	return ss.currentIndex
}

func (ss *PhysicsState) CurrentSet() *domain.PhysicsSet {
	if ss.currentIndex < 0 || ss.currentIndex >= len(ss.PhysicsSets) {
		return nil
	}

	return ss.PhysicsSets[ss.currentIndex]
}

// SaveSet セット情報を保存
func (ss *PhysicsState) SaveSet(jsonPath string) error {
	if strings.ToLower(filepath.Ext(jsonPath)) != ".json" {
		// 拡張子が.jsonでない場合は付与
		jsonPath += ".json"
	}

	// セット情報をJSONに変換してファイルダイアログで選択した箇所に保存
	if output, err := json.Marshal(ss.PhysicsSets); err == nil && len(output) > 0 {
		if err := os.WriteFile(jsonPath, output, 0644); err == nil {
			mlog.I(mi18n.T("物理焼き込みセット保存成功", map[string]any{"Path": jsonPath}))
		} else {
			mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
			return err
		}
	} else {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	return nil
}

// LoadSet セット情報を読み込む
func (ss *PhysicsState) LoadSet(jsonPath string) error {
	// セット情報をJSONから読み込んでセット情報を更新
	if input, err := os.ReadFile(jsonPath); err == nil && len(input) > 0 {
		if err := json.Unmarshal(input, &ss.PhysicsSets); err == nil {
			mlog.I(mi18n.T("物理焼き込みセット読込成功", map[string]any{"Path": jsonPath}))
		} else {
			mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
			return err
		}
	} else {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return err
	}

	return nil
}

// LoadModel 元モデルを読み込む
func (physicsState *PhysicsState) LoadModel(
	cw *controller.ControlWindow, path string,
) error {
	physicsState.SetWidgetEnabled(false)

	// オプションクリア
	physicsState.ClearOptions()

	if err := physicsState.CurrentSet().LoadModel(path); err != nil {
		return err
	}

	cw.StoreModel(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalModel)
	cw.StoreModel(1, physicsState.CurrentIndex(), physicsState.CurrentSet().PhysicsBakedModel)

	cw.StoreMotion(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalMotion)
	if physicsState.CurrentSet().OriginalMotion != nil {
		if copiedMotion, err := physicsState.CurrentSet().OriginalMotion.Copy(); err == nil {
			cw.StoreMotion(1, physicsState.CurrentIndex(), copiedMotion)
		}
	}

	for n := range physicsState.PhysicsSets {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}

	physicsState.OutputMotionIndexEdit.SetValue(1.0)
	physicsState.OutputMotionIndexEdit.SetRange(1.0, 2.0)

	physicsState.OutputModelPicker.ChangePath(physicsState.CurrentSet().OutputModelPath)
	physicsState.SetWidgetEnabled(true)

	return nil
}

// LoadMotion 物理焼き込みモーションを読み込む
func (physicsState *PhysicsState) LoadMotion(
	cw *controller.ControlWindow, path string, isClear bool,
) error {
	physicsState.SetWidgetEnabled(false)

	// オプションクリア
	if isClear {
		physicsState.ClearOptions()
	}

	if err := physicsState.CurrentSet().LoadMotion(path); err != nil {
		return err
	}

	if physicsState.CurrentSet().OriginalMotion != nil {
		cw.StoreMotion(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalMotion)
	}

	if physicsState.CurrentSet().OutputMotion != nil {
		cw.StoreMotion(1, physicsState.CurrentIndex(), physicsState.CurrentSet().OutputMotion)
	}

	for n := range physicsState.PhysicsSets {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}

	physicsState.OutputMotionIndexEdit.SetValue(1.0)
	physicsState.OutputMotionIndexEdit.SetRange(1.0, 2.0)

	// モーションプレイヤーのリセット
	if physicsState.CurrentSet().OriginalMotion != nil {
		physicsState.Player.Reset(physicsState.CurrentSet().OriginalMotion.MaxFrame())
		physicsState.StartFrameEdit.SetRange(0, float64(physicsState.CurrentSet().OriginalMotion.MaxFrame()))
		physicsState.EndFrameEdit.SetRange(0, float64(physicsState.CurrentSet().OriginalMotion.MaxFrame()))
		physicsState.EndFrameEdit.SetValue(float64(physicsState.CurrentSet().OriginalMotion.MaxFrame()))
	}

	physicsState.OutputMotionPicker.SetPath(physicsState.CurrentSet().OutputMotionPath)
	physicsState.SetWidgetEnabled(true)

	return nil
}

// SetWidgetEnabled 物理焼き込み有効無効設定
func (physicsState *PhysicsState) SetWidgetEnabled(enabled bool) {
	physicsState.AddSetButton.SetEnabled(enabled)
	physicsState.ResetSetButton.SetEnabled(enabled)
	physicsState.SaveSetButton.SetEnabled(enabled)
	physicsState.LoadSetButton.SetEnabled(enabled)

	physicsState.OriginalMotionPicker.SetEnabled(enabled)
	physicsState.OriginalModelPicker.SetEnabled(enabled)
	physicsState.OutputMotionPicker.SetEnabled(enabled)
	physicsState.OutputModelPicker.SetEnabled(enabled)

	physicsState.Player.SetEnabled(enabled)

	physicsState.SaveModelButton.SetEnabled(enabled)
	physicsState.SaveMotionButton.SetEnabled(enabled)

	physicsState.SetPhysicsOptionEnabled(enabled)
}

func (physicsState *PhysicsState) SetPhysicsOptionEnabled(enabled bool) {
	physicsState.GravityEdit.SetEnabled(enabled)
	physicsState.MaxSubStepsEdit.SetEnabled(enabled)

	physicsState.StartFrameEdit.SetEnabled(enabled)
	physicsState.EndFrameEdit.SetEnabled(enabled)
	physicsState.OutputMotionIndexEdit.SetEnabled(enabled)
}
