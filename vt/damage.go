package vt

// Damage represents a damaged area.
type Damage interface {
	// Bounds returns the bounds of the damaged area.
	Bounds() Rectangle
}

// CellDamage represents a damaged cell.
type CellDamage struct {
	X, Y  int
	Width int
}

// Bounds returns the bounds of the damaged area.
func (d CellDamage) Bounds() Rectangle {
	return Rect(d.X, d.Y, d.Width, 1)
}

// RectDamage represents a damaged rectangle.
type RectDamage Rectangle

// Bounds returns the bounds of the damaged area.
func (d RectDamage) Bounds() Rectangle {
	return Rectangle(d)
}

// X returns the x-coordinate of the damaged area.
func (d RectDamage) X() int {
	return Rectangle(d).X()
}

// Y returns the y-coordinate of the damaged area.
func (d RectDamage) Y() int {
	return Rectangle(d).Y()
}

// Width returns the width of the damaged area.
func (d RectDamage) Width() int {
	return Rectangle(d).Width()
}

// Height returns the height of the damaged area.
func (d RectDamage) Height() int {
	return Rectangle(d).Height()
}

// ScreenDamage represents a damaged screen.
type ScreenDamage struct {
	Width, Height int
}

// Bounds returns the bounds of the damaged area.
func (d ScreenDamage) Bounds() Rectangle {
	return Rect(0, 0, d.Width, d.Height)
}

// MoveDamage represents a moved area.
// The area is moved from the source to the destination.
type MoveDamage struct {
	Src, Dst Rectangle
}

// ScrollDamage represents a scrolled area.
// The area is scrolled by the given deltas.
type ScrollDamage struct {
	Rectangle
	Dx, Dy int
}
