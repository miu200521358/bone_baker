package ui

import (
	"fmt"
	"math"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/walk/pkg/declarative"
	"github.com/miu200521358/walk/pkg/walk"
)

// 固定レイアウト定数
const (
	headerHeight       = 20   // ヘッダー行の高さ
	rowHeight          = 50   // 各台形行の高さ
	trapezoidHeight    = 40   // 台形自体の高さ
	trapezoidMarginTop = 10   // 台形上マージン
	framesPer150px     = 100  // 150pxあたりのフレーム数（100F = 150px）
	pixelsPerFrame     = 1.5  // 1フレームあたりのピクセル数
	defaultMaxFrame    = 8000 // デフォルトの最大フレーム数
	cornerThreshold    = 10.0 // 四隅判定の閾値（ピクセル）
)

var trapezoidFillColor = walk.ColorSteelBlue       // 通常の台形色
var trapezoidHoverColor = walk.ColorDarkSlateBlue  // ホバー中の台形色
var trapezoidBorderColor = walk.ColorDarkSlateBlue // 台形の輪郭線
var backgroundColor = walk.ColorWhite              // 背景色
var borderColor = walk.ColorDarkGray               // 台形の枠線色
var fontColor = walk.RGB(30, 30, 30)               // フォント色

// RigidBodyTable グラフィカルモデル物理設定テーブル
type RigidBodyTable struct {
	*walk.CustomWidget
	records      []*entity.RigidBodyRecord
	maxFrame     float32
	store        *WidgetStore
	hoveredIndex int // ホバー中の台形インデックス (-1: なし)
}

// createRigidBodyTable グラフィカルモデル物理テーブル作成
func createRigidBodyTable(store *WidgetStore) declarative.Widget {
	return declarative.ScrollView{
		Layout:  declarative.HBox{},
		MinSize: declarative.Size{Width: 200, Height: 180},
		MaxSize: declarative.Size{Width: 200, Height: 180},
		Children: []declarative.Widget{
			declarative.CustomWidget{
				AssignTo: &store.RigidBodyTableWidget,
				MinSize:  declarative.Size{Width: defaultMaxFrame * pixelsPerFrame, Height: 180},
				Paint: func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
					return drawGraphicalRigidBodyTable(canvas, updateBounds, store)
				},
				OnMouseDown: func(x, y int, button walk.MouseButton) {
					handleRigidBodyTableMouseDown(x, y, button, store)
				},
				OnMouseMove: func(x, y int, button walk.MouseButton) {
					handleRigidBodyTableMouseMove(x, y, button, store)
				},
			},
		},
	}
}

// handleRigidBodyTableMouseDown マウスクリックイベントハンドラ
func handleRigidBodyTableMouseDown(x, y int, button walk.MouseButton, store *WidgetStore) {
	if button != walk.LeftButton {
		return
	}

	// 現在のセットからモデル物理レコードを取得
	currentSet := store.currentSet()
	if currentSet == nil || len(currentSet.RigidBodyRecords) == 0 {
		return
	}

	records := currentSet.RigidBodyRecords
	maxFrame := store.maxFrame()
	if maxFrame <= 0 {
		maxFrame = defaultMaxFrame
	}

	// クリック位置の台形を判定
	clickedIndex := getTrapezoidAtPosition(x, y, records, maxFrame)
	if clickedIndex >= 0 && clickedIndex < len(records) {
		// クリックされたレコードを取得
		record := records[clickedIndex]
		// ダイアログ表示（編集モード）
		dialog := NewRigidBodyTableViewDialog(store)
		dialog.Show(record, clickedIndex)
	}
}

