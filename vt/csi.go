package vt

import (
	"log"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/charmbracelet/x/cellbuf"
)

var spaceCell = cellbuf.Cell{
	Content: " ",
	Width:   1,
}

// handleCsi handles a CSI escape sequences.
func (t *Terminal) handleCsi(seq []byte) {
	// params := t.parser.Params[:t.parser.ParamsLen]
	cmd := t.parser.Cmd
	switch cmd { // cursor
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'X', 'd', 'e':
		t.handleCursor()
	case 'm': // SGR - Select Graphic Rendition
		t.handleSgr()
	case 'J':
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
		t.scr.cur.Style = style
	case 'g': // TBC - Tab Clear
		var param int
		if t.parser.ParamsLen > 0 {
			param = ansi.Param(t.parser.Params[0]).Param(0)
		}

		switch param {
		case 0:
			t.tabstops.Reset(t.scr.cur.Pos.X)
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

		x, y := t.scr.cur.Pos.X, t.scr.cur.Pos.Y
		t.scr.buf.InsertCell(x, y, n)
	case 'P': // DCH - Delete Character
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(0); param > 0 {
				n = param
			}
		}

		x, y := t.scr.cur.Pos.X, t.scr.cur.Pos.Y
		t.scr.buf.DeleteCell(x, y, n)
	case 'r': // DECSTBM - Set Top and Bottom Margins
		log.Printf("scrolling region %d, %d", t.parser.Params[0], t.parser.Params[1])
		if t.parser.ParamsLen == 2 {
			top := ansi.Param(t.parser.Params[0]).Param(1)
			bottom := ansi.Param(t.parser.Params[1]).Param(t.Height())
			if top >= bottom {
				break
			}

			t.scrollregion.Min.Y = top - 1
			t.scrollregion.Max.Y = bottom - 1
		} else {
			t.scrollregion.Min.Y = 0
			t.scrollregion.Max.Y = t.Height() - 1
		}

		t.scr.moveCursor(t.scrollregion.Min.X, t.scrollregion.Min.Y)
	default:
		log.Printf("unhandled CSI: %q", seq)
	}
}
