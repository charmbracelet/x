package cellbuf

import (
	"bytes"

	"github.com/charmbracelet/x/ansi"
)

// Render returns a string representation of the buffer with ANSI escape
// sequences. Use [ansi.Strip] to remove them.
func (b *Buffer) Render() string {
	var pen Style
	var link Hyperlink
	var buf bytes.Buffer
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			if cell, err := b.At(x, y); err == nil && cell.Width > 0 {
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
		buf.WriteString("\r\n")
	}
	return buf.String()
}
