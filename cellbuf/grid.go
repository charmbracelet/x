package cellbuf

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// Segment represents a continuous segment of cells with the same style
// attributes and hyperlink.
type Segment = Cell

// Grid represents an interface for a grid of cells that can be written to and
// read from.
type Grid interface {
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

// SetContentAt writes the given data to the grid starting from the given
// position and with the given width and height.
func (m WidthMethod) SetContentAt(b Grid, c string, x, y, w, h int) []int {
	return setContent(b, c, x, y, w, h, m, strings.ReplaceAll, utf8.DecodeRuneInString)
}

// SetContent writes the given data to the grid starting from the first cell.
func (m WidthMethod) SetContent(g Grid, content string) []int {
	return m.SetContentAt(g, content, 0, 0, g.Width(), Height(content))
}

// Render returns a string representation of the grid with ANSI escape sequences.
// Use [ansi.Strip] to remove them.
func Render(g Grid) string {
	var buf bytes.Buffer
	height := g.Height()
	for y := 0; y < height; y++ {
		_, line := RenderLine(g, y)
		buf.WriteString(line)
		if y < height-1 {
			buf.WriteString("\r\n")
		}
	}
	return buf.String()
}

// RenderLine returns a string representation of the yth line of the grid along
// with the width of the line.
func RenderLine(g Grid, n int) (w int, line string) {
	var pen Style
	var link Link
	var buf bytes.Buffer
	var pendingLine string
	var pendingWidth int // this ignores space cells until we hit a non-space cell
	for x := 0; x < g.Width(); x++ {
		if cell, ok := g.Cell(x, n); ok && cell.Width > 0 {
			if cell.Style.Empty() && !pen.Empty() {
				buf.WriteString(ansi.ResetStyle) //nolint:errcheck
				pen.Reset()
			}
			if !cell.Style.Equal(pen) {
				seq := cell.Style.DiffSequence(pen)
				buf.WriteString(seq) // nolint:errcheck
				pen = cell.Style
			}

			// Write the URL escape sequence
			if cell.Link != link && link.URL != "" {
				buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
				link.Reset()
			}
			if cell.Link != link {
				buf.WriteString(ansi.SetHyperlink(cell.Link.URL, cell.Link.URLID)) //nolint:errcheck
				link = cell.Link
			}

			// We only write the cell content if it's not empty. If it is, we
			// append it to the pending line and width to be evaluated later.
			if cell.Style.Empty() && len(strings.TrimSpace(cell.Content)) == 0 {
				pendingLine += cell.Content
				pendingWidth += cell.Width
			} else {
				buf.WriteString(pendingLine + cell.Content)
				w += pendingWidth + cell.Width
				pendingWidth = 0
				pendingLine = ""
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
func Fill(g Grid, c Cell) {
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			g.SetCell(x, y, c) //nolint:errcheck
		}
	}
}

// Equal returns whether two grids are equal.
func Equal(a, b Grid) bool {
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
