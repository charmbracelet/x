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
		if t.scr.cur.X > 0 {
			t.scr.cur.X--
		}
	case ansi.HT: // HT - Horizontal Tab
		t.scr.cur.X = t.tabstops.Next(t.scr.cur.X)
	case ansi.LF:
		// LF - Line Feed
		if t.scr.cur.Y < t.scr.Height()-1 {
			t.scr.cur.Y++
		} else {
			t.scr.ScrollUp(1)
		}
	case ansi.CR: // CR - Carriage Return
		t.scr.cur.X = 0
	case ansi.HTS: // HTS - Horizontal Tab Set
		t.tabstops.Set(t.scr.cur.X)
	case ansi.RI: // RI - Reverse Index
		if t.scr.cur.Y > 0 {
			t.scr.cur.Y--
		} else {
			t.scr.ScrollDown(1)
		}
	case ansi.SO: // SO - Shift Out
	// TODO: Handle Shift Out
	case ansi.SI: // SI - Shift In
	// TODO: Handle Shift In
	default:
		log.Printf("unhandled control: %q", r)
	}
}
