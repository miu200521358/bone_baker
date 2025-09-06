package ui

import (
	"github.com/miu200521358/mlib_go/pkg/config/mconfig"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
)

// loadMotion 物理焼き込みモーションを読み込む
func (s *WidgetStore) loadMotion(cw *controller.ControlWindow, path string) error {
	s.setWidgetEnabled(false)

	if err := s.loadUsecase.LoadMotion(s.CurrentSet(), path); err != nil {
		return err
	}

	// UI反映処理
	currentSet := s.CurrentSet()
	if currentSet.OriginalMotion != nil {
		cw.StoreMotion(0, s.CurrentIndex, currentSet.OriginalMotion)
	}
	if currentSet.OutputMotion != nil {
		cw.StoreMotion(1, s.CurrentIndex, currentSet.OutputMotion)
	}

	// 履歴クリア処理
	for n := range s.BakeSets {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}

	// s.BakedHistoryIndexEdit.SetValue(1.0)
	// s.BakedHistoryIndexEdit.SetRange(1.0, 2.0)

	// モーションプレイヤーのリセット
	s.Player.Reset(s.MaxFrame())

	s.OutputMotionPicker.SetPath(s.CurrentSet().OutputMotionPath)
	s.setWidgetEnabled(true)

	return nil
}

func (s *WidgetStore) loadModel(cw *controller.ControlWindow, path string) error {
	s.setWidgetEnabled(false)

	if err := s.loadUsecase.LoadModel(s.CurrentSet(), path); err != nil {
		return err
	}

	// UI反映処理
	currentSet := s.CurrentSet()
	cw.StoreModel(0, s.CurrentIndex, currentSet.OriginalModel)
	cw.StoreModel(1, s.CurrentIndex, currentSet.BakedModel)

	// 履歴クリア処理
	for n := range s.BakeSets {
		cw.ClearDeltaMotion(0, n)
		cw.ClearDeltaMotion(1, n)
		cw.SetSaveDeltaIndex(0, 0)
		cw.SetSaveDeltaIndex(1, 0)
	}

	// s.BakedHistoryIndexEdit.SetValue(1.0)
	// s.BakedHistoryIndexEdit.SetRange(1.0, 2.0)

	s.OutputModelPicker.ChangePath(s.CurrentSet().OutputModelPath)
	s.setWidgetEnabled(true)

	return nil
}

func (s *WidgetStore) loadBakeSets(filePath string) {
	s.setWidgetEnabled(false)
	mconfig.SaveUserConfig("physics_set_path", filePath, 1)

	for n := range 2 {
		for m := range s.NavToolBar.Actions().Len() {
			s.mWidgets.Window().StoreModel(n, m, nil)
			s.mWidgets.Window().StoreMotion(n, m, nil)
		}
	}

	s.ResetStore()
	var err error
	s.BakeSets, err = s.loadUsecase.LoadFile(filePath)
	if err != nil {
		return
	}

	for range len(s.BakeSets) - 1 {
		s.AddAction()
	}

	physicsWorldMotion := s.mWidgets.Window().LoadPhysicsWorldMotion(0)

	for index := range s.BakeSets {
		s.ChangeCurrentAction(index)
		s.OriginalModelPicker.SetForcePath(s.BakeSets[index].OriginalModelPath)
		s.OriginalMotionPicker.SetForcePath(s.BakeSets[index].OriginalMotionPath)

		physicsRecords := s.BakeSets[index].PhysicsRecords
		newPhysicsTable := NewPhysicsTableModelWithRecords(physicsRecords)
		s.PhysicsTableView.SetModel(newPhysicsTable)

		s.physicsUsecase.ApplyPhysicsWorldMotion(
			physicsWorldMotion,
			newPhysicsTable.Records,
			s.BakeSets[index].OriginalModel,
		)
	}

	s.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	s.mWidgets.Window().TriggerPhysicsReset()

	s.CurrentIndex = 0
	s.setWidgetEnabled(true)
}