// handleRigidBodyTableMouseMove マウス移動イベントハンドラ
func handleRigidBodyTableMouseMove(x, y int, button walk.MouseButton, store *WidgetStore) {
	// 現在のセットからモデル物理レコードを取得
	currentSet := store.currentSet()
	if currentSet == nil || len(currentSet.RigidBodyRecords) == 0 {
		// データがない場合はツールチップをクリア
		if store.RigidBodyTableWidget != nil {
			store.RigidBodyTableWidget.SetToolTipText("")
		}
		return
	}

	records := currentSet.RigidBodyRecords
	maxFrame := store.maxFrame()
	if maxFrame <= 0 {
		maxFrame = defaultMaxFrame
	}

	// ホバー位置の台形を判定
	trapezoidIndex := getTrapezoidAtPosition(x, y, records, maxFrame)
	if trapezoidIndex < 0 {
		// 台形外の場合はツールチップをクリア
		if store.RigidBodyTableWidget != nil {
			store.RigidBodyTableWidget.SetToolTipText("")
		}
		return
	}

	record := records[trapezoidIndex]

	// 台形の座標を計算
	rowStartY := headerHeight + trapezoidIndex*rowHeight
	y1 := float32(rowStartY + trapezoidMarginTop)                   // 上辺
	y2 := float32(rowStartY + trapezoidMarginTop + trapezoidHeight) // 下辺
	x1 := record.StartFrame * pixelsPerFrame                        // 下部始点
	x2 := record.MaxStartFrame * pixelsPerFrame                     // 上部始点
	x3 := record.MaxEndFrame * pixelsPerFrame                       // 上部終点
	x4 := record.EndFrame * pixelsPerFrame                          // 下部終点

	// 四隅判定とツールチップ設定
	tooltipText := getTrapezoidTooltipText(float32(x), float32(y), record, x1, x2, x3, x4, y1, y2)

	if store.RigidBodyTableWidget != nil {
		store.RigidBodyTableWidget.SetToolTipText(tooltipText)
	}
}

// getTrapezoidAtPosition 指定座標にある台形のインデックス取得
func getTrapezoidAtPosition(x, y int, records []*entity.RigidBodyRecord, maxFrame float32) int {
	if len(records) == 0 {
		return -1
	}

	for i, record := range records {
		// 固定サイズレイアウトでの座標計算
		rowStartY := headerHeight + i*rowHeight
		y1 := float32(rowStartY + trapezoidMarginTop)                   // 上辺
		y2 := float32(rowStartY + trapezoidMarginTop + trapezoidHeight) // 下辺

		// Y座標チェック
		if float32(y) < y1 || float32(y) > y2 {
			continue
		}

		// X座標チェック（台形内部判定）
		x1 := record.StartFrame * pixelsPerFrame    // 下部始点
		x2 := record.MaxStartFrame * pixelsPerFrame // 上部始点
		x3 := record.MaxEndFrame * pixelsPerFrame   // 上部終点
		x4 := record.EndFrame * pixelsPerFrame      // 下部終点

		if isInsideTrapezoidPosition(float32(x), float32(y), x1, x2, x3, x4, y1, y2) {
			return i
		}
	}

	return -1
}

// isInsideTrapezoidPosition 台形内部判定
func isInsideTrapezoidPosition(x, y, x1, x2, x3, x4, y1, y2 float32) bool {
	// 簡単な矩形範囲チェック（厳密な台形判定は複雑なため）
	minX := float32(math.Min(math.Min(float64(x1), float64(x2)), math.Min(float64(x3), float64(x4))))
	maxX := float32(math.Max(math.Max(float64(x1), float64(x2)), math.Max(float64(x3), float64(x4))))

	return x >= minX && x <= maxX && y >= y1 && y <= y2
}

