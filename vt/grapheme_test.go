package vt

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// screenState renders the first row plus the cursor column, which together are
// what the issue compares between an emulator fed whole and one fed in pieces.
func screenState(e *Emulator, width int) (string, uv.Position) {
	var b strings.Builder
	for y := range e.Height() {
		for x := range width {
			if cell := e.CellAt(x, y); cell != nil {
				b.WriteString(cell.Content)
			}
			b.WriteByte('|')
		}
		b.WriteByte('\n')
	}
	return b.String(), e.CursorPosition()
}

// TestGraphemeClusterAcrossWrites is the reproduction from charmbracelet/x#935:
// the same bytes in the same order must land the same way however the writer
// happened to chop them up. A pipe, socket, PTY or tmux control client can
// split one visible character across two reads, and the emulator is the only
// place that can put it back together.
//
// Every case is split at every byte boundary and compared against the whole,
// which also pins that a split cannot change where the line wraps.
//
// A cluster whose finished width does not fit the last column is left out: the
// unsplit path writes an oversized cell there and the screen blanks it, which
// happens on main too (a single write of "ab" plus a two-column emoji on a
// four-column screen loses the emoji) and is not this change to make.
func TestGraphemeClusterAcrossWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		text  string
	}{
		{name: "heart with variation selector", width: 20, text: "\u2764\ufe0f vs \u2764"},
		{name: "combining accent on ASCII", width: 20, text: "cafe\u0301 vs cafe"},
		{name: "emoji with skin tone", width: 20, text: "\U0001F44D\U0001F3FB ok"},
		{name: "emoji ZWJ sequence", width: 20, text: "\U0001F469\u200d\U0001F4BB ok"},
		{name: "regional indicator flag", width: 20, text: "\U0001F1FA\U0001F1F8 ok"},
		// Two flags running together: a split can land mid-pair, so the second
		// write starts with the tail of one cluster and continues into the next.
		{name: "adjacent regional indicator flags", width: 20, text: "\U0001F1FA\U0001F1F8\U0001F1EC\U0001F1E7"},
		{name: "three flags", width: 20, text: "\U0001F1FA\U0001F1F8\U0001F1EC\U0001F1E7\U0001F1EF\U0001F1F5"},
		{name: "conjoining hangul jamo", width: 20, text: "\u1100\u1161\u11A8 ok"},
		// Narrow enough that the cluster decides where the line wraps.
		{name: "widening cluster at the wrap point", width: 4, text: "ab\U0001F44D\U0001F3FBc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			whole := NewEmulator(tc.width, 3)
			mustWrite(t, whole, tc.text)
			wantCells, wantCursor := screenState(whole, tc.width)

			for i := 1; i < len(tc.text); i++ {
				split := NewEmulator(tc.width, 3)
				mustWrite(t, split, tc.text[:i])
				mustWrite(t, split, tc.text[i:])

				gotCells, gotCursor := screenState(split, tc.width)
				if gotCells != wantCells {
					t.Errorf("split at byte %d: cells = %q, want %q", i, gotCells, wantCells)
				}
				if gotCursor != wantCursor {
					t.Errorf("split at byte %d: cursor = %v, want %v", i, gotCursor, wantCursor)
				}
			}
		})
	}
}

// TestGraphemeMergeIsBounded stops a stream that never stops joining from
// growing one cell without limit; each continuation rescans everything already
// in the cell, so an unbounded cluster is quadratic as well as large.
func TestGraphemeMergeIsBounded(t *testing.T) {
	t.Parallel()

	e := NewEmulator(20, 2)
	mustWrite(t, e, "\U0001F469")
	for range 500 {
		mustWrite(t, e, "\u200d")
		mustWrite(t, e, "\U0001F469")
	}

	cell := e.CellAt(0, 0)
	if cell == nil {
		t.Fatal("no cell at 0,0")
	}
	if len(cell.Content) > maxClusterBytes {
		t.Errorf("cell grew to %d bytes, want at most %d", len(cell.Content), maxClusterBytes)
	}
	// And it did keep merging up to the bound rather than stopping early.
	if len(cell.Content) < maxClusterBytes/2 {
		t.Errorf("cell holds only %d bytes, want it filled towards the bound", len(cell.Content))
	}
}

// TestGraphemeCombiningMarkJoinsBase covers the second case in the issue: a
// combining mark has to attach to the character before it even when that
// character took the ASCII fast path and was already committed to a cell.
func TestGraphemeCombiningMarkJoinsBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantCur int
	}{
		// Written with escapes on purpose: the decomposed and precomposed
		// forms are indistinguishable as literals but are not the same bytes.
		{name: "NFD e with combining acute", input: "e\u0301", want: "e\u0301", wantCur: 1},
		{name: "precomposed e acute", input: "\u00e9", want: "\u00e9", wantCur: 1},
		{name: "heart with variation selector", input: "\u2764\ufe0f", want: "\u2764\ufe0f", wantCur: 2},
		{name: "emoji with skin tone", input: "\U0001F44D\U0001F3FB", want: "\U0001F44D\U0001F3FB", wantCur: 2},
		{name: "adjacent ASCII must not merge", input: "ab", want: "a", wantCur: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := NewEmulator(20, 2)
			if _, err := e.Write([]byte(tc.input)); err != nil {
				t.Fatalf("writing: %v", err)
			}

			cell := e.CellAt(0, 0)
			if cell == nil {
				t.Fatal("no cell at 0,0")
			}
			if got := cell.String(); got != tc.want {
				t.Errorf("cell = %q, want %q", got, tc.want)
			}
			if got := e.CursorPosition().X; got != tc.wantCur {
				t.Errorf("cursor x = %d, want %d", got, tc.wantCur)
			}
		})
	}
}

