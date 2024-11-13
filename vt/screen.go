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
	return s.SetCell(x, y, c)
}

// SetCell sets the cell at the given x, y position.
// It returns true if the cell was set successfully.
func (s *Screen) SetCell(x, y int, c Cell) bool {
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
func (s *Screen) Fill(c Cell, rects ...Rectangle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf.Fill(c, rects...)
}

// setCursorX sets the cursor X position.
func (s *Screen) setCursorX(x int) {
	s.mu.Lock()
	s.cur.X = x
	s.mu.Unlock()
}

// setCursorY sets the cursor Y position.
func (s *Screen) setCursorY(y int) {
	s.mu.Lock()
	s.cur.Y = y
	s.mu.Unlock()
}

// setCursor sets the cursor position.
func (s *Screen) setCursor(x, y int) {
	s.mu.Lock()
	s.cur.X = x
	s.cur.Y = y
	s.mu.Unlock()
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

// ScrollUp scrolls the content up n lines within the given region.
// Lines scrolled past the top margin are lost.
func (s *Screen) ScrollUp(n int) {
	s.mu.Lock()
	s.scrollUp(n, s.scroll)
	s.mu.Unlock()
}

func (s *Screen) scrollUp(n int, rect Rectangle) {
	if n <= 0 {
		return
	}

	if n > rect.Height() {
		n = rect.Height()
	}

	if rect == s.buf.Bounds() {
		// OPTIM: for scrolling the whole screen.
		// Move lines up, dropping the top n lines
		s.buf.lines = s.buf.lines[n:]
		for i := 0; i < n; i++ {
			s.buf.lines = append(s.buf.lines, make(Line, s.buf.Width()))
		}
	} else {
		// Copy lines up within region
		for i := rect.Min.Y; i < rect.Max.Y-n; i++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				c, _ := s.buf.Cell(x, i+n)
				s.buf.SetCell(x, i, c)
			}
		}
	}

	// Clear the bottom n lines of the region
	s.buf.Clear(Rect(rect.Min.X, rect.Max.Y-n, rect.Max.X, rect.Max.Y))
}

// ScrollDown scrolls the content down n lines within the given region.
// Lines scrolled past the bottom margin are lost.
func (s *Screen) ScrollDown(n int) {
	s.mu.Lock()
	s.scrollDown(n, s.scroll)
	s.mu.Unlock()
}

func (s *Screen) scrollDown(n int, rect Rectangle) {
	if n <= 0 {
		return
	}

	if n > rect.Height() {
		n = rect.Height()
	}

	if rect == s.buf.Bounds() {
		// OPTIM: for scrolling the whole screen.
		// Move lines down, dropping the bottom n lines
		s.buf.lines = s.buf.lines[:len(s.buf.lines)-n]
		for i := 0; i < n; i++ {
			s.buf.lines = append([]Line{make(Line, s.buf.Width())}, s.buf.lines...)
		}
	} else {
		// Copy lines down within region
		for i := rect.Max.Y - 1; i >= rect.Min.Y+n; i-- {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				c, _ := s.buf.Cell(x, i-n)
				s.buf.SetCell(x, i, c)
			}
		}
	}

	// Clear the top n lines of the region
	s.buf.Clear(Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+n))
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
