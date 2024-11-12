package vt

import (
	"github.com/charmbracelet/x/cellbuf"
)

// Screen represents a virtual terminal screen.
type Screen struct {
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
	return s.buf.Bounds()
}

// Cell implements cellbuf.Screen.
func (s *Screen) Cell(x int, y int) (Cell, bool) {
	return s.buf.Cell(x, y)
}

// Draw implements cellbuf.Screen.
func (s *Screen) Draw(x int, y int, c Cell) bool {
	return s.buf.Draw(x, y, c)
}

// Height implements cellbuf.Grid.
func (s *Screen) Height() int {
	return s.buf.Height()
}

// Resize implements cellbuf.Grid.
func (s *Screen) Resize(width int, height int) {
	s.buf.Resize(width, height)
	s.scroll = s.Bounds()
}

// Width implements cellbuf.Grid.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// Clear clears the screen or part of it.
func (s *Screen) Clear(rects ...Rectangle) {
	s.buf.Clear(rects...)
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c Cell, rects ...Rectangle) {
	s.buf.Fill(c, rects...)
}

// Pos returns the cursor position.
func (s *Screen) Pos() (int, int) {
	return s.cur.X, s.cur.Y
}

// moveCursor moves the cursor.
func (s *Screen) moveCursor(x, y int) {
	s.cur.X = x
	s.cur.Y = y
}

// Cursor returns the cursor.
func (s *Screen) Cursor() Cursor {
	return s.cur
}

// SaveCursor saves the cursor.
func (s *Screen) SaveCursor() {
	s.saved = s.cur
}

// RestoreCursor restores the cursor.
func (s *Screen) RestoreCursor() {
	s.cur = s.saved
}

// ShowCursor shows the cursor.
func (s *Screen) ShowCursor() {
	s.cur.Hidden = false
}

// HideCursor hides the cursor.
func (s *Screen) HideCursor() {
	s.cur.Hidden = true
}

// InsertCell inserts n blank characters at the cursor position pushing out
// cells to the right and out of the screen.
func (s *Screen) InsertCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y

	s.buf.InsertCell(x, y, n, s.scroll)
}

// DeleteCell deletes n cells at the cursor position moving cells to the left.
// This has no effect if the cursor is outside the scroll region.
func (s *Screen) DeleteCell(n int) {
	if n <= 0 {
		return
	}

	x, y := s.cur.X, s.cur.Y

	s.buf.DeleteCell(x, y, n, s.scroll)
}

// ScrollUp scrolls the screen up n lines within the scroll region.
// Lines scrolled past the top margin are lost.
func (s *Screen) ScrollUp(n int) {
	if n <= 0 {
		return
	}

	if n > s.scroll.Height() {
		n = s.scroll.Height()
	}

	if s.scroll == s.Bounds() {
		// OPTIM: for scrolling the whole screen.
		// Move lines up, dropping the top n lines
		copy(s.buf.lines[0:], s.buf.lines[n:])
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

	if n > s.scroll.Height() {
		n = s.scroll.Height()
	}

	if s.scroll == s.Bounds() {
		// OPTIM: for scrolling the whole screen.
		// Move lines down, dropping the bottom n lines
		copy(s.buf.lines[n:], s.buf.lines[0:len(s.buf.lines)-n])
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
