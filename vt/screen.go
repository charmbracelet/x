package vt

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/ordered"
)

// Screen represents a virtual terminal screen.
type Screen struct {
	// cb is the callbacks struct to use.
	cb *Callbacks
	// The buffer of the screen.
	buf *uv.RenderBuffer
	// The cur of the screen.
	cur, saved Cursor
	// scroll is the scroll region.
	scroll uv.Rectangle
	// scrollback is the scrollback buffer for lines scrolled off the top.
	scrollback *Scrollback
	// rotate and rotateUsed are scratch space for moving lines.
	rotate     []uv.Line
	rotateUsed []int
	// reference makes the screen use the plain cell by cell implementations
	// instead of the fast paths, for tests to compare them.
	reference bool
	// used holds, for each line, a bound on the columns that can hold a
	// character: every cell at or past used[y] is [uv.EmptyCell]. Scrollback
	// and blanking lines use it to skip the empty cells instead of going
	// over the whole line, which for the short lines of most output is most
	// of the work. Every write marks the columns it may change, and anything
	// unusual marks the whole line.
	used []int
}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := Screen{
		buf:        uv.NewRenderBuffer(w, h),
		scrollback: NewScrollback(DefaultScrollbackSize),
	}
	s.scroll = s.buf.Bounds()
	s.used = make([]int, h)
	return &s
}

// Reset resets the screen.
// It clears the screen, sets the cursor to the top left corner, reset the
// cursor styles, and resets the scroll region.
func (s *Screen) Reset() {
	s.buf.Clear()
	clear(s.used)
	s.cur = Cursor{}
	s.saved = Cursor{}
	s.scroll = s.buf.Bounds()
	s.buf.Touched = nil
}

// Bounds returns the bounds of the screen.
func (s *Screen) Bounds() uv.Rectangle {
	return s.buf.Bounds()
}

// Touched returns touched lines in the screen buffer.
func (s *Screen) Touched() []*uv.LineData {
	return s.buf.Touched
}

// ClearTouched clears the touched state.
func (s *Screen) ClearTouched() {
	s.buf.Touched = nil
}

// CellAt returns the cell at the given x, y position.
func (s *Screen) CellAt(x int, y int) *uv.Cell {
	return s.buf.CellAt(x, y)
}

// SetCell sets the cell at the given x, y position.
func (s *Screen) SetCell(x, y int, c *uv.Cell) {
	s.buf.SetCell(x, y, c)
	if c != nil {
		s.markUsed(y, x+max(c.Width, 1))
	}
}

// markUsed records that the cells of line y before column end may be
// written.
func (s *Screen) markUsed(y, end int) {
	if y >= 0 && y < len(s.used) {
		s.used[y] = max(s.used[y], end)
	}
}

// markAllUsed records that every line may be written across its whole width.
func (s *Screen) markAllUsed() {
	for y := range s.used {
		s.used[y] = s.buf.Width()
	}
}

// Height returns the height of the screen.
func (s *Screen) Height() int {
	return s.buf.Height()
}

// Resize resizes the screen.
func (s *Screen) Resize(width int, height int) {
	if s.buf == nil {
		s.buf = uv.NewRenderBuffer(width, height)
	} else {
		s.buf.Resize(width, height)
		s.buf.Touched = nil
	}
	s.used = make([]int, s.buf.Height())
	s.markAllUsed()
	s.scroll = s.buf.Bounds()
}

// Width returns the width of the screen.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// Clear clears the screen with blank cells.
func (s *Screen) Clear() {
	s.ClearArea(s.Bounds())
}

// ClearWithScrollback saves all non-empty lines to scrollback before clearing.
// This is used for operations like ED 2 (erase screen) where content should
// be preserved in history.
func (s *Screen) ClearWithScrollback() {
	if s.scrollback != nil {
		// Save all lines that have content before clearing
		for y := 0; y < s.buf.Height(); y++ {
			line := s.buf.Line(y)
			if line != nil && !s.isLineEmpty(line) {
				s.scrollback.Push(line)
			}
		}
	}
	s.Clear()
}

// isLineEmpty returns true if the line contains only empty/space cells.
func (s *Screen) isLineEmpty(line uv.Line) bool {
	for _, cell := range line {
		if cell.Content != "" && cell.Content != " " {
			return false
		}
	}
	return true
}

// ClearArea clears the given area.
func (s *Screen) ClearArea(area uv.Rectangle) {
	s.FillArea(nil, area)
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c *uv.Cell) {
	s.FillArea(c, s.Bounds())
}

