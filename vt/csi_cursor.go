package vt

import (
	"github.com/charmbracelet/x/ansi"
)

func (t *Terminal) handleCursor() {
	width, height := t.Width(), t.Height()
	n := 1
	if param, ok := t.parser.Param(0, 1); ok {
		n = param
	}

	x, y := t.scr.CursorPosition()
	switch t.parser.Cmd() {
	case 'A':
		// Cursor Up [ansi.CUU]
		t.scr.moveCursor(0, -n)
	case 'B':
		// Cursor Down [ansi.CUD]
		t.scr.moveCursor(0, n)
	case 'C':
		// Cursor Forward [ansi.CUF]
		t.scr.moveCursor(n, 0)
	case 'D':
		// Cursor Backward [ansi.CUB]
		t.scr.moveCursor(-n, 0)
	case 'E':
		// Cursor Next Line [ansi.CNL]
		t.scr.moveCursor(0, n)
		t.carriageReturn()
	case 'F':
		// Cursor Previous Line [ansi.CPL]
		t.scr.moveCursor(0, -n)
		t.carriageReturn()
	case 'G':
		// Cursor Horizontal Absolute [ansi.CHA]
		t.scr.setCursor(min(width-1, n-1), y, false)
	case 'H':
		// Cursor Position [ansi.CUP]
		row, _ := t.parser.Param(0, 1)
		col, _ := t.parser.Param(1, 1)
		y = min(height-1, row-1)
		x = min(width-1, col-1)
		t.setCursorPosition(x, y)
	case 'I':
		// Cursor Horizontal Tabulation [ansi.CHT]
		scroll := t.scr.ScrollRegion()
		for i := 0; i < n; i++ {
			ts := t.tabstops.Next(x)
			if ts >= scroll.Max.X {
				break
			}
			x = ts
		}
		t.scr.setCursor(x, y, false)
	case '`':
		// Horizontal Position Absolute [ansi.HPA]
		t.setCursorPosition(min(width-1, n-1), y)
	case 'a':
		// Horizontal Position Relative [ansi.HPR]
		t.setCursorPosition(min(width-1, x+n), y)
	case 'e':
		// Vertical Position Relative [ansi.VPR]
		t.setCursorPosition(x, min(height-1, y+n))
	case 'f':
		// Horizontal and Vertical Position [ansi.HVP]
		row, _ := t.parser.Param(0, 1)
		col, _ := t.parser.Param(1, 1)
		y = min(height-1, row-1)
		x = min(width-1, col-1)
		t.scr.setCursor(x, y, false)
	case 'd':
		// Vertical Position Absolute [ansi.VPA]
		t.setCursorPosition(x, min(height-1, n-1))
	}
}

// setCursorPosition sets the cursor position. This respects [ansi.DECOM],
// Origin Mode. This performs the same function as [ansi.CUP].
func (t *Terminal) setCursorPosition(x, y int) {
	mode, ok := t.modes[ansi.DECOM]
	margins := ok && mode.IsSet()
	t.scr.setCursor(x, y, margins)
}

// carriageReturn moves the cursor to the leftmost column. If [ansi.DECOM] is
// set, the cursor is set to the left margin. If not, and the cursor is on or
// to the right of the left margin, the cursor is set to the left margin.
// Otherwise, the cursor is set to the leftmost column of the screen.
// This performs the same function as [ansi.CR].
func (t *Terminal) carriageReturn() {
	mode, ok := t.modes[ansi.DECOM]
	margins := ok && mode.IsSet()
	x, y := t.scr.CursorPosition()
	if margins {
		t.scr.setCursor(0, y, true)
	} else if region := t.scr.ScrollRegion(); region.Contains(Pos(x, y)) {
		t.scr.setCursor(region.Min.X, y, false)
	} else {
		t.scr.setCursor(0, y, false)
	}
}