// getTrapezoidTooltipText 台形のツールチップテキストを取得
func getTrapezoidTooltipText(x, y float32, record *entity.RigidBodyRecord, x1, x2, x3, x4, y1, y2 float32) string {
	// 四隅判定
	// x1, y2: 下部開始点 (StartFrame)
	// x2, y1: 上部開始点 (MaxStartFrame)
	// x3, y1: 上部終了点 (MaxEndFrame)
	// x4, y2: 下部終了点 (EndFrame)

	// 下部開始点 (StartFrame)
	if isNearCorner(x, y, x1, y2, cornerThreshold) {
		return fmt.Sprintf("%s: %.0fF", mi18n.T("開始フレーム"), record.StartFrame)
	}

	// 上部開始点 (MaxStartFrame)
	if isNearCorner(x, y, x2, y1, cornerThreshold) {
		return fmt.Sprintf("%s: %.0fF", mi18n.T("最大開始フレーム"), record.MaxStartFrame)
	}

	// 上部終了点 (MaxEndFrame)
	if isNearCorner(x, y, x3, y1, cornerThreshold) {
		return fmt.Sprintf("%s: %.0fF", mi18n.T("最大終了フレーム"), record.MaxEndFrame)
	}

	// 下部終了点 (EndFrame)
	if isNearCorner(x, y, x4, y2, cornerThreshold) {
		return fmt.Sprintf("%s: %.0fF", mi18n.T("終了フレーム"), record.EndFrame)
	}

	// 四隅以外の台形内部の場合はItemNames()を実行
	return record.ItemNames()
}

