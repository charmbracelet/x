package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Segment represents a continuous segment of cells with the same style
// attributes and hyperlink.
type Segment struct {
	Style   Style
	Link    Link
	Content string
	Width   int
}

// Paint writes the given data to the canvas. If rect is not nil, it only
// writes to the rectangle.
func Paint(d Window, content string) []int {
	return PaintRect(d, content, d.Bounds())
}

// PaintRect writes the given data to the canvas starting from the given
// rectangle.
func PaintRect(d Window, content string, rect Rectangle) []int {
	return setContent(d, content, WcWidth, rect)
}

func renderLine(d *Buffer, n int, opt Options) (w int, line string) {
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
			if cell.Equal(&BlankCell) {
				pendingLine += cell.String()
				pendingWidth += cell.Width
			} else {
				writePending()
				buf.WriteString(cell.String())
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
