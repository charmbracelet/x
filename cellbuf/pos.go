package cellbuf

import (
	"strconv"
)

// A Position is an X, Y coordinate pair. The axes increase right and down.
type Position struct {
	X, Y int
}

// String returns a string representation of p like "(3,4)".
func (p Position) String() string {
	return "(" + strconv.Itoa(p.X) + "," + strconv.Itoa(p.Y) + ")"
}

// In reports whether p is in r.
func (p Position) In(r Rectangle) bool {
	return r.Start.X <= p.X && p.X < r.End.X &&
		r.Start.Y <= p.Y && p.Y < r.End.Y
}

// Eq reports whether p and q are equal.
func (p Position) Equal(q Position) bool {
	return p == q
}

// Pos is a shorthand for Position{X, Y}.
func Pos(x, y int) Position {
	return Position{x, y}
}

// Rectangle represents a rectangle in a grid.
type Rectangle struct {
	Start Position
	End   Position
}

// Width returns the width of r.
func (r Rectangle) Width() int {
	return r.End.X - r.Start.X
}

// Height returns the height of r.
func (r Rectangle) Height() int {
	return r.End.Y - r.Start.Y
}

// String returns a string representation of r like "(3,4)-(6,5)".
func (r Rectangle) String() string {
	return r.Start.String() + "-" + r.End.String()
}

// Size returns r's width and height.
func (r Rectangle) Size() (int, int) {
	return r.Width(), r.Height()
}

// Empty reports whether the rectangle contains no points.
func (r Rectangle) Empty() bool {
	return r.Start.X >= r.End.X || r.Start.Y >= r.End.Y
}

// Equal reports whether r and s contain the same set of points. All empty
// rectangles are considered equal.
func (r Rectangle) Equal(s Rectangle) bool {
	return r == s || r.Empty() && s.Empty()
}

// In reports whether every point in r is in s.
func (r Rectangle) In(s Rectangle) bool {
	if r.Empty() {
		return true
	}
	// Note that r.Max is an exclusive bound for r, so that r.In(s)
	// does not require that r.Max.In(s).
	return s.Start.X <= r.Start.X && r.End.X <= s.End.X &&
		s.Start.Y <= r.Start.Y && r.End.Y <= s.End.Y
}

// Rect is a shorthand for Rectangle{Start: Position{X1, Y1}, End: Position{X2, Y2}}.
func Rect(x0, y0, x1, y1 int) Rectangle {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return Rectangle{Start: Pos(x0, y0), End: Pos(x1, y1)}
}
