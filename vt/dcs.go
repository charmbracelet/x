package vt

import "github.com/charmbracelet/x/ansi"

// handleDcs handles a DCS escape sequence.
func (t *Terminal) handleDcs(cmd ansi.Cmd, params ansi.Params, data []byte) {
	if !t.handlers.handleDcs(cmd, params, data) {
		t.logf("unhandled sequence: DCS %q %q", paramsString(cmd, params), data)
	}
}

// handleApc handles an APC escape sequence.
func (t *Terminal) handleApc(data []byte) {
	if !t.handlers.handleApc(data) {
		t.logf("unhandled sequence: APC %q", data)
	}
}