// TestGraphemeMergeIntoWideBase covers a continuation that arrives in a later
// write than its wide base. Reaching the base means stepping back over the
// blank cell that trails a double-width one.
func TestGraphemeMergeIntoWideBase(t *testing.T) {
	t.Parallel()

	e := NewEmulator(10, 2)
	mustWrite(t, e, "\U0001F44D") // thumbs up, two columns
	mustWrite(t, e, "\U0001F3FB") // skin tone, in its own write

	base := e.CellAt(0, 0)
	if base == nil || base.Content != "\U0001F44D\U0001F3FB" {
		t.Errorf("base cell = %q, want the modified emoji", cellContent(base))
	}
	if base != nil && base.Width != 2 {
		t.Errorf("base width = %d, want 2", base.Width)
	}
	// The trailing blank of the wide cell must survive as a blank.
	if spacer := e.CellAt(1, 0); spacer == nil || spacer.Content != "" || spacer.Width != 0 {
		t.Errorf("spacer cell = %q width %v, want an empty zero-width cell", cellContent(spacer), spacerWidth(spacer))
	}
	if got := e.CursorPosition().X; got != 2 {
		t.Errorf("cursor x = %d, want 2", got)
	}
}

// TestGraphemeNoMergeWhenCursorMoved pins that a continuation only folds into a
// cell the cursor is actually sitting just past. Positioning the cursor into
// the middle of a wide cell must not rewrite that cell.
func TestGraphemeNoMergeWhenCursorMoved(t *testing.T) {
	t.Parallel()

	e := NewEmulator(10, 2)
	mustWrite(t, e, "\U0001F44D")
	mustWrite(t, e, "\x1b[1;2H") // column 2, inside the wide cell
	mustWrite(t, e, "\u0301")

	// Writing into the second half of a wide cell destroys it, which is the
	// screen's own rule; what must not happen is the mark folding into it.
	if base := e.CellAt(0, 0); base != nil && base.Content == "\U0001F44D\u0301" {
		t.Error("the mark merged into a cell the cursor was not sitting just past")
	}
}

// TestGraphemeNoMergeAcrossClusterBreak pins the rule the merge rests on: two
// pieces are folded together only when Unicode says they are one cluster, never
// merely because a continuation looked likely. A joiner invites anything to
// follow it, so ordinary text after one is the case that has to be refused.
func TestGraphemeNoMergeAcrossClusterBreak(t *testing.T) {
	t.Parallel()

	e := NewEmulator(10, 2)
	mustWrite(t, e, "\U0001F469") // woman
	mustWrite(t, e, "\u200d")     // joiner: the next rune may or may not belong
	mustWrite(t, e, "a")          // it does not

	if base := e.CellAt(0, 0); base == nil || base.Content != "\U0001F469\u200d" {
		t.Errorf("base cell = %q, want the joiner sequence left as it was", cellContent(base))
	}
	// Refusing to merge must not swallow the text: it stands as its own cell.
	if next := e.CellAt(2, 0); next == nil || next.Content != "a" {
		t.Errorf("cell after the base = %q, want %q", cellContent(next), "a")
	}
}

// TestGraphemeMergeAtLastColumn covers a continuation that does not widen its
// base but arrives while the cursor is parked on the last column, where it
// stops advancing. The merge still has to happen there.
func TestGraphemeMergeAtLastColumn(t *testing.T) {
	t.Parallel()

	e := NewEmulator(3, 2)
	mustWrite(t, e, "abe")    // 'e' lands in the last column
	mustWrite(t, e, "\u0301") // its accent arrives afterwards

	if base := e.CellAt(2, 0); base == nil || base.Content != "e\u0301" {
		t.Errorf("last cell = %q, want the accented e", cellContent(base))
	}

	// The merged cell still holds the last column, so the next character has
	// to wrap exactly as it would have without the merge.
	mustWrite(t, e, "z")
	if got := e.CursorPosition(); got.Y != 1 {
		t.Errorf("cursor after the next character = %v, want it wrapped to row 1", got)
	}
	if c := e.CellAt(0, 1); c == nil || c.Content != "z" {
		t.Errorf("wrapped cell = %q, want %q", cellContent(c), "z")
	}
}

