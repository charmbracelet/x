package cellbuf

// Buffer is a 2D grid of cells representing a screen or terminal.
type Buffer struct {
	cells []Cell
	width int
}

// Width returns the width of the buffer.
func (b *Buffer) Width() int {
	return b.width
}

// Height returns the height of the buffer.
func (b *Buffer) Height() int {
	if b.width == 0 {
		return 0
	}
	return len(b.cells) / b.width
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

// Set sets the cell at the given x, y position.
func (b *Buffer) Set(x, y int, c Cell) (v bool) {
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
	return true
}

// Clone returns a deep copy of the buffer.
func (b *Buffer) Clone() *Buffer {
	var clone Buffer
	clone.width = b.width
	clone.cells = make([]Cell, len(b.cells))
	copy(clone.cells, b.cells)
	return &clone
}

// Resize resizes the buffer to the given width and height. It grows the buffer
// if necessary and fills the new cells with space cells. Otherwise, it
// truncates the buffer.
func (b *Buffer) Resize(width, height int) {
	b.width = width
	if area := width * height; len(b.cells) < area {
		ln := len(b.cells)
		b.cells = append(b.cells, make([]Cell, area-ln)...)
		// Fill the buffer with space cells
		for i := ln; i < area; i++ {
			b.cells[i] = spaceCell
		}
	} else if len(b.cells) > area {
		// Truncate the buffer if necessary
		b.cells = b.cells[:area]
	}
}
