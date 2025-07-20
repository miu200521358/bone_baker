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
	AddSetButton         *widget.MPushButton  // セット追加ボタン
	ResetSetButton       *widget.MPushButton  // セットリセットボタン
	SaveSetButton        *widget.MPushButton  // セット保存ボタン
	LoadSetButton        *widget.MPushButton  // セット読込ボタン
	NavToolBar           *walk.ToolBar        // セットツールバー
	currentIndex         int                  // 現在のインデックス
	OriginalMotionPicker *widget.FilePicker   // 元モーション
	PhysicsModelPicker   *widget.FilePicker   // 物理焼き込み先モデル
	OutputMotionPicker   *widget.FilePicker   // 出力モーション
	OutputModelPicker    *widget.FilePicker   // 出力モデル
	SaveButton           *widget.MPushButton  // 保存ボタン
	Player               *widget.MotionPlayer // モーションプレイヤー
	GravitySlider        *widget.TextSlider   // 重力スライダー
	PhysicsSets          []*domain.PhysicsSet `json:"physics_sets"` // 物理焼き込みセット
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
	ss.OriginalMotionPicker.ChangePath(ss.CurrentSet().OriginalMotionPath)
	ss.PhysicsModelPicker.ChangePath(ss.CurrentSet().PhysicsModelPath)
	ss.OutputMotionPicker.ChangePath(ss.CurrentSet().OutputMotionPath)
	ss.OutputModelPicker.ChangePath(ss.CurrentSet().OutputModelPath)
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

// LoadOriginalModel 元モデルを読み込む
func (physicsState *PhysicsState) LoadOriginalModel(
	cw *controller.ControlWindow, path string,
) error {
	physicsState.SetPhysicsEnabled(false)

	// オプションクリア
	physicsState.ClearOptions()

	if err := physicsState.CurrentSet().LoadOriginalModel(path); err != nil {
		return err
	}

	cw.StoreModel(1, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalModel)

	cw.StoreMotion(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OutputMotion)
	cw.StoreMotion(1, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalMotion)

	physicsState.SetPhysicsEnabled(true)

	return nil
}

// LoadPhysicsModel 物理焼き込み先モデルを読み込む
func (physicsState *PhysicsState) LoadPhysicsModel(
	cw *controller.ControlWindow, path string,
) error {
	physicsState.SetPhysicsEnabled(false)

	// オプションクリア
	physicsState.ClearOptions()

	if err := physicsState.CurrentSet().LoadPhysicsModel(path); err != nil {
		return err
	}

	cw.StoreModel(0, physicsState.CurrentIndex(), physicsState.CurrentSet().PhysicsModel)

	cw.StoreMotion(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OutputMotion)
	cw.StoreMotion(1, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalMotion)

	physicsState.OutputModelPicker.SetPath(physicsState.CurrentSet().OutputModelPath)
	physicsState.OutputMotionPicker.SetPath(physicsState.CurrentSet().OutputMotionPath)

	physicsState.SetPhysicsEnabled(true)

	return nil
}

// LoadPhysicsMotion 物理焼き込みモーションを読み込む
func (physicsState *PhysicsState) LoadPhysicsMotion(
	cw *controller.ControlWindow, path string, isClear bool,
) error {
	physicsState.SetPhysicsEnabled(false)

	// オプションクリア
	if isClear {
		physicsState.ClearOptions()
	}

	if err := physicsState.CurrentSet().LoadMotion(path); err != nil {
		return err
	}

	cw.StoreMotion(0, physicsState.CurrentIndex(), physicsState.CurrentSet().OutputMotion)
	cw.StoreMotion(1, physicsState.CurrentIndex(), physicsState.CurrentSet().OriginalMotion)

	if physicsState.CurrentSet().OriginalMotion != nil {
		physicsState.Player.Reset(physicsState.CurrentSet().OriginalMotion.MaxFrame())
	}

	physicsState.OutputMotionPicker.SetPath(physicsState.CurrentSet().OutputMotionPath)

	physicsState.SetPhysicsEnabled(true)

	return nil
}

// SetPhysicsEnabled 物理焼き込み有効無効設定
func (physicsState *PhysicsState) SetPhysicsEnabled(enabled bool) {
	physicsState.AddSetButton.SetEnabled(enabled)
	physicsState.ResetSetButton.SetEnabled(enabled)
	physicsState.SaveSetButton.SetEnabled(enabled)
	physicsState.LoadSetButton.SetEnabled(enabled)

	physicsState.OriginalMotionPicker.SetEnabled(enabled)
	physicsState.PhysicsModelPicker.SetEnabled(enabled)
	physicsState.OutputMotionPicker.SetEnabled(enabled)
	physicsState.OutputModelPicker.SetEnabled(enabled)

	physicsState.Player.SetEnabled(enabled)
	physicsState.GravitySlider.SetEnabled(enabled)

	physicsState.SetPhysicsOptionEnabled(enabled)
}

func (physicsState *PhysicsState) SetPhysicsOptionEnabled(enabled bool) {
	physicsState.SaveButton.SetEnabled(enabled)
}
