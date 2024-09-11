package cellbuf

import (
	"errors"
)

// Buffer is a 2D grid of cells representing a screen or terminal.
type Buffer struct {
	cells         []Cell
	width, height int
	method        WidthMethod // Defaults to WcWidth
}

// NewBuffer creates a new Buffer with the given width and height.
func NewBuffer(width, height int) *Buffer {
	cells := make([]Cell, width*height)
	b := &Buffer{cells: cells, width: width, height: height}
	b.Fill(spaceCell)
	return b
}

// SetWidthMethod sets the display width calculation method.
func (b *Buffer) SetWidthMethod(method WidthMethod) {
	b.method = method
}

// Reset resets the buffer to the default state.
//
// This is a syntactic sugar for Fill(spaceCell).
func (b *Buffer) Reset() {
	b.Fill(spaceCell)
}

// Size returns the width and height of the buffer.
func (b *Buffer) Size() (width, height int) {
	return b.width, b.height
}

// Resize resizes the buffer to the given width and height.
func (b *Buffer) Resize(width, height int) {
	if width == b.width && height == b.height {
		return
	}

	// Truncate or extend the buffer
	area := width * height
	if area > len(b.cells) {
		newcells := make([]Cell, area-len(b.cells))
		for i := range newcells {
			newcells[i] = spaceCell
		}
		b.cells = append(b.cells, newcells...)
	}

	b.width, b.height = width, height
}

// Free frees extra memory used by the buffer.
func (b *Buffer) Free() {
	area := b.width * b.height
	if area < len(b.cells) {
		b.cells = b.cells[:area]
	}
}

// IsClear returns true if the buffer is empty with only space cells.
func (b *Buffer) IsClear() bool {
	for j := 0; j < b.height; j++ {
		for i := 0; i < b.width; i++ {
			if c, err := b.At(i, j); err == nil && !c.Equal(spaceCell) {
				return false
			}
		}
	}
	return true
}

// ErrOutOfBounds is returned when the given x, y position is out of bounds.
var ErrOutOfBounds = errors.New("out of bounds")

// At returns the cell at the given x, y position.
func (b *Buffer) At(x, y int) (Cell, error) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
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
	for j := 0; j < b.height; j++ {
		for i := 0; i < b.width; i++ {
			b.Set(i, j, c) //nolint:errcheck
		}
	}
}

// FillFunc fills the buffer with the given function.
func (b *Buffer) FillFunc(f func(c Cell) Cell) {
	for j := 0; j < b.height; j++ {
		for i := 0; i < b.width; i++ {
			idx := j*b.width + i
			if c, err := b.At(i, j); err == nil {
				b.cells[idx] = f(c)
			}
		}
	}
}

// FillRange fills the buffer with the given Cell in the given range.
func (b *Buffer) FillRange(x, y, w, h int, c Cell) {
	for i := y; i < y+h; i++ {
		for j := x; j < x+w; j++ {
			b.Set(j, i, c) //nolint:errcheck
		}
	}
}

// FillRangeFunc fills the buffer with the given function in the given range.
func (b *Buffer) FillRangeFunc(x, y, w, h int, f func(c Cell) Cell) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			b.SetFunc(i, j, f) //nolint:errcheck
		}
	}
}

// Set sets the cell at the given x, y position.
func (b *Buffer) Set(x, y int, c Cell) error {
	if x > b.width-1 || y > b.height-1 {
		return ErrOutOfBounds
	}
	idx := y*b.width + x
	if idx < 0 || idx >= len(b.cells) {
		return ErrOutOfBounds
	}

	b.cells[idx] = c

	return nil
}

// SetFunc sets the cell at the given x, y position using a function.
func (b *Buffer) SetFunc(x, y int, f func(c Cell) Cell) error {
	c, err := b.At(x, y)
	if err != nil {
		return err
	}
	c = f(c)
	return b.Set(x, y, c)
}