// FillArea fills the given area with the given cell.
func (s *Screen) FillArea(c *uv.Cell, area uv.Rectangle) {
	if !s.reference && area.Min.X <= 0 && area.Max.X >= s.buf.Width() && (c == nil || c.Width <= 1) {
		s.fillLines(c, area)
	} else {
		s.buf.FillArea(c, area)
		if c != nil && !isPlainEmpty(c) {
			for y := max(area.Min.Y, 0); y < min(area.Max.Y, s.buf.Height()); y++ {
				s.markUsed(y, area.Max.X)
			}
		}
	}
	s.touchArea(area)
}

// fillLines fills whole lines at once, which needs none of the fixups for
// wide cells cut in half.
func (s *Screen) fillLines(c *uv.Cell, area uv.Rectangle) {
	fill := uv.EmptyCell
	if c != nil {
		fill = *c
	}
	used := s.buf.Width()
	if isPlainEmpty(&fill) {
		used = 0
	}
	for y := max(area.Min.Y, 0); y < min(area.Max.Y, s.buf.Height()); y++ {
		fillLine(s.buf.Lines[y], fill)
		s.used[y] = used
	}
}

// fillLine sets every cell of the line to c.
func fillLine(line uv.Line, c uv.Cell) {
	if len(line) == 0 {
		return
	}
	line[0] = c
	// Fill by doubling, which copies whole runs of cells at once.
	for filled := 1; filled < len(line); filled *= 2 {
		copy(line[filled:], line[:filled])
	}
}

// setHorizontalMargins sets the horizontal margins.
func (s *Screen) setHorizontalMargins(left, right int) {
	s.scroll.Min.X = left
	s.scroll.Max.X = right
}

// setVerticalMargins sets the vertical margins.
func (s *Screen) setVerticalMargins(top, bottom int) {
	s.scroll.Min.Y = top
	s.scroll.Max.Y = bottom
}

// setCursorX sets the cursor X position. If margins is true, the cursor is
// only set if it is within the scroll margins.
func (s *Screen) setCursorX(x int, margins bool) {
	s.setCursor(x, s.cur.Y, margins)
}

// setCursor sets the cursor position. If margins is true, the cursor is only
// set if it is within the scroll margins. This follows how [ansi.CUP] works.
func (s *Screen) setCursor(x, y int, margins bool) {
	old := s.cur.Position
	if !margins {
		y = ordered.Clamp(y, 0, s.buf.Height()-1)
		x = ordered.Clamp(x, 0, s.buf.Width()-1)
	} else {
		y = ordered.Clamp(s.scroll.Min.Y+y, s.scroll.Min.Y, s.scroll.Max.Y-1)
		x = ordered.Clamp(s.scroll.Min.X+x, s.scroll.Min.X, s.scroll.Max.X-1)
	}
	s.cur.X, s.cur.Y = x, y

	if s.cb.CursorPosition != nil && (old.X != x || old.Y != y) {
		s.cb.CursorPosition(old, uv.Pos(x, y))
	}
}

// moveCursor moves the cursor by the given x and y deltas. If the cursor
// position is inside the scroll region, it is bounded by the scroll region.
// Otherwise, it is bounded by the screen bounds.
// This follows how [ansi.CUU], [ansi.CUD], [ansi.CUF], [ansi.CUB], [ansi.CNL],
// [ansi.CPL].
func (s *Screen) moveCursor(dx, dy int) {
	scroll := s.scroll
	old := s.cur.Position
	if old.X < scroll.Min.X {
		scroll.Min.X = 0
	}
	if old.X >= scroll.Max.X {
		scroll.Max.X = s.buf.Width()
	}

	pt := uv.Pos(s.cur.X+dx, s.cur.Y+dy)

	var x, y int
	if old.In(scroll) {
		y = ordered.Clamp(pt.Y, scroll.Min.Y, scroll.Max.Y-1)
		x = ordered.Clamp(pt.X, scroll.Min.X, scroll.Max.X-1)
	} else {
		y = ordered.Clamp(pt.Y, 0, s.buf.Height()-1)
		x = ordered.Clamp(pt.X, 0, s.buf.Width()-1)
	}

	s.cur.X, s.cur.Y = x, y

	if s.cb.CursorPosition != nil && (old.X != x || old.Y != y) {
		s.cb.CursorPosition(old, uv.Pos(x, y))
	}
}

// Cursor returns the cursor.
func (s *Screen) Cursor() Cursor {
	return s.cur
}

// CursorPosition returns the cursor position.
func (s *Screen) CursorPosition() (x, y int) {
	return s.cur.X, s.cur.Y
}

// ScrollRegion returns the scroll region.
func (s *Screen) ScrollRegion() uv.Rectangle {
	return s.scroll
}

