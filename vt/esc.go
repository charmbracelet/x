package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleEsc handles an escape sequence.
func (t *Terminal) handleEsc(seq []byte) {
	cmd := t.parser.Cmd
	switch cmd {
	case 'H': // HTS - Horizontal Tab Set
		t.tabstops.Set(t.scr.cur.X)
	case 'M': // RI - Reverse Index
		if t.scr.cur.Y > 0 {
			t.scr.cur.Y--
		} else {
			t.scr.ScrollDown(1)
		}
	case '=': // DECKPAM - Keypad Application Mode
		t.pmodes[ansi.NumericKeypadMode] = ModeSet
	case '>': // DECKPNM - Keypad Numeric Mode
		t.pmodes[ansi.NumericKeypadMode] = ModeReset
	case '7': // DECSC - Save Cursor
		t.scr.SaveCursor()
	case '8': // DECRC - Restore Cursor
		t.scr.RestoreCursor()
	default:
		t.logf("unhandled ESC: %q", seq)
	}
}
