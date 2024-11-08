package cellbuf

import (
	"fmt"
	"image"
)

// Position represents an x, y position.
type Position image.Point

// String returns a string representation of the position.
func (p Position) String() string {
	return image.Point(p).String()
}

// Pos is a shorthand for Position{X: x, Y: y}.
func Pos(x, y int) Position {
	return Position{X: x, Y: y}
}

// Rectange represents a rectangle.
type Rectangle struct {
	X, Y, Width, Height int
}

// String returns a string representation of the rectangle.
func (r Rectangle) String() string {
	return fmt.Sprintf("(%d,%d)-(%d,%d)", r.X, r.Y, r.X+r.Width, r.Y+r.Height)
}

// Bounds returns the rectangle as an image.Rectangle.
func (r Rectangle) Bounds() image.Rectangle {
	return image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height)
}

// Contains reports whether the rectangle contains the given point.
func (r Rectangle) Contains(p Position) bool {
	return image.Point(p).In(r.Bounds())
}

// Rect is a shorthand for Rectangle.
func Rect(x, y, w, h int) Rectangle {
	return Rectangle{X: x, Y: y, Width: w, Height: h}
}
