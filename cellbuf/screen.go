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

// Screen represents a screen grid of cells.
type Screen interface {
	// Width returns the width of the grid.
	Width() int

	// Height returns the height of the grid.
	Height() int

	// Cell returns the cell at the given position.
	Cell(x, y int) (Cell, bool)

	// Draw writes a cell to the grid at the given position. It returns true if
	// the cell was written successfully.
	Draw(x, y int, c Cell) bool
}

// Paint writes the given data to the canvas. If rect is not nil, it only
// writes to the rectangle. Otherwise, it writes to the whole canvas.
func Paint(d Screen, m Method, content string, rect *Rectangle) []int {
	if rect == nil {
		r := Rect(0, 0, d.Width(), d.Height())
		rect = &r
	}
	return setContent(d, content, m, *rect)
}

// RenderOptions represents options for rendering a canvas.
type RenderOptions struct {
	// Profile is the color profile to use when rendering the canvas.
	Profile colorprofile.Profile
}

// RenderOption is a function that configures a RenderOptions.
type RenderOption func(*RenderOptions)

// WithRenderProfile sets the color profile to use when rendering the canvas.
func WithRenderProfile(p colorprofile.Profile) RenderOption {
	return func(o *RenderOptions) {
		o.Profile = p
	}
}

// Render returns a string representation of the grid with ANSI escape sequences.
func Render(d Screen, opts ...RenderOption) string {
	var opt RenderOptions
	for _, o := range opts {
		o(&opt)
	}
	var buf bytes.Buffer
	height := d.Height()
	for y := 0; y < height; y++ {
		_, line := renderLine(d, y, opt)
		buf.WriteString(line)
		if y < height-1 {
			buf.WriteString("\r\n")
		}
	}
	return buf.String()
}

// RenderLine returns a string representation of the yth line of the grid along
// with the width of the line.
func RenderLine(d Screen, n int, opts ...RenderOption) (w int, line string) {
	var opt RenderOptions
	for _, o := range opts {
		o(&opt)
	}
	return renderLine(d, n, opt)
}

func renderLine(d Screen, n int, opt RenderOptions) (w int, line string) {
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
			cellStyle := ConvertStyle(cell.Style, opt.Profile)
			cellLink := ConvertLink(cell.Link, opt.Profile)
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

// Fill fills the canvas with the given cell. If rect is not nil, it only fills
// the rectangle. Otherwise, it fills the whole canvas.
func Fill(d Screen, c Cell, rect *Rectangle) {
	if rect == nil {
		r := Rect(0, 0, d.Width(), d.Height())
		rect = &r
	}

	for y := rect.Y(); y < rect.Y()+rect.Height(); y++ {
		for x := rect.X(); x < rect.X()+rect.Width(); x += c.Width {
			d.Draw(x, y, c) //nolint:errcheck
		}
	}
}

// Clear clears the canvas with space cells. If rect is not nil, it only clears
// the rectangle. Otherwise, it clears the whole canvas.
func Clear(d Screen, rect *Rectangle) {
	Fill(d, spaceCell, rect)
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

// InsertLine inserts a new line at the given position. If rect is not nil, it
// only inserts the line in the rectangle. Otherwise, it inserts the line in the
// whole screen.
//
// It pushes the lines below down and fills the new line with space cells.
func InsertLine(s Screen, y, n int, rect *Rectangle) {
	if n <= 0 {
		return
	}
	if rect == nil {
		r := Rect(0, 0, s.Width(), s.Height())
		rect = &r
	}

	for i := 0; i < n; i++ {
		for x := rect.X(); x < rect.X()+rect.Width(); x++ {
			for j := rect.Y() + rect.Height() - 1; j > y; j-- {
				c, _ := s.Cell(x, j-1)
				s.Draw(x, j, c) //nolint:errcheck
			}
			s.Draw(x, y, spaceCell) //nolint:errcheck
		}
	}
}

// ScrollUp scrolls the screen up by n lines. If rect is not nil, it only
// scrolls the rectangle. Otherwise, it scrolls the whole canvas.
//
// It pushes the top lines out and fills the new lines with space cells.
func ScrollUp(s Screen, n int, rect *Rectangle) {
	if n <= 0 {
		return
	}
	if rect == nil {
		r := Rect(0, 0, s.Width(), s.Height())
		rect = &r
	}

	for i := 0; i < n; i++ {
		for x := rect.X(); x < rect.X()+rect.Width(); x++ {
			for y := rect.Y(); y < rect.Y()+rect.Height()-1; y++ {
				c, _ := s.Cell(x, y+1)
				s.Draw(x, y, c) //nolint:errcheck
			}
			s.Draw(x, rect.Y()+rect.Height()-1, spaceCell) //nolint:errcheck
		}
	}
}

// ScrollDown scrolls the screen down by n lines. If rect is not nil, it only
// scrolls the rectangle. Otherwise, it scrolls the whole canvas.
//
// It pushes the bottom lines out and fills the new lines with space cells.
func ScrollDown(s Screen, n int, rect *Rectangle) {
	if n <= 0 {
		return
	}
	if rect == nil {
		r := Rect(0, 0, s.Width(), s.Height())
		rect = &r
	}

	for i := 0; i < n; i++ {
		for x := rect.X(); x < rect.X()+rect.Width(); x++ {
			for y := rect.Y() + rect.Height() - 1; y > rect.Y(); y-- {
				c, _ := s.Cell(x, y-1)
				s.Draw(x, y, c) //nolint:errcheck
			}
			s.Draw(x, rect.Y(), spaceCell) //nolint:errcheck
		}
	}
}

// InsertCell inserts n blank cell at the given position moving the cells on
// the same line after the column to the right. This will push the cells out of
// the screen if necessary i.e. the cells will be lost.
func InsertCell(s Screen, x, y, n int) {
	if n <= 0 {
		return
	}
	for i := s.Width() - 1; i >= x; i-- {
		c, _ := s.Cell(i, y)
		s.Draw(i+n, y, c) //nolint:errcheck
	}
	for i := 0; i < n; i++ {
		s.Draw(x+i, y, spaceCell) //nolint:errcheck
	}
}

// DeleteCell deletes n cells at the given position moving the cells on the same
// line to the left and adding space cells at the end.
func DeleteCell(s Screen, x, y, n int) {
	if n <= 0 {
		return
	}
	for i := x; i < s.Width(); i++ {
		if i+n < s.Width() {
			c, _ := s.Cell(i+n, y)
			s.Draw(i, y, c) //nolint:errcheck
		} else {
			s.Draw(i, y, spaceCell) //nolint:errcheck
		}
	}
}
