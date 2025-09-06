package ui

import (
	"fmt"

	"github.com/miu200521358/bone_baker/pkg/application/usecase"
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	pRepository "github.com/miu200521358/bone_baker/pkg/infrastructure/repository"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/mlib_go/pkg/interface/controller/widget"
	"github.com/miu200521358/walk/pkg/walk"
)

type WidgetStore struct {
	mWidgets                  *controller.MWidgets    // ウィジェット管理
	NavToolBar                *walk.ToolBar           // 設定ツールバー
	CurrentIndex              int                     // 現在のインデックス
	AddSetButton              *widget.MPushButton     // 設定追加ボタン
	ResetSetButton            *widget.MPushButton     // 設定リセットボタン
	SaveSetButton             *widget.MPushButton     // 設定保存ボタン
	LoadSetButton             *widget.MPushButton     // 設定読込ボタン
	OriginalModelPicker       *widget.FilePicker      // 物理焼き込み先モデル
	OriginalMotionPicker      *widget.FilePicker      // 物理焼き込み対象モーション
	OutputMotionPicker        *widget.FilePicker      // 出力モーション
	OutputModelPicker         *widget.FilePicker      // 出力モデル
	BakedHistoryIndexEdit     *walk.NumberEdit        // 出力モーションインデックスプルダウン
	BakeHistoryClearButton    *widget.MPushButton     // 焼き込み履歴クリアボタン
	SaveModelButton           *widget.MPushButton     // モデル保存ボタン
	SaveMotionButton          *widget.MPushButton     // モーション保存ボタン
	Player                    *widget.MotionPlayer    // モーションプレイヤー
	AddPhysicsButton          *widget.MPushButton     // 物理設定追加ボタン
	PhysicsTableView          *walk.TableView         // 物理設定テーブル
	AddRigidBodyPhysicsButton *widget.MPushButton     // 剛体物理追加ボタン
	RigidBodyTableView        *walk.TableView         // 剛体物理テーブル
	AddOutputButton           *widget.MPushButton     // 出力設定追加ボタン
	OutputTableView           *walk.TableView         // 出力定義テーブル
	BakeSets                  []*entity.BakeSet       `json:"bake_sets"`       // ボーン焼き込みセット
	PhysicsRecords            []*entity.PhysicsRecord `json:"physics_records"` // 物理設定レコード

	loadUsecase    *usecase.LoadUsecase
	saveUsecase    *usecase.SaveUsecase
	physicsUsecase *usecase.PhysicsUsecase
}

func NewWidgetStore(mWidgets *controller.MWidgets) *WidgetStore {
	fileRepo := pRepository.NewFileRepository()

	return &WidgetStore{
		mWidgets:       mWidgets,
		BakeSets:       make([]*entity.BakeSet, 0),
		CurrentIndex:   -1,
		loadUsecase:    usecase.NewLoadUsecase(fileRepo),
		saveUsecase:    usecase.NewSaveUsecase(fileRepo),
		physicsUsecase: usecase.NewPhysicsUsecase(),
	}
}

func (s *WidgetStore) Window() *controller.ControlWindow {
	return s.mWidgets.Window()
}

func (s *WidgetStore) AddAction() {
	index := s.NavToolBar.Actions().Len()

	action := s.newAction(index)
	s.NavToolBar.Actions().Add(action)
	s.changeCurrentAction(index)
}

func (s *WidgetStore) newAction(index int) *walk.Action {
	action := walk.NewAction()
	action.SetCheckable(true)
	action.SetExclusive(true)
	action.SetText(fmt.Sprintf(" No. %d ", index+1))

	action.Triggered().Attach(func() {
		s.changeCurrentAction(index)
	})

	return action
}

func (s *WidgetStore) resetStore() {
	// 一旦全部削除
	for range s.NavToolBar.Actions().Len() {
		index := s.NavToolBar.Actions().Len() - 1
		s.BakeSets[index].Clear()
		s.NavToolBar.Actions().RemoveAt(index)
	}

	s.BakeSets = make([]*entity.BakeSet, 0)
	s.CurrentIndex = -1

	// 1セット追加
	s.BakeSets = append(s.BakeSets, entity.NewBakeSet(len(s.BakeSets)))
	s.AddAction()
}

func (s *WidgetStore) changeCurrentAction(index int) {
	// 一旦すべてのチェックを外す
	for i := range s.NavToolBar.Actions().Len() {
		s.NavToolBar.Actions().At(i).SetChecked(false)
	}

	// 該当INDEXのみチェックON
	s.CurrentIndex = index
	s.NavToolBar.Actions().At(index).SetChecked(true)

	// 物理焼き込み設定の情報を表示
	s.OriginalModelPicker.ChangePath(s.currentSet().OriginalModelPath)
	s.OriginalMotionPicker.ChangePath(s.currentSet().OriginalMotionPath)
	s.OutputModelPicker.ChangePath(s.currentSet().OutputModelPath)
	s.OutputMotionPicker.ChangePath(s.currentSet().OutputMotionPath)

	// TODO 他のも復元
}

func (s *WidgetStore) maxFrame() float32 {
	maxFrame := float32(0)
	for _, physicsSet := range s.BakeSets {
		if physicsSet.OriginalMotion != nil && maxFrame < physicsSet.OriginalMotion.MaxFrame() {
			maxFrame = physicsSet.OriginalMotion.MaxFrame()
		}
	}

	return maxFrame + 1
}

func (s *WidgetStore) currentSet() *entity.BakeSet {
	if s.CurrentIndex < 0 || s.CurrentIndex >= len(s.BakeSets) {
		return nil
	}

	return s.BakeSets[s.CurrentIndex]
}

func (s *WidgetStore) widgetList() []controller.IMWidget {
	return []controller.IMWidget{
		s.AddSetButton,
		s.ResetSetButton,
		s.LoadSetButton,
		s.SaveSetButton,
		s.OriginalModelPicker,
		s.OriginalMotionPicker,
		s.OutputModelPicker,
		s.OutputMotionPicker,
		s.Player,
		// s.BakeHistoryClearButton,
		s.SaveModelButton,
		s.SaveMotionButton,
		s.AddPhysicsButton,
		// s.AddRigidBodyPhysicsButton,
		// s.AddOutputButton,
	}
}