// TestGraphemeMergeChainAcrossWrites covers a cluster that keeps growing: a ZWJ
// emoji sequence delivered one rune per write, which is what a slow pipe looks
// like. Each continuation has to fold into the result of the last one, so the
// anchor must follow the merged cell rather than the originally written base.
func TestGraphemeMergeChainAcrossWrites(t *testing.T) {
	t.Parallel()

	// Woman + ZWJ + laptop: "woman technologist".
	parts := []string{"\U0001F469", "\u200d", "\U0001F4BB"}
	want := strings.Join(parts, "")

	whole := NewEmulator(10, 2)
	mustWrite(t, whole, want)

	piecemeal := NewEmulator(10, 2)
	for _, part := range parts {
		mustWrite(t, piecemeal, part)
	}

	for _, tc := range []struct {
		name string
		e    *Emulator
	}{{"whole", whole}, {"piecemeal", piecemeal}} {
		cell := tc.e.CellAt(0, 0)
		if cell == nil || cell.Content != want {
			t.Errorf("%s: cell = %q, want %q", tc.name, cellContent(cell), want)
		}
	}
	if whole.CursorPosition().X != piecemeal.CursorPosition().X {
		t.Errorf("cursor x: whole = %d, piecemeal = %d",
			whole.CursorPosition().X, piecemeal.CursorPosition().X)
	}
}

// TestGraphemeNoMergeAfterCellOverwritten covers an anchor that is no longer
// what was written there. A continuation must not fold into whatever happens to
// occupy the position now.
func TestGraphemeNoMergeAfterCellOverwritten(t *testing.T) {
	t.Parallel()

	e := NewEmulator(10, 2)
	mustWrite(t, e, "\u2764")                         // the anchor
	e.SetCell(0, 0, &uv.Cell{Content: "Z", Width: 1}) // something else took it
	mustWrite(t, e, "\ufe0f")

	if base := e.CellAt(0, 0); base == nil || base.Content != "Z" {
		t.Errorf("base cell = %q, want the overwriting content untouched", cellContent(base))
	}
}

// TestGraphemeNoMergeWhenCursorClamped covers the last column with autowrap
// off, where the cursor stops advancing and stays on top of the cell just
// written. The continuation must not be folded into the cell to its left.
func TestGraphemeNoMergeWhenCursorClamped(t *testing.T) {
	t.Parallel()

	e := NewEmulator(3, 2)
	mustWrite(t, e, "\x1b[?7l") // autowrap off
	mustWrite(t, e, "ab\u2764") // heart in the last column, cursor clamped there
	mustWrite(t, e, "\ufe0f")

	if c := e.CellAt(1, 0); c == nil || c.Content != "b" {
		t.Errorf("cell 1 = %q, want %q -- the mark folded into the wrong cell", cellContent(c), "b")
	}
}

// TestGraphemeMergeAtLineEdge covers a continuation that would widen its base
// past the last column. The base is already placed, so it is left as it is
// rather than reflowed.
func TestGraphemeMergeAtLineEdge(t *testing.T) {
	t.Parallel()

	e := NewEmulator(3, 2)
	mustWrite(t, e, "ab\u2764") // heart lands in the last column, one wide
	mustWrite(t, e, "\ufe0f")   // would make it two wide, which does not fit

	if base := e.CellAt(2, 0); base == nil || base.Content != "\u2764" {
		t.Errorf("base cell = %q, want the unwidened heart", cellContent(base))
	}
	for x, want := range map[int]string{0: "a", 1: "b"} {
		if c := e.CellAt(x, 0); c == nil || c.Content != want {
			t.Errorf("cell %d = %q, want %q", x, cellContent(c), want)
		}
	}
}

func mustWrite(t *testing.T, e *Emulator, s string) {
	t.Helper()
	if _, err := e.Write([]byte(s)); err != nil {
		t.Fatalf("writing %q: %v", s, err)
	}
}

func cellContent(c *uv.Cell) string {
	if c == nil {
		return "<nil>"
	}
	return c.Content
}

func spacerWidth(c *uv.Cell) any {
	if c == nil {
		return "<nil>"
	}
	return c.Width
}

// TestGraphemeNoMergeAfterScreenChange covers an anchor that a sequence has
// invalidated. A cleared or scrolled cell can hold a blank that matches the
// remembered content by coincidence, so the anchor has to be dropped when a
// sequence runs rather than trusted to compare unequal.
func TestGraphemeNoMergeAfterScreenChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		seq  string
	}{
		{name: "erase display", seq: "\x1b[2J"},
		{name: "erase line", seq: "\x1b[2K"},
		{name: "scroll up", seq: "\x1b[S"},
		{name: "alternate screen", seq: "\x1b[?1049h"},
		{name: "full reset", seq: "\x1bc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := NewEmulator(6, 3)
			mustWrite(t, e, "ab ")  // a blank the anchor could be confused with
			mustWrite(t, e, tc.seq) // whatever it did, the anchor is stale now
			mustWrite(t, e, "\u0301")

			// The mark belongs wherever the cursor is, never folded back into
			// the cell the anchor used to name.
			if c := e.CellAt(2, 0); c != nil && c.Content == " \u0301" {
				t.Error("the mark folded into a cell the sequence had already changed")
			}
		})
	}
}
