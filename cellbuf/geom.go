package cellbuf

import (
	"image"
)

// Position represents an x, y position.
type Position = image.Point

// Pos is a shorthand for Position{X: x, Y: y}.
func Pos(x, y int) Position {
	return image.Pt(x, y)
}

// Rectange represents a rectangle.
type Rectangle struct {
	image.Rectangle
}

// Contains reports whether the rectangle contains the given point.
func (r Rectangle) Contains(p Position) bool {
	return p.In(r.Bounds())
}

// Width returns the width of the rectangle.
func (r Rectangle) Width() int {
	return r.Rectangle.Dx()
}

// Height returns the height of the rectangle.
func (r Rectangle) Height() int {
	return r.Rectangle.Dy()
}

// X returns the starting x position of the rectangle.
// This is equivalent to Min.X.
func (r Rectangle) X() int {
	return r.Min.X
}

// Y returns the starting y position of the rectangle.
// This is equivalent to Min.Y.
func (r Rectangle) Y() int {
	return r.Min.Y
}

// Rect is a shorthand for Rectangle.
func Rect(x, y, w, h int) Rectangle {
	return Rectangle{image.Rect(x, y, x+w, y+h)}
}
