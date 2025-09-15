package ui

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/interface/controller"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

func NewBakePage(mWidgets *controller.MWidgets) declarative.TabPage {
	var bakeTab *walk.TabPage

	store := NewWidgetStore(mWidgets)

	// ウィジェット群を作成
	store.createPlayerWidget()
	store.createFilePickerWidgets()
	store.createButtonWidgets()

	mWidgets.Widgets = append(mWidgets.Widgets, store.widgetList()...)
	mWidgets.SetOnLoaded(func() {
		store.BakeSets = append(store.BakeSets, entity.NewBakeSet(len(store.BakeSets)))
		store.AddAction()
		store.AddPhysicsButton.SetEnabled(false)
		store.AddWindButton.SetEnabled(false)
		store.AddRigidBodyButton.SetEnabled(false)
		store.AddOutputButton.SetEnabled(false)
		store.SaveModelButton.SetEnabled(false)
		store.SaveMotionButton.SetEnabled(false)
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
					store.AddSetButton.Widgets(),
					store.ResetSetButton.Widgets(),
					store.LoadSetButton.Widgets(),
					store.SaveSetButton.Widgets(),
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
						AssignTo:           &store.NavToolBar,
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
					store.OriginalModelPicker.Widgets(),
					store.OriginalMotionPicker.Widgets(),
					declarative.VSeparator{},
					declarative.Composite{
						Layout:  declarative.HBox{},
						MinSize: declarative.Size{Width: 200, Height: 40},
						MaxSize: declarative.Size{Width: 2560, Height: 40},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("ワールド物理設定テーブル"),
								ToolTipText: mi18n.T("ワールド物理設定テーブル説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.ILT(mi18n.T("ワールド物理設定テーブル"), mi18n.T("ワールド物理設定テーブル説明"))
								},
							},
							declarative.HSpacer{},
							store.AddPhysicsButton.Widgets(),
						},
					},
					createPhysicsTableView(store),
					declarative.Composite{
						Layout:  declarative.HBox{},
						MinSize: declarative.Size{Width: 200, Height: 40},
						MaxSize: declarative.Size{Width: 2560, Height: 40},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("モデル物理設定テーブル"),
								ToolTipText: mi18n.T("モデル物理設定テーブル説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.ILT(mi18n.T("モデル物理設定テーブル"), mi18n.T("モデル物理設定テーブル説明"))
								},
							},
							declarative.HSpacer{},
							store.AddRigidBodyButton.Widgets(),
						},
					},
					createRigidBodyTable(store),
					declarative.Composite{
						Layout:  declarative.HBox{},
						MinSize: declarative.Size{Width: 200, Height: 40},
						MaxSize: declarative.Size{Width: 2560, Height: 40},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("風設定テーブル"),
								ToolTipText: mi18n.T("風設定テーブル説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.ILT(mi18n.T("風設定テーブル"), mi18n.T("風設定テーブル説明"))
								},
							},
							declarative.HSpacer{},
							store.AddWindButton.Widgets(),
						},
					},
					createWindTableView(store),
					declarative.VSeparator{},
					declarative.Composite{
						Layout:   declarative.Grid{Columns: 4},
						Children: store.createBakedHistoryWidgets(),
					},
					declarative.Composite{
						Layout:  declarative.HBox{},
						MinSize: declarative.Size{Width: 200, Height: 40},
						MaxSize: declarative.Size{Width: 2560, Height: 40},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:        mi18n.T("出力設定テーブル"),
								ToolTipText: mi18n.T("出力設定テーブル説明"),
								OnMouseDown: func(x, y int, button walk.MouseButton) {
									mlog.ILT(mi18n.T("出力設定テーブル"), mi18n.T("出力設定テーブル説明"))
								},
							},
							declarative.HSpacer{},
							store.AddOutputButton.Widgets(),
						},
					},
					createOutputTableView(store),
					declarative.VSeparator{},
					store.OutputModelPicker.Widgets(),
					store.SaveModelButton.Widgets(),
					declarative.VSeparator{},
					store.OutputMotionPicker.Widgets(),
					store.SaveMotionButton.Widgets(),
				},
			},
			store.Player.Widgets(),
		},
	}
}
