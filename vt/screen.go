package vt

import (
	"sync"

	"github.com/charmbracelet/x/cellbuf"
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
	scroll Rectangle
	// mutex for the screen.
	mu sync.RWMutex
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
	s.mu.Lock()
	s.buf.Clear()
	s.cur = Cursor{}
	s.saved = Cursor{}
	s.scroll = s.buf.Bounds()
	s.mu.Unlock()
}

// Bounds returns the bounds of the screen.
func (s *Screen) Bounds() Rectangle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Bounds()
}

// Cell returns the cell at the given x, y position.
func (s *Screen) Cell(x int, y int) *Cell {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Cell(x, y)
}

// SetCell sets the cell at the given x, y position.
// It returns true if the cell was set successfully.
func (s *Screen) SetCell(x, y int, c *Cell) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.buf.SetCell(x, y, c)
	if v && s.cb.Damage != nil {
		width := 1
		if c != nil {
			width = c.Width
		}
		s.cb.Damage(CellDamage{x, y, width})
	}
	return v
}

// Height returns the height of the screen.
func (s *Screen) Height() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Height()
}

// Resize resizes the screen.
func (s *Screen) Resize(width int, height int) {
	s.mu.Lock()
	s.buf.Resize(width, height)
	s.scroll = s.buf.Bounds()
	if s.cb != nil && s.cb.Damage != nil {
		s.cb.Damage(ScreenDamage{width, height})
	}
	s.mu.Unlock()
}

// Width returns the width of the screen.
func (s *Screen) Width() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Width()
}

// Clear clears the screen or part of it.
func (s *Screen) Clear(rects ...Rectangle) {
	s.mu.Lock()
	if len(rects) == 0 {
		s.buf.Clear()
	} else {
		for _, r := range rects {
			s.buf.ClearRect(r)
		}
	}
	if s.cb.Damage != nil {
		for _, r := range rects {
			s.cb.Damage(RectDamage(r))
		}
	}
	s.mu.Unlock()
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c *Cell, rects ...Rectangle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(rects) == 0 {
		s.buf.Fill(c)
	} else {
		for _, r := range rects {
			s.buf.FillRect(c, r)
		}
	}
	if s.cb.Damage != nil {
		for _, r := range rects {
			s.cb.Damage(RectDamage(r))
		}
	}
}

// setHorizontalMargins sets the horizontal margins.
func (s *Screen) setHorizontalMargins(left, right int) {
	s.mu.Lock()
	s.scroll.Min.X = left
	s.scroll.Max.X = right
	s.mu.Unlock()
}

// setVerticalMargins sets the vertical margins.
func (s *Screen) setVerticalMargins(top, bottom int) {
	s.mu.Lock()
	s.scroll.Min.Y = top
	s.scroll.Max.Y = bottom
	s.mu.Unlock()
}

// setCursorX sets the cursor X position. If margins is true, the cursor is
// only set if it is within the scroll margins.
func (s *Screen) setCursorX(x int, margins bool) {
	s.setCursor(x, s.cur.Y, margins)
}

// setCursorY sets the cursor Y position. If margins is true, the cursor is
// only set if it is within the scroll margins.
func (s *Screen) setCursorY(y int, margins bool) {
	s.setCursor(s.cur.X, y, margins)
}

// setCursor sets the cursor position. If margins is true, the cursor is only
// set if it is within the scroll margins. This follows how [ansi.CUP] works.
func (s *Screen) setCursor(x, y int, margins bool) {
	s.mu.Lock()
	old := s.cur.Position
	if !margins {
		y = clamp(y, 0, s.buf.Height()-1)
		x = clamp(x, 0, s.buf.Width()-1)
	} else {
		y = clamp(s.scroll.Min.Y+y, s.scroll.Min.Y, s.scroll.Max.Y-1)
		x = clamp(s.scroll.Min.X+x, s.scroll.Min.X, s.scroll.Max.X-1)
	}
	s.cur.X, s.cur.Y = x, y
	s.mu.Unlock()
	if s.cb.CursorPosition != nil && (old.X != x || old.Y != y) {
		s.cb.CursorPosition(old, cellbuf.Pos(x, y))
	}
}

