package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// handleUtf8 handles a UTF-8 characters.
func (t *Terminal) handleUtf8(seq []byte, width int, r rune, rw int) {
	cur := t.scr.cur
	x, y := cur.Pos.X, cur.Pos.Y
	if autowrap, ok := t.pmodes[ansi.AutowrapMode]; ok && autowrap.IsSet() {
		if x+width > t.scr.Width() {
			x = 0
			y++
		}
	}

	t.scr.Draw(x, y, cellbuf.Cell{
		Style:   t.scr.cur.Pen,
		Link:    cellbuf.Link{},
		Content: string(seq),
		Width:   width,
	})
	t.scr.moveCursor(x+width, y)
}
