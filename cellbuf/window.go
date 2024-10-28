package cellbuf

import "image"

// Window is a 2D grid of cells representing a window in a terminal screen.
type Window struct {
	buf                 *Buffer
	children            []*Window
	x, y, width, height int
}

var _ Screen = &Window{}

// Cell implements Grid.
func (w *Window) Cell(x int, y int) (Cell, bool) {
	if x < 0 || y < 0 || x >= w.width || y >= w.height {
		return Cell{}, false
	}
	return w.buf.Cell(w.x+x, w.y+y)
}

// Resize implements Grid.
func (w *Window) Resize(width int, height int) {
	w.width = width
	w.height = height
}

// SetCell implements Grid.
func (w *Window) SetCell(x int, y int, c Cell) bool {
	if x < 0 || y < 0 || x >= w.x+w.width || y >= w.x+w.height {
		return false
	}
	return w.buf.SetCell(w.x+x, w.y+y, c)
}

// X returns the x position of the window.
func (w *Window) X() int {
	return w.x
}

// Y returns the y position of the window.
func (w *Window) Y() int {
	return w.y
}

// SetX sets the x position of the window.
func (w *Window) SetX(x int) {
	for _, c := range w.children {
		c.x -= w.x
		c.x += x
	}
	w.x = x
}

// SetY sets the y position of the window.
func (w *Window) SetY(y int) {
	for _, c := range w.children {
		c.y -= w.y
		c.y += y
	}
	w.y = y
}

// Move moves the window to the given x, y position.
func (w *Window) Move(x, y int) {
	w.SetX(x)
	w.SetY(y)
}

// Width returns the width of the window.
func (w *Window) Width() int {
	return w.width
}

// Height returns the height of the window.
func (w *Window) Height() int {
	return w.height
}

// Bounds returns the bounds of the window.
func (w *Window) Bounds() image.Rectangle {
	return image.Rect(w.x, w.y, w.x+w.width, w.y+w.height)
}

// InBounds returns whether the given x, y position is within the window.
func (w *Window) InBounds(x, y int) bool {
	return x >= w.x && x < w.x+w.width && y >= w.y && y < w.y+w.height
}

// Child returns a child window of the window whose origin is at the given x, y
// relative to the window. It returns false if the child window is out of bounds.
func (w *Window) Child(x, y, width, height int) *Window {
	c := &Window{
		buf:    w.buf,
		x:      w.x + x,
		y:      w.y + y,
		width:  width,
		height: height,
	}
	w.children = append(w.children, c)
	return c
}

// NewRootWindow returns a new window that represents the entire grid.
func NewRootWindow(buf *Buffer) *Window {
	c := &Window{
		buf:    buf,
		width:  buf.Width(),
		height: buf.Height(),
	}
	return c
}
