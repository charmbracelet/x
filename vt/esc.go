package vt

// handleEsc handles an escape sequence.
func (t *Terminal) handleEsc(seq []byte) {
	cmd := t.parser.Cmd
	switch cmd {
	case '7': // DECSC - Save Cursor
		t.scr.SaveCursor()
	case '8': // DECRC - Restore Cursor
		t.scr.RestoreCursor()
	}
}