// isNearCorner 指定座標が角の近くにあるかを判定
func isNearCorner(x, y, cornerX, cornerY, threshold float32) bool {
	dx := x - cornerX
	dy := y - cornerY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	return distance <= threshold
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

// drawGraphicalRigidBodyTable グラフィカルモデル物理テーブル描画
func drawGraphicalRigidBodyTable(canvas *walk.Canvas, bounds walk.Rectangle, store *WidgetStore) error {
	// 背景をクリア
	brush, err := walk.NewSolidColorBrush(backgroundColor)
	if err != nil {
		return err
	}
	defer brush.Dispose()
	canvas.FillRectanglePixels(brush, bounds)

	// 現在のセットからモデル物理レコードを取得
	currentSet := store.currentSet()
	if currentSet == nil || len(currentSet.RigidBodyRecords) == 0 {
		// データがない場合は説明文を表示
		font, _ := walk.NewFont("MS UI Gothic", 10, 0)
		if font != nil {
			defer font.Dispose()
			canvas.DrawText(mi18n.T("モデル物理設定レコードがありません"), font, fontColor,
				walk.Rectangle{X: 10, Y: 10, Width: bounds.Width - 20, Height: 30}, walk.TextLeft)
		}
		return nil
	}

	records := currentSet.RigidBodyRecords
	maxFrame := store.maxFrame()
	if maxFrame <= 0 {
		maxFrame = defaultMaxFrame
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
	gridPen, err := walk.NewCosmeticPen(walk.PenSolid, borderColor)
	if err != nil {
		return err
	}
	defer gridPen.Dispose()

	// 横線（固定サイズレイアウト）
	// ヘッダー行の下辺
	canvas.DrawLine(gridPen, walk.Point{X: 0, Y: headerHeight}, walk.Point{X: bounds.Width, Y: headerHeight})

	// 各台形行の下辺
	for i := 0; i < len(records); i++ {
		y := headerHeight + (i+1)*rowHeight
		canvas.DrawLine(gridPen, walk.Point{X: 0, Y: y}, walk.Point{X: bounds.Width, Y: y})
	}

	// 縦線（100F刻み）
	for frame := float32(0); frame <= maxFrame; frame += framesPer150px {
		x := int(frame * pixelsPerFrame)
		canvas.DrawLine(gridPen, walk.Point{X: x, Y: 0}, walk.Point{X: x, Y: bounds.Height})

		// フレーム番号表示（ヘッダー行の中央に配置）
		if frame > 0 {
			font, _ := walk.NewFont("MS UI Gothic", 8, 0)
			if font != nil {
				defer font.Dispose()
				canvas.DrawText(fmt.Sprintf("%.0fF", frame), font, fontColor,
					walk.Rectangle{X: x - 15, Y: 5, Width: 30, Height: headerHeight - 10}, walk.TextCenter)
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

	for i, record := range records {
		// 台形の座標計算（固定サイズレイアウト）
		rowStartY := headerHeight + i*rowHeight
		y1 := rowStartY + trapezoidMarginTop                   // 上辺
		y2 := rowStartY + trapezoidMarginTop + trapezoidHeight // 下辺
		x1 := int(record.StartFrame * pixelsPerFrame)          // 下部始点
		x2 := int(record.MaxStartFrame * pixelsPerFrame)       // 上部始点
		x3 := int(record.MaxEndFrame * pixelsPerFrame)         // 上部終点
		x4 := int(record.EndFrame * pixelsPerFrame)            // 下部終点

		// 台形の塗りつぶし（水平線を密に描画して塗りつぶし効果を実現）
		fillPen, err := walk.NewCosmeticPen(walk.PenSolid, trapezoidFillColor)
		if err != nil {
			continue
		}
		defer fillPen.Dispose()

		// 台形の塗りつぶし処理
		for y := y1; y <= y2; y++ {
			// 現在のy座標での台形の左端と右端を計算
			progress := float32(y-y1) / float32(y2-y1) // 0.0から1.0の進行度

			// 左辺の計算（上部始点から下部始点への線形補間）
			leftX := int(float32(x2) + progress*(float32(x1-x2)))
			// 右辺の計算（上部終点から下部終点への線形補間）
			rightX := int(float32(x3) + progress*(float32(x4-x3)))

			// 水平線を描画
			if rightX > leftX {
				canvas.DrawLine(fillPen, walk.Point{X: leftX, Y: y}, walk.Point{X: rightX, Y: y})
			}
		}

		// 台形の輪郭線を描画
		pen, err := walk.NewCosmeticPen(walk.PenSolid, trapezoidBorderColor)
		if err != nil {
			continue
		}
		defer pen.Dispose()

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

		// 台形の下罫線を強調表示
		bottomLinePen, err := walk.NewCosmeticPen(walk.PenSolid, borderColor)
		if err != nil {
			continue
		}
		defer bottomLinePen.Dispose()
		canvas.DrawLine(bottomLinePen, walk.Point{X: x1, Y: y2}, walk.Point{X: x4, Y: y2})

		// レコード番号表示
		font, _ := walk.NewFont("MS UI Gothic", 9, 0)
		if font != nil {
			defer font.Dispose()
			text := fmt.Sprintf("%02d", i+1)
			textY := y1 + (y2-y1)/2 - 10
			canvas.DrawText(text, font, fontColor,
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
	brush, err := walk.NewSolidColorBrush(backgroundColor)
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
	gridPen, err := walk.NewCosmeticPen(walk.PenSolid, borderColor)
	if err != nil {
		return err
	}
	defer gridPen.Dispose()

	// 横線（固定サイズレイアウト）
	// ヘッダー行の下辺
	canvas.DrawLine(gridPen, walk.Point{X: 0, Y: headerHeight}, walk.Point{X: bounds.Width, Y: headerHeight})

	// 各台形行の下辺
	for i := 0; i < len(g.records); i++ {
		y := headerHeight + (i+1)*rowHeight
		canvas.DrawLine(gridPen, walk.Point{X: 0, Y: y}, walk.Point{X: bounds.Width, Y: y})
	}

	// 縦線（100F刻み）
	for frame := float32(0); frame <= g.maxFrame; frame += framesPer150px {
		x := int(frame * pixelsPerFrame)
		canvas.DrawLine(gridPen, walk.Point{X: x, Y: 0}, walk.Point{X: x, Y: bounds.Height})

		// フレーム番号表示（ヘッダー行の中央に配置）
		if frame > 0 {
			font := g.Font()
			if font == nil {
				font, _ = walk.NewFont("MS UI Gothic", 8, 0)
			}
			canvas.DrawText(fmt.Sprintf("%.0fF", frame), font, fontColor,
				walk.Rectangle{X: x - 15, Y: 5, Width: 30, Height: headerHeight - 10}, walk.TextCenter)
		}
	}

	return nil
}

// drawTrapezoids 台形描画
func (g *RigidBodyTable) drawTrapezoids(canvas *walk.Canvas, bounds walk.Rectangle) error {
	if len(g.records) == 0 {
		return nil
	}

	for i, record := range g.records {
		// 台形の色（ホバー中は強調）
		fillColor := trapezoidFillColor
		if i == g.hoveredIndex {
			fillColor = trapezoidHoverColor
		}

		// 台形の座標計算（固定サイズレイアウト）
		rowStartY := headerHeight + i*rowHeight
		y1 := rowStartY + trapezoidMarginTop                   // 上辺
		y2 := rowStartY + trapezoidMarginTop + trapezoidHeight // 下辺
		x1 := int(record.StartFrame * pixelsPerFrame)          // 下部始点
		x2 := int(record.MaxStartFrame * pixelsPerFrame)       // 上部始点
		x3 := int(record.MaxEndFrame * pixelsPerFrame)         // 上部終点
		x4 := int(record.EndFrame * pixelsPerFrame)            // 下部終点

		// 台形の塗りつぶし（水平線を密に描画して塗りつぶし効果を実現）
		fillPen, err := walk.NewCosmeticPen(walk.PenSolid, fillColor)
		if err != nil {
			continue
		}
		defer fillPen.Dispose()

		// 台形の塗りつぶし処理
		for y := y1; y <= y2; y++ {
			// 現在のy座標での台形の左端と右端を計算
			progress := float32(y-y1) / float32(y2-y1) // 0.0から1.0の進行度

			// 左辺の計算（上部始点から下部始点への線形補間）
			leftX := int(float32(x2) + progress*(float32(x1-x2)))
			// 右辺の計算（上部終点から下部終点への線形補間）
			rightX := int(float32(x3) + progress*(float32(x4-x3)))

			// 水平線を描画
			if rightX > leftX {
				canvas.DrawLine(fillPen, walk.Point{X: leftX, Y: y}, walk.Point{X: rightX, Y: y})
			}
		}

		// 台形の輪郭線を描画
		pen, err := walk.NewCosmeticPen(walk.PenSolid, trapezoidBorderColor)
		if err != nil {
			continue
		}
		defer pen.Dispose()

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

		// 台形の下罫線を強調表示
		bottomLinePen, err := walk.NewCosmeticPen(walk.PenSolid, borderColor)
		if err != nil {
			continue
		}
		defer bottomLinePen.Dispose()
		canvas.DrawLine(bottomLinePen, walk.Point{X: x1, Y: y2}, walk.Point{X: x4, Y: y2})

		// レコード番号表示
		font := g.Font()
		if font == nil {
			font, _ = walk.NewFont("MS UI Gothic", 9, 0)
		}
		text := fmt.Sprintf("No.%d", i+1)
		textY := y1 + (y2-y1)/2 - 8
		canvas.DrawText(text, font, fontColor,
			walk.Rectangle{X: 5, Y: textY, Width: 50, Height: 16}, walk.TextLeft)
	}

	return nil
}

// getTrapezoidAt 指定座標にある台形のインデックス取得
func (g *RigidBodyTable) getTrapezoidAt(x, y int, bounds walk.Rectangle) int {
	if len(g.records) == 0 {
		return -1
	}

	for i, record := range g.records {
		// 固定サイズレイアウトでの座標計算
		rowStartY := headerHeight + i*rowHeight
		y1 := float32(rowStartY + trapezoidMarginTop)                   // 上辺
		y2 := float32(rowStartY + trapezoidMarginTop + trapezoidHeight) // 下辺

		// Y座標チェック
		if float32(y) < y1 || float32(y) > y2 {
			continue
		}

		// X座標チェック（台形内部判定）
		x1 := record.StartFrame * pixelsPerFrame    // 下部始点
		x2 := record.MaxStartFrame * pixelsPerFrame // 上部始点
		x3 := record.MaxEndFrame * pixelsPerFrame   // 上部終点
		x4 := record.EndFrame * pixelsPerFrame      // 下部終点

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
