package vt

import (
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// printableASCII reports whether r is a character that can never join the
// grapheme cluster beside it. No ASCII character is a combining mark, a ZWJ, a
// spacing mark or a regional indicator, and CR and LF are not printable, so a
// cluster boundary between two of these is certain without asking the
// segmenter.
func printableASCII(r rune) bool {
	return r >= ansi.SP && r < ansi.DEL
}

// handlePrint handles printable characters.
//
// A character is held back rather than printed straight away, because the one
// after it may be a combining mark that belongs to the same cluster. Printing
// "e" from "e\u0301" before the mark arrives puts the base in one cell and
// leaves the mark to become a zero-width cell of its own, which the next write
// or erase then destroys.
func (e *Emulator) handlePrint(r rune) {
	// Two printable ASCII characters always have a boundary between them, so
	// the held character can go out with no segmentation. This is the common
	// case by a wide margin and the reason it is worth special-casing.
	if printableASCII(r) && len(e.grapheme) == 1 && printableASCII(rune(e.grapheme[0])) {
		e.handleGrapheme(string(e.grapheme[0]), 1)
		e.grapheme = utf8.AppendRune(e.grapheme[:0], r)
		return
	}

	e.grapheme = utf8.AppendRune(e.grapheme, r)
	e.emitFinishedGraphemes()
}

// emitFinishedGraphemes prints every cluster in the buffer that is known to be
// finished, which is every cluster but the last. Any character still to come
// could be a combining mark that extends it, so the last one waits for either
// the next character or a flush.
func (e *Emulator) emitFinishedGraphemes() {
	for len(e.grapheme) > 0 {
		cluster, width := ansi.FirstGraphemeCluster(e.grapheme, ansi.GraphemeWidth)
		if len(cluster) == 0 {
			// Not reachable for well-formed input, but this loop runs over
			// bytes an application did not choose, so refuse to spin.
			return
		}
		if len(cluster) >= len(e.grapheme) {
			// Only one cluster so far, and it may still grow.
			return
		}
		e.handleGrapheme(string(cluster), width)
		// Compact rather than reslice, so the buffer keeps reusing the space
		// it already has instead of walking off the end of its array.
		e.grapheme = e.grapheme[:copy(e.grapheme, e.grapheme[len(cluster):])]
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
	graphemes := e.grapheme
	for len(graphemes) > 0 {
		cluster, width := ansi.FirstGraphemeCluster(graphemes, method)
		if len(cluster) == 0 {
			break
		}
		e.handleGrapheme(string(cluster), width)
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
	if e.atPhantom && awm {
		// moves cursor down similar to [Terminal.linefeed] except it doesn't
		// respects [ansi.LNM] mode.
		// This will reset the phantom state i.e. pending wrap state.
		e.index()
		_, y = e.scr.CursorPosition()
		x = 0
	}

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

	e.scr.SetCell(x, y, &cell)

	// Handle phantom state at the end of the line
	e.atPhantom = awm && x >= e.scr.Width()-1
	if !e.atPhantom {
		x += cell.Width
	}

	// NOTE: We don't reset the phantom state here, we handle it up above.
	e.scr.setCursor(x, y, false)
}
