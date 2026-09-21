package vt

import "github.com/charmbracelet/x/ansi"

// handleDcs handles a DCS escape sequence.
func (e *Emulator) handleDcs(cmd ansi.Cmd, params ansi.Params, data []byte) {
	e.endCluster() // A DCS sequence ends any cluster in progress.
	if !e.handlers.handleDcs(cmd, params, data) {
		e.logf("unhandled sequence: DCS %q %q", paramsString(cmd, params), data)
	}
}

// handleApc handles an APC escape sequence.
func (e *Emulator) handleApc(data []byte) {
	e.endCluster() // An APC sequence ends any cluster in progress.
	if !e.handlers.handleApc(data) {
		e.logf("unhandled sequence: APC %q", data)
	}
}

// handleSos handles an SOS escape sequence.
func (e *Emulator) handleSos(data []byte) {
	e.endCluster() // A SOS sequence ends any cluster in progress.
	if !e.handlers.handleSos(data) {
		e.logf("unhandled sequence: SOS %q", data)
	}
}

// handlePm handles a PM escape sequence.
func (e *Emulator) handlePm(data []byte) {
	e.endCluster() // A PM sequence ends any cluster in progress.
	if !e.handlers.handlePm(data) {
		e.logf("unhandled sequence: PM %q", data)
	}
}
