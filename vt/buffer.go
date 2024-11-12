package vt

import (
	"slices"

	"github.com/charmbracelet/x/cellbuf"
)

// Position represents a position in the terminal.
type Position = cellbuf.Position

// Pos returns a new position.
func Pos(x int, y int) Position {
	return cellbuf.Pos(x, y)
}

// Rectangle represents a rectangle in the terminal.
type Rectangle = cellbuf.Rectangle

// Rect returns a new rectangle.
func Rect(x, y, w, h int) Rectangle {
	return cellbuf.Rect(x, y, w, h)
}

// Cell represents a cell in the terminal.
type Cell = cellbuf.Cell

// Line represents a line in the terminal.
type Line []Cell

// Width returns the width of the line.
func (l Line) Width() int {
	return len(l)
}

// Buffer is a 2D grid of cells representing a screen or terminal.
type Buffer struct {
	lines []Line
}

var _ cellbuf.Screen = &Buffer{}

// Cell implements cellbuf.Screen.
func (b *Buffer) Cell(x int, y int) (cellbuf.Cell, bool) {
	if y < 0 || y >= len(b.lines) {
		return cellbuf.Cell{}, false
	}
	if x < 0 || x >= b.lines[y].Width() {
		return cellbuf.Cell{}, false
	}
	return b.lines[y][x], true
}

// Draw implements cellbuf.Screen.
func (b *Buffer) Draw(x int, y int, c cellbuf.Cell) bool {
	return b.SetCell(x, y, c)
}

// SetCell sets the cell at the given x, y position.
func (b *Buffer) SetCell(x, y int, c Cell) bool {
	if y < 0 || y >= len(b.lines) {
		return false
	}
	if x < 0 || x >= b.lines[y].Width() {
		return false
	}

	// When a wide cell is partially overwritten, we need
	// to fill the rest of the cell with space cells to
	// avoid rendering issues.
	prev := b.lines[y][x]
	if prev.Width > 1 {
		// Writing to the first wide cell
		for j := 0; j < prev.Width && x+j < b.lines[y].Width(); j++ {
			newCell := prev
			newCell.Content = " "
			newCell.Width = 1
			b.lines[y][x+j] = newCell
		}
	} else if prev.Width == 0 {
		// Writing to wide cell placeholders
		for j := 1; j < 4 && x-j >= 0; j++ {
			wide := b.lines[y][x-j]
			if wide.Width > 1 {
				for k := 0; k < wide.Width; k++ {
					newCell := wide
					newCell.Content = " "
					newCell.Width = 1
					b.lines[y][x-j+k] = newCell
				}
				break
			}
		}
	}

	b.lines[y][x] = c

	// Mark wide cells with emptyCell zero width
	// We set the wide cell down below
	if c.Width > 1 {
		for j := 1; j < c.Width && x+j < b.lines[y].Width(); j++ {
			b.lines[y][x+j] = Cell{}
		}
	}

	return true
}

// Height implements cellbuf.Screen.
func (b *Buffer) Height() int {
	return len(b.lines)
}

// Width implements cellbuf.Screen.
func (b *Buffer) Width() int {
	if len(b.lines) == 0 {
		return 0
	}
	return b.lines[0].Width()
}

// Bounds returns the bounds of the buffer.
func (b *Buffer) Bounds() Rectangle {
	return Rect(0, 0, b.Width(), b.Height())
}

// Resize resizes the buffer to the given width and height.
func (b *Buffer) Resize(width int, height int) {
	if width == 0 || height == 0 {
		b.lines = nil
		return
	}

	if height > len(b.lines) {
		for i := len(b.lines); i < height; i++ {
			b.lines = append(b.lines, make(Line, width))
		}
	} else if height < len(b.lines) {
		b.lines = b.lines[:height]
	}

	if width > b.Width() {
		line := make(Line, width-b.Width())
		// Fill the new cells with space cells
		for i := range line {
			line[i] = spaceCell
		}

		for i := range b.lines {
			b.lines[i] = append(b.lines[i], line...)
		}
	} else if width < b.Width() {
		for i := range b.lines {
			b.lines[i] = b.lines[i][:width]
		}
	}
}

// fill fills the buffer with the given cell and rectangle.
func (b *Buffer) fill(c cellbuf.Cell, rect cellbuf.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x += c.Width {
			b.Draw(x, y, c) //nolint:errcheck
		}
	}
}

// Fill fills the buffer with the given cell and rectangle.
func (b *Buffer) Fill(c cellbuf.Cell, rects ...cellbuf.Rectangle) {
	if len(rects) == 0 {
		b.fill(c, b.Bounds())
		return
	}
	for _, rect := range rects {
		b.fill(c, rect)
	}
}

// Clear clears the buffer with space cells and rectangle.
func (b *Buffer) Clear(rects ...cellbuf.Rectangle) {
	if len(rects) == 0 {
		b.Fill(spaceCell, b.Bounds())
		return
	}
	for _, rect := range rects {
		b.Fill(spaceCell, rect)
	}
}

