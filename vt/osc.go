package vt

import (
	"bytes"
	"image/color"
	"io"

	"github.com/charmbracelet/x/ansi"
)

// handleOsc handles an OSC escape sequence.
func (t *Terminal) handleOsc(cmd int, data []byte) {
	t.flushGrapheme() // Flush any pending grapheme before handling OSC sequences.
	if !t.handlers.handleOsc(cmd, data) {
		t.logf("unhandled sequence: OSC %q", data)
	}
}

func (t *Terminal) handleTitle(cmd int, data []byte) {
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 2 {
		// Invalid, ignore
		return
	}
	switch cmd {
	case 0: // Set window title and icon name
		name := string(parts[1])
		t.iconName, t.title = name, name
		if t.cb.Title != nil {
			t.cb.Title(name)
		}
		if t.cb.IconName != nil {
			t.cb.IconName(name)
		}
	case 1: // Set icon name
		name := string(parts[1])
		t.iconName = name
		if t.cb.IconName != nil {
			t.cb.IconName(name)
		}
	case 2: // Set window title
		name := string(parts[1])
		t.title = name
		if t.cb.Title != nil {
			t.cb.Title(name)
		}
	}
}

func (t *Terminal) handleDefaultColor(cmd int, data []byte) {
	var setCol func(color.Color)
	var col color.Color

	parts := bytes.Split(data, []byte{';'})
	if len(parts) == 0 {
		// Invalid, ignore
		return
	}

	switch cmd {
	case 10, 11, 12:
		if len(parts) != 2 {
			// Invalid, ignore
			return
		}

		var enc func(string) string
		if s := string(parts[1]); s == "?" {
			switch cmd {
			case 10:
				enc = ansi.SetForegroundColor
				col = t.ForegroundColor()
				if col == nil {
					col = defaultFg
				}
			case 11:
				enc = ansi.SetBackgroundColor
				col = t.BackgroundColor()
				if col == nil {
					col = defaultBg
				}
			case 12:
				enc = ansi.SetCursorColor
				col = t.CursorColor()
				if col == nil {
					col = defaultCur
				}
			}

			if enc != nil && col != nil {
				xrgb := ansi.XRGBColor{Color: col}
				io.WriteString(t.pw, enc(xrgb.String())) //nolint:errcheck
			}
		} else {
			col = ansi.XParseColor(string(parts[1]))
			if col == nil {
				return
			}
		}
	case 110, 111, 112:
		col = nil
	}

	switch cmd {
	case 10, 110: // Set/Reset foreground color
		setCol = t.SetForegroundColor
		if t.cb.ForegroundColor != nil && t.fgColor != col {
			t.cb.ForegroundColor(col)
		}
	case 11, 111: // Set/Reset background color
		setCol = t.SetBackgroundColor
		if t.cb.BackgroundColor != nil && t.bgColor != col {
			t.cb.BackgroundColor(col)
		}
	case 12, 112: // Set/Reset cursor color
		setCol = t.SetCursorColor
		if t.cb.CursorColor != nil && t.curColor != col {
			t.cb.CursorColor(col)
		}
	}

	setCol(col)
}

func (t *Terminal) handleWorkingDirectory(cmd int, data []byte) {
	if cmd != 7 {
		// Invalid, ignore
		return
	}

	// The data is the working directory path.
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 2 {
		// Invalid, ignore
		return
	}

	path := string(parts[1])
	t.cwd = path

	if t.cb.WorkingDirectory != nil {
		t.cb.WorkingDirectory(path)
	}
}

func (t *Terminal) handleHyperlink(cmd int, data []byte) {
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 3 || cmd != 8 {
		// Invalid, ignore
		return
	}

	t.scr.cur.Link.URL = string(parts[1])
	t.scr.cur.Link.Params = string(parts[2])
}
