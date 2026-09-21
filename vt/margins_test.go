package vt

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestMarginsOutsideThePage covers margins set past the edge of the screen.
// DECSTBM and DECSLRM take their bounds straight from the sequence, and a
// program emitting one larger than the page used to leave a scroll region
// pointing outside the buffer; the next scroll or line insertion then indexed
// off the end and took the host process down with it. The bytes come from
// whatever is running in the terminal, so this has to be survivable.
func TestMarginsOutsideThePage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		width, height int
		input         string
	}{
		{
			name:  "DECSTBM bottom past the last row, then scroll up",
			width: 4, height: 2, input: "\x1b[1;9r\x1b[S",
		},
		{
			name:  "DECSTBM bottom past the last row, then scroll down",
			width: 4, height: 2, input: "\x1b[1;9r\x1b[T",
		},
		{
			name:  "DECSTBM top past the last row",
			width: 4, height: 2, input: "\x1b[8;9r\x1b[S",
		},
		{
			name:  "DECSLRM right past the last column, then scroll",
			width: 2, height: 3, input: "\x1b[?69h\x1b[2;4s\x1b[S",
		},
		{
			name:  "DECSLRM right past the last column, then insert line",
			width: 2, height: 3, input: "\x1b[?69h\x1b[1;5s\x1b[L",
		},
		{
			name:  "DECSLRM right past the last column, then delete line",
			width: 2, height: 3, input: "\x1b[?69h\x1b[1;5s\x1b[M",
		},
		{
			name:  "DECSLRM left past the last column",
			width: 2, height: 3, input: "\x1b[?69h\x1b[7;9s\x1b[S",
		},
		{
			name:  "DECSTBM entirely past the page",
			width: 6, height: 3, input: "\x1b[9;10r\n\n\n\n",
		},
		{
			name:  "DECSLRM entirely past the page, in origin mode",
			width: 3, height: 3, input: "\x1b[?69h\x1b[?6h\x1b[7;9sabc",
		},
		{
			name:  "both margins past the page, with text and a line feed",
			width: 3, height: 2, input: "\x1b[?69h\x1b[1;9s\x1b[1;9rabc\ndef\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := NewEmulator(tc.width, tc.height)
			// A panic here is the bug; the write must simply survive.
			if _, err := e.Write([]byte(tc.input)); err != nil {
				t.Fatalf("writing: %v", err)
			}

			// Whatever the sequence asked for, the scroll region has to stay
			// inside the buffer or the next operation on it indexes off the
			// end, and it has to keep at least one row and column: an empty
			// region parked past the last line is in bounds but nothing can
			// scroll within it, which garbles everything printed afterwards.
			scroll := e.scr.scroll
			if scroll.Min.X < 0 || scroll.Max.X > tc.width || scroll.Min.X >= scroll.Max.X {
				t.Errorf("horizontal scroll region [%d,%d) is not a region inside a %d-column screen",
					scroll.Min.X, scroll.Max.X, tc.width)
			}
			if scroll.Min.Y < 0 || scroll.Max.Y > tc.height || scroll.Min.Y >= scroll.Max.Y {
				t.Errorf("vertical scroll region [%d,%d) is not a region inside a %d-row screen",
					scroll.Min.Y, scroll.Max.Y, tc.height)
			}
		})
	}
}

// TestCountedSequencesAreBounded covers sequences that take a repeat or erase
// count. The count comes from the byte stream, so it can name far more cells
// than the screen has, and the work used to be proportional to the number
// rather than to the screen: a handful of bytes could hold the emulator for
// seconds, which for a program rendering someone else's output is enough to
// wedge it.
func TestCountedSequencesAreBounded(t *testing.T) {
	t.Parallel()

	// Large enough that spending the count is unmistakable next to the budget,
	// while still parsing as a single parameter.
	const huge = "888888889"

	tests := []struct {
		name  string
		input string
	}{
		{name: "ECH erase character", input: "\x1b[" + huge + "X"},
		{name: "REP repeat character", input: "a\x1b[" + huge + "b"},
		{name: "CHT forward tabs", input: "\x1b[" + huge + "I"},
		{name: "CBT backward tabs", input: "\x1b[" + huge + "Z"},
		{name: "ICH insert characters", input: "\x1b[" + huge + "@"},
		{name: "DCH delete characters", input: "\x1b[" + huge + "P"},
		{name: "IL insert lines", input: "\x1b[" + huge + "L"},
		{name: "DL delete lines", input: "\x1b[" + huge + "M"},
		{name: "SU scroll up", input: "\x1b[" + huge + "S"},
		{name: "SD scroll down", input: "\x1b[" + huge + "T"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := NewEmulator(4, 3)
			// The error is read after the wait rather than reported from the
			// goroutine: on timeout the subtest ends while the write is still
			// running, and reporting from it then takes down the whole binary.
			var writeErr error
			done := make(chan struct{})
			go func() {
				defer close(done)
				_, writeErr = e.Write([]byte(tc.input))
			}()

			select {
			case <-done:
				if writeErr != nil {
					t.Errorf("writing: %v", writeErr)
				}
			case <-time.After(500 * time.Millisecond):
				// Still enormous headroom: every one of these takes
				// microseconds once the work is bounded by the screen rather
				// than by the count, and seconds when it is not.
				t.Fatal("the sequence spent its count instead of stopping at the screen")
			}
		})
	}
}

// TestRepeatCountMatchesTheWholeCount pins that capping REP is invisible. The
// count is bounded so a huge one cannot be spent, but the screen and the cursor
// have to end up where writing every repeat would have left them: past a full
// screen the result repeats every Width characters, so keeping that remainder
// is what makes the cap unobservable. Compared against actually writing the
// characters, which is the behaviour REP stands in for.
func TestRepeatCountMatchesTheWholeCount(t *testing.T) {
	t.Parallel()

	const width, height = 4, 3

	// A count of zero is left out: this emulator repeats nothing for it,
	// where ECMA-48 makes an omitted or zero parameter mean one. That is a
	// separate question from bounding a large one.
	for _, n := range []int{1, 3, 4, 11, 12, 13, 200, 1001} {
		repeated := NewEmulator(width, height)
		if _, err := repeated.Write([]byte(fmt.Sprintf("a\x1b[%db", n))); err != nil {
			t.Fatalf("n=%d: writing the repeat: %v", n, err)
		}

		written := NewEmulator(width, height)
		if _, err := written.Write([]byte(strings.Repeat("a", n+1))); err != nil {
			t.Fatalf("n=%d: writing the characters: %v", n, err)
		}

		if got, want := repeated.CursorPosition(), written.CursorPosition(); got != want {
			t.Errorf("n=%d: cursor = %v, want %v", n, got, want)
		}
		for y := range height {
			for x := range width {
				got, want := repeated.CellAt(x, y), written.CellAt(x, y)
				if got == nil || want == nil {
					continue
				}
				if got.Content != want.Content {
					t.Errorf("n=%d: cell (%d,%d) = %q, want %q", n, x, y, got.Content, want.Content)
				}
			}
		}
	}
}
