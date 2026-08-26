package vt

import (
	"fmt"
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

// rowBoundary describes how a physical row begins. Keeping this as one value
// makes impossible states such as "soft and truncated" unrepresentable.
type rowBoundary uint8

const (
	boundaryHard rowBoundary = iota
	boundarySoft
	boundaryTruncatedHead
)

func (b rowBoundary) valid() bool {
	return b == boundaryHard || b == boundarySoft || b == boundaryTruncatedHead
}

// semanticRow is the value projection used when rows cross the visible-grid
// and scrollback boundary. The contained line is always cloned at ownership
// boundaries.
type semanticRow struct {
	line     uv.Line
	boundary rowBoundary
}

func cloneSemanticRow(row semanticRow) semanticRow {
	return semanticRow{line: slices.Clone(row.line), boundary: row.boundary}
}

// semanticBuffer is the sole mutation owner for visible cells and their row
// boundaries. The backing RenderBuffer may be wider than width on the
// alternate screen so clipped cells survive a shrink-grow cycle.
type semanticBuffer struct {
	cells      *uv.RenderBuffer
	width      int
	boundaries []rowBoundary
}

func newSemanticBuffer(width, height int) *semanticBuffer {
	return &semanticBuffer{
		cells:      uv.NewRenderBuffer(width, height),
		width:      width,
		boundaries: make([]rowBoundary, height),
	}
}

func (b *semanticBuffer) bounds() uv.Rectangle {
	return uv.Rect(0, 0, b.width, b.height())
}

func (b *semanticBuffer) widthValue() int { return b.width }

func (b *semanticBuffer) height() int {
	if b == nil || b.cells == nil {
		return 0
	}
	return b.cells.Height()
}

func (b *semanticBuffer) visibleLine(y int) uv.Line {
	if b == nil || b.cells == nil {
		return nil
	}
	line := b.cells.Line(y)
	if len(line) > b.width {
		line = line[:b.width]
	}
	return line
}

func (b *semanticBuffer) cellAt(x, y int) *uv.Cell {
	if x < 0 || x >= b.width {
		return nil
	}
	cell := b.cells.CellAt(x, y)
	if cell == nil {
		return nil
	}
	return cell.Clone()
}

func (b *semanticBuffer) setCell(x, y int, cell *uv.Cell) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height() {
		return
	}
	b.cells.SetCell(x, y, cell)
}

func (b *semanticBuffer) touched() []*uv.LineData {
	if b == nil || b.cells == nil {
		return nil
	}
	result := make([]*uv.LineData, len(b.cells.Touched))
	for i, line := range b.cells.Touched {
		if line != nil {
			copy := *line
			result[i] = &copy
		}
	}
	return result
}

func (b *semanticBuffer) clearTouched() { b.cells.Touched = nil }

func (b *semanticBuffer) touchLine(x, y, n int) { b.cells.TouchLine(x, y, n) }

func (b *semanticBuffer) reset() {
	b.cells.Clear()
	clear(b.boundaries)
	b.cells.Touched = nil
}

func (b *semanticBuffer) clearArea(area uv.Rectangle) {
	area = area.Intersect(b.bounds())
	if area.Empty() {
		return
	}
	b.cells.ClearArea(area)
	b.hardenFullyReplacedRows(area)
}

func (b *semanticBuffer) fillArea(cell *uv.Cell, area uv.Rectangle) {
	area = area.Intersect(b.bounds())
	if area.Empty() {
		return
	}
	b.cells.FillArea(cell, area)
	b.hardenFullyReplacedRows(area)
}

func (b *semanticBuffer) hardenFullyReplacedRows(area uv.Rectangle) {
	if area.Min.X != 0 || area.Max.X < b.width {
		return
	}
	for y := area.Min.Y; y < area.Max.Y; y++ {
		b.setBoundary(y, boundaryHard)
		// A continuation cannot cross a row whose contents were replaced.
		b.setBoundary(y+1, boundaryHard)
	}
}

func (b *semanticBuffer) insertCells(x, y, n int, cell *uv.Cell, area uv.Rectangle) {
	b.cells.InsertCellArea(x, y, n, cell, area.Intersect(b.bounds()))
}

func (b *semanticBuffer) deleteCells(x, y, n int, cell *uv.Cell, area uv.Rectangle) {
	b.cells.DeleteCellArea(x, y, n, cell, area.Intersect(b.bounds()))
}

func (b *semanticBuffer) insertRows(y, n int, cell *uv.Cell, area uv.Rectangle) {
	area = area.Intersect(b.bounds())
	b.cells.InsertLineArea(y, n, cell, area)
	b.shiftBoundariesDown(y, n, area)
}

func (b *semanticBuffer) deleteRows(y, n int, cell *uv.Cell, area uv.Rectangle) {
	area = area.Intersect(b.bounds())
	b.cells.DeleteLineArea(y, n, cell, area)
	b.shiftBoundariesUp(y, n, area)
}

