package vt

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// newReferenceEmulator returns an emulator that takes none of the fast
// paths, to check them against.
func newReferenceEmulator(w, h int) *Emulator {
	e := NewEmulator(w, h)
	e.reference = true
	e.scrs[0].reference = true
	e.scrs[1].reference = true
	return e
}

// fuzzText is the non-ASCII text the streams are made of: wide and isolated
// characters, combining marks, emoji sequences, Hangul jamo that join,
// prepend characters, and invalid or truncated UTF-8.
var fuzzText = []string{
	"é", "e\u0301", "\u0301", "世", "界", "─", "│", "╭", "█", "▒", "⣿", "한", "글",
	"\u1100", "\u1161", "\u11a8", "\u1100\u1161\u11a8", "\u1100가", "\u0600", "\u0600a",
	"👍", "👍🏽", "🇺🇸", "🇺", "\u200d", "👨\u200d👩\u200d👧", "©", "©\ufe0f", "\ufe0f",
	"\u0915\u094d\u0937", "\u094d", "ア", "\u3099", "Ω", "ß", "\u00ad",
	"\xff", "\xc0\xaf", "\x80", "\xe2\x82", "\xed\xa0\x80", "\xf0\x9f", "\xe4",
}

// fuzzStream returns a random stream of what programs write to a terminal.
func fuzzStream(r *rand.Rand, tokens int) []byte {
	var b bytes.Buffer
	word := func(n int) {
		for range n {
			b.WriteByte(byte('a' + r.IntN(26)))
		}
	}
	for range tokens {
		switch r.IntN(24) {
		case 0, 1, 2, 3:
			word(1 + r.IntN(12))
		case 4:
			word(40 + r.IntN(160)) // wraps
		case 5:
			b.WriteString([]string{" ", "  ", "\t", "\b", "\x7f"}[r.IntN(5)])
		case 6, 7:
			b.WriteString([]string{"\r\n", "\n", "\r", "\v", "\f", "\x85", "\x84", "\x8d"}[r.IntN(8)])
		case 8, 9, 10:
			b.WriteString(fuzzText[r.IntN(len(fuzzText))])
		case 11:
			// Runs of isolated characters, the TUI fast path.
			for range 1 + r.IntN(30) {
				b.WriteString([]string{"─", "│", "█", "⣿", "世", "한", "a"}[r.IntN(7)])
			}
		case 12:
			fmt.Fprintf(&b, "\x1b[%d%c", r.IntN(30), "ABCDEFGHJKLMPSTX@bdIZ`ae"[r.IntN(24)])
		case 13:
			fmt.Fprintf(&b, "\x1b[%d;%dH", r.IntN(30), r.IntN(30))
		case 14:
			fmt.Fprintf(&b, "\x1b[%d;%dr", r.IntN(12), r.IntN(14))
		case 15:
			if r.IntN(2) == 0 {
				fmt.Fprintf(&b, "\x1b[?69h\x1b[%d;%ds", r.IntN(12), r.IntN(24))
			} else {
				b.WriteString("\x1b[?69l")
			}
		case 16:
			b.WriteString([]string{
				"\x1b[0m", "\x1b[1;31m", "\x1b[7m", "\x1b[4:3m", "\x1b[41m", "\x1b[22;39;49m",
				fmt.Sprintf("\x1b[38;5;%dm", r.IntN(256)),
				fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r.IntN(256), r.IntN(256), r.IntN(256)),
			}[r.IntN(8)])
		case 17:
			b.WriteString([]string{
				"\x1b[?7l", "\x1b[?7h", "\x1b[20h", "\x1b[20l", "\x1b[?6h", "\x1b[?6l",
				"\x1b[?1049h", "\x1b[?1049l", "\x1b[?25l", "\x1b[?25h", "\x1b[4h", "\x1b[4l",
			}[r.IntN(12)])
		case 18:
			b.WriteString([]string{
				"\x1b7", "\x1b8", "\x1bD", "\x1bM", "\x1bE", "\x1b(0", "\x1b(B", "\x1b)0",
				"\x0e", "\x0f", "\x1bN", "\x1bO", "\x8e", "\x1b#8", "\x1b[!p", "\x1bH",
			}[r.IntN(16)])
		case 19:
			fmt.Fprintf(&b, "\x1b]0;title %d\x07", r.IntN(100))
		case 20:
			fmt.Fprintf(&b, "\x1b[%db", r.IntN(20))
		case 21:
			// Lines of output, the scrolling fast path.
			for range 1 + r.IntN(40) {
				word(r.IntN(30))
				b.WriteString("\r\n")
			}
		case 22:
			b.WriteString("\x1b[2J\x1b[H")
		default:
			if r.IntN(20) == 0 {
				b.WriteString("\x1bc")
			}
		}
	}
	return b.Bytes()
}

