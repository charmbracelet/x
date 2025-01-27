package vt

import "github.com/charmbracelet/x/ansi"

// handleDcs handles a DCS escape sequence.
func (t *Terminal) handleDcs(cmd ansi.Cmd, params ansi.Params, data []byte) {
	t.logf("unhandled sequence: DCS %q %q", paramsString(cmd, params), data)
}
