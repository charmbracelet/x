package vt

import (
	"bytes"
	"fmt"
)

// ReattachSnapshot renders retained terminal state as xterm-compatible ANSI.
// A hard boundary is emitted as CRLF. A soft-wrap continuation is emitted
// without a line break so the receiving terminal creates an isWrapped row at
// its own width and can reflow it on later resizes.
func (e *Emulator) ReattachSnapshot() []byte {
	includeScrollback := !e.IsAltScreen()
	rows := e.scr.snapshotRows(includeScrollback)

	var buf bytes.Buffer
	for i, row := range rows {
		buf.WriteString(row.Line.Render())
		if i+1 < len(rows) && !rows[i+1].Wrapped {
			buf.WriteString("\r\n")
		}
	}

	x, y := e.scr.CursorPosition()
	_, _ = fmt.Fprintf(&buf, "\x1b[%d;%dH\x1b[K", y+1, x+1)
	return buf.Bytes()
}
