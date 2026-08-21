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

// firstCombining is the lowest rune that can continue a grapheme cluster on its
// own account: everything below it starts a new one. Latin-1 and ASCII text
// therefore never reaches the expensive test.
const firstCombining = 0x0300

// maxClusterBytes bounds how much one cell may accumulate by merging. The
// longest clusters in real text are subdivision flags and long emoji ZWJ
// sequences, tens of bytes; without a bound a stream delivering joiners one
// rune per write grows a single cell without limit, and each continuation
// rescans everything already there.
const maxClusterBytes = 256

// continuesCluster reports whether content could extend the cluster before it,
// rather than starting one of its own. It is only a cheap gate on the exact
// test in [Emulator.mergeIntoPrevious], which is the authority: an allowlist of
// continuation classes cannot be built from the unicode tables, because they
// and the grapheme-break data do not agree on every rune, so this keeps only
// ASCII and Latin-1 away and defers everything else.
func continuesCluster(content string) bool {
	r, _ := utf8.DecodeRuneInString(content)
	return r >= firstCombining
}

// mergeIntoPrevious folds content into the grapheme written just before it
// when the two are one cluster, and reports whether it did.
//
// The emulator cannot wait to find out whether a cluster is finished: that is
// only known once the rune after it arrives, so the base is written as soon as
// it is seen and a continuation has to be applied to it after the fact.
// Without this a cluster split across two [Emulator.Write] calls — which a
// pipe, socket, PTY or tmux control client may do at any byte — lands as two
// cells, and a combining mark following an ASCII base never attaches at all,
// because the ASCII fast path in [Emulator.handlePrint] has already committed
// it.
//
// The base is tracked rather than inferred from the cursor. At the last column
// the cursor stops advancing — pending wrap with autowrap on, clamped without
// it — so "the cell before the cursor" is not the cell just written, and
// inferring it there merges into the wrong one.
func (e *Emulator) mergeIntoPrevious(content string) bool {
	if e.lastGrapheme == "" || !continuesCluster(content) {
		return false
	}

	baseX, baseY := e.lastGraphemeX, e.lastGraphemeY
	prev := e.scr.CellAt(baseX, baseY)
	if prev == nil || prev.Content != e.lastGrapheme {
		return false // something has overwritten it since
	}
	if len(prev.Content)+len(content) > maxClusterBytes {
		return false
	}

	// The cursor has to be where writing that grapheme left it, or the base is
	// no longer what this continuation belongs to. A pending wrap leaves it on
	// the base instead of past it.
	x, y := e.scr.CursorPosition()
	wantX := baseX + prev.Width
	if e.atPhantom {
		wantX = baseX
	}
	// The cursor never leaves the screen: a cell reaching the last column
	// leaves it there rather than one past the end.
	if maxX := e.scr.Width() - 1; wantX > maxX {
		wantX = maxX
	}
	if x != wantX || y != baseY {
		return false
	}

	// Built in a buffer the emulator keeps: every character above Latin-1 comes
	// through here, and concatenating a fresh string for each would allocate
	// once per cell just to ask the question.
	e.mergeBuf = append(e.mergeBuf[:0], prev.Content...)
	e.mergeBuf = append(e.mergeBuf, content...)
	cluster, width := ansi.FirstGraphemeCluster(e.mergeBuf, ansi.GraphemeWidth)
	// Nothing of content belongs to the cluster prev ends.
	if len(cluster) <= len(prev.Content) {
		return false
	}
	// Some of it does but not all, as when a write starts with the tail of one
	// flag and continues into the next: keep the part that belongs and let the
	// rest be placed normally.
	rest := string(e.mergeBuf[len(cluster):])
	// Growing the cell must not push it off the line; leave that to be written
	// as its own cell rather than reflowing what is already placed.
	if baseX+width > e.scr.Width() {
		return false
	}

	merged := *prev
	merged.Content = string(cluster)
	merged.Width = width
	e.scr.SetCell(baseX, baseY, &merged)
	e.lastGrapheme = merged.Content

	// Leave the cursor exactly as writing this cell from scratch would have,
	// so the screen cannot depend on where the writer split its input.
	newX := baseX
	e.atPhantom = e.isModeSet(ansi.ModeAutoWrap) && baseX >= e.scr.Width()-1
	if !e.atPhantom {
		newX = baseX + width
	}
	e.scr.setCursor(newX, baseY, false)

	if rest != "" {
		e.handleGrapheme(rest, ansi.StringWidth(rest))
	}

	return true
}

// endCluster ends any cluster in progress, both the runes still buffered and
// the one already on the screen. Anything that is not a printable character
// breaks a grapheme cluster, and it may also move the cursor or rewrite the
// cell the merge anchor points at, so the anchor cannot outlive it.
func (e *Emulator) endCluster() {
	e.flushGrapheme()
	e.lastGrapheme = ""
}

// handleGrapheme handles UTF-8 graphemes.
func (e *Emulator) handleGrapheme(content string, width int) {
	if e.mergeIntoPrevious(content) {
		return
	}

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
	e.lastGrapheme, e.lastGraphemeX, e.lastGraphemeY = cell.Content, x, y

	// Handle phantom state at the end of the line
	e.atPhantom = awm && x >= e.scr.Width()-1
	if !e.atPhantom {
		x += cell.Width
	}

	// NOTE: We don't reset the phantom state here, we handle it up above.
	e.scr.setCursor(x, y, false)
}
