package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleEsc handles an escape sequence.
func (t *Terminal) handleEsc(seq ansi.EscSequence) {
	switch cmd := t.parser.Cmd(); cmd {
	case 'H': // Horizontal Tab Set [ansi.HTS]
		t.horizontalTabSet()
	case 'M': // Reverse Index [ansi.RI]
		t.reverseIndex()
	case '=': // Keypad Application Mode [ansi.DECKPAM]
		t.setMode(ansi.NumericKeypadMode, ansi.ModeSet)
	case '>': // Keypad Numeric Mode [ansi.DECKPNM]
		t.setMode(ansi.NumericKeypadMode, ansi.ModeReset)
	case '7': // Save Cursor [ansi.DECSC]
		t.scr.SaveCursor()
	case '8': // Restore Cursor [ansi.DECRC]
		t.scr.RestoreCursor()
	case 'c': // Reset Initial State [ansi.RIS]
		t.fullReset()
	case '~': // Locking Shift 1 Right [ansi.LS1R]
		t.gr = 1
	case 'n': // Locking Shift G2 [ansi.LS2]
		t.gl = 2
	case '}': // Locking Shift 2 Right [ansi.LS2R]
		t.gr = 2
	case 'o': // Locking Shift G3 [ansi.LS3]
		t.gl = 3
	case '|': // Locking Shift 3 Right [ansi.LS3R]
		t.gr = 3
	default:
		switch inter := cmd.Intermediate(); inter {
		case '(', ')', '*', '+': // Select Character Set [ansi.SCS]
			set := inter - '('
			switch cmd.Command() {
			case 'A': // UK Character Set
				t.charsets[set] = UK
			case 'B': // USASCII Character Set
				t.charsets[set] = nil // USASCII is the default
			case '0': // Special Drawing Character Set
				t.charsets[set] = SpecialDrawing
			default:
				t.logf("unknown character set: %q", seq)
			}
		default:
			t.logf("unhandled ESC: %q", seq)
		}
	}
}

// fullReset performs a full terminal reset as in [ansi.RIS].
func (t *Terminal) fullReset() {
	t.scrs[0].Reset()
	t.scrs[1].Reset()
	t.resetTabStops()

	// TODO: Do we reset all modes here? Investigate.
	t.resetModes()

	t.gl, t.gr = 0, 1
	t.gsingle = 0
	t.charsets = [4]CharSet{}
	t.atPhantom = false
}
