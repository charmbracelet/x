package vt

// handleCsi handles a CSI escape sequences.
func (t *Terminal) handleCsi(seq []byte) {
	params := t.parser.Params[:t.parser.ParamsLen]
	cmd := t.parser.Cmd
	switch cmd {
	case 'm': // SGR - Select Graphic Rendition
		t.handleSgr(params)
	}
}

// handleSgr handles SGR escape sequences.
func (t *Terminal) handleSgr(params []int) {
}
