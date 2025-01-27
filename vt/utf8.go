package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

// handleUtf8 handles a UTF-8 characters.
func (t *Terminal) handleUtf8(r rune) {
	var width int
	var content string
	width = runewidth.RuneWidth(r)
	content = string(r)

	x, y := t.scr.CursorPosition()
	if t.atPhantom || x+width > t.scr.Width() {
		// moves cursor down similar to [Terminal.linefeed] except it doesn't
		// respects [ansi.LNM] mode.
		// This will rest the phantom state i.e. pending wrap state.
		t.index()
		_, y = t.scr.CursorPosition()
		x = 0
	}

	// Handle character set mappings
	if len(content) == 1 {
		var charset CharSet
		c := content[0]
		if t.gsingle > 1 && t.gsingle < 4 {
			charset = t.charsets[t.gsingle]
			t.gsingle = 0
		} else if c < 128 {
			charset = t.charsets[t.gl]
		} else {
			charset = t.charsets[t.gr]
		}

		if charset != nil {
			if r, ok := charset[c]; ok {
				content = r
			}
		}
	}

	cell := &Cell{
		Style: t.scr.cursorPen(),
		Link:  Link{}, // TODO: Link support
		// FIXME: This is incorrect and ignores combining characters
		Rune:  firstRune(content),
		Width: width,
	}

	if t.scr.SetCell(x, y, cell) {
		t.lastChar = r
	}

	// Handle phantom state at the end of the line
	if x+width >= t.scr.Width() {
		if t.isModeSet(ansi.AutoWrapMode) {
			t.atPhantom = true
		}
	} else {
		x += width
	}

	// NOTE: We don't reset the phantom state here, we handle it up above.
	t.scr.setCursor(x, y, false)
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}
