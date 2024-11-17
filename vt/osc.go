package vt

import (
	"bytes"
	"image/color"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

// handleOsc handles an OSC escape sequence.
func (t *Terminal) handleOsc(seq ansi.OscSequence) {
	switch cmd := t.parser.Cmd(); cmd {
	case 0, 1, 2:
		parts := bytes.Split(t.parser.Data(), []byte{';'})
		if len(parts) != 2 {
			// Invalid, ignore
			return
		}
		switch cmd {
		case 0: // Set window title and icon name
			name := string(parts[1])
			t.iconName, t.title = name, name
			if t.Callbacks.Title != nil {
				t.Callbacks.Title(name)
			}
			if t.Callbacks.IconName != nil {
				t.Callbacks.IconName(name)
			}
		case 1: // Set icon name
			name := string(parts[1])
			t.iconName = name
			if t.Callbacks.IconName != nil {
				t.Callbacks.IconName(name)
			}
		case 2: // Set window title
			name := string(parts[1])
			t.title = name
			if t.Callbacks.Title != nil {
				t.Callbacks.Title(name)
			}
		}
	case 10, 11, 12, 110, 111, 112:
		var setCol func(color.Color)
		var col color.Color

		parts := bytes.Split(t.parser.Data(), []byte{';'})
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

			var enc func(color.Color) string
			if s := string(parts[1]); s == "?" {
				switch cmd {
				case 10:
					enc = ansi.SetForegroundColor
					col = t.ForegroundColor()
				case 11:
					enc = ansi.SetBackgroundColor
					col = t.BackgroundColor()
				case 12:
					enc = ansi.SetCursorColor
					col = t.CursorColor()
				}

				if enc != nil && col != nil {
					t.buf.WriteString(enc(ansi.XRGBColorizer{Color: col}))
				}
			} else {
				col := xParseColor(string(parts[1]))
				if col == nil {
					return
				}
			}
		case 110:
			col = defaultFg
		case 111:
			col = defaultBg
		case 112:
			col = defaultCur
		}

		switch cmd {
		case 10, 110: // Set/Reset foreground color
			setCol = t.SetForegroundColor
		case 11, 111: // Set/Reset background color
			setCol = t.SetBackgroundColor
		case 12, 112: // Set/Reset cursor color
			setCol = t.SetCursorColor
		}

		setCol(col)
	default:
		t.logf("unhandled OSC: %s", seq)
	}
}

type shiftable interface {
	~uint | ~uint16 | ~uint32 | ~uint64
}

func shift[T shiftable](x T) T {
	if x > 0xff {
		x >>= 8
	}
	return x
}

func xParseColor(s string) color.Color {
	switch {
	case strings.HasPrefix(s, "#"):
		c, err := colorful.Hex(s)
		if err != nil {
			return nil
		}

		return c
	case strings.HasPrefix(s, "rgb:"):
		parts := strings.Split(s[4:], "/")
		if len(parts) != 3 {
			return nil
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)

		return color.RGBA{uint8(shift(r)), uint8(shift(g)), uint8(shift(b)), 255} //nolint:gosec
	case strings.HasPrefix(s, "rgba:"):
		parts := strings.Split(s[5:], "/")
		if len(parts) != 4 {
			return nil
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)
		a, _ := strconv.ParseUint(parts[3], 16, 32)

		return color.RGBA{uint8(shift(r)), uint8(shift(g)), uint8(shift(b)), uint8(shift(a))} //nolint:gosec
	}
	return nil
}
