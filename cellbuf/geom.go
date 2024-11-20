package cellbuf

import (
	"github.com/charmbracelet/x/vt"
)

// Position represents an x, y position.
type Position = vt.Position

// Pos is a shorthand for Position{X: x, Y: y}.
func Pos(x, y int) Position {
	return vt.Pos(x, y)
}

// Rectange represents a rectangle.
type Rectangle = vt.Rectangle

// Rect is a shorthand for Rectangle.
func Rect(x, y, w, h int) Rectangle {
	return vt.Rect(x, y, w, h)
}
