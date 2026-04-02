package vt

import uv "github.com/charmbracelet/ultraviolet"

type reflowLine struct {
	cells []uv.Cell
}

type reflowPosition struct {
	logical int
	offset  int
}

func (s *Screen) resizeWider(width, height int, cursorPhantom bool) {
	if s.buf == nil {
		s.buf = uv.NewRenderBuffer(width, height)
		s.scroll = s.buf.Bounds()
		return
	}

	oldWidth, oldHeight := s.buf.Width(), s.buf.Height()
	if width <= oldWidth {
		s.resizePlain(width, height)
		return
	}

	if oldWidth <= 0 || oldHeight <= 0 {
		s.resizePlain(width, height)
		return
	}

	logical, cursorPos, savedPos := captureReflowState(s, oldWidth, oldHeight, cursorPhantom)
	wrapped, cursor, saved := wrapReflowState(logical, width, cursorPos, savedPos)

	next := uv.NewRenderBuffer(width, height)
	for y := 0; y < len(wrapped) && y < height; y++ {
		copy(next.Lines[y], wrapped[y])
	}

	s.buf = next
	s.scroll = s.buf.Bounds()
	s.cur.X, s.cur.Y = clampReflowCursor(cursor, width, height)
	s.saved.X, s.saved.Y = clampReflowCursor(saved, width, height)
	s.buf.Touched = nil
}

func (s *Screen) resizePlain(width, height int) {
	s.buf.Resize(width, height)
	s.buf.Touched = nil
	s.scroll = s.buf.Bounds()
}

func captureReflowState(s *Screen, width, height int, cursorPhantom bool) ([]reflowLine, reflowPosition, reflowPosition) {
	logical := make([]reflowLine, 0, height)
	rowLogical := make([]int, height)
	rowBase := make([]int, height)
	curRow := clampRow(s.cur.Y, height)
	savedRow := clampRow(s.saved.Y, height)

	for y := 0; y < height; y++ {
		preserveCols := make([]int, 0, 2)
		if y == curRow {
			col := s.cur.X
			if cursorPhantom {
				col = width
			}
			preserveCols = append(preserveCols, col)
		}
		if y == savedRow {
			preserveCols = append(preserveCols, s.saved.X)
		}

		line := s.buf.Line(y)
		cells := captureReflowCells(line, width, preserveCols...)
		// Treat a full-width row as a soft-wrap continuation when widening.
		if y == 0 || !screenLineUsesFullWidth(s.buf.Line(y-1), width) {
			logical = append(logical, reflowLine{cells: cells})
			rowLogical[y] = len(logical) - 1
			rowBase[y] = 0
			continue
		}

		rowLogical[y] = rowLogical[y-1]
		rowBase[y] = rowBase[y-1] + width
		logical[rowLogical[y]].cells = append(logical[rowLogical[y]].cells, cells...)
	}

	cursorCol := s.cur.X
	if cursorPhantom {
		cursorCol = width
	}
	return logical,
		reflowPosition{logical: rowLogical[curRow], offset: rowBase[curRow] + max(cursorCol, 0)},
		reflowPosition{logical: rowLogical[savedRow], offset: rowBase[savedRow] + max(s.saved.X, 0)}
}

func wrapReflowState(logical []reflowLine, width int, cursorPos, savedPos reflowPosition) ([]uv.Line, uv.Position, uv.Position) {
	if width <= 0 {
		width = 1
	}

	wrappedCounts := make([]int, len(logical))
	rows := make([]uv.Line, 0, len(logical))
	for i, line := range logical {
		wrappedLine := wrapReflowLine(line.cells, width)
		wrappedCounts[i] = len(wrappedLine)
		rows = append(rows, wrappedLine...)
	}

	cursor := reflowWrappedPosition(wrappedCounts, cursorPos, width)
	saved := reflowWrappedPosition(wrappedCounts, savedPos, width)
	return rows, cursor, saved
}

func reflowWrappedPosition(wrappedCounts []int, pos reflowPosition, width int) uv.Position {
	if len(wrappedCounts) == 0 {
		return uv.Pos(0, 0)
	}

	row := 0
	if pos.logical < 0 {
		pos.logical = 0
	}
	if pos.logical >= len(wrappedCounts) {
		pos.logical = len(wrappedCounts) - 1
	}
	for i := 0; i < pos.logical; i++ {
		row += wrappedCounts[i]
	}
	if pos.offset < 0 {
		pos.offset = 0
	}
	return uv.Pos(pos.offset%width, row+pos.offset/width)
}

func captureReflowCells(line uv.Line, width int, preserveCols ...int) []uv.Cell {
	if width <= 0 {
		return nil
	}

	limit := reflowLineEnd(line, width, preserveCols...)
	if limit == 0 {
		return nil
	}

	cells := make([]uv.Cell, 0, limit)
	col := 0
	for col < limit {
		cell := line.At(col)
		if cell == nil || cell.Width == 0 {
			col++
			continue
		}
		cells = append(cells, *cell)
		col += max(cell.Width, 1)
	}
	return cells
}

func reflowLineEnd(line uv.Line, width int, preserveCols ...int) int {
	end := 0
	col := 0
	for col < width {
		cell := line.At(col)
		if cell == nil {
			col++
			continue
		}
		if cell.Width == 0 {
			col++
			continue
		}
		cellWidth := max(cell.Width, 1)
		if cell.Content != "" || (!cell.IsZero() && !cell.Equal(&uv.EmptyCell)) {
			end = max(end, col+cellWidth)
		}
		col += cellWidth
	}
	for _, preserve := range preserveCols {
		if preserve < 0 {
			continue
		}
		end = max(end, min(preserve+1, width))
	}
	return end
}

func screenLineUsesFullWidth(line uv.Line, width int) bool {
	if line == nil || width <= 0 {
		return false
	}
	cell := line.At(width - 1)
	if cell == nil {
		return false
	}
	if cell.Width == 0 {
		return true
	}
	// Match tmux: any non-empty last column is treated as a soft wrap when
	// widening, even though right-aligned full-width rows can be false positives.
	return cell.Content != "" || (!cell.IsZero() && !cell.Equal(&uv.EmptyCell))
}

func wrapReflowLine(cells []uv.Cell, width int) []uv.Line {
	row := uv.NewLine(width)
	rows := make([]uv.Line, 0, 1)
	col := 0

	appendRow := func() {
		rows = append(rows, row)
		row = uv.NewLine(width)
		col = 0
	}

	if len(cells) == 0 {
		appendRow()
		return rows
	}

	for _, cell := range cells {
		cellWidth := max(cell.Width, 1)
		if col > 0 && col+cellWidth > width {
			appendRow()
		}
		row.Set(col, &cell)
		col += cellWidth
		if col >= width {
			appendRow()
		}
	}
	if col > 0 {
		appendRow()
	}
	return rows
}

func clampReflowCursor(pos uv.Position, width, height int) (int, int) {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	x := pos.X
	y := pos.Y
	if x < 0 {
		x = 0
	}
	if x >= width {
		x = width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= height {
		y = height - 1
	}
	return x, y
}

func clampRow(row, height int) int {
	if row < 0 {
		return 0
	}
	if row >= height {
		return height - 1
	}
	return row
}
