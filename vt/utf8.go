package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// handleUtf8 handles a UTF-8 characters.
func (t *Terminal) handleUtf8(seq []byte, width int) {
	var autowrap bool
	x, y := t.scr.CursorPosition()
	if mode, ok := t.pmodes[ansi.AutowrapMode]; ok && mode.IsSet() {
		autowrap = true
	}

	// Handle wide chars at the edge - wrap them entirely
	if autowrap && x+width > t.scr.Width() {
		x = 0
		y++
		// Only scroll if we're past the last line
		if y >= t.scr.Height() {
			t.scr.ScrollUp(1)
			y = t.scr.Height() - 1
		}
	}

	cell := cellbuf.Cell{
		Style:   t.scr.cur.Pen,
		Link:    cellbuf.Link{}, // TODO: Link support
		Content: string(seq),
		Width:   width,
	}

	if t.scr.Draw(x, y, cell) && t.Damage != nil {
		t.Damage(CellDamage{X: x, Y: y, Cell: cell})
	}

	t.scr.setCursor(x+width, y)
}
