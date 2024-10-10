package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

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
	var pen CellStyle
	var link CellLink
	var buf bytes.Buffer
	for x := 0; x < g.Width(); x++ {
		if cell, err := g.At(x, n); err == nil && cell.Width > 0 {
			if cell.Style.IsEmpty() && !pen.IsEmpty() {
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

			w += cell.Width
			buf.WriteString(cell.Content)
		}
	}
	if link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
	}
	if !pen.IsEmpty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
	}
	return w, strings.TrimRight(buf.String(), " ") // Trim trailing spaces
}
