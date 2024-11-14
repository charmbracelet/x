package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// handleCsi handles a CSI escape sequences.
func (t *Terminal) handleCsi(seq ansi.CsiSequence) {
	switch t.parser.Cmd() { // cursor
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
	case ansi.Cmd('?', 0, 'W'): // DECST8C - Set Tab at Every 8 Columns
		if params := t.parser.Params(); len(params) == 1 && params[0] == 5 {
			t.resetTabStops()
		}
	case ansi.Cmd(0, ' ', 'q'): // DECSCUSR - Set Cursor Style
		style := 1
		if param, ok := t.parser.Param(0, 0); ok {
			style = param
		}
		t.scr.cur.Style = CursorStyle((style / 2) + 1)
		t.scr.cur.Steady = style%2 != 1
	case 'g': // TBC - Tab Clear
		var value int
		if param, ok := t.parser.Param(0, 0); ok {
			value = param
		}

		switch value {
		case 0:
			t.tabstops.Reset(t.scr.cur.X)
		case 3:
			t.tabstops.Clear()
		}
	case '@': // ICH - Insert Character
		n := 1
		if param, ok := t.parser.Param(0, 1); ok {
			n = param
		}

		t.scr.InsertCell(n)
	case 'P': // DCH - Delete Character
		n := 1
		if param, ok := t.parser.Param(0, 1); ok {
			n = param
		}

		t.scr.DeleteCell(n)
	default:
		t.logf("unhandled CSI: %q", seq)
	}
}
