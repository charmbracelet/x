package vt

import "github.com/charmbracelet/x/cellbuf"

// Damage represents a damaged area.
type Damage interface {
	// Bounds returns the bounds of the damaged area.
	Bounds() cellbuf.Rectangle
}

// CellDamage represents a damaged cell.
type CellDamage struct {
	Cell cellbuf.Cell
	X, Y int
}

// Bounds returns the bounds of the damaged area.
func (d CellDamage) Bounds() cellbuf.Rectangle {
	return cellbuf.Rect(d.X, d.Y, d.Cell.Width, 1)
}

// RectDamage represents a damaged rectangle.
type RectDamage cellbuf.Rectangle

// Bounds returns the bounds of the damaged area.
func (d RectDamage) Bounds() cellbuf.Rectangle {
	return cellbuf.Rectangle(d)
}

// X returns the x-coordinate of the damaged area.
func (d RectDamage) X() int {
	return cellbuf.Rectangle(d).X()
}

// Y returns the y-coordinate of the damaged area.
func (d RectDamage) Y() int {
	return cellbuf.Rectangle(d).Y()
}

// Width returns the width of the damaged area.
func (d RectDamage) Width() int {
	return cellbuf.Rectangle(d).Width()
}

// Height returns the height of the damaged area.
func (d RectDamage) Height() int {
	return cellbuf.Rectangle(d).Height()
}

// ScreenDamage represents a damaged screen.
type ScreenDamage struct {
	Width, Height int
}

// Bounds returns the bounds of the damaged area.
func (d ScreenDamage) Bounds() cellbuf.Rectangle {
	return cellbuf.Rect(0, 0, d.Width, d.Height)
}
