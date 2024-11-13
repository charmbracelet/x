package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

func (t *Terminal) handleScreen() {
	var count int
	if t.parser.ParamsLen > 0 {
		count = ansi.Param(t.parser.Params[0]).Param(0)
	}

	w, h := t.Width(), t.Height()
	_, y := t.scr.CursorPosition()

	cmd := ansi.Cmd(t.parser.Cmd)
	switch cmd.Command() {
	case 'J':
		switch count {
		case 0: // Erase screen below (including cursor)
			rect := cellbuf.Rect(0, y, w, h-y)
			t.scr.Clear(rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 1: // Erase screen above (including cursor)
			rect := cellbuf.Rect(0, 0, w, y+1)
			t.scr.Clear(rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 2: // erase screen
			fallthrough
		case 3: // erase display
			// TODO: Scrollback buffer support?
			t.scr.Clear()
			if t.Damage != nil {
				t.Damage(ScreenDamage{w, h})
			}
		}
	case 'L': // IL - Insert Line
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		t.scr.InsertLine(n)
		// Move the cursor to the left margin.
		t.scr.setCursorX(0, true)

	case 'M': // DL - Delete Line
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		t.scr.DeleteLine(n)
		// Move the cursor to the left margin.
		t.scr.setCursorX(0, true)

	case 'X':
		// ECH - Erase Character
		// It clears character attributes as well but not colors.
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}
		x, y := t.scr.CursorPosition()
		c := blankCell
		c.Style = t.scr.cur.Pen
		c.Style.Attrs = 0
		rect := Rect(x, y, n, 1)
		t.scr.Fill(c, rect)
		// ECH does not move the cursor.

	case 'r': // DECSTBM - Set Top and Bottom Margins
		if t.parser.ParamsLen == 2 {
			top := ansi.Param(t.parser.Params[0]).Param(1)
			bottom := ansi.Param(t.parser.Params[1]).Param(t.Height())
			if top > bottom {
				top, bottom = bottom, top
			}

			// Rect is [x, y) which means y is exclusive. So the top margin
			// is the top of the screen minus one.
			t.scr.scroll.Min.Y = top - 1
			t.scr.scroll.Max.Y = bottom
		} else {
			// Rect is [x, y) which means y is exclusive. So the bottom margin
			// is the height of the screen.
			t.scr.scroll.Min.Y = 0
			t.scr.scroll.Max.Y = t.Height()
		}

		// Move the cursor to the top-left of the screen or scroll region
		// depending on [ansi.DECOM].
		t.setCursorPosition(0, 0)
	}
}

func (t *Terminal) handleLine() {
	var count int
	if t.parser.ParamsLen > 0 {
		count = ansi.Param(t.parser.Params[0]).Param(0)
	}

	cmd := ansi.Cmd(t.parser.Cmd)
	switch cmd.Command() {
	case 'K': // EL - Erase in Line
		// NOTE: Erase Line (EL) erases all character attributes but not cell
		// bg color.
		x, y := t.scr.CursorPosition()
		w := t.scr.Width()
		c := blankCell
		c.Style = t.scr.cur.Pen
		c.Style.Attrs = 0

		switch count {
		case 0: // Erase from cursor to end of line
			rect := cellbuf.Rect(x, y, w-x, 1)
			t.scr.Fill(c, rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 1: // Erase from start of line to cursor
			rect := cellbuf.Rect(0, y, x+1, 1)
			t.scr.Fill(c, rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 2: // Erase entire line
			rect := cellbuf.Rect(0, y, w, 1)
			t.scr.Fill(c, rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		}
	case 'S': // SU - Scroll Up
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		t.scr.ScrollUp(n)
	case 'T': // SD - Scroll Down
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		t.scr.ScrollDown(n)
	}
}
