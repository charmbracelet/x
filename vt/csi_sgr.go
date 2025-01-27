package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// handleSgr handles SGR escape sequences.
// handleSgr handles Select Graphic Rendition (SGR) escape sequences.
func (t *Terminal) handleSgr(params ansi.Params) {
	cellbuf.ReadStyle(params, &t.scr.cur.Pen)
}
