package ui

import (
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
)

// loadMotion 物理焼き込みモーションを読み込む
func (s *WidgetStore) loadMotion(cw *controller.ControlWindow, path string) error {
	s.setWidgetEnabled(false)

	if err := s.loadUsecase.LoadMotion(s.currentSet(), path); err != nil {
		return err
	}

	// UI反映処理
	currentSet := s.currentSet()
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
	s.Player.Reset(s.maxFrame())

	s.OutputMotionPicker.SetPath(s.currentSet().OutputMotionPath)
	s.setWidgetEnabled(true)

	return nil
}

func (s *WidgetStore) loadModel(cw *controller.ControlWindow, path string) error {
	s.setWidgetEnabled(false)

	if err := s.loadUsecase.LoadModel(s.currentSet(), path); err != nil {
		return err
	}

	// UI反映処理
	currentSet := s.currentSet()
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

	s.OutputModelPicker.ChangePath(s.currentSet().OutputModelPath)
	s.setWidgetEnabled(true)

	return nil
}

func (s *WidgetStore) saveBakeSets(filePath string) error {
	return s.saveUsecase.SaveFile(s.BakeSets, s.PhysicsRecords, filePath)
}

func (s *WidgetStore) loadBakeSets(filePath string) {
	s.setWidgetEnabled(false)

	for n := range 2 {
		for m := range s.NavToolBar.Actions().Len() {
			s.mWidgets.Window().StoreModel(n, m, nil)
			s.mWidgets.Window().StoreMotion(n, m, nil)
		}
	}

	s.resetStore()
	var err error
	s.BakeSets, s.PhysicsRecords, err = s.loadUsecase.LoadFile(filePath)
	if err != nil {
		return
	}

	for range len(s.BakeSets) - 1 {
		s.AddAction()
	}

	physicsWorldMotion := s.mWidgets.Window().LoadPhysicsWorldMotion(0)

	for index := range s.BakeSets {
		s.changeCurrentAction(index)
		s.OriginalModelPicker.SetForcePath(s.BakeSets[index].OriginalModelPath)
		s.OriginalMotionPicker.SetForcePath(s.BakeSets[index].OriginalMotionPath)
	}

	newPhysicsTableModel := NewPhysicsTableModelWithRecords(s.PhysicsRecords)
	s.PhysicsTableView.SetModel(newPhysicsTableModel)

	s.physicsUsecase.ApplyPhysicsWorldMotion(
		physicsWorldMotion,
		newPhysicsTableModel.Records,
	)

	s.mWidgets.Window().StorePhysicsWorldMotion(0, physicsWorldMotion)
	s.mWidgets.Window().TriggerPhysicsReset()

	s.CurrentIndex = 0
	s.setWidgetEnabled(true)
}
