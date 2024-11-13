package vt

import (
	"sync"

	"github.com/charmbracelet/x/cellbuf"
)

// Screen represents a virtual terminal screen.
type Screen struct {
	mu sync.RWMutex
	// The buffer of the screen.
	buf Buffer
	// The cur of the screen.
	cur, saved Cursor
	// scroll is the scroll region.
	scroll Rectangle
}

var _ cellbuf.Screen = &Screen{}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := new(Screen)
	s.Resize(w, h)
	return s
}

// Bounds returns the bounds of the screen.
func (s *Screen) Bounds() Rectangle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Bounds()
}

// Cell implements cellbuf.Screen.
func (s *Screen) Cell(x int, y int) (Cell, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Cell(x, y)
}

// Draw implements cellbuf.Screen.
func (s *Screen) Draw(x int, y int, c Cell) bool {
	return s.SetCell(x, y, &c)
}

// SetCell sets the cell at the given x, y position.
// It returns true if the cell was set successfully.
func (s *Screen) SetCell(x, y int, c *Cell) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.SetCell(x, y, c)
}

// Height implements cellbuf.Grid.
func (s *Screen) Height() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Height()
}

// Resize implements cellbuf.Grid.
func (s *Screen) Resize(width int, height int) {
	s.mu.Lock()
	s.buf.Resize(width, height)
	s.scroll = s.buf.Bounds()
	s.mu.Unlock()
}

// Width implements cellbuf.Grid.
func (s *Screen) Width() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.Width()
}

// Clear clears the screen or part of it.
func (s *Screen) Clear(rects ...Rectangle) {
	s.mu.Lock()
	s.buf.Clear(rects...)
	s.mu.Unlock()
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c *Cell, rects ...Rectangle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf.Fill(c, rects...)
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
	defer s.mu.Unlock()
	if !margins {
		s.cur.Y = clamp(y, 0, s.buf.Height()-1)
		s.cur.X = clamp(x, 0, s.buf.Width()-1)
	} else {
		s.cur.Y = clamp(s.scroll.Min.Y+y, s.scroll.Min.Y, s.scroll.Max.Y-1)
		s.cur.X = clamp(s.scroll.Min.X+x, s.scroll.Min.X, s.scroll.Max.X-1)
	}
}

// moveCursor moves the cursor by the given x and y deltas. If the cursor
// position is inside the scroll region, it is bounded by the scroll region.
// Otherwise, it is bounded by the screen bounds.
// This follows how [ansi.CUU], [ansi.CUD], [ansi.CUF], [ansi.CUB], [ansi.CNL],
// [ansi.CPL].
func (s *Screen) moveCursor(dx, dy int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pt := Pos(s.cur.X+dx, s.cur.Y+dy)
	if s.scroll.Contains(pt) {
		s.cur.Y = clamp(pt.Y, s.scroll.Min.Y, s.scroll.Max.Y-1)
		s.cur.X = clamp(pt.X, s.scroll.Min.X, s.scroll.Max.X-1)
	} else {
		s.cur.Y = clamp(pt.Y, 0, s.buf.Height()-1)
		s.cur.X = clamp(pt.X, 0, s.buf.Width()-1)
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
	defer s.mu.Unlock()
	s.saved = s.cur
}

// RestoreCursor restores the cursor.
func (s *Screen) RestoreCursor() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cur = s.saved
}

// setCursorHidden sets the cursor hidden.
func (s *Screen) setCursorHidden(hidden bool) {
	s.mu.Lock()
	s.cur.Hidden = hidden
	s.mu.Unlock()
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

	s.buf.InsertCell(x, y, n, s.blankCell(), s.scroll)
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

	s.buf.DeleteCell(x, y, n, s.blankCell(), s.scroll)
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
func (s *Screen) InsertLine(n int) {
	if n <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	_, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < s.scroll.Min.Y || y >= s.scroll.Max.Y {
		return
	}

	s.buf.InsertLine(y, n, s.blankCell(), s.scroll)
}

// DeleteLine deletes n lines at the cursor position Y coordinate.
// Only operates if cursor is within scroll region. Lines below cursor Y
// are moved up, with blank lines inserted at the bottom of scroll region.
func (s *Screen) DeleteLine(n int) {
	if n <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	_, y := s.cur.X, s.cur.Y

	// Only operate if cursor Y is within scroll region
	if y < s.scroll.Min.Y || y >= s.scroll.Max.Y {
		return
	}

	s.buf.DeleteLine(y, n, s.blankCell(), s.scroll)
}

// blankCell returns the cursor blank cell with the background color set to the
// current pen background color. If the pen background color is nil, the return
// value is nil.
func (s *Screen) blankCell() (c *Cell) {
	if s.cur.Pen.Bg == nil {
		return
	}

	c = new(Cell)
	*c = blankCell
	c.Style.Bg = s.cur.Pen.Bg
	return
}
