package vt

import (
	"log"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

func (t *Terminal) handleScreen() {
	var count int
	if t.parser.ParamsLen > 0 {
		count = ansi.Param(t.parser.Params[0]).Param(0)
	}

	scr := t.scr
	cur := scr.Cursor()
	w, h := scr.Width(), scr.Height()
	_, y := cur.Pos.X, cur.Pos.Y

	cmd := ansi.Cmd(t.parser.Cmd)
	switch cmd.Command() {
	case 'J':
		switch count {
		case 0: // Erase screen below (including cursor)
			rect := cellbuf.Rect(0, y, w, h-y)
			t.scr.Clear(&rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 1: // Erase screen above (including cursor)
			rect := cellbuf.Rect(0, 0, w, y+1)
			t.scr.Clear(&rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 2: // erase screen
			fallthrough
		case 3: // erase display
			// TODO: Scrollback buffer support?
			rect := cellbuf.Rect(0, 0, w, h)
			t.scr.Clear(&rect)
			if t.Damage != nil {
				t.Damage(ScreenDamage{w, h})
			}
		}
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
		// NOTE: Erase Line (EL) is a bit confusing. Erasing the line erases
		// the characters on the line while applying the cursor pen style
		// like background color and so on. The cursor position is not changed.
		cur := t.scr.Cursor()
		x, y := cur.Pos.X, cur.Pos.Y
		w := t.scr.Width()
		switch count {
		case 0: // Erase from cursor to end of line
			c := spaceCell
			c.Style = t.scr.cur.Pen
			rect := cellbuf.Rect(x, y, w-x, 1)
			t.scr.Fill(c, &rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 1: // Erase from start of line to cursor
			c := spaceCell
			c.Style = t.scr.cur.Pen
			rect := cellbuf.Rect(0, y, x+1, 1)
			t.scr.Fill(c, &rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 2: // Erase entire line
			c := spaceCell
			c.Style = t.scr.cur.Pen
			rect := cellbuf.Rect(0, y, w, 1)
			t.scr.Fill(c, &rect)
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

		sr := t.scrollregion
		log.Printf("SU: scrolling region %d, %d", sr.Min.Y, sr.Max.Y)
		t.scr.ScrollUp(n, &sr)
	case 'T': // SD - Scroll Down
		n := 1
		if t.parser.ParamsLen > 0 {
			if param := ansi.Param(t.parser.Params[0]).Param(1); param > 0 {
				n = param
			}
		}

		sr := t.scrollregion
		t.scr.ScrollDown(n, &sr)
	}
}
