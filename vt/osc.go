package vt

// handleOsc handles an OSC escape sequence.
func (t *Terminal) handleOsc([]byte) {
	cmd := t.parser.Cmd
	switch cmd {
	case 0: // Set window title and icon name
		name := string(t.parser.Data[:t.parser.DataLen])
		t.iconName, t.title = name, name
	case 1: // Set icon name
		name := string(t.parser.Data[:t.parser.DataLen])
		t.iconName = name
	case 2: // Set window title
		name := string(t.parser.Data[:t.parser.DataLen])
		t.title = name
	}
}
