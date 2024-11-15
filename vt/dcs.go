package vt

import "github.com/charmbracelet/x/ansi"

// handleDcs handles a DCS escape sequence.
func (t *Terminal) handleDcs(seq ansi.DcsSequence) {
	mark, inter, cmd := seq.Cmd.Marker(), seq.Cmd.Intermediate(), seq.Cmd.Command()
	t.logf("unhandled DCS: (%c, %c, %c)", mark, inter, cmd)
}
