package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleControl handles a control character.
func (t *Terminal) handleControl(r ansi.ControlCode) {
	switch r {
	case ansi.NUL: // Null [ansi.NUL]
		// Ignored
	case ansi.BEL: // Bell [ansi.BEL]
		if t.Callbacks.Bell != nil {
			t.Callbacks.Bell()
		}
	case ansi.BS: // Backspace [ansi.BS]
		// This acts like [ansi.CUB]
		t.moveCursor(-1, 0)
	case ansi.HT: // Horizontal Tab [ansi.HT]
		t.nextTab(1)
	case ansi.VT: // Vertical Tab [ansi.VT]
		fallthrough
	case ansi.FF: // Form Feed [ansi.FF]
		fallthrough
	case ansi.LF: // Line Feed [ansi.LF]
		t.linefeed()
	case ansi.CR: // Carriage Return [ansi.CR]
		t.carriageReturn()
	case ansi.HTS: // Horizontal Tab Set [ansi.HTS]
		t.horizontalTabSet()
	case ansi.RI: // Reverse Index [ansi.RI]
		t.reverseIndex()
	case ansi.SO: // Shift Out [ansi.SO]
		t.gl = 1
	case ansi.SI: // Shift In [ansi.SI]
		t.gl = 0
	case ansi.IND: // Index [ansi.IND]
		t.index()
	case ansi.SS2: // Single Shift 2 [ansi.SS2]
		t.gsingle = 2
	case ansi.SS3: // Single Shift 3 [ansi.SS3]
		t.gsingle = 3
	default:
		t.logf("unhandled control: %q", r)
	}
}

// linefeed is the same as [index], except that it respects [ansi.LNM] mode.
func (t *Terminal) linefeed() {
	t.index()
	if t.isModeSet(ansi.LineFeedNewLineMode) {
		t.carriageReturn()
	}
}

// index moves the cursor down one line, scrolling up if necessary. This
// always resets the phantom state i.e. pending wrap state.
func (t *Terminal) index() {
	x, y := t.scr.CursorPosition()
	scroll := t.scr.ScrollRegion()
	// TODO: Handle scrollback whenever we add it.
	if y == scroll.Max.Y-1 && x >= scroll.Min.X && x < scroll.Max.X {
		t.scr.ScrollUp(1)
	} else if y < scroll.Max.Y-1 || !scroll.Contains(Pos(x, y)) {
		t.scr.moveCursor(0, 1)
	}
	t.atPhantom = false
}

// horizontalTabSet sets a horizontal tab stop at the current cursor position.
func (t *Terminal) horizontalTabSet() {
	x, _ := t.scr.CursorPosition()
	t.tabstops.Set(x)
}

// reverseIndex moves the cursor up one line, or scrolling down. This does not
// reset the phantom state i.e. pending wrap state.
func (t *Terminal) reverseIndex() {
	x, y := t.scr.CursorPosition()
	scroll := t.scr.ScrollRegion()
	if y == scroll.Min.Y && x >= scroll.Min.X && x < scroll.Max.X {
		t.scr.ScrollDown(1)
	} else {
		t.scr.moveCursor(0, -1)
	}
}
