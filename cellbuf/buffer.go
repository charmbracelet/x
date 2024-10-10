package cellbuf

// Buffer is a 2D grid of cells representing a screen or terminal.
type Buffer struct {
	cells  []Cell
	width  int
	method WidthMethod // Defaults to WcWidth
}

// NewBuffer creates a new Buffer with the given width and height.
func NewBuffer(width int, method WidthMethod) *Buffer {
	b := &Buffer{
		cells:  make([]Cell, width),
		width:  width,
		method: method,
	}
	return b
}

// SetMethod sets the width method used by the buffer.
func (b *Buffer) SetMethod(method WidthMethod) {
	b.method = method
}

// Method returns the width method used by the buffer.
func (b *Buffer) Method() WidthMethod {
	return b.method
}

// Equal returns true if the buffer is equal to the other buffer.
func (b *Buffer) Equal(o *Buffer) bool {
	if b.width != o.width {
		return false
	}
	if len(b.cells) != len(o.cells) {
		return false
	}
	for i := range b.cells {
		if !b.cells[i].Equal(o.cells[i]) {
			return false
		}
	}
	return true
}

// Width returns the width of the buffer.
func (b *Buffer) Width() int {
	return b.width
}

// Height returns the height of the buffer.
func (b *Buffer) Height() int {
	return len(b.cells) / b.width
}

// Size returns the width and height of the buffer.
func (b *Buffer) Size() (width, height int) {
	height = len(b.cells) / b.width
	return b.width, height
}

// At returns the cell at the given x, y position.
func (b *Buffer) At(x, y int) (Cell, error) {
	if b.width == 0 {
		return Cell{}, ErrOutOfBounds
	}
	height := len(b.cells) / b.width
	if x < 0 || x >= b.width || y < 0 || y >= height {
		return Cell{}, ErrOutOfBounds
	}
	idx := y*b.width + x
	if idx < 0 || idx >= len(b.cells) {
		return Cell{}, ErrOutOfBounds
	}
	return b.cells[idx], nil
}

// Fill fills the buffer with the given style and rune.
func (b *Buffer) Fill(c Cell) {
	if b.width == 0 {
		return
	}
	height := len(b.cells) / b.width
	for j := 0; j < height; j++ {
		for i := 0; i < b.width; i++ {
			b.Set(i, j, c) //nolint:errcheck
		}
	}
}

// Set sets the cell at the given x, y position.
func (b *Buffer) Set(x, y int, c Cell) {
	if b.width == 0 {
		return
	}
	height := len(b.cells) / b.width
	if x > b.width-1 || y > height-1 {
		return
	}
	idx := y*b.width + x
	if idx < 0 || idx >= len(b.cells) {
		return
	}

	b.cells[idx] = c
}

// SetFunc sets the cell at the given x, y position using a function.
func (b *Buffer) SetFunc(x, y int, f func(c Cell) Cell) {
	c, err := b.At(x, y)
	if err != nil {
		return
	}
	b.Set(x, y, f(c))
}

// lastInLine returns true if the cell is the last non-space cell in the line.
func (b *Buffer) lastInLine(x, y int) bool {
	for i := x + 1; i < b.width; i++ {
		if cell, err := b.At(i, y); err == nil && !cell.Equal(spaceCell) {
			return false
		}
	}
	return true
}

// Clone returns a deep copy of the buffer.
func (b *Buffer) Clone() *Buffer {
	clone := NewBuffer(b.width, b.method)
	clone.cells = make([]Cell, len(b.cells))
	copy(clone.cells, b.cells)
	return clone
}
