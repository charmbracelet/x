package vt

// CursorStyle represents a cursor style.
type CursorStyle int

// Cursor styles.
const (
	CursorBlock CursorStyle = iota
	CursorUnderline
	CursorBar
)

// Cursor represents a cursor in a terminal.
type Cursor struct {
	Position

	Pen    Style
	Style  CursorStyle
	Steady bool // Not blinking
	Hidden bool
}
