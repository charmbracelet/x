package vt

import "github.com/charmbracelet/uv"

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
	Pen  uv.Style
	Link uv.Link

	uv.Position

	Style  CursorStyle
	Steady bool // Not blinking
	Hidden bool
}
