package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Render returns a string representation of the buffer with ANSI escape
// sequences. Use [ansi.Strip] to remove them.
func (b *Buffer) Render() string {
	var buf bytes.Buffer
	height := len(b.cells) / b.width
	for y := 0; y < height; y++ {
		buf.WriteString(b.RenderLine(y))
		buf.WriteString("\r\n")
	}
	return buf.String()
}

// RenderLine returns a string representation of the buffer n line with ANSI escape
// sequences. Use [ansi.Strip] to remove them.
func (b *Buffer) RenderLine(n int) string {
	var pen CellStyle
	var link CellLink
	var buf bytes.Buffer
	for x := 0; x < b.width; x++ {
		if cell, err := b.At(x, n); err == nil && cell.Width > 0 {
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

			buf.WriteString(cell.Content)
		}
	}
	if link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
	}
	if !pen.IsEmpty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
	}
	return strings.TrimRight(buf.String(), " ") // Trim trailing spaces
}