// SaveCursor saves the cursor.
func (s *Screen) SaveCursor() {
	s.saved = s.cur
}

// RestoreCursor restores the cursor.
func (s *Screen) RestoreCursor() {
	old := s.cur.Position
	s.cur = s.saved

	if s.cb.CursorPosition != nil && (old.X != s.cur.X || old.Y != s.cur.Y) {
		s.cb.CursorPosition(old, s.cur.Position)
	}
}

// setCursorHidden sets the cursor hidden.
func (s *Screen) setCursorHidden(hidden bool) {
	changed := s.cur.Hidden != hidden
	s.cur.Hidden = hidden
	if changed && s.cb.CursorVisibility != nil {
		s.cb.CursorVisibility(!hidden)
	}
}

// setCursorStyle sets the cursor style.
func (s *Screen) setCursorStyle(style CursorStyle, blink bool) {
	changed := s.cur.Style != style || s.cur.Steady != !blink
	s.cur.Style = style
	s.cur.Steady = !blink
	if changed && s.cb.CursorStyle != nil {
		s.cb.CursorStyle(style, !blink)
	}
}

// cursorPen returns the cursor pen.
func (s *Screen) cursorPen() uv.Style {
	return s.cur.Pen
}

// cursorLink returns the cursor link.
func (s *Screen) cursorLink() uv.Link {
	return s.cur.Link
}

// ShowCursor shows the cursor.
func (s *Screen) ShowCursor() {
	s.setCursorHidden(false)
}

// HideCursor hides the cursor.
func (s *Screen) HideCursor() {
	s.setCursorHidden(true)
}

// InsertCell inserts n blank characters at the cursor position pushing out
// cells to the right and out of the screen.
func (s *Screen) InsertCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y
	s.buf.InsertCellArea(x, y, n, s.blankCell(), s.scroll)
	s.markUsed(y, s.buf.Width())
}

// DeleteCell deletes n cells at the cursor position moving cells to the left.
// This has no effect if the cursor is outside the scroll region.
func (s *Screen) DeleteCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y
	s.buf.DeleteCellArea(x, y, n, s.blankCell(), s.scroll)
	s.markUsed(y, s.buf.Width())
}

// ScrollUp scrolls the content up n lines within the given region. Lines
// scrolled past the top margin are lost. This is equivalent to [ansi.SU] which
// moves the cursor to the top margin and performs a [ansi.DL] operation.
// Unlike them, it leaves the cursor where it is.
func (s *Screen) ScrollUp(n int) {
	s.deleteLines(s.scroll.Min.Y, n)
}

// ScrollDown scrolls the content down n lines within the given region. Lines
// scrolled past the bottom margin are lost. This is equivalent to [ansi.SD]
// which moves the cursor to top margin and performs a [ansi.IL] operation.
// Unlike them, it leaves the cursor where it is.
func (s *Screen) ScrollDown(n int) {
	s.insertLines(s.scroll.Min.Y, n)
}

// InsertLine inserts n blank lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved down, with those past bottom margin being discarded.
// It returns true if the operation was successful.
func (s *Screen) InsertLine(n int) bool {
	if n <= 0 || !s.cur.In(s.scroll) {
		return false
	}

	s.insertLines(s.cur.Y, n)
	return true
}

// DeleteLine deletes n lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved up, with blank lines inserted at the bottom of scroll region.
// If scrollback is enabled and cursor is at top of scroll region, lines
// are saved to the scrollback buffer before deletion.
// It returns true if the operation was successful.
func (s *Screen) DeleteLine(n int) bool {
	if n <= 0 || !s.cur.In(s.scroll) {
		return false
	}

	s.deleteLines(s.cur.Y, n)
	return true
}

// fullWidth reports whether the scroll region spans the whole width of the
// screen, so that scrolling it moves whole lines.
func (s *Screen) fullWidth() bool {
	return s.scroll.Min.X == 0 && s.scroll.Max.X == s.buf.Width()
}

// insertLines inserts n blank lines at line y of the scroll region, moving
// the lines below down and dropping the ones moved past the bottom margin.
func (s *Screen) insertLines(y, n int) {
	scroll := s.scroll
	if n <= 0 || y < scroll.Min.Y || y >= scroll.Max.Y {
		return
	}
	if s.reference || !s.fullWidth() {
		s.buf.InsertLineArea(y, n, s.blankCell(), scroll)
		s.markAllUsed()
		return
	}

	n = min(n, scroll.Max.Y-y)
	// Move the lines that fall off the bottom to the top and blank them.
	s.rotateLines(y, scroll.Max.Y, scroll.Max.Y-y-n)
	for i := y; i < y+n; i++ {
		s.blankLine(i)
	}
	s.touchLines(y, scroll.Max.Y)
}

