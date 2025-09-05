package page

import "github.com/miu200521358/mlib_go/pkg/interface/controller"

// LoadMotion 物理焼き込みモーションを読み込む
func (s *WidgetStore) LoadMotion(cw *controller.ControlWindow, path string) error {
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

func (s *WidgetStore) LoadModel(cw *controller.ControlWindow, path string) error {
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
