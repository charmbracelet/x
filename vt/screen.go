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
	return s.buf.Draw(x, y, c)
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
	s.mu.Unlock()
	s.scroll = s.Bounds()
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
	defer s.mu.Unlock()
	s.buf.Clear(rects...)
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

	s.buf.InsertCell(x, y, n, s.scroll)
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

	s.buf.DeleteCell(x, y, n, s.scroll)
}

// ScrollUp scrolls the screen up n lines within the scroll region.
// Lines scrolled past the top margin are lost.
func (s *Screen) ScrollUp(n int) {
	if n <= 0 {
		return
	}

	s.mu.RLock()
	if n > s.scroll.Height() {
		n = s.scroll.Height()
	}
	s.mu.RUnlock()

	if s.scroll == s.Bounds() {
		s.mu.Lock()
		// OPTIM: for scrolling the whole screen.
		// Move lines up, dropping the top n lines
		s.buf.lines = s.buf.lines[n:]
		for i := 0; i < n; i++ {
			s.buf.lines = append(s.buf.lines, make(Line, s.buf.Width()))
		}
		s.mu.Unlock()
	} else {
		// Copy lines up within scroll region
		for i := s.scroll.Min.Y; i < s.scroll.Max.Y-n; i++ {
			for x := s.scroll.Min.X; x < s.scroll.Max.X; x++ {
				c, _ := s.Cell(x, i+n)
				s.Draw(x, i, c)
			}
		}
	}

	// Clear the bottom n lines of the scroll region
	s.Clear(Rect(s.scroll.Min.X, s.scroll.Max.Y-n, s.scroll.Max.X, s.scroll.Max.Y))
}

// ScrollDown scrolls the screen down n lines within the scroll region.
// Lines scrolled past the bottom margin are lost.
func (s *Screen) ScrollDown(n int) {
	if n <= 0 {
		return
	}

	s.mu.RLock()
	if n > s.scroll.Height() {
		n = s.scroll.Height()
	}
	s.mu.RUnlock()

	if s.scroll == s.Bounds() {
		s.mu.Lock()
		// OPTIM: for scrolling the whole screen.
		// Move lines down, dropping the bottom n lines
		s.buf.lines = s.buf.lines[:len(s.buf.lines)-n]
		for i := 0; i < n; i++ {
			s.buf.lines = append([]Line{make(Line, s.buf.Width())}, s.buf.lines...)
		}
		s.mu.Unlock()
	} else {
		// Copy lines down within scroll region
		for i := s.scroll.Max.Y - 1; i >= s.scroll.Min.Y+n; i-- {
			for x := s.scroll.Min.X; x < s.scroll.Max.X; x++ {
				c, _ := s.Cell(x, i-n)
				s.Draw(x, i, c)
			}
		}
	}

	// Clear the top n lines of the scroll region
	s.Clear(Rect(s.scroll.Min.X, s.scroll.Min.Y, s.scroll.Max.X, s.scroll.Min.Y+n))
}
