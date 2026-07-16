package vt

import (
	"bytes"
	"fmt"
)

// SnapshotError reports an invalid semantic terminal state. A failed snapshot
// never contains partial ANSI bytes.
type SnapshotError struct {
	Cause error
}

func (e *SnapshotError) Error() string { return "vt reattach snapshot: " + e.Cause.Error() }

func (e *SnapshotError) Unwrap() error { return e.Cause }

// ReattachSnapshot renders retained terminal state as xterm-compatible ANSI.
// A hard boundary is emitted as CRLF. A soft-wrap continuation is emitted
// without a line break so the receiving terminal creates an isWrapped row at
// its own width and can reflow it on later resizes.
func (e *Emulator) ReattachSnapshot() ([]byte, error) {
	includeScrollback := !e.IsAltScreen()
	rows, err := e.scr.snapshotRows(includeScrollback)
	if err != nil {
		return nil, &SnapshotError{Cause: err}
	}

	var buf bytes.Buffer
	for i, row := range rows {
		buf.WriteString(row.line.Render())
		if i+1 < len(rows) && rows[i+1].boundary != boundarySoft {
			buf.WriteString("\r\n")
		}
	}

	x, y := e.scr.CursorPosition()
	_, _ = fmt.Fprintf(&buf, "\x1b[%d;%dH\x1b[K", y+1, x+1)
	return buf.Bytes(), nil
}
