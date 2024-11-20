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

// Buffer represents a screen grid of cells.
type Buffer interface {
	// Width returns the width of the grid.
	Width() int

	// Height returns the height of the grid.
	Height() int

	// Cell returns the cell at the given position. If the cell is out of
	// bounds, it returns nil.
	Cell(x, y int) *Cell

	// SetCell writes a cell to the grid at the given position. It returns true
	// if the cell was written successfully. If the cell is nil, a blank cell
	// is written.
	SetCell(x, y int, c *Cell) bool
}

// Resizable is an interface for buffers that can be resized.
type Resizable interface {
	// Resize resizes the buffer to the given width and height.
	Resize(width, height int)
}

// Paint writes the given data to the canvas. If rect is not nil, it only
// writes to the rectangle. Otherwise, it writes to the whole canvas.
func Paint(d Buffer, m Method, content string, rect *Rectangle) []int {
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
func Render(d Buffer, opts ...RenderOption) string {
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
func RenderLine(d Buffer, n int, opts ...RenderOption) (w int, line string) {
	var opt RenderOptions
	for _, o := range opts {
		o(&opt)
	}
	return renderLine(d, n, opt)
}

func renderLine(d Buffer, n int, opt RenderOptions) (w int, line string) {
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
		if cell := d.Cell(x, n); cell != nil && cell.Width > 0 {
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
			if cell.Equal(&spaceCell) {
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
func Fill(d Buffer, c *Cell, rects ...Rectangle) {
	if len(rects) == 0 {
		fill(d, c, Rect(0, 0, d.Width(), d.Height()))
		return
	}
	for _, rect := range rects {
		fill(d, c, rect)
	}
}

func fill(d Buffer, c *Cell, rect Rectangle) {
	cellWidth := 1
	if c != nil {
		cellWidth = c.Width
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x += cellWidth {
			d.SetCell(x, y, c) //nolint:errcheck
		}
	}
}

// Clear clears the canvas with space cells. If rect is not nil, it only clears
// the rectangle. Otherwise, it clears the whole canvas.
func Clear(d Buffer, rects ...Rectangle) {
	if len(rects) == 0 {
		fill(d, nil, Rect(0, 0, d.Width(), d.Height()))
		return
	}
	for _, rect := range rects {
		fill(d, nil, rect)
	}
}

// Equal returns whether two grids are equal.
func Equal(a, b Buffer) bool {
	if a.Width() != b.Width() || a.Height() != b.Height() {
		return false
	}
	for y := 0; y < a.Height(); y++ {
		for x := 0; x < a.Width(); x++ {
			ca := a.Cell(x, y)
			cb := b.Cell(x, y)
			if ca != nil && cb != nil && !ca.Equal(cb) {
				return false
			}
		}
	}
	return true
}
