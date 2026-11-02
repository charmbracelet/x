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

// isolatedRune reports whether r is a character that, next to another such
// character or printable ASCII, is always a grapheme cluster of its own: it
// is neither Extend, ZWJ, SpacingMark, Prepend, a regional indicator, a
// Hangul jamo, Extended_Pictographic, nor part of an Indic conjunct. These
// are the non-ASCII characters programs draw the most, and finding their
// clusters needs no segmenting.
func isolatedRune(r rune) bool {
	switch {
	case r >= 0x2500 && r <= 0x259F: // Box Drawing, Block Elements
		return true
	case r >= 0x2800 && r <= 0x28FF: // Braille Patterns
		return true
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified Ideographs
		return true
	case r >= 0xAC00 && r <= 0xD7A3: // Hangul Syllables (LV and LVT)
		return true
	}
	return false
}

// handlePrint handles printable characters.
//
// A character is held back rather than printed straight away, because the one
// after it may be a combining mark that belongs to the same cluster. Printing
// "e" from "e\u0301" before the mark arrives puts the base in one cell and
// leaves the mark to become a zero-width cell of its own, which the next write
// or erase then destroys.
func (e *Emulator) handlePrint(r rune) {
	// No ASCII character can extend the cluster before it: none of them is
	// Extend, ZWJ, SpacingMark, Prepend, a regional indicator, or part of an
	// Indic conjunct. So an ASCII character arriving is proof that whatever is
	// buffered is finished, and the buffer can go out.
	//
	// This is also what keeps the buffer from being re-segmented as it grows.
	// Asking where the clusters are on every character costs the length of the
	// buffer each time, which is quadratic over a long run of combining marks
	// that never resolves into more than one cluster.
	if printableASCII(r) {
		// Two printable ASCII characters in a row is the common case by a wide
		// margin, and needs no segmenting at all.
		if len(e.grapheme) == 1 && printableASCII(rune(e.grapheme[0])) {
			e.handleGrapheme(asciiStrings[e.grapheme[0]], 1)
			e.grapheme = utf8.AppendRune(e.grapheme[:0], r)
			return
		}
		e.flushGrapheme()
	} else if !e.reference && isolatedRune(r) && len(e.grapheme) > 0 {
		// Likewise, an isolated character can't extend a single isolated or
		// ASCII character before it, which is then a cluster of its own.
		if prev, n := utf8.DecodeRune(e.grapheme); n == len(e.grapheme) &&
			(printableASCII(prev) || isolatedRune(prev)) {
			e.printCluster(e.grapheme)
			e.grapheme = e.grapheme[:0]
		}
	}

	e.grapheme = utf8.AppendRune(e.grapheme, r)
}

// flushGrapheme flushes the current grapheme buffer, if any, and handles the
// grapheme as a single unit.
func (e *Emulator) flushGrapheme() {
	if len(e.grapheme) == 0 {
		return
	}

	// A single character is a single cluster.
	if _, n := utf8.DecodeRune(e.grapheme); n == len(e.grapheme) && !e.reference {
		e.printCluster(e.grapheme)
		e.grapheme = e.grapheme[:0]
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
		e.handleGrapheme(e.cluster(cluster, width).content, width)
		graphemes = graphemes[len(cluster):]
	}
	e.grapheme = e.grapheme[:0] // Reset the grapheme buffer.
}

// maxClusters bounds how many grapheme clusters are interned.
const maxClusters = 4096

// clusterInfo is an interned grapheme cluster and its width.
type clusterInfo struct {
	content string
	width   int
}

// printCluster prints a complete grapheme cluster.
func (e *Emulator) printCluster(cluster []byte) {
	if len(cluster) == 1 && printableASCII(rune(cluster[0])) {
		e.handleGrapheme(asciiStrings[cluster[0]], 1)
		return
	}
	c := e.cluster(cluster, -1)
	e.handleGrapheme(c.content, c.width)
}

// cluster returns the grapheme cluster as a string, with its width, computing
// the width when it is negative. Clusters are interned, since programs draw
// the same few non-ASCII characters (borders, blocks, symbols) over and over,
// and each would otherwise cost an allocation and a width lookup.
func (e *Emulator) cluster(cluster []byte, width int) clusterInfo {
	if c, ok := e.clusters[string(cluster)]; ok {
		return c
	}
	if width < 0 {
		_, width = ansi.FirstGraphemeCluster(cluster, ansi.GraphemeWidth)
	}
	c := clusterInfo{content: string(cluster), width: width}
	if len(e.clusters) < maxClusters {
		if e.clusters == nil {
			e.clusters = make(map[string]clusterInfo)
		}
		e.clusters[c.content] = c
	}
	return c
}

// asciiStrings holds a string for every ASCII character, so printing one
// doesn't allocate.
var asciiStrings = func() (s [128]string) {
	for i := range s {
		s[i] = string(rune(i))
	}
	return s
}()

// printASCII prints a run of printable ASCII characters received in the
// ground state. It does what calling [Emulator.handlePrint] for each of them
// does, but writes the cells straight into the screen.
func (e *Emulator) printASCII(run []byte) {
	// No ASCII character extends the cluster before it, so whatever is
	// pending is complete.
	e.flushGrapheme()

	// The last character is held back like handlePrint does, since a
	// combining mark may follow it.
	last := len(run) - 1
	if last > 0 {
		e.printASCIICells(run[:last])
	}
	e.grapheme = append(e.grapheme[:0], run[last])
}

// printASCIICells prints printable ASCII characters, each in its own cell.
func (e *Emulator) printASCIICells(chars []byte) {
	scr := e.scr
	width := scr.Width()
	if e.gsingle != 0 || e.charsets[e.gl] != nil || width <= 0 {
		// Character set mappings apply, or there are no cells to write.
		for _, c := range chars {
			e.handleGrapheme(asciiStrings[c], 1)
		}
		return
	}

	awm := e.autoWrap
	cell := uv.Cell{Width: 1, Style: scr.cursorPen(), Link: scr.cursorLink()}
	for _, c := range chars {
		x, y := scr.CursorPosition()
		if e.atPhantom && awm {
			e.index()
			_, y = scr.CursorPosition()
			x = 0
		}

		cell.Content = asciiStrings[c]
		if y < len(scr.buf.Lines) && x < len(scr.buf.Lines[y]) && scr.buf.Lines[y][x].Width == 1 {
			// Overwriting a narrow cell needs none of the fixups for wide
			// cells, and only marks the line changed when the cell changes.
			if old := &scr.buf.Lines[y][x]; !old.Equal(&cell) {
				*old = cell
				scr.buf.TouchLine(x, y, 1)
			}
		} else {
			// This also covers a cursor left outside of the screen, by a
			// resize or a restored cursor, where nothing is written.
			scr.SetCell(x, y, &cell)
		}

		e.atPhantom = awm && x >= width-1
		if !e.atPhantom {
			x++
		}
		scr.setCursor(x, y, false)
	}
	e.lastChar = rune(chars[len(chars)-1])
}

// handleGrapheme handles UTF-8 graphemes.
func (e *Emulator) handleGrapheme(content string, width int) {
	awm := e.autoWrap
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

	// A single shift applies to the next character, whatever that character
	// turns out to be. Take it now, so it cannot leak onto the one after this
	// when this is a cluster no charset has a mapping for.
	single := e.gsingle
	e.gsingle = 0

	// Handle character set mappings
	if len(content) == 1 { //nolint:nestif
		var charset CharSet
		c := content[0]
		if single > 1 && single < 4 {
			charset = e.charsets[single]
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
