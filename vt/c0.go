package vt

import "github.com/charmbracelet/x/ansi"

// handleControl handles a control character.
func (t *Terminal) handleControl(r rune) {
	switch r {
	case ansi.BEL: // BEL - Bell
		if t.Bell != nil {
			t.Bell()
		}
	case ansi.BS: // BS - Backspace
		if t.scr.cur.Pos.X > 0 {
			t.scr.cur.Pos.X--
		}
	case ansi.HT: // HT - Horizontal Tab
	case ansi.LF: // LF - Line Feed
		if t.scr.cur.Pos.Y < t.scr.Height()-1 {
			t.scr.cur.Pos.Y++
		}
	case ansi.CR: // CR - Carriage Return
		t.scr.cur.Pos.X = 0
	}
}
