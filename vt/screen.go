package vt

import (
	"github.com/charmbracelet/uv"
)

// Screen represents a virtual terminal screen.
type Screen struct {
	// cb is the callbacks struct to use.
	cb *Callbacks
	// The buffer of the screen.
	buf Buffer
	// The cur of the screen.
	cur, saved Cursor
	// scroll is the scroll region.
	scroll uv.Rectangle
}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := new(Screen)
	s.Resize(w, h)
	return s
}

// Reset resets the screen.
// It clears the screen, sets the cursor to the top left corner, reset the
// cursor styles, and resets the scroll region.
func (s *Screen) Reset() {
	s.buf.Clear()
	s.cur = Cursor{}
	s.saved = Cursor{}
	s.scroll = s.buf.Bounds()
}

// Bounds returns the bounds of the screen.
func (s *Screen) Bounds() uv.Rectangle {
	return s.buf.Bounds()
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
	s.buf.Resize(width, height)
	s.scroll = s.buf.Bounds()
}

// Width returns the width of the screen.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// Clear clears the screen or part of it.
func (s *Screen) Clear(rects ...uv.Rectangle) {
	if len(rects) == 0 {
		s.buf.Clear()
	} else {
		for _, r := range rects {
			s.buf.ClearArea(r)
		}
	}
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c *uv.Cell, rects ...uv.Rectangle) {
	if len(rects) == 0 {
		s.buf.Fill(c)
	} else {
		for _, r := range rects {
			s.buf.FillArea(c, r)
		}
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

// setCursorY sets the cursor Y position. If margins is true, the cursor is
// only set if it is within the scroll margins.
func (s *Screen) setCursorY(y int, margins bool) { //nolint:unused
	s.setCursor(s.cur.X, y, margins)
}

// setCursor sets the cursor position. If margins is true, the cursor is only
// set if it is within the scroll margins. This follows how [ansi.CUP] works.
func (s *Screen) setCursor(x, y int, margins bool) {
	old := s.cur.Position
	if !margins {
		y = clamp(y, 0, s.buf.Height()-1)
		x = clamp(x, 0, s.buf.Width()-1)
	} else {
		y = clamp(s.scroll.Min.Y+y, s.scroll.Min.Y, s.scroll.Max.Y-1)
		x = clamp(s.scroll.Min.X+x, s.scroll.Min.X, s.scroll.Max.X-1)
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
		y = clamp(pt.Y, scroll.Min.Y, scroll.Max.Y-1)
		x = clamp(pt.X, scroll.Min.X, scroll.Max.X-1)
	} else {
		y = clamp(pt.Y, 0, s.buf.Height()-1)
		x = clamp(pt.X, 0, s.buf.Width()-1)
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
	s.cur.Hidden = hidden

	if s.cb.CursorVisibility != nil {
		s.cb.CursorVisibility(!hidden)
	}
}

// setCursorStyle sets the cursor style.
func (s *Screen) setCursorStyle(style CursorStyle, blink bool) {
	s.cur.Style = style
	s.cur.Steady = !blink

	if s.cb.CursorStyle != nil {
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

	s.buf.DeleteLineArea(y, n, s.blankCell(), scroll)

	return true
}

// blankCell returns the cursor blank cell with the background color set to the
// current pen background color. If the pen background color is nil, the return
// value is nil.
func (s *Screen) blankCell() (c *uv.Cell) {
	if s.cur.Pen.Bg == nil {
		return
	}

	c = new(uv.Cell)
	*c = uv.EmptyCell
	c.Style.Bg = s.cur.Pen.Bg
	return
}
