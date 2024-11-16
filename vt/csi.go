package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleCsi handles a CSI escape sequences.
func (t *Terminal) handleCsi(seq ansi.CsiSequence) {
	switch t.parser.Cmd() { // cursor
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'a', 'd', 'e', 'f', '`':
		t.handleCursor()
	case 'm': // Select Graphic Rendition [ansi.SGR]
		t.handleSgr()
	case 'J', 'L', 'M', 'X', 'r', 's':
		t.handleScreen()
	case 'K', 'S', 'T':
		t.handleLine()
	case ansi.Cmd(0, 0, 'h'), ansi.Cmd('?', 0, 'h'): // Set Mode [ansi.SM]
		fallthrough
	case ansi.Cmd(0, 0, 'l'), ansi.Cmd('?', 0, 'l'): // Reset Mode [ansi.RM]
		t.handleMode()
	case ansi.Cmd('?', 0, 'W'): // Set Tab at Every 8 Columns [ansi.DECST8C]
		if params := t.parser.Params(); len(params) == 1 && params[0] == 5 {
			t.resetTabStops()
		}
	case ansi.Cmd(0, ' ', 'q'): // Set Cursor Style [ansi.DECSCUSR]
		style := 1
		if param, ok := t.parser.Param(0, 0); ok {
			style = param
		}
		t.scr.cur.Style = CursorStyle((style / 2) + 1)
		t.scr.cur.Steady = style%2 != 1
	case 'g': // Tab Clear [ansi.TBC]
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
	case '@': // Insert Character [ansi.ICH]
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}

		t.scr.InsertCell(n)
	case 'P': // Delete Character [ansi.DCH]
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}

		t.scr.DeleteCell(n)
	default:
		t.logf("unhandled CSI: %q", seq)
	}
}