// insertLine inserts a new line at the given position.
func (b *Buffer) InsertLine(y, n int, rects ...Rectangle) {
	if len(rects) == 0 {
		b.insertLineInRect(y, n, b.Bounds())
	}
	for _, rect := range rects {
		b.insertLineInRect(y, n, rect)
	}
}

// insertLineInRect inserts new lines at the given position within the rectangle bounds.
// Only cells within the rectangle's horizontal bounds are affected.
func (b *Buffer) insertLineInRect(y, n int, rect Rectangle) {
	if n <= 0 || y < rect.Min.Y || y >= rect.Max.Y || y >= b.Height() {
		return
	}

	// Limit number of lines to insert to available space
	if y+n > rect.Max.Y {
		n = rect.Max.Y - y
	}

	// For each line we need to insert
	for i := 0; i < n; i++ {
		// Create a new line copying cells from outside the rectangle bounds
		newLine := make(Line, b.Width())
		copy(newLine, b.lines[y])

		// Clear only the cells within rectangle bounds
		for x := rect.Min.X; x < rect.Max.X && x < len(newLine); x++ {
			newLine[x] = spaceCell
		}

		// Insert the new line
		b.lines = slices.Insert(b.lines, y, newLine)
	}

	// Remove excess lines that got pushed below rectangle bounds
	if y+n < b.Height() && rect.Max.Y < b.Height() {
		b.lines = slices.Delete(b.lines, rect.Max.Y, rect.Max.Y+n)
	}
}

// DeleteLine deletes n lines at the given position in the rectangle.
func (b *Buffer) DeleteLine(y, n int) {
	if n <= 0 || y < 0 || y >= b.Height() {
		return
	}

	// Delete n lines at the given y position.
	b.lines = slices.Delete(b.lines, y, y+n)
}

// InsertCell inserts new cells at the given position within the specified rectangles.
// If no rectangles are specified, it inserts cells in the entire buffer.
func (b *Buffer) InsertCell(x, y, n int, rects ...Rectangle) {
	if len(rects) == 0 {
		b.insertCellInRect(x, y, n, b.Bounds())
		return
	}
	for _, rect := range rects {
		b.insertCellInRect(x, y, n, rect)
	}
}

// insertCellInRect inserts new cells at the given position within the rectangle bounds.
// Only cells within the rectangle's bounds are affected, following terminal ICH behavior.
func (b *Buffer) insertCellInRect(x, y, n int, rect Rectangle) {
	if n <= 0 || y < rect.Min.Y || y >= rect.Max.Y || y >= b.Height() ||
		x < rect.Min.X || x >= rect.Max.X || x >= b.Width() {
		return
	}

	// Create a new line preserving cells outside rectangle
	newLine := make(Line, b.Width())
	copy(newLine, b.lines[y])

	// Calculate how many positions we can actually insert
	remainingSpace := rect.Max.X - x
	if n > remainingSpace {
		n = remainingSpace
	}

	// Shift cells to the right starting from the right edge
	for i := rect.Max.X - 1; i >= x+n; i-- {
		newLine[i] = b.lines[y][i-n]
	}

	// Insert spaces at the insertion point
	for i := x; i < x+n; i++ {
		newLine[i] = spaceCell
	}

	b.lines[y] = newLine
}

// DeleteCell deletes cells at the given position within the specified rectangles.
// If no rectangles are specified, it deletes cells in the entire buffer.
func (b *Buffer) DeleteCell(x, y, n int, rects ...Rectangle) {
	if len(rects) == 0 {
		b.deleteCellInRect(x, y, n, b.Bounds())
		return
	}
	for _, rect := range rects {
		b.deleteCellInRect(x, y, n, rect)
	}
}

// deleteCellInRect deletes cells at the given position within the rectangle bounds.
// Only cells within the rectangle's bounds are affected, following terminal DCH behavior.
func (b *Buffer) deleteCellInRect(x, y, n int, rect Rectangle) {
	if n <= 0 || y < rect.Min.Y || y >= rect.Max.Y || y >= b.Height() ||
		x < rect.Min.X || x >= rect.Max.X || x >= b.Width() {
		return
	}

	// Create a new line preserving cells outside rectangle
	newLine := make(Line, b.Width())
	copy(newLine, b.lines[y])

	// Calculate how many positions we can actually delete
	remainingCells := rect.Max.X - x
	if n > remainingCells {
		n = remainingCells
	}

	// Shift the remaining cells to the left
	for i := x; i < rect.Max.X-n; i++ {
		if i+n < rect.Max.X {
			newLine[i] = b.lines[y][i+n]
		}
	}

	// Fill the vacated positions with spaces
	for i := rect.Max.X - n; i < rect.Max.X; i++ {
		newLine[i] = spaceCell
	}

	b.lines[y] = newLine
}