// TestMarginsPastTheScreen checks that margins set beyond the edge of the
// screen are clamped to it, like xterm does, rather than letting scrolling
// index lines that don't exist.
func TestMarginsPastTheScreen(t *testing.T) {
	for _, e := range []*Emulator{NewEmulator(10, 5), newReferenceEmulator(10, 5)} {
		_, _ = e.WriteString("\x1b[2;100r")
		if got := e.scr.ScrollRegion(); got != uv.Rect(0, 1, 10, 4) {
			t.Fatalf("expected the bottom margin clamped to the screen, got %v", got)
		}
		_, _ = e.WriteString("\x1b[?69h\x1b[3;100s")
		if got := e.scr.ScrollRegion(); got != uv.Rect(2, 1, 8, 4) {
			t.Fatalf("expected the right margin clamped to the screen, got %v", got)
		}
		// Scrolling and printing at the margins must not panic.
		_, _ = e.WriteString("\x1b[5S\x1b[5T\x1b[3L\x1b[3M\x1b[5;10Hhello\r\n\r\n\r\n")
		// A region entirely past the screen is ignored.
		_, _ = e.WriteString("\x1b[r\x1b[7;9r")
		if got := e.scr.ScrollRegion(); got.Min.Y != 0 || got.Max.Y != 5 {
			t.Fatalf("expected a region past the screen to be ignored, got %v", got)
		}
	}
}

// TestFastPathsMatchReference runs random streams through an emulator that
// takes the fast paths and one that doesn't, and checks they end up in the
// same state after every write.
func TestFastPathsMatchReference(t *testing.T) {
	iterations := 300
	if testing.Short() {
		iterations = 30
	}
	for seed := range uint64(iterations) {
		r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
		w, h := 1+r.IntN(24), 1+r.IntN(12)
		fast, ref := NewEmulator(w, h), newReferenceEmulator(w, h)
		fast.SetScrollbackSize(1 + r.IntN(50))
		ref.SetScrollbackSize(fast.Scrollback().MaxLines())

		stream := fuzzStream(r, 300)
		for len(stream) > 0 {
			n := min(len(stream), 1+r.IntN(64))
			chunk := stream[:n]
			stream = stream[n:]

			if r.IntN(40) == 0 {
				w, h = 1+r.IntN(24), 1+r.IntN(12)
				fast.Resize(w, h)
				ref.Resize(w, h)
			}

			before, refBefore := snapshotScreen(fast), snapshotScreen(ref)
			fast.scr.ClearTouched()
			ref.scr.ClearTouched()
			_, _ = fast.Write(chunk)
			_, _ = ref.Write(chunk)

			if diff := compareEmulators(fast, ref); diff != "" {
				t.Fatalf("seed %d: after writing %q: %s", seed, chunk, diff)
			}
			// Changes must be marked as touched wherever the reference marks
			// them.
			refMissed := untouchedChanges(ref, refBefore)
			for cell := range untouchedChanges(fast, before) {
				if !refMissed[cell] {
					t.Fatalf("seed %d: after writing %q: cell %v changed but was not touched", seed, chunk, cell)
				}
			}
			for _, e := range []*Emulator{fast, ref} {
				if msg := checkUsedBounds(e); msg != "" {
					t.Fatalf("seed %d: after writing %q: %s", seed, chunk, msg)
				}
				if e.autoWrap != e.isModeSet(ansi.ModeAutoWrap) ||
					e.lineFeedNewLine != e.isModeSet(ansi.ModeLineFeedNewLine) ||
					e.originMode != e.isModeSet(ansi.ModeOrigin) {
					t.Fatalf("seed %d: cached modes out of step after writing %q", seed, chunk)
				}
			}
		}
	}
}

// screenSnapshot is the active screen of an emulator at some point.
type screenSnapshot struct {
	scr   *Screen
	w, h  int
	lines [][]uv.Cell
}

