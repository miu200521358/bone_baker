package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/infrastructure"
	"github.com/miu200521358/bone_baker/pkg/usecase"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

func NewBakePage(mWidgets *controller.MWidgets) declarative.TabPage {
	var bakeTab *walk.TabPage

	// Repository パターンの依存性注入
	bakeSetRepository := infrastructure.NewFileBakeSetRepository()
	modelRepository := infrastructure.NewPmxModelRepository()
	motionRepository := infrastructure.NewVmdMotionRepository()
	bakeUsecase := usecase.NewBakeUsecase(bakeSetRepository, modelRepository, motionRepository)
	bakeState := NewBakeState(bakeUsecase)

	// WidgetFactoryを使用してウィジェット作成
	widgetFactory := NewWidgetFactory(bakeState, mWidgets)

	// ウィジェット群を作成
	widgetFactory.CreatePlayerWidget()
	widgetFactory.CreateFilePickerWidgets()
	widgetFactory.CreateButtonWidgets()

	mWidgets.Widgets = append(mWidgets.Widgets, bakeState.Player, bakeState.OriginalMotionPicker,
		bakeState.OriginalModelPicker, bakeState.OutputMotionPicker,
		bakeState.OutputModelPicker, bakeState.AddSetButton, bakeState.ResetSetButton,
		bakeState.LoadSetButton, bakeState.SaveSetButton, bakeState.SaveMotionButton,
		bakeState.SaveModelButton, bakeState.AddPhysicsButton, bakeState.AddOutputButton,
		bakeState.BakeHistoryClearButton)
	mWidgets.SetOnLoaded(func() {
		bakeState.BakeSets = append(bakeState.BakeSets, domain.NewPhysicsSet(len(bakeState.BakeSets)))
		bakeState.AddAction()
		bakeState.AddPhysicsButton.SetEnabled(false)
		bakeState.AddOutputButton.SetEnabled(false)
		bakeState.SaveModelButton.SetEnabled(false)
		bakeState.SaveMotionButton.SetEnabled(false)
	})

	return declarative.TabPage{
		Title:    mi18n.T("焼き込み"),
		AssignTo: &bakeTab,
		Layout:   declarative.VBox{},
		Background: declarative.SolidColorBrush{
			Color: controller.ColorTabBackground,
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout:  declarative.HBox{},
				MinSize: declarative.Size{Width: 200, Height: 40},
				MaxSize: declarative.Size{Width: 5120, Height: 40},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					bakeState.AddSetButton.Widgets(),
					bakeState.ResetSetButton.Widgets(),
					bakeState.LoadSetButton.Widgets(),
					bakeState.SaveSetButton.Widgets(),
				},
			},
			// セットスクロール
			declarative.ScrollView{
				Layout:        declarative.VBox{},
				MinSize:       declarative.Size{Width: 200, Height: 40},
				MaxSize:       declarative.Size{Width: 5120, Height: 40},
				VerticalFixed: true,
				Children: []declarative.Widget{
					// ナビゲーション用ツールバー
					declarative.ToolBar{
						AssignTo:           &bakeState.NavToolBar,
						MinSize:            declarative.Size{Width: 200, Height: 25},
						MaxSize:            declarative.Size{Width: 5120, Height: 25},
						DefaultButtonWidth: 200,
						Orientation:        walk.Horizontal,
						ButtonStyle:        declarative.ToolBarButtonTextOnly,
					},
				},
			},
			// セットごとの焼き込み内容
			declarative.ScrollView{
				Layout:  declarative.VBox{},
				MinSize: declarative.Size{Width: 126, Height: 350},
				MaxSize: declarative.Size{Width: 2560, Height: 5120},
				Children: []declarative.Widget{
					bakeState.OriginalModelPicker.Widgets(),
					bakeState.OriginalMotionPicker.Widgets(),
					declarative.VSeparator{},
					declarative.Composite{
						Layout:  declarative.HBox{},
						MinSize: declarative.Size{Width: 200, Height: 40},
						MaxSize: declarative.Size{Width: 2560, Height: 40},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("物理設定テーブル"),
								ToolTipText: mi18n.T("物理設定テーブル説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.ILT(mi18n.T("物理設定テーブル"), mi18n.T("物理設定テーブル説明"))
								},
							},
							declarative.HSpacer{},
							bakeState.AddPhysicsButton.Widgets(),
						},
					},
					widgetFactory.CreatePhysicsTableView(),
					declarative.VSeparator{},
					bakeState.OutputModelPicker.Widgets(),
					bakeState.SaveModelButton.Widgets(),
					declarative.VSeparator{},
					bakeState.OutputMotionPicker.Widgets(),
					declarative.Composite{
						Layout:   declarative.Grid{Columns: 4},
						Children: widgetFactory.CreateBakedHistoryWidgets(),
					},
					declarative.Composite{
						Layout:  declarative.HBox{},
						MinSize: declarative.Size{Width: 200, Height: 40},
						MaxSize: declarative.Size{Width: 2560, Height: 40},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("焼き込み保存設定テーブル"),
								ToolTipText: mi18n.T("焼き込み保存設定テーブル説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.ILT(mi18n.T("焼き込み保存設定テーブル"), mi18n.T("焼き込み保存設定テーブル説明"))
								},
							},
							declarative.HSpacer{},
							bakeState.AddOutputButton.Widgets(),
						},
					},
					widgetFactory.CreateOutputTableView(),
					bakeState.SaveMotionButton.Widgets(),
				},
			},
			bakeState.Player.Widgets(),
		},
	}
}