// deleteLines deletes n lines at line y of the scroll region, moving the
// lines below up and adding blank lines at the bottom margin. Lines deleted
// from the top of a scroll region that spans the whole width go to the
// scrollback buffer.
func (s *Screen) deleteLines(y, n int) {
	scroll := s.scroll
	if n <= 0 || y < scroll.Min.Y || y >= scroll.Max.Y {
		return
	}
	n = min(n, scroll.Max.Y-y)
	fullWidth := s.fullWidth()

	// Save lines that will be deleted, since they would otherwise be lost
	// during scroll up operations.
	if fullWidth && s.scrollback != nil && y == scroll.Min.Y {
		for i := range n {
			s.scrollback.pushRow(s.buf.Lines[y+i], s.used[y+i])
		}
	}

	if s.reference || !fullWidth {
		s.buf.DeleteLineArea(y, n, s.blankCell(), scroll)
		s.markAllUsed()
		return
	}

	// Move the deleted lines to the bottom and blank them.
	s.rotateLines(y, scroll.Max.Y, n)
	for i := scroll.Max.Y - n; i < scroll.Max.Y; i++ {
		s.blankLine(i)
	}
	s.touchLines(y, scroll.Max.Y)
}

// rotateLines rotates the lines in [top, bottom) up by n lines, so the line
// at top+n ends up at top and the first n lines at the bottom. Moving the
// lines themselves is much cheaper than copying their cells.
func (s *Screen) rotateLines(top, bottom, n int) {
	lines := s.buf.Lines[top:bottom]
	if n <= 0 || n >= len(lines) {
		return
	}
	s.rotate = append(s.rotate[:0], lines[:n]...)
	copy(lines, lines[n:])
	copy(lines[len(lines)-n:], s.rotate)
	clear(s.rotate)

	used := s.used[top:bottom]
	s.rotateUsed = append(s.rotateUsed[:0], used[:n]...)
	copy(used, used[n:])
	copy(used[len(used)-n:], s.rotateUsed)
}

// blankLine fills line y with blank cells.
func (s *Screen) blankLine(y int) {
	line := s.buf.Lines[y]
	if c := s.blankCell(); c != nil {
		// Cells that take the background color aren't empty.
		fillLine(line, *c)
		s.used[y] = len(line)
		return
	}
	// Past the bound the line is empty already.
	end := min(s.used[y], len(line))
	for i := range line[:end] {
		line[i] = uv.EmptyCell
	}
	s.used[y] = 0
}

// isPlainEmpty reports whether the cell is [uv.EmptyCell].
func isPlainEmpty(c *uv.Cell) bool {
	return c.Content == " " && c.Width == 1 && isBlank(c)
}

// touchLines marks the lines in [top, bottom) as changed across the whole
// width. It does what calling TouchLine for each line does.
func (s *Screen) touchLines(top, bottom int) {
	b := s.buf
	width := b.Width()
	bottom = min(bottom, len(b.Lines))
	if len(b.Touched) < bottom {
		b.Touched = append(b.Touched, make([]*uv.LineData, bottom-len(b.Touched))...)
	}
	for y := max(top, 0); y < bottom; y++ {
		t := b.Touched[y]
		if t == nil {
			b.Touched[y] = &uv.LineData{FirstCell: 0, LastCell: width}
			continue
		}
		t.FirstCell = min(t.FirstCell, 0)
		t.LastCell = max(t.LastCell, width)
	}
}

// blankCell returns the cursor blank cell with the background color set to the
// current pen background color. If the pen background color is nil, the return
// value is nil.
func (s *Screen) blankCell() *uv.Cell {
	if s.cur.Pen.Bg == nil {
		return nil
	}

	c := uv.EmptyCell
	c.Style.Bg = s.cur.Pen.Bg
	return &c
}

// touchArea marks all lines in the given area as touched.
func (s *Screen) touchArea(area uv.Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		s.buf.TouchLine(area.Min.X, y, area.Max.X-area.Min.X)
	}
}

// Scrollback returns the screen's scrollback buffer.
func (s *Screen) Scrollback() *Scrollback {
	return s.scrollback
}

// SetScrollback sets the screen's scrollback buffer.
// Pass nil to disable scrollback.
func (s *Screen) SetScrollback(sb *Scrollback) {
	s.scrollback = sb
}

// SetScrollbackSize sets the maximum number of lines in the scrollback buffer.
func (s *Screen) SetScrollbackSize(maxLines int) {
	if s.scrollback == nil {
		s.scrollback = NewScrollback(maxLines)
	} else {
		s.scrollback.SetMaxLines(maxLines)
	}
}