// moveCursor moves the cursor by the given x and y deltas. If the cursor
// position is inside the scroll region, it is bounded by the scroll region.
// Otherwise, it is bounded by the screen bounds.
// This follows how [ansi.CUU], [ansi.CUD], [ansi.CUF], [ansi.CUB], [ansi.CNL],
// [ansi.CPL].
func (s *Screen) moveCursor(dx, dy int) {
	s.mu.Lock()
	scroll := s.scroll
	old := s.cur.Position
	if old.X < scroll.Min.X {
		scroll.Min.X = 0
	}
	if old.X >= scroll.Max.X {
		scroll.Max.X = s.buf.Width()
	}

	pt := cellbuf.Pos(s.cur.X+dx, s.cur.Y+dy)

	var x, y int
	if old.In(scroll) {
		y = clamp(pt.Y, scroll.Min.Y, scroll.Max.Y-1)
		x = clamp(pt.X, scroll.Min.X, scroll.Max.X-1)
	} else {
		y = clamp(pt.Y, 0, s.buf.Height()-1)
		x = clamp(pt.X, 0, s.buf.Width()-1)
	}

	s.cur.X, s.cur.Y = x, y
	s.mu.Unlock()
	if s.cb.CursorPosition != nil && (old.X != x || old.Y != y) {
		s.cb.CursorPosition(old, cellbuf.Pos(x, y))
	}
}

// Cursor returns the cursor.
func (s *Screen) Cursor() Cursor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cur
}

// CursorPosition returns the cursor position.
func (s *Screen) CursorPosition() (x, y int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cur.X, s.cur.Y
}

// ScrollRegion returns the scroll region.
func (s *Screen) ScrollRegion() Rectangle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scroll
}

// SaveCursor saves the cursor.
func (s *Screen) SaveCursor() {
	s.mu.Lock()
	s.saved = s.cur
	s.mu.Unlock()
}

// RestoreCursor restores the cursor.
func (s *Screen) RestoreCursor() {
	s.mu.Lock()
	old := s.cur.Position
	s.cur = s.saved
	s.mu.Unlock()
	if s.cb.CursorPosition != nil && (old.X != s.cur.X || old.Y != s.cur.Y) {
		s.cb.CursorPosition(old, s.cur.Position)
	}
}

// setCursorHidden sets the cursor hidden.
func (s *Screen) setCursorHidden(hidden bool) {
	s.mu.Lock()
	wasHidden := s.cur.Hidden
	s.cur.Hidden = hidden
	s.mu.Unlock()
	if s.cb.CursorVisibility != nil && wasHidden != hidden {
		s.cb.CursorVisibility(!hidden)
	}
}

// setCursorStyle sets the cursor style.
func (s *Screen) setCursorStyle(style CursorStyle, blink bool) {
	s.mu.Lock()
	s.cur.Style = style
	s.cur.Steady = !blink
	s.mu.Unlock()
	if s.cb.CursorStyle != nil {
		s.cb.CursorStyle(style, !blink)
	}
}

// cursorPen returns the cursor pen.
func (s *Screen) cursorPen() Style {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cur.Pen
}

// cursorLink returns the cursor link.
func (s *Screen) cursorLink() Link {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

	s.mu.Lock()
	defer s.mu.Unlock()
	x, y := s.cur.X, s.cur.Y

	s.buf.InsertCellRect(x, y, n, s.blankCell(), s.scroll)
	if s.cb.Damage != nil {
		s.cb.Damage(RectDamage(cellbuf.Rect(x, y, s.scroll.Dx()-x, 1)))
	}
}

// DeleteCell deletes n cells at the cursor position moving cells to the left.
// This has no effect if the cursor is outside the scroll region.
func (s *Screen) DeleteCell(n int) {
	if n <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	x, y := s.cur.X, s.cur.Y

	s.buf.DeleteCellRect(x, y, n, s.blankCell(), s.scroll)
	if s.cb.Damage != nil {
		s.cb.Damage(RectDamage(cellbuf.Rect(x, y, s.scroll.Dx()-x, 1)))
	}
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

	s.mu.Lock()
	defer s.mu.Unlock()
	x, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < s.scroll.Min.Y || y >= s.scroll.Max.Y ||
		x < s.scroll.Min.X || x >= s.scroll.Max.X {
		return false
	}

	s.buf.InsertLineRect(y, n, s.blankCell(), s.scroll)
	if s.cb.Damage != nil {
		rect := s.scroll
		rect.Min.Y = y
		rect.Max.Y += n
		s.cb.Damage(RectDamage(rect))
	}

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

	s.mu.Lock()
	defer s.mu.Unlock()
	scroll := s.scroll
	x, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < scroll.Min.Y || y >= scroll.Max.Y ||
		x < scroll.Min.X || x >= scroll.Max.X {
		return false
	}

	s.buf.DeleteLineRect(y, n, s.blankCell(), scroll)
	if s.cb.Damage != nil {
		rect := scroll
		rect.Min.Y = y
		rect.Max.Y += n
		s.cb.Damage(RectDamage(rect))
	}

	return true
}

// blankCell returns the cursor blank cell with the background color set to the
// current pen background color. If the pen background color is nil, the return
// value is nil.
func (s *Screen) blankCell() (c *Cell) {
	if s.cur.Pen.Bg == nil {
		return
	}

	c = new(Cell)
	*c = cellbuf.BlankCell
	c.Style.Bg = s.cur.Pen.Bg
	return
}
