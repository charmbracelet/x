package cellbuf

// Window represents a terminal window that can be written to. A window has a
// width and height, and starts at the given x, y position.
type Window interface {
	Width() int
	Height() int
	AbsX() int
	AbsY() int
	At(x, y int) (Cell, error)
	Set(x, y int, c Cell)
	SetContent(content string) []int

	// Child returns a new window that is a child of the current window. The
	// child window starts at the given x, y position and has the given width
	// and height.
	// A child window cannot be larger than the parent nor can it start outside
	// of the parent window.
	Child(x, y, width, height int) (Window, error)
}

type childWindow struct {
	parent Window
	buf    Grid
	x, y   int // relative to parent
	w, h   int
}

var _ Window = &childWindow{}

// newChildWindow creates a new child window.
func newChildWindow(buf Grid, parent Window, x, y, width, height int) (*childWindow, error) {
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return nil, ErrOutOfBounds
	}
	if x+width > parent.Width() || y+height > parent.Height() {
		return nil, ErrOutOfBounds
	}
	if width > parent.Width() || height > parent.Height() {
		return nil, ErrOutOfBounds
	}
	if x+width < 0 || y+height < 0 {
		return nil, ErrOutOfBounds
	}
	return &childWindow{
		buf:    buf,
		parent: parent,
		x:      x, y: y,
		w: width, h: height,
	}, nil
}

// AbsX implements Window.
func (c *childWindow) AbsX() int {
	return c.parent.AbsX() + c.x
}

// AbsY implements Window.
func (c *childWindow) AbsY() int {
	return c.parent.AbsY() + c.y
}

// Child implements Window.
func (c *childWindow) Child(x int, y int, width int, height int) (Window, error) {
	return newChildWindow(c.buf, c.parent, x, y, width, height)
}

// Height implements Window.
func (c *childWindow) Height() int {
	return c.h
}

// At implements Window.
func (c *childWindow) At(x int, y int) (Cell, error) {
	return c.parent.At(x+c.x, y+c.y)
}

// Set implements Window.
func (c *childWindow) Set(x int, y int, cell Cell) {
	c.parent.Set(x+c.x, y+c.y, cell)
}

// SetContent implements Window.
func (c *childWindow) SetContent(content string) []int {
	return setStringContent(c.buf, content, c.AbsX(), c.AbsY(), c.w, c.h, c.buf.Method())
}

// Width implements Window.
func (c *childWindow) Width() int {
	return c.w
}
