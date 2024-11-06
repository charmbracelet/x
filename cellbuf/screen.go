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

// Drawable represents a drawable grid of cells.
type Drawable interface {
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
func Paint(d Drawable, m Method, content string, rect *Rectangle) []int {
	if rect == nil {
		rect = &Rectangle{0, 0, d.Width(), d.Height()}
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
func Render(d Drawable, opts ...RenderOption) string {
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
func RenderLine(d Drawable, n int, opts ...RenderOption) (w int, line string) {
	var opt RenderOptions
	for _, o := range opts {
		o(&opt)
	}
	return renderLine(d, n, opt)
}

func renderLine(d Drawable, n int, opt RenderOptions) (w int, line string) {
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
			cellStyle := cell.Style.Convert(opt.Profile)
			cellLink := cell.Link.Convert(opt.Profile)
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
func Fill(d Drawable, c Cell, rect *Rectangle) {
	if rect == nil {
		rect = &Rectangle{0, 0, d.Width(), d.Height()}
	}

	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x += c.Width {
			d.Draw(x, y, c) //nolint:errcheck
		}
	}
}

// Clear clears the canvas with space cells. If rect is not nil, it only clears
// the rectangle. Otherwise, it clears the whole canvas.
func Clear(d Drawable, rect *Rectangle) {
	Fill(d, spaceCell, rect)
}

// Equal returns whether two grids are equal.
func Equal(a, b Drawable) bool {
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
