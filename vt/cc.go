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
		if t.Bell != nil {
			t.Bell()
		}
	case ansi.BS: // Backspace [ansi.BS]
		t.scr.moveCursor(-1, 0)
	case ansi.HT: // Horizontal Tab [ansi.HT]
		x, _ := t.scr.CursorPosition()
		x = t.tabstops.Next(x)
		t.scr.setCursorX(x, false)
	case ansi.LF:
		// Line Feed [ansi.LF]
		x, y := t.scr.CursorPosition()
		scroll := t.scr.ScrollRegion()
		if y == scroll.Max.Y-1 && x >= scroll.Min.X && x < scroll.Max.X {
			t.scr.ScrollUp(1)
		} else {
			t.scr.moveCursor(0, 1)
		}
	case ansi.CR: // Carriage Return [ansi.CR]
		t.carriageReturn()
	case ansi.HTS: // Horizontal Tab Set [ansi.HTS]
		t.horizontalTabSet()
	case ansi.RI: // Reverse Index [ansi.RI]
		t.reverseIndex()
	case ansi.SO: // Shift Out [ansi.SO]
	// TODO: Handle Shift Out
	case ansi.SI: // Shift In [ansi.SI]
	// TODO: Handle Shift In
	default:
		t.logf("unhandled control: %q", r)
	}
}

// horizontalTabSet sets a horizontal tab stop at the current cursor position.
func (t *Terminal) horizontalTabSet() {
	x, _ := t.scr.CursorPosition()
	t.tabstops.Set(x)
}

// reverseIndex moves the cursor up one line, or scrolling down.
func (t *Terminal) reverseIndex() {
	x, y := t.scr.CursorPosition()
	scroll := t.scr.ScrollRegion()
	if y == scroll.Min.Y && x >= scroll.Min.X && x < scroll.Max.X {
		t.scr.ScrollDown(1)
	} else {
		t.scr.moveCursor(0, -1)
	}
}
