package vt

import (
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// handlePrint handles printable characters.
func (e *Emulator) handlePrint(r rune) {
	if r >= ansi.SP && r < ansi.DEL {
		if len(e.grapheme) > 0 {
			// If we have a grapheme buffer, flush it before handling the ASCII character.
			e.flushGrapheme()
		}
		e.handleGrapheme(string(r), 1)
	} else {
		e.grapheme = append(e.grapheme, r)
	}
}

// flushGrapheme flushes the current grapheme buffer, if any, and handles the
// grapheme as a single unit.
func (e *Emulator) flushGrapheme() {
	if len(e.grapheme) == 0 {
		return
	}

	// XXX: We always use [ansi.GraphemeWidth] here to report accurate widths
	// and it's up to the caller to decide how to handle Unicode vs non-Unicode
	// modes.
	method := ansi.GraphemeWidth
	graphemes := string(e.grapheme)
	for len(graphemes) > 0 {
		cluster, width := ansi.FirstGraphemeCluster(graphemes, method)
		e.handleGrapheme(cluster, width)
		graphemes = graphemes[len(cluster):]
	}
	e.grapheme = e.grapheme[:0] // Reset the grapheme buffer.
}

// handleGrapheme handles UTF-8 graphemes.
func (e *Emulator) handleGrapheme(content string, width int) {
	awm := e.isModeSet(ansi.ModeAutoWrap)
	cell := uv.Cell{
		Content: content,
		Width:   width,
		Style:   e.scr.cursorPen(),
		Link:    e.scr.cursorLink(),
	}

	x, y := e.scr.CursorPosition()

	// Handle character set mappings
	if len(content) == 1 { //nolint:nestif
		var charset CharSet
		c := content[0]
		if e.gsingle > 1 && e.gsingle < 4 {
			charset = e.charsets[e.gsingle]
			e.gsingle = 0
		} else if c < 128 {
			charset = e.charsets[e.gl]
		} else {
			charset = e.charsets[e.gr]
		}

		if charset != nil {
			if r, ok := charset[c]; ok {
				cell.Content = r
				cell.Width = 1
			}
		}
	}

	if cell.Width == 1 && len(content) == 1 {
		e.lastChar, _ = utf8.DecodeRuneInString(content)
	}

	// Wrap once the cell is final, since the character set mapping above may
	// have changed its width. Either a wrap is already pending, or the cell is
	// wider than the room left on the line — a two-column character with one
	// column to go belongs on the next line whole, not written over the edge.
	// A cell too wide for the screen itself is placed regardless; no line
	// would hold it and looking for one would only lose it.
	scrWidth := e.scr.Width()
	if awm && (e.atPhantom || (x+cell.Width > scrWidth && cell.Width <= scrWidth)) {
		// moves cursor down similar to [Terminal.linefeed] except it doesn't
		// respects [ansi.LNM] mode.
		// This will reset the phantom state i.e. pending wrap state.
		e.index()
		_, y = e.scr.CursorPosition()
		x = 0
	}

	e.scr.SetCell(x, y, &cell)

	// Handle phantom state at the end of the line. The cell occupies the
	// columns from x to x+Width, so it is the far edge of a wide one that
	// decides whether the line is full, not the column it starts in.
	e.atPhantom = awm && x+cell.Width >= scrWidth
	if e.atPhantom {
		// Pending wrap leaves the cursor on the cell, and for a wide one that
		// is its last column, not its first: everything that reads the cursor
		// back — a resize, a backspace, a cursor report — expects the column
		// the character reached.
		x += cell.Width - 1
	} else {
		x += cell.Width
	}

	// NOTE: We don't reset the phantom state here, we handle it up above.
	e.scr.setCursor(x, y, false)
}
