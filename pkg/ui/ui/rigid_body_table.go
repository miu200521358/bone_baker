package ui

import (
	"fmt"
	"math"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// RigidBodyTable グラフィカル剛体設定テーブル
type RigidBodyTable struct {
	*walk.CustomWidget
	records      []*entity.RigidBodyRecord
	maxFrame     float32
	store        *WidgetStore
	hoveredIndex int // ホバー中の台形インデックス (-1: なし)
}

// createRigidBodyTable グラフィカル剛体テーブル作成
func createRigidBodyTable(store *WidgetStore) declarative.Widget {
	return declarative.CustomWidget{
		AssignTo: &store.RigidBodyTableWidget,
		MinSize:  declarative.Size{Width: 200, Height: 300},
		Paint: func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
			return drawGraphicalRigidBodyTable(canvas, updateBounds, store)
		},
	}
}

func createRigidBodyTableViewDialog(store *WidgetStore, isAdd bool) func() {
	return func() {
		var record *entity.RigidBodyRecord
		recordIndex := -1
		switch isAdd {
		case true:
			if store.currentSet().OriginalMotion == nil {
				record = entity.NewRigidBodyRecord(0, 0, store.currentSet().OriginalModel)
			} else {
				record = entity.NewRigidBodyRecord(
					store.currentSet().OriginalMotion.MinFrame(),
					store.currentSet().OriginalMotion.MaxFrame(),
					store.currentSet().OriginalModel)
			}
		case false:
			// record = store.currentSet().RigidBodyRecords[store.RigidBodyTableView.CurrentIndex()]
			// recordIndex = store.RigidBodyTableView.CurrentIndex()
		}
		dialog := NewRigidBodyTableViewDialog(store)
		dialog.Show(record, recordIndex)
	}
}

// drawGraphicalRigidBodyTable グラフィカル剛体テーブル描画
func drawGraphicalRigidBodyTable(canvas *walk.Canvas, bounds walk.Rectangle, store *WidgetStore) error {
	// 背景をクリア
	brush, err := walk.NewSolidColorBrush(walk.Color(0xFFFFFF))
	if err != nil {
		return err
	}
	defer brush.Dispose()
	canvas.FillRectanglePixels(brush, bounds)

	// 現在のセットから剛体レコードを取得
	currentSet := store.currentSet()
	if currentSet == nil || len(currentSet.RigidBodyRecords) == 0 {
		// データがない場合は説明文を表示
		font, _ := walk.NewFont("MS UI Gothic", 10, 0)
		if font != nil {
			defer font.Dispose()
			canvas.DrawText("剛体設定レコードがありません", font, walk.Color(0x666666),
				walk.Rectangle{X: 10, Y: 10, Width: bounds.Width - 20, Height: 30}, walk.TextLeft)
		}
		return nil
	}

	records := currentSet.RigidBodyRecords
	maxFrame := store.maxFrame()
	if maxFrame <= 0 {
		maxFrame = 1000 // デフォルト値
	}

	// グリッド描画
	if err := drawGrid(canvas, bounds, records, maxFrame); err != nil {
		return err
	}

	// 台形描画
	if err := drawTrapezoids(canvas, bounds, records, maxFrame); err != nil {
		return err
	}

	return nil
}

// drawGrid グリッド描画（グローバル関数）
func drawGrid(canvas *walk.Canvas, bounds walk.Rectangle, records []*entity.RigidBodyRecord, maxFrame float32) error {
	gridPen, err := walk.NewCosmeticPen(walk.PenSolid, walk.Color(0xCCCCCC))
	if err != nil {
		return err
	}
	defer gridPen.Dispose()

	// 横線（各レコード境界）
	rowHeight := float32(bounds.Height) / float32(len(records)+1) // +1 for header
	for i := 0; i <= len(records); i++ {
		y := int(float32(i) * rowHeight)
		canvas.DrawLine(gridPen, walk.Point{X: 0, Y: y}, walk.Point{X: bounds.Width, Y: y})
	}

	// 縦線（100F刻み）
	frameWidth := float32(bounds.Width) / maxFrame
	for frame := float32(0); frame <= maxFrame; frame += 100 {
		x := int(frame * frameWidth)
		canvas.DrawLine(gridPen, walk.Point{X: x, Y: 0}, walk.Point{X: x, Y: bounds.Height})

		// フレーム番号表示
		if frame > 0 {
			font, _ := walk.NewFont("MS UI Gothic", 8, 0)
			if font != nil {
				defer font.Dispose()
				canvas.DrawText(fmt.Sprintf("%.0fF", frame), font, walk.Color(0x000000),
					walk.Rectangle{X: x - 15, Y: 0, Width: 30, Height: 20}, walk.TextCenter)
			}
		}
	}

	return nil
}

