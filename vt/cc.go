package vt

import (
	"log"

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
		x, y := t.scr.CursorPosition()
		if x > 0 {
			x--
		}

		t.scr.setCursor(x, y)
	case ansi.HT: // HT - Horizontal Tab
		x, y := t.scr.CursorPosition()
		x = t.tabstops.Next(x)
		t.scr.setCursor(x, y)
	case ansi.LF:
		// LF - Line Feed
		x, y := t.scr.CursorPosition()
		if y < t.scr.Height()-1 {
			y++
		} else {
			t.scr.ScrollUp(1)
		}
		t.scr.setCursor(x, y)
	case ansi.CR: // CR - Carriage Return
		_, y := t.scr.CursorPosition()
		t.scr.setCursor(0, y)
	case ansi.HTS: // HTS - Horizontal Tab Set
		x, y := t.scr.CursorPosition()
		t.tabstops.Set(x)
		t.scr.setCursor(x, y)
	case ansi.RI: // RI - Reverse Index
		x, y := t.scr.CursorPosition()
		if y > 0 {
			y--
		} else {
			t.scr.ScrollDown(1)
		}
		t.scr.setCursor(x, y)
	case ansi.SO: // SO - Shift Out
	// TODO: Handle Shift Out
	case ansi.SI: // SI - Shift In
	// TODO: Handle Shift In
	default:
		log.Printf("unhandled control: %q", r)
	}
}
