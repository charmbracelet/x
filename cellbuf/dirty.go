package cellbuf

// DirtyCell represents a cell with a dirty flag.
type DirtyCell struct {
	Cell
	Dirty bool
}

// DirtyBuffer is a safe cell buffer that tracks dirty cells.
type DirtyBuffer struct {
	cells  []DirtyCell
	width  int
	method WidthMethod
}

var _ Grid = &DirtyBuffer{}

// NewDirtyBuffer creates a new DirtyBuffer with the given width and method.
func NewDirtyBuffer(width int, method WidthMethod) *DirtyBuffer {
	d := &DirtyBuffer{
		cells:  make([]DirtyCell, width),
		width:  width,
		method: method,
	}
	return d
}

// Method returns the width method used by the buffer.
func (d *DirtyBuffer) Method() WidthMethod {
	return d.method
}

// Commit marks all cells as clean.
func (d *DirtyBuffer) Commit() {
	for i := range d.cells {
		d.cells[i].Dirty = false
	}
}

// IsDirty returns true if the cell at the given position is dirty.
func (d *DirtyBuffer) IsDirty(x, y int) bool {
	idx := y*d.width + x
	if idx < 0 || idx >= len(d.cells) {
		return false
	}
	return d.cells[idx].Dirty
}

// SetContent writes the given data to the buffer starting from the first cell.
func (d *DirtyBuffer) SetContent(content string) []int {
	height := Height(content)
	if area := d.width * height; len(d.cells) < area {
		ln := len(d.cells)
		d.cells = append(d.cells, make([]DirtyCell, area-ln)...)
		// Fill the buffer with space cells
		for i := ln; i < area; i++ {
			d.cells[i].Cell = spaceCell
		}
	} else if len(d.cells) > area {
		// Truncate the buffer if necessary
		d.cells = d.cells[:area]
	}

	return setStringContent(d, content, 0, 0, d.width, height, d.method)
}

// SetWidth sets the width of the buffer.
func (d *DirtyBuffer) SetWidth(width int) {
	d.width = width
}

// Width returns the width of the buffer.
func (d *DirtyBuffer) Width() int {
	return d.width
}

// Height returns the height of the buffer.
func (d *DirtyBuffer) Height() int {
	return len(d.cells) / d.width
}

// At implements Grid.
func (d *DirtyBuffer) At(x int, y int) (Cell, error) {
	idx := y*d.width + x
	if idx < 0 || idx >= len(d.cells) {
		return Cell{}, ErrOutOfBounds
	}
	return d.cells[idx].Cell, nil
}

// Set implements Grid.
func (d *DirtyBuffer) Set(x int, y int, c Cell) {
	idx := y*d.width + x
	if idx < 0 || idx >= len(d.cells) {
		return
	}

	if d.cells[idx].Cell.Equal(c) {
		return
	}

	d.cells[idx].Cell = c
	d.cells[idx].Dirty = true
}