// drawTrapezoids 台形描画（グローバル関数）
func drawTrapezoids(canvas *walk.Canvas, bounds walk.Rectangle, records []*entity.RigidBodyRecord, maxFrame float32) error {
	if len(records) == 0 {
		return nil
	}

	rowHeight := float32(bounds.Height) / float32(len(records)+1)
	frameWidth := float32(bounds.Width) / maxFrame

	for i, record := range records {
		// 台形の色
		pen, err := walk.NewCosmeticPen(walk.PenSolid, walk.Color(0x4A90E2))
		if err != nil {
			continue
		}
		defer pen.Dispose()

		// 台形の座標計算
		y1 := int(float32(i+1) * rowHeight)          // 上辺
		y2 := int(float32(i+2) * rowHeight)          // 下辺
		x1 := int(record.StartFrame * frameWidth)    // 下部始点
		x2 := int(record.MaxStartFrame * frameWidth) // 上部始点
		x3 := int(record.MaxEndFrame * frameWidth)   // 上部終点
		x4 := int(record.EndFrame * frameWidth)      // 下部終点

		// 台形描画
		points := []walk.Point{
			{X: x1, Y: y2}, // 下部始点
			{X: x2, Y: y1}, // 上部始点
			{X: x3, Y: y1}, // 上部終点
			{X: x4, Y: y2}, // 下部終点
		}

		// 台形を線で描画
		for j := 0; j < len(points); j++ {
			next := (j + 1) % len(points)
			canvas.DrawLine(pen, points[j], points[next])
		}

		// レコード番号表示
		font, _ := walk.NewFont("MS UI Gothic", 9, 0)
		if font != nil {
			defer font.Dispose()
			text := fmt.Sprintf("No.%d", i+1)
			textY := y1 + (y2-y1)/2 - 8
			canvas.DrawText(text, font, walk.Color(0x000000),
				walk.Rectangle{X: 5, Y: textY, Width: 50, Height: 16}, walk.TextLeft)
		}
	}

	return nil
}

// SetMaxFrame 最大フレーム設定
func (g *RigidBodyTable) SetMaxFrame(maxFrame float32) {
	g.maxFrame = maxFrame
	g.Invalidate()
}

// CurrentIndex 現在選択中のインデックス取得
func (g *RigidBodyTable) CurrentIndex() int {
	return g.hoveredIndex
}

// onPaint 描画処理
func (g *RigidBodyTable) onPaint(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
	// 背景をクリア
	bounds := g.ClientBounds()
	brush, err := walk.NewSolidColorBrush(walk.Color(0xFFFFFF))
	if err != nil {
		return err
	}
	defer brush.Dispose()
	canvas.FillRectanglePixels(brush, bounds)

	if g.maxFrame <= 0 || len(g.records) == 0 {
		return nil
	}

	// グリッド描画
	if err := g.drawGrid(canvas, bounds); err != nil {
		return err
	}

	// 台形描画
	if err := g.drawTrapezoids(canvas, bounds); err != nil {
		return err
	}

	return nil
}

// onMouseMove マウス移動処理
func (g *RigidBodyTable) onMouseMove(x, y int, button walk.MouseButton) {
	// ホバー中の台形を判定
	bounds := g.ClientBounds()
	oldHoveredIndex := g.hoveredIndex
	g.hoveredIndex = g.getTrapezoidAt(x, y, bounds)

	// ホバー状態が変わった場合は再描画
	if oldHoveredIndex != g.hoveredIndex {
		g.Invalidate()

		// ツールチップ設定
		if g.hoveredIndex >= 0 && g.hoveredIndex < len(g.records) {
			g.SetToolTipText(g.records[g.hoveredIndex].ItemNames())
		} else {
			g.SetToolTipText("")
		}
	}
}

// onMouseDown マウスクリック処理
func (g *RigidBodyTable) onMouseDown(x, y int, button walk.MouseButton) {
	if button != walk.LeftButton {
		return
	}

	bounds := g.ClientBounds()
	clickedIndex := g.getTrapezoidAt(x, y, bounds)

	if clickedIndex >= 0 && clickedIndex < len(g.records) {
		// クリックされたレコードを取得
		record := g.store.currentSet().RigidBodyRecords[clickedIndex]
		// ダイアログ表示（編集モード）
		dialog := NewRigidBodyTableViewDialog(g.store)
		dialog.Show(record, clickedIndex)
	}
}

