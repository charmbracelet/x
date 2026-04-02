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
}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := Screen{
		buf:        uv.NewRenderBuffer(w, h),
		scrollback: NewScrollback(DefaultScrollbackSize),
	}
	s.scroll = s.buf.Bounds()
	return &s
}

// Reset resets the screen.
// It clears the screen, sets the cursor to the top left corner, reset the
// cursor styles, and resets the scroll region.
func (s *Screen) Reset() {
	s.buf.Clear()
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
		count := 0
		// Save all lines that have content before clearing
		for y := 0; y < s.buf.Height(); y++ {
			line := s.buf.Line(y)
			if line != nil && !s.isLineEmpty(line) {
				s.scrollback.Push(line)
				count++
			}
		}
		if count > 0 && s.cb != nil && s.cb.ScrollbackPush != nil {
			s.cb.ScrollbackPush(count, s.buf.Width())
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
	s.buf.ClearArea(area)
	s.touchArea(area)
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c *uv.Cell) {
	s.FillArea(c, s.Bounds())
}

// FillArea fills the given area with the given cell.
func (s *Screen) FillArea(c *uv.Cell, area uv.Rectangle) {
	s.buf.FillArea(c, area)
	s.touchArea(area)
}

// setHorizontalMargins sets the horizontal margins.
func (s *Screen) setHorizontalMargins(left, right int) {
	if s.buf == nil || s.buf.Width() <= 0 {
		s.scroll.Min.X = 0
		s.scroll.Max.X = 0
		return
	}
	left = ordered.Clamp(left, 0, s.buf.Width()-1)
	right = ordered.Clamp(right, left+1, s.buf.Width())
	s.scroll.Min.X = left
	s.scroll.Max.X = right
}

// setVerticalMargins sets the vertical margins.
func (s *Screen) setVerticalMargins(top, bottom int) {
	if s.buf == nil || s.buf.Height() <= 0 {
		s.scroll.Min.Y = 0
		s.scroll.Max.Y = 0
		return
	}
	top = ordered.Clamp(top, 0, s.buf.Height()-1)
	bottom = ordered.Clamp(bottom, top+1, s.buf.Height())
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
}

// DeleteCell deletes n cells at the cursor position moving cells to the left.
// This has no effect if the cursor is outside the scroll region.
func (s *Screen) DeleteCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y
	s.buf.DeleteCellArea(x, y, n, s.blankCell(), s.scroll)
}

// ScrollUp scrolls the content up n lines within the given region. Lines
// scrolled past the top margin are lost. This is equivalent to [ansi.SU] which
// moves the cursor to the top margin and performs a [ansi.DL] operation.
func (s *Screen) ScrollUp(n int) {
	x, y := s.CursorPosition()
	s.setCursor(s.cur.X, 0, true)
	s.DeleteLine(n)
	s.setCursor(x, y, false)
}

// ScrollDown scrolls the content down n lines within the given region. Lines
// scrolled past the bottom margin are lost. This is equivalent to [ansi.SD]
// which moves the cursor to top margin and performs a [ansi.IL] operation.
func (s *Screen) ScrollDown(n int) {
	x, y := s.CursorPosition()
	s.setCursor(s.cur.X, 0, true)
	s.InsertLine(n)
	s.setCursor(x, y, false)
}

// InsertLine inserts n blank lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved down, with those past bottom margin being discarded.
// It returns true if the operation was successful.
func (s *Screen) InsertLine(n int) bool {
	if n <= 0 {
		return false
	}

	x, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < s.scroll.Min.Y || y >= s.scroll.Max.Y ||
		x < s.scroll.Min.X || x >= s.scroll.Max.X {
		return false
	}

	s.buf.InsertLineArea(y, n, s.blankCell(), s.scroll)

	return true
}

// DeleteLine deletes n lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved up, with blank lines inserted at the bottom of scroll region.
// If scrollback is enabled and cursor is at top of scroll region, lines
// are saved to the scrollback buffer before deletion.
// It returns true if the operation was successful.
func (s *Screen) DeleteLine(n int) bool {
	if n <= 0 {
		return false
	}

	scroll := s.scroll
	x, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < scroll.Min.Y || y >= scroll.Max.Y ||
		x < scroll.Min.X || x >= scroll.Max.X {
		return false
	}

	// Save lines to scrollback if we're at the top of the scroll region
	// and the scroll region uses the full width (typical terminal scroll).
	// This captures lines that would be lost during scroll up operations.
	if s.scrollback != nil && y == scroll.Min.Y &&
		scroll.Min.X == 0 && scroll.Max.X == s.buf.Width() {
		// Save lines that will be deleted
		linesToSave := min(n, scroll.Max.Y-y)
		s.scrollback.PushN(s.buf, y, linesToSave)
		if linesToSave > 0 && s.cb != nil && s.cb.ScrollbackPush != nil {
			s.cb.ScrollbackPush(linesToSave, s.buf.Width())
		}
	}

	s.buf.DeleteLineArea(y, n, s.blankCell(), scroll)

	return true
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