func snapshotScreen(e *Emulator) screenSnapshot {
	s := screenSnapshot{scr: e.scr, w: e.Width(), h: e.Height()}
	for y := range s.h {
		s.lines = append(s.lines, slices.Clone(e.scr.buf.Lines[y]))
	}
	return s
}

// untouchedChanges returns the cells that changed since the snapshot without
// being in a touched part of their line.
func untouchedChanges(e *Emulator, before screenSnapshot) map[[2]int]bool {
	missed := map[[2]int]bool{}
	if e.scr != before.scr || e.Width() != before.w || e.Height() != before.h {
		return missed // A different screen or size: everything changed.
	}
	after := snapshotScreen(e)
	touched := e.scr.Touched()
	for y := range after.lines {
		for x := range after.lines[y] {
			if after.lines[y][x] == before.lines[y][x] {
				continue
			}
			if y >= len(touched) || touched[y] == nil ||
				x < touched[y].FirstCell || x >= touched[y].LastCell {
				missed[[2]int{x, y}] = true
			}
		}
	}
	return missed
}

// compareEmulators describes how the state of two emulators differs.
func compareEmulators(a, b *Emulator) string {
	if (a.scr == &a.scrs[1]) != (b.scr == &b.scrs[1]) {
		return "alternate screen differs"
	}
	for i := range a.scrs {
		sa, sb := &a.scrs[i], &b.scrs[i]
		if sa.Width() != sb.Width() || sa.Height() != sb.Height() {
			return fmt.Sprintf("screen %d size differs", i)
		}
		for y := range sa.Height() {
			for x := range sa.Width() {
				if ca, cb := *sa.CellAt(x, y), *sb.CellAt(x, y); ca != cb {
					return fmt.Sprintf("screen %d cell %d,%d: %#v != %#v", i, x, y, ca, cb)
				}
			}
		}
		if sa.cur != sb.cur || sa.saved != sb.saved {
			return fmt.Sprintf("screen %d cursor: %+v != %+v", i, sa.cur, sb.cur)
		}
		if sa.scroll != sb.scroll {
			return fmt.Sprintf("screen %d scroll region: %v != %v", i, sa.scroll, sb.scroll)
		}
		if la, lb := sa.scrollback.Len(), sb.scrollback.Len(); la != lb {
			return fmt.Sprintf("screen %d scrollback length: %d != %d", i, la, lb)
		}
		for y := range sa.scrollback.Len() {
			la, lb := sa.scrollback.Line(y), sb.scrollback.Line(y)
			if len(la) != len(lb) {
				return fmt.Sprintf("screen %d scrollback line %d length: %d != %d", i, y, len(la), len(lb))
			}
			for x := range la {
				if la[x] != lb[x] {
					return fmt.Sprintf("screen %d scrollback cell %d,%d: %#v != %#v", i, x, y, la[x], lb[x])
				}
			}
		}
	}
	switch {
	case a.atPhantom != b.atPhantom:
		return fmt.Sprintf("phantom state: %t != %t", a.atPhantom, b.atPhantom)
	case a.lastChar != b.lastChar:
		return fmt.Sprintf("last character: %q != %q", a.lastChar, b.lastChar)
	case a.gl != b.gl || a.gr != b.gr || a.gsingle != b.gsingle ||
		fmt.Sprint(a.charsets) != fmt.Sprint(b.charsets):
		return "character sets differ"
	case !bytes.Equal(a.grapheme, b.grapheme):
		return fmt.Sprintf("pending cluster: %q != %q", a.grapheme, b.grapheme)
	case a.parser.State() != b.parser.State():
		return "parser state differs"
	}
	return ""
}

// checkUsedBounds checks the invariant of the written width of lines: all the
// cells past it are empty cells.
func checkUsedBounds(e *Emulator) string {
	for i := range e.scrs {
		scr := &e.scrs[i]
		if len(scr.used) != scr.Height() {
			return fmt.Sprintf("screen %d: %d bounds for %d lines", i, len(scr.used), scr.Height())
		}
		for y, line := range scr.buf.Lines {
			for x := scr.used[y]; x < len(line); x++ {
				if !isPlainEmpty(&line[x]) {
					return fmt.Sprintf("screen %d: cell %d,%d is %#v but the line is bounded at %d", i, x, y, line[x], scr.used[y])
				}
			}
		}
	}
	return ""
}
