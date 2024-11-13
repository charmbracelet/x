package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// handleCsi handles a CSI escape sequences.
func (t *Terminal) handleCsi(seq []byte) {
	cmd := t.parser.Cmd
	switch cmd { // cursor
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'a', 'd', 'e', 'f', '`':
		t.handleCursor()
	case 'm': // SGR - Select Graphic Rendition
		t.handleSgr()
	case 'J', 'L', 'M', 'X', 'r':
		t.handleScreen()
	case 'K', 'S', 'T':
		t.handleLine()
	case 'l', 'h', 'l' | '?'<<parser.MarkerShift, 'h' | '?'<<parser.MarkerShift:
		t.handleMode()
	case 'W' | '?'<<parser.MarkerShift: // DECST8C - Set Tab at Every 8 Columns
		if t.parser.ParamsLen == 1 && t.parser.Params[0] == 5 {
			t.resetTabStops()
		}
	case 'q' | ' '<<parser.IntermedShift: // DECSCUSR - Set Cursor Style
		style := 1
		if t.parser.ParamsLen > 0 {
			style = ansi.Param(t.parser.Params[0]).Param(0)
		}
		t.scr.cur.Style = CursorStyle((style / 2) + 1)
		t.scr.cur.Steady = style%2 != 1
	case 'g': // TBC - Tab Clear
		var param int
		if t.parser.ParamsLen > 0 {
			param = ansi.Param(t.parser.Params[0]).Param(0)
		}

		switch param {
		case 0:
			t.tabstops.Reset(t.scr.cur.X)
		case 3:
			t.tabstops.Clear()
		}
	case '@': // ICH - Insert Character
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		t.scr.InsertCell(n)
	case 'P': // DCH - Delete Character
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		t.scr.DeleteCell(n)
	default:
		t.logf("unhandled CSI: %q", seq)
	}
}
