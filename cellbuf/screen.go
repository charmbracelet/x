package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/colorprofile"
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
func Render(d Screen) string {
	return RenderWithProfile(d, colorprofile.TrueColor)
}

// RenderWithProfile returns a string representation of the grid with ANSI escape
// sequences converting styles and colors to the given color profile.
func RenderWithProfile(d Screen, p colorprofile.Profile) string {
	var buf bytes.Buffer
	height := d.Height()
	for y := 0; y < height; y++ {
		_, line := RenderLineWithProfile(d, y, p)
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
	return RenderLineWithProfile(d, n, colorprofile.TrueColor)
}

// RenderLineWithProfile returns a string representation of the nth line of the
// grid along with the width of the line converting styles and colors to the
// given color profile.
func RenderLineWithProfile(d Screen, n int, p colorprofile.Profile) (w int, line string) {
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
			// Convert the cell's style and link to the given color profile.
			cellStyle := cell.Style.Convert(p)
			cellLink := cell.Link.Convert(p)
			if cellStyle.Empty() && !pen.Empty() {
				writePending()
				buf.WriteString(ansi.ResetStyle) //nolint:errcheck
				pen.Reset()
			}
			if !cellStyle.Equal(pen) {
				writePending()
				seq := cellStyle.DiffSequence(pen)
				buf.WriteString(seq) // nolint:errcheck
				pen = cellStyle
			}

			// Write the URL escape sequence
			if cellLink != link && link.URL != "" {
				writePending()
				buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
				link.Reset()
			}
			if cellLink != link {
				writePending()
				buf.WriteString(ansi.SetHyperlink(cellLink.URL, cellLink.URLID)) //nolint:errcheck
				link = cellLink
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
