package vt

import "github.com/charmbracelet/x/cellbuf"

// Screen represents a virtual terminal screen.
type Screen struct {
	// The buffer of the screen.
	buf cellbuf.Buffer
	// The cur of the screen.
	cur, saved Cursor
}

var _ cellbuf.Grid = &Screen{}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := new(Screen)
	s.buf.Resize(w, h)
	return s
}

// At implements cellbuf.Grid.
func (s *Screen) At(x int, y int) (cellbuf.Cell, error) {
	return s.buf.At(x, y)
}

// Height implements cellbuf.Grid.
func (s *Screen) Height() int {
	return s.buf.Height()
}

// Resize implements cellbuf.Grid.
func (s *Screen) Resize(width int, height int) {
	s.buf.Resize(width, height)
}

// Set implements cellbuf.Grid.
func (s *Screen) Set(x int, y int, c cellbuf.Cell) bool {
	return s.buf.Set(x, y, c)
}

// Width implements cellbuf.Grid.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// Pos returns the cursor position.
func (s *Screen) Pos() (int, int) {
	return s.cur.Pos.X, s.cur.Pos.Y
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
	s.cur.Visible = true
}

// HideCursor hides the cursor.
func (s *Screen) HideCursor() {
	s.cur.Visible = false
}

// moveCursor moves the cursor.
func (s *Screen) moveCursor(x, y int) {
	s.cur.Pos.X = x
	s.cur.Pos.Y = y
}