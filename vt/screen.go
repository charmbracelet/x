package vt

import "github.com/charmbracelet/x/cellbuf"

// Screen represents a virtual terminal screen.
type Screen struct {
	// The buffer of the screen.
	buf cellbuf.Buffer
	// The cur of the screen.
	cur, saved Cursor

	// tabstop is the list of tab stops.
	tabstops TabStops
}

var _ cellbuf.Screen = &Screen{}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := new(Screen)
	s.buf.Resize(w, h)
	s.tabstops = DefaultTabStops(w)
	return s
}

// Cell implements cellbuf.Screen.
func (s *Screen) Cell(x int, y int) (cellbuf.Cell, bool) {
	return s.buf.Cell(x, y)
}

// Draw implements cellbuf.Screen.
func (s *Screen) Draw(x int, y int, c cellbuf.Cell) bool {
	return s.buf.Draw(x, y, c)
}

// Height implements cellbuf.Grid.
func (s *Screen) Height() int {
	return s.buf.Height()
}

// Resize implements cellbuf.Grid.
func (s *Screen) Resize(width int, height int) {
	s.buf.Resize(width, height)
	s.tabstops = DefaultTabStops(width)
}

// Width implements cellbuf.Grid.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// Clear clears the screen or part of it.
func (s *Screen) Clear(rect *cellbuf.Rectangle) {
	s.buf.Clear(rect)
}

// Fill fills the screen or part of it.
func (s *Screen) Fill(c cellbuf.Cell, rect *cellbuf.Rectangle) {
	s.buf.Fill(c, rect)
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
