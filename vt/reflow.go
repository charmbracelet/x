package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

type logicalPosition struct {
	line   int
	offset int
}

// resize reflows the screen and its scrollback while preserving the cursor's
// position in the logical line. It reports whether the cursor remains in the
// pending-wrap position at the new width.
func (s *Screen) resize(width, height int, cursorPastEnd bool) bool {
	if s.buf == nil {
		s.resizeBuffer(width, height)
		return false
	}
	if width == s.buf.Width() && height == s.buf.Height() {
		s.scroll = s.buf.Bounds()
		return cursorPastEnd
	}

	scrollbackLen := 0
	if s.scrollback != nil {
		scrollbackLen = len(s.scrollback.lines)
	}

	lastScreenLine := max(s.cur.Y, s.saved.Y)
	for y := s.buf.Height() - 1; y >= 0; y-- {
		if !s.isLineEmpty(s.buf.Line(y)) {
			lastScreenLine = max(lastScreenLine, y)
			break
		}
	}

	lines := make([]uv.Line, 0, scrollbackLen+lastScreenLine+1)
	wrapped := make([]bool, 0, cap(lines))
	wrapWidth := make([]int, 0, cap(lines))
	if s.scrollback != nil {
		lines = append(lines, s.scrollback.lines...)
		wrapped = append(wrapped, s.scrollback.wrapped...)
		wrapWidth = append(wrapWidth, s.scrollback.wrapWidth...)
	}
	for y := 0; y <= lastScreenLine && y < s.buf.Height(); y++ {
		lines = append(lines, s.buf.Line(y))
		wrapped = append(wrapped, s.wrapped[y])
		wrapWidth = append(wrapWidth, s.wrapWidth[y])
	}

	curOffset := s.cur.X
	if cursorPastEnd {
		cell := s.buf.CellAt(s.cur.X, s.cur.Y)
		if cell != nil {
			curOffset += max(cell.Width, 1)
		} else {
			curOffset++
		}
	}
	logical, cur, saved := makeLogicalLines(
		lines,
		wrapped,
		wrapWidth,
		uv.Pos(curOffset, scrollbackLen+s.cur.Y),
		uv.Pos(s.saved.X, scrollbackLen+s.saved.Y),
	)

	physical, rewrapped, rewrapWidth, lineStarts := reflow(logical, width)
	curPos := reflowPosition(logical[cur.line], width, cur.offset)
	curPos.Y += lineStarts[cur.line]
	savedPos := reflowPosition(logical[saved.line], width, saved.offset)
	savedPos.Y += lineStarts[saved.line]

	start := max(0, len(physical)-height)
	start = min(start, curPos.Y)
	start = max(start, curPos.Y-height+1)
	if s.scrollback != nil {
		s.scrollback.replace(physical[:start], rewrapped[:start], rewrapWidth[:start])
	}

	s.buf = uv.NewRenderBuffer(width, height)
	s.wrapped = make([]bool, height)
	s.wrapWidth = make([]int, height)
	for y := start; y < len(physical) && y-start < height; y++ {
		s.buf.Lines[y-start] = slices.Clone(physical[y])
		s.wrapped[y-start] = rewrapped[y]
		s.wrapWidth[y-start] = rewrapWidth[y]
	}
	s.buf.Touched = nil
	s.scroll = s.buf.Bounds()

	s.cur.X, s.cur.Y, cursorPastEnd = resizedCursor(curPos, start, width, height)
	s.saved.X, s.saved.Y, _ = resizedCursor(savedPos, start, width, height)
	return cursorPastEnd
}

func makeLogicalLines(lines []uv.Line, wrapped []bool, wrapWidth []int, cur, saved uv.Position) (
	logical []uv.Line,
	curPos logicalPosition,
	savedPos logicalPosition,
) {
	logical = append(logical, nil)
	for y, line := range lines {
		lineIndex := len(logical) - 1
		base := len(logical[lineIndex])
		if y == cur.Y {
			curPos = logicalPosition{lineIndex, base + cur.X}
		}
		if y == saved.Y {
			savedPos = logicalPosition{lineIndex, base + saved.X}
		}

		continues := y < len(wrapped) && wrapped[y]
		last := lineContentWidth(line)
		if continues || wrapWidth[y] > 0 {
			last = effectiveLineWidth(line, wrapWidth[y])
		}
		if y == cur.Y {
			last = max(last, cur.X)
		}
		if y == saved.Y {
			last = max(last, saved.X)
		}
		last = min(last, len(line))
		logical[lineIndex] = append(logical[lineIndex], slices.Clone(line[:last])...)

		if !continues && y < len(lines)-1 {
			logical = append(logical, nil)
		}
	}
	return logical, curPos, savedPos
}

