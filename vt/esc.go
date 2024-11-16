package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleEsc handles an escape sequence.
func (t *Terminal) handleEsc(seq ansi.EscSequence) {
	switch t.parser.Cmd() {
	case 'H': // Horizontal Tab Set [ansi.HTS]
		t.horizontalTabSet()
	case 'M': // Reverse Index [ansi.RI]
		t.reverseIndex()
	case '=': // Keypad Application Mode [ansi.DECKPAM]
		t.setMode(ansi.NumericKeypadMode, ModeSet)
	case '>': // Keypad Numeric Mode [ansi.DECKPNM]
		t.setMode(ansi.NumericKeypadMode, ModeReset)
	case '7': // Save Cursor [ansi.DECSC]
		t.scr.SaveCursor()
	case '8': // Restore Cursor [ansi.DECRC]
		t.scr.RestoreCursor()
	case ansi.Cmd(0, '(', 'B'): // G0 Character Set
	// TODO: Handle G0 Character Set
	default:
		t.logf("unhandled ESC: %q", seq)
	}
}