func (b *semanticBuffer) rows(y, n int) []semanticRow {
	if y < 0 || y >= b.height() || n <= 0 {
		return nil
	}
	n = min(n, b.height()-y)
	rows := make([]semanticRow, n)
	for i := range n {
		rows[i] = semanticRow{
			line:     slices.Clone(b.visibleLine(y + i)),
			boundary: b.boundary(y + i),
		}
	}
	return rows
}

func (b *semanticBuffer) allRows() []semanticRow { return b.rows(0, b.height()) }

func (b *semanticBuffer) boundary(y int) rowBoundary {
	if y < 0 || y >= len(b.boundaries) {
		return boundaryHard
	}
	return b.boundaries[y]
}

func (b *semanticBuffer) setBoundary(y int, boundary rowBoundary) {
	if y < 0 || y >= len(b.boundaries) {
		return
	}
	b.boundaries[y] = boundary
}

func (b *semanticBuffer) shiftBoundariesDown(y, n int, area uv.Rectangle) {
	maxY := min(area.Max.Y, len(b.boundaries))
	n = min(n, maxY-y)
	if n <= 0 || y < 0 || y >= maxY {
		return
	}
	if area.Min.X != 0 || area.Max.X != b.width {
		for row := y; row < maxY; row++ {
			b.boundaries[row] = boundaryHard
		}
		return
	}
	copy(b.boundaries[y+n:maxY], b.boundaries[y:maxY-n])
	clear(b.boundaries[y : y+n])
}

func (b *semanticBuffer) shiftBoundariesUp(y, n int, area uv.Rectangle) {
	maxY := min(area.Max.Y, len(b.boundaries))
	n = min(n, maxY-y)
	if n <= 0 || y < 0 || y >= maxY {
		return
	}
	if area.Min.X != 0 || area.Max.X != b.width {
		for row := y; row < maxY; row++ {
			b.boundaries[row] = boundaryHard
		}
		return
	}
	copy(b.boundaries[y:maxY-n], b.boundaries[y+n:maxY])
	clear(b.boundaries[maxY-n : maxY])
}

func (b *semanticBuffer) replace(width, height int, rows []semanticRow) {
	next := uv.NewRenderBuffer(width, height)
	boundaries := make([]rowBoundary, height)
	for y := 0; y < min(height, len(rows)); y++ {
		copy(next.Line(y), rows[y].line)
		boundaries[y] = rows[y].boundary
	}
	next.Touched = nil
	b.cells = next
	b.width = width
	b.boundaries = boundaries
}

// resizePhysical resizes an alternate-screen grid without semantic grouping.
// It returns the row shift applied to cursor-bearing coordinates.
func (b *semanticBuffer) resizePhysical(width, height, cursorY int) int {
	if width <= 0 || height <= 0 {
		return 0
	}
	storageWidth := b.cells.Width()
	if width > storageWidth {
		b.cells.Resize(width, b.height())
		storageWidth = width
	}
	b.width = width

	oldHeight := b.height()
	if height == oldHeight {
		b.cells.Touched = nil
		return 0
	}

	start, end, shift := 0, oldHeight, 0
	if height < oldHeight {
		remove := oldHeight - height
		belowCursor := max(0, oldHeight-1-cursorY)
		removeBelow := min(remove, belowCursor)
		removeAbove := remove - removeBelow
		start = removeAbove
		end = oldHeight - removeBelow
		shift = -removeAbove
	}

	next := uv.NewRenderBuffer(storageWidth, height)
	boundaries := make([]rowBoundary, height)
	for dst, src := 0, start; dst < height && src < end; dst, src = dst+1, src+1 {
		copy(next.Line(dst), b.cells.Line(src))
		boundaries[dst] = b.boundary(src)
	}
	if start > 0 && len(boundaries) > 0 && boundaries[0] == boundarySoft {
		boundaries[0] = boundaryTruncatedHead
	}
	next.Touched = nil
	b.cells = next
	b.boundaries = boundaries
	return shift
}

func (b *semanticBuffer) string() string { return uv.Lines(b.visibleLines()).String() }

func (b *semanticBuffer) render() string { return uv.Lines(b.visibleLines()).Render() }

func (b *semanticBuffer) visibleLines() uv.Lines {
	lines := make(uv.Lines, b.height())
	for y := range lines {
		lines[y] = b.visibleLine(y)
	}
	return lines
}

func (b *semanticBuffer) validate() error {
	if len(b.boundaries) != b.height() {
		return fmt.Errorf("row boundary count %d does not match height %d", len(b.boundaries), b.height())
	}
	for y, boundary := range b.boundaries {
		if !boundary.valid() {
			return fmt.Errorf("row %d has invalid boundary %d", y, boundary)
		}
	}
	return nil
}
