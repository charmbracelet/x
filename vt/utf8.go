package vt

import (
	"log"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// handleUtf8 handles a UTF-8 characters.
func (t *Terminal) handleUtf8(seq []byte, width int) {
	var autowrap bool
	cur := t.scr.cur
	x, y := cur.Pos.X, cur.Pos.Y
	if mode, ok := t.pmodes[ansi.AutowrapMode]; ok && mode.IsSet() {
		autowrap = true
	}

	if autowrap && x+width > t.scr.Width() {
		x = 0
		y++
		sr := t.scrollregion
		log.Printf("utf8: scrolling region %d, %d", sr.Min.Y, sr.Max.Y)
		t.scr.ScrollUp(1, &sr)
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

	t.scr.moveCursor(x+width, y)
}
