package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// handleEsc handles an escape sequence.
func (t *Terminal) handleEsc(seq []byte) {
	cmd := t.parser.Cmd
	switch cmd {
	case 'H': // Horizontal Tab Set [ansi.HTS]
		t.horizontalTabSet()
	case 'M': // Reverse Index [ansi.RI]
		t.reverseIndex()
	case '=': // Keypad Application Mode [ansi.DECKPAM]
		t.pmodes[ansi.NumericKeypadMode] = ModeSet
	case '>': // Keypad Numeric Mode [ansi.DECKPNM]
		t.pmodes[ansi.NumericKeypadMode] = ModeReset
	case '7': // Save Cursor [ansi.DECSC]
		t.scr.SaveCursor()
	case '8': // Restore Cursor [ansi.DECRC]
		t.scr.RestoreCursor()
	case 'B' | '('<<parser.IntermedShift: // G0 Character Set
	// TODO: Handle G0 Character Set
	default:
		t.logf("unhandled ESC: %q", seq)
	}
}
