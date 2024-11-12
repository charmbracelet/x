package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleControl handles a control character.
func (t *Terminal) handleControl(r rune) {
	switch r {
	case ansi.NUL: // NUL - Null
		// Ignored
	case ansi.BEL: // BEL - Bell
		if t.Bell != nil {
			t.Bell()
		}
	case ansi.BS: // BS - Backspace
		x, _ := t.scr.CursorPosition()
		if x > 0 {
			x--
		}

		t.scr.setCursorX(x)
	case ansi.HT: // HT - Horizontal Tab
		x, _ := t.scr.CursorPosition()
		x = t.tabstops.Next(x)
		t.scr.setCursorX(x)
	case ansi.LF:
		// LF - Line Feed
		_, y := t.scr.CursorPosition()
		if y < t.scr.Height()-1 {
			t.scr.setCursorY(y + 1)
		} else {
			t.scr.ScrollUp(1)
		}
	case ansi.CR: // CR - Carriage Return
		t.scr.setCursorX(0)
	case ansi.HTS: // HTS - Horizontal Tab Set
		x, _ := t.scr.CursorPosition()
		t.tabstops.Set(x)
		t.scr.setCursorX(x)
	case ansi.RI: // RI - Reverse Index
		_, y := t.scr.CursorPosition()
		if y > 0 {
			y--
		} else {
			t.scr.ScrollDown(1)
		}
		t.scr.setCursorY(y)
	case ansi.SO: // SO - Shift Out
	// TODO: Handle Shift Out
	case ansi.SI: // SI - Shift In
	// TODO: Handle Shift In
	default:
		t.logf("unhandled control: %q", r)
	}
}
