package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/wcwidth"
)

// handleUtf8 handles a UTF-8 characters.
func (t *Terminal) handleUtf8(seq ansi.Sequence) {
	var width int
	switch seq := seq.(type) {
	case ansi.Rune:
		width = wcwidth.RuneWidth(rune(seq))
	case ansi.Grapheme:
		width = seq.Width
	}

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

	cell := &Cell{
		Style:   t.scr.cur.Pen,
		Link:    Link{}, // TODO: Link support
		Content: seq.String(),
		Width:   width,
	}

	if t.scr.SetCell(x, y, cell) && t.Damage != nil {
		t.Damage(CellDamage{X: x, Y: y})
	}

	// TODO: Is this correct?
	t.scr.setCursor(x+width, y, true)
}
