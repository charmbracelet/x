package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Segment represents a continuous segment of cells with the same style
// attributes and hyperlink.
type Segment = Cell

// Screen represents an interface for a grid of cells that can be written to
// and read from.
type Screen interface {
	// Width returns the width of the grid.
	Width() int

	// Height returns the height of the grid.
	Height() int

	// SetCell writes a cell to the grid at the given position. It returns true
	// if the cell was written successfully.
	SetCell(x, y int, c Cell) bool

	// Cell returns the cell at the given position.
	Cell(x, y int) (Cell, bool)

	// Resize resizes the grid to the given width and height.
	Resize(width, height int)
}

// SetContent writes the given data to the grid starting from the first cell.
func SetContent(d Screen, m Method, content string) []int {
	return setContent(d, content, m)
}

// Render returns a string representation of the grid with ANSI escape sequences.
// Use [ansi.Strip] to remove them.
func Render(d Screen) string {
	var buf bytes.Buffer
	height := d.Height()
	for y := 0; y < height; y++ {
		_, line := RenderLine(d, y)
		buf.WriteString(line)
		if y < height-1 {
			buf.WriteString("\r\n")
		}
	}
	return buf.String()
}

// RenderLine returns a string representation of the yth line of the grid along
// with the width of the line.
func RenderLine(d Screen, n int) (w int, line string) {
	var pen Style
	var link Link
	var buf bytes.Buffer
	var pendingLine string
	var pendingWidth int // this ignores space cells until we hit a non-space cell

	writePending := func() {
		// If there's no pending line, we don't need to do anything.
		if len(pendingLine) == 0 {
			return
		}
		buf.WriteString(pendingLine)
		w += pendingWidth
		pendingWidth = 0
		pendingLine = ""
	}

	for x := 0; x < d.Width(); x++ {
		if cell, ok := d.Cell(x, n); ok && cell.Width > 0 {
			if cell.Style.Empty() && !pen.Empty() {
				writePending()
				buf.WriteString(ansi.ResetStyle) //nolint:errcheck
				pen.Reset()
			}
			if !cell.Style.Equal(pen) {
				writePending()
				seq := cell.Style.DiffSequence(pen)
				buf.WriteString(seq) // nolint:errcheck
				pen = cell.Style
			}

			// Write the URL escape sequence
			if cell.Link != link && link.URL != "" {
				writePending()
				buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
				link.Reset()
			}
			if cell.Link != link {
				writePending()
				buf.WriteString(ansi.SetHyperlink(cell.Link.URL, cell.Link.URLID)) //nolint:errcheck
				link = cell.Link
			}

			// We only write the cell content if it's not empty. If it is, we
			// append it to the pending line and width to be evaluated later.
			if cell.Equal(spaceCell) {
				pendingLine += cell.Content
				pendingWidth += cell.Width
			} else {
				writePending()
				buf.WriteString(cell.Content)
				w += cell.Width
			}
		}
	}
	if link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
	}
	if !pen.Empty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
	}
	return w, strings.TrimRight(buf.String(), " ") // Trim trailing spaces
}

// Fill fills the grid with the given cell.
func Fill(d Screen, c Cell) {
	for y := 0; y < d.Height(); y++ {
		for x := 0; x < d.Width(); x++ {
			d.SetCell(x, y, c) //nolint:errcheck
		}
	}
}

// Equal returns whether two grids are equal.
func Equal(a, b Screen) bool {
	if a.Width() != b.Width() || a.Height() != b.Height() {
		return false
	}
	for y := 0; y < a.Height(); y++ {
		for x := 0; x < a.Width(); x++ {
			ca, _ := a.Cell(x, y)
			cb, _ := b.Cell(x, y)
			if !ca.Equal(cb) {
				return false
			}
		}
	}
	return true
}
