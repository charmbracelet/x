package cellbuf

import (
	"image"
)

// Position represents an x, y position.
type Position image.Point

// Point returns the position as an image.Point.
func (p Position) Point() image.Point {
	return image.Point(p)
}

// String returns a string representation of the position.
func (p Position) String() string {
	return image.Point(p).String()
}

// Pos is a shorthand for Position{X: x, Y: y}.
func Pos(x, y int) Position {
	return Position{X: x, Y: y}
}

// Rectange represents a rectangle.
type Rectangle image.Rectangle

// String returns a string representation of the rectangle.
func (r Rectangle) String() string {
	return image.Rectangle(r).String()
}

// Bounds returns the rectangle as an image.Rectangle.
func (r Rectangle) Bounds() image.Rectangle {
	return image.Rectangle(r).Bounds()
}

// Contains reports whether the rectangle contains the given point.
func (r Rectangle) Contains(p Position) bool {
	return image.Point(p).In(r.Bounds())
}

// Width returns the width of the rectangle.
func (r Rectangle) Width() int {
	return image.Rectangle(r).Dx()
}

// Height returns the height of the rectangle.
func (r Rectangle) Height() int {
	return image.Rectangle(r).Dy()
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
	return Rectangle{Min: image.Point{X: x, Y: y}, Max: image.Point{X: x + w, Y: y + h}}
}
