package vt

import (
	"unicode/utf8"

	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

// handlePrint handles printable characters.
func (t *Terminal) handlePrint(r rune) {
	t.handleGrapheme(string(r), runewidth.RuneWidth(r))
}

// handleGrapheme handles UTF-8 graphemes.
func (t *Terminal) handleGrapheme(content string, width int) {
	if !t.isModeSet(ansi.GraphemeClusteringMode) {
		width = runewidth.StringWidth(content)
	}

	cell := uv.Cell{
		Content: content,
		Width:   width,
	}
	x, y := t.scr.CursorPosition()
	if t.atPhantom || x+cell.Width > t.scr.Width() {
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
				cell.Content = string(r)
				cell.Width = 1
			}
		}
	}

	cell.Style = t.scr.cursorPen()
	cell.Link = t.scr.cursorLink()

	t.scr.SetCell(x, y, &cell)
	if cell.Width == 1 && len(content) == 1 {
		t.lastChar, _ = utf8.DecodeRuneInString(content)
	}

	// Handle phantom state at the end of the line
	if x+cell.Width >= t.scr.Width() {
		if t.isModeSet(ansi.AutoWrapMode) {
			t.atPhantom = true
		}
	} else {
		x += cell.Width
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