// drawGrid グリッド描画
func (g *RigidBodyTable) drawGrid(canvas *walk.Canvas, bounds walk.Rectangle) error {
	gridPen, err := walk.NewCosmeticPen(walk.PenSolid, walk.Color(0xCCCCCC))
	if err != nil {
		return err
	}
	defer gridPen.Dispose()

	// 横線（各レコード境界）
	rowHeight := float32(bounds.Height) / float32(len(g.records)+1) // +1 for header
	for i := 0; i <= len(g.records); i++ {
		y := int(float32(i) * rowHeight)
		canvas.DrawLine(gridPen, walk.Point{X: 0, Y: y}, walk.Point{X: bounds.Width, Y: y})
	}

	// 縦線（100F刻み）
	frameWidth := float32(bounds.Width) / g.maxFrame
	for frame := float32(0); frame <= g.maxFrame; frame += 100 {
		x := int(frame * frameWidth)
		canvas.DrawLine(gridPen, walk.Point{X: x, Y: 0}, walk.Point{X: x, Y: bounds.Height})

		// フレーム番号表示
		if frame > 0 {
			font := g.Font()
			if font == nil {
				font, _ = walk.NewFont("MS UI Gothic", 8, 0)
			}
			canvas.DrawText(fmt.Sprintf("%.0fF", frame), font, walk.Color(0x000000),
				walk.Rectangle{X: x - 15, Y: 0, Width: 30, Height: 20}, walk.TextCenter)
		}
	}

	return nil
}

// drawTrapezoids 台形描画
func (g *RigidBodyTable) drawTrapezoids(canvas *walk.Canvas, bounds walk.Rectangle) error {
	if len(g.records) == 0 {
		return nil
	}

	rowHeight := float32(bounds.Height) / float32(len(g.records)+1)
	frameWidth := float32(bounds.Width) / g.maxFrame

	for i, record := range g.records {
		// 台形の色（ホバー中は強調）
		fillColor := walk.Color(0x4A90E2) // 青系
		if i == g.hoveredIndex {
			fillColor = walk.Color(0x357ABD) // より濃い青
		}

		brush, err := walk.NewSolidColorBrush(fillColor)
		if err != nil {
			continue
		}
		defer brush.Dispose()

		pen, err := walk.NewCosmeticPen(walk.PenSolid, walk.Color(0x2E5A87))
		if err != nil {
			continue
		}
		defer pen.Dispose()

		// 台形の座標計算
		y1 := int(float32(i+1) * rowHeight)          // 上辺
		y2 := int(float32(i+2) * rowHeight)          // 下辺
		x1 := int(record.StartFrame * frameWidth)    // 下部始点
		x2 := int(record.MaxStartFrame * frameWidth) // 上部始点
		x3 := int(record.MaxEndFrame * frameWidth)   // 上部終点
		x4 := int(record.EndFrame * frameWidth)      // 下部終点

		// 台形描画
		points := []walk.Point{
			{X: x1, Y: y2}, // 下部始点
			{X: x2, Y: y1}, // 上部始点
			{X: x3, Y: y1}, // 上部終点
			{X: x4, Y: y2}, // 下部終点
		}

		// 台形を線で描画（Walkライブラリに多角形塗りつぶしがないため線で近似）
		for i := 0; i < len(points); i++ {
			next := (i + 1) % len(points)
			canvas.DrawLine(pen, points[i], points[next])
		}

		// レコード番号表示
		font := g.Font()
		if font == nil {
			font, _ = walk.NewFont("MS UI Gothic", 9, 0)
		}
		text := fmt.Sprintf("No.%d", i+1)
		textY := y1 + (y2-y1)/2 - 8
		canvas.DrawText(text, font, walk.Color(0x000000),
			walk.Rectangle{X: 5, Y: textY, Width: 50, Height: 16}, walk.TextLeft)
	}

	return nil
}

// getTrapezoidAt 指定座標にある台形のインデックス取得
func (g *RigidBodyTable) getTrapezoidAt(x, y int, bounds walk.Rectangle) int {
	if len(g.records) == 0 {
		return -1
	}

	rowHeight := float32(bounds.Height) / float32(len(g.records)+1)
	frameWidth := float32(bounds.Width) / g.maxFrame

	for i, record := range g.records {
		y1 := float32(i+1) * rowHeight
		y2 := float32(i+2) * rowHeight

		// Y座標チェック
		if float32(y) < y1 || float32(y) > y2 {
			continue
		}

		// X座標チェック（台形内部判定）
		x1 := record.StartFrame * frameWidth    // 下部始点
		x2 := record.MaxStartFrame * frameWidth // 上部始点
		x3 := record.MaxEndFrame * frameWidth   // 上部終点
		x4 := record.EndFrame * frameWidth      // 下部終点

		if g.isInsideTrapezoid(float32(x), float32(y), x1, x2, x3, x4, y1, y2) {
			return i
		}
	}

	return -1
}

// isInsideTrapezoid 台形内部判定
func (g *RigidBodyTable) isInsideTrapezoid(x, y, x1, x2, x3, x4, y1, y2 float32) bool {
	// 簡単な矩形範囲チェック（厳密な台形判定は複雑なため）
	minX := float32(math.Min(math.Min(float64(x1), float64(x2)), math.Min(float64(x3), float64(x4))))
	maxX := float32(math.Max(math.Max(float64(x1), float64(x2)), math.Max(float64(x3), float64(x4))))

	return x >= minX && x <= maxX && y >= y1 && y <= y2
}