func lineContentWidth(line uv.Line) int {
	for i := len(line) - 1; i >= 0; i-- {
		cell := &line[i]
		if !cell.IsZero() && !cell.Equal(&uv.EmptyCell) {
			return min(len(line), i+max(cell.Width, 1))
		}
	}
	return 0
}

func effectiveLineWidth(line uv.Line, wrapWidth int) int {
	return max(wrapWidth, lineContentWidth(line))
}

func reflow(logical []uv.Line, width int) (lines []uv.Line, wrapped []bool, wrapWidth []int, starts []int) {
	starts = make([]int, 0, len(logical))
	for _, line := range logical {
		rows, widths := reflowLine(line, width)
		starts = append(starts, len(lines))
		lines = append(lines, rows...)
		for y := range rows {
			continues := y < len(rows)-1
			wrapped = append(wrapped, continues)
			if continues || widths[y] > lineContentWidth(rows[y]) {
				wrapWidth = append(wrapWidth, widths[y])
			} else {
				wrapWidth = append(wrapWidth, 0)
			}
		}
	}
	return lines, wrapped, wrapWidth, starts
}

func reflowLine(cells uv.Line, width int) ([]uv.Line, []int) {
	rows := []uv.Line{uv.NewLine(width)}
	widths := []int{0}
	x, y := 0, 0
	for i := 0; i < len(cells); {
		cell := &cells[i]
		cellWidth := max(cell.Width, 1)
		if x > 0 && x+cellWidth > width {
			widths[y] = x
			rows = append(rows, uv.NewLine(width))
			widths = append(widths, 0)
			x, y = 0, y+1
		}

		rows[y].Set(x, cell)
		x += cellWidth
		widths[y] = x
		i += cellWidth
		if x == width && i < len(cells) {
			rows = append(rows, uv.NewLine(width))
			widths = append(widths, 0)
			x, y = 0, y+1
		}
	}
	return rows, widths
}

func reflowPosition(cells uv.Line, width, offset int) uv.Position {
	offset = min(offset, len(cells))
	x, y := 0, 0
	for i := 0; i < len(cells); {
		cellWidth := max(cells[i].Width, 1)
		if x > 0 && x+cellWidth > width {
			x, y = 0, y+1
		}
		if offset >= i && offset < i+cellWidth {
			return uv.Pos(x+offset-i, y)
		}
		x += cellWidth
		i += cellWidth
		if x == width && i < len(cells) {
			x, y = 0, y+1
		}
		if offset == i {
			return uv.Pos(x, y)
		}
	}
	return uv.Pos(x, y)
}

func resizedCursor(pos uv.Position, start, width, height int) (x, y int, pastEnd bool) {
	y = pos.Y - start
	if y < 0 {
		y = 0
	}
	if y >= height {
		y = height - 1
	}
	if pos.X >= width {
		return width - 1, y, true
	}
	return max(pos.X, 0), y, false
}

func (s *Scrollback) replace(lines []uv.Line, wrapped []bool, wrapWidth []int) {
	if len(lines) > s.maxLines {
		lines = lines[len(lines)-s.maxLines:]
		wrapped = wrapped[len(wrapped)-s.maxLines:]
		wrapWidth = wrapWidth[len(wrapWidth)-s.maxLines:]
	}
	s.lines = make([]uv.Line, len(lines))
	for i, line := range lines {
		last := lineContentWidth(line)
		if wrapped[i] || wrapWidth[i] > 0 {
			last = min(len(line), max(last, wrapWidth[i]))
		}
		s.lines[i] = slices.Clone(line[:last])
	}
	s.wrapped = slices.Clone(wrapped)
	s.wrapWidth = slices.Clone(wrapWidth)
}
