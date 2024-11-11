package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

func (t *Terminal) handleCursor() {
	p := t.parser
	width, height := t.scr.Width(), t.scr.Height()
	cmd := ansi.Cmd(p.Cmd)
	n := 1
	if p.ParamsLen > 0 {
		if param := ansi.Param(p.Params[0]).Param(1); param > 0 {
			n = param
		}
	}

	x, y := t.scr.cur.Pos.X, t.scr.cur.Pos.Y
	switch cmd.Command() {
	case 'A':
		// CUU - Cursor Up
		y = max(0, y-n)
	case 'B':
		// CUD - Cursor Down
		y = min(height-1, y+n)
	case 'C':
		// CUF - Cursor Forward
		x = min(width-1, x+n)
	case 'D':
		// CUB - Cursor Back
		x = max(0, x-n)
	case 'E':
		// CNL - Cursor Next Line
		y = min(height-1, y+n)
		x = 0
	case 'F':
		// CPL - Cursor Previous Line
		y = max(0, y-n)
		x = 0
	case 'G':
		// CHA - Cursor Character Absolute
		x = min(width-1, max(0, n-1))
	case 'H', 'f':
		// CUP - Cursor Position
		// HVP - Horizontal and Vertical Position
		if p.ParamsLen >= 2 {
			row, col := ansi.Param(p.Params[0]).Param(1), ansi.Param(p.Params[1]).Param(1)
			y = min(height-1, max(0, row-1))
			x = min(width-1, max(0, col-1))
		} else {
			x, y = 0, 0
		}
	case 'I':
		// CHT - Cursor Forward Tabulation
		for i := 0; i < n; i++ {
			x = t.tabstops.Next(x)
		}
	case 'X':
		// ECH - Erase Character
		c := spaceCell
		c.Style = t.scr.cur.Pen
		rect := cellbuf.Rect(x, y, n, 1)
		t.scr.Fill(c, &rect)
		x = min(width-1, x+n)
	case 'd':
		// VPA - Vertical Line Position Absolute
		y = min(height-1, max(0, n-1))
	case 'e':
		// VPR - Vertical Line Position Relative
		y = min(height-1, max(0, y+n))
	}

	t.scr.moveCursor(x, y)
}
