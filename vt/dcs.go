package vt

import "github.com/charmbracelet/x/ansi"

// handleDcs handles a DCS escape sequence.
func (t *Terminal) handleDcs(seq ansi.DcsSequence) {
	t.logf("unhandled DCS: %q", seq)
}
