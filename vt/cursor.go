package vt

import (
	"image"
)

// Cursor represents a cursor in a terminal.
type Cursor struct {
	Pen     Style
	Pos     image.Point
	Style   int
	Visible bool
}
