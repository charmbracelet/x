package vt

import (
	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/x/ansi"
)

// handleSgr handles SGR escape sequences.
// handleSgr handles Select Graphic Rendition (SGR) escape sequences.
func (t *Terminal) handleSgr(params ansi.Params) {
	uv.ReadStyle(params, &t.scr.cur.Pen)
}
