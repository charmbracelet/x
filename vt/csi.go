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

	case 'c': // Primary Device Attributes [ansi.DA1]
		n, _ := t.parser.Param(0, 0)
		if n != 0 {
			break
		}

		// Do we fully support VT220?
		t.buf.WriteString(ansi.PrimaryDeviceAttributes(
			62, // VT220
			1,  // 132 columns
			6,  // Selective Erase
			22, // ANSI color
		))

	case ansi.Cmd('>', 0, 'c'): // Secondary Device Attributes [ansi.DA2]
		n, _ := t.parser.Param(0, 0)
		if n != 0 {
			break
		}

		// Do we fully support VT220?
		t.buf.WriteString(ansi.SecondaryDeviceAttributes(
			1,  // VT220
			10, // Version 1.0
			0,  // ROM Cartridge is always zero
		))

	case 'n': // Device Status Report [ansi.DSR]
		n, ok := t.parser.Param(0, 1)
		if !ok || n == 0 {
			break
		}

		switch n {
		case 5: // Operating Status
			// We're always ready ;)
			// See: https://vt100.net/docs/vt510-rm/DSR-OS.html
			t.buf.WriteString(ansi.DeviceStatusReport(ansi.DECStatus(0)))
		case 6: // Cursor Position Report [ansi.CPR]
			t.buf.WriteString(ansi.CursorPositionReport(t.scr.cur.X+1, t.scr.cur.Y+1))
		}

	case ansi.Cmd('?', 0, 'n'): // Device Status Report (DEC) [ansi.DSR]
		n, ok := t.parser.Param(0, 1)
		if !ok || n == 0 {
			break
		}

		switch n {
		case 6: // Extended Cursor Position Report [ansi.DECXCPR]
			t.buf.WriteString(ansi.ExtendedCursorPositionReport(t.scr.cur.X+1, t.scr.cur.Y+1, 0)) // We don't support page numbers
		}

	default:
		t.logf("unhandled CSI: %q", seq)
	}
}
