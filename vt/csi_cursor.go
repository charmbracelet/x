package vt

import (
	"github.com/charmbracelet/x/ansi"
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

	x, y := t.scr.cur.X, t.scr.cur.Y
	switch cmd.Command() {
	case 'A':
		// CUU - Cursor Up
		y = max(0, y-n)
	case 'B', 'e':
		// CUD - Cursor Down
		// VPR - Vertical Position Relative
		y = min(height-1, y+n)
	case 'C', 'a':
		// CUF - Cursor Forward
		// HPR - Horizontal Position Relative
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
	case 'G', '`':
		// CHA - Cursor Character Absolute
		// HPA - Horizontal Position Absolute
		x = min(width-1, n-1)
	case 'H', 'f':
		// CUP - Cursor Position
		// HVP - Horizontal and Vertical Position
		if p.ParamsLen >= 2 {
			row, col := ansi.Param(p.Params[0]).Param(1), ansi.Param(p.Params[1]).Param(1)
			y = min(height-1, row-1)
			x = min(width-1, col-1)
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
		// It clears character attributes as well but not colors.
		c := spaceCell
		c.Style = t.scr.cur.Pen
		c.Style.Attrs = 0
		rect := Rect(x, y, n, 1)
		t.scr.Fill(c, rect)
		// ECH does not move the cursor.
	case 'd':
		// VPA - Vertical Line Position Absolute
		y = min(height-1, max(0, n-1))
	}

	t.scr.moveCursor(x, y)
}
