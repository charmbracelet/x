package vt

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// A combining mark belongs to the character before it, however that character
// arrived. The printable-ASCII path used to write its cell immediately, which
// left a following mark with nowhere to go: it became a zero-width cell of its
// own, and the next write or erase destroyed it.
func TestCombiningMarkJoinsPrecedingASCII(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  []string
	}{
		{"bare", "bcé", []string{"b", "c", "é"}},
		{"then erase", "bcé\x1b[K", []string{"b", "c", "é"}},
		{"then another character", "bcéx", []string{"b", "c", "é", "x"}},
		{"several in a row", "áb́c", []string{"á", "b́", "c"}},
		{"mark on a wide character", "世́x", []string{"世́", "", "x"}},
		{"emoji with a modifier", "\U0001f44b\U0001f3ffx", []string{"\U0001f44b\U0001f3ff", "", "x"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 10, 1)
			term.WriteString(tc.input)
			for x, want := range tc.want {
				cell := term.CellAt(x, 0)
				if cell == nil {
					t.Fatalf("cell %d: nil", x)
				}
				if cell.Content != want {
					t.Errorf("cell %d: got %q, want %q", x, cell.Content, want)
				}
			}
		})
	}
}

// A character is held back until the next one arrives, so it has to be flushed
// when the write ends or a caller reading the screen would not see it.
func TestLastCharacterIsFlushedWhenTheWriteEnds(t *testing.T) {
	term := newTestTerminal(t, 10, 1)
	term.WriteString("hi")
	for x, want := range []string{"h", "i"} {
		if cell := term.CellAt(x, 0); cell == nil || cell.Content != want {
			t.Errorf("cell %d: got %v, want %q", x, cell, want)
		}
	}
}

// A single shift applies to one character, and the character it applies to may
// be a cluster no charset has a mapping for. Taking the shift only when the
// mapping happens left it set, and it landed on the character after.
func TestSingleShiftDoesNotLeakPastACluster(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  []string
	}{
		// SS2, then a cluster, then a plain character. The cluster has no
		// mapping; the "b" must not be drawn from the shifted set.
		{"cluster then plain", "\x1b*0\x8eáb", []string{"á", "b"}},
		// The same shift with a mappable character, to show it still applies.
		{"plain then plain", "\x1b*0\x8eab", []string{"▒", "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 10, 1)
			term.WriteString(tc.input)
			for x, want := range tc.want {
				if cell := term.CellAt(x, 0); cell == nil || cell.Content != want {
					t.Errorf("cell %d: got %v, want %q", x, cell, want)
				}
			}
		})
	}
}

// A run of combining marks is one cluster that keeps growing. Asking where its
// boundaries are on every character costs the length of the buffer each time,
// which is quadratic: 64KB of marks took six seconds before this was a single
// segmentation at the end of the run.
//
// The budget below is thousands of times the linear cost and a fraction of the
// quadratic one, so it says which of the two is in place without being a
// stopwatch on the machine it runs on.
func TestLongClusterRunStaysLinear(t *testing.T) {
	input := "a" + strings.Repeat("́", 32*1024)

	done := make(chan time.Duration, 1)
	go func() {
		term := NewEmulator(80, 24)
		start := time.Now()
		term.WriteString(input)
		done <- time.Since(start)
	}()

	select {
	case elapsed := <-done:
		t.Logf("%d bytes in %v", len(input), elapsed)
	case <-time.After(10 * time.Second):
		t.Fatalf("writing %d bytes of one cluster did not finish in 10s", len(input))
	}
}

func TestClusterWithoutBaseIsDropped(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
	}{
		{"combining mark after a control", "ab\r\u0301"},
		{"lone ZWJ after a control", "ab\r\u200d"},
		{"variation selector after a control", "ab\r\ufe0f"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 10, 1)
			term.WriteString(tc.input)
			if got := rowText(term, 0); got != "ab" {
				t.Errorf("row 0: got = %q, want = %q", got, "ab")
			}
			if c := term.CellAt(0, 0); c == nil || c.Width != 1 || c.Content != "a" {
				t.Errorf("cell 0: got = %+v, want a", c)
			}
			if x, _ := term.scr.CursorPosition(); x != 0 {
				t.Errorf("cursor x: got = %d, want = 0", x)
			}
		})
	}
}

// A PTY read can end anywhere, including inside a grapheme cluster. A
// cluster split across two Writes lands in one cell, as it does when it
// arrives in one.
func TestClusterAcrossWrites(t *testing.T) {
	for _, tc := range []struct {
		name   string
		chunks []string
		want   []string // cell contents from column 0
	}{
		{"ZWJ sequence cut after a ZWJ", []string{"\U0001F468\u200d", "\U0001F469\u200d\U0001F467x"}, []string{"\U0001F468\u200d\U0001F469\u200d\U0001F467", "", "x"}},
		{"combining mark in the next write", []string{"ae", "\u0301b"}, []string{"a", "e\u0301", "b"}},
		{"rune split across writes", []string{"e\xcc", "\x81b"}, []string{"e\u0301", "b"}},
		{"flag split between its regional indicators", []string{"\U0001F1FA", "\U0001F1F8x"}, []string{"\U0001F1FA\U0001F1F8", "", "x"}},
		{"wide base at the end of the line", []string{"abcdefgh世", "\u0301"}, []string{"a", "b", "c", "d", "e", "f", "g", "h", "世\u0301", ""}},
		{"insert mode shifts once", []string{"xyz\x1b[1;1H\x1b[4he", "\u0301"}, []string{"e\u0301", "x", "y", "z"}},
		{"not joined after a control", []string{"e", "\r", "\u0301"}, []string{"e"}},
		{"not joined after a sequence that does not move the cursor", []string{"ae", "\x1b[1m", "\u0301"}, []string{"a", "e", " "}},
		{"not joined to an ASCII character", []string{"e", "f"}, []string{"e", "f"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 10, 1)
			for _, c := range tc.chunks {
				term.WriteString(c)
			}
			for x, want := range tc.want {
				if c := term.CellAt(x, 0); c == nil || c.Content != want {
					t.Errorf("cell %d: got %+v, want %q", x, c, want)
				}
			}
		})
	}

	t.Run("carried cluster is bounded", func(t *testing.T) {
		term := newTestTerminal(t, 10, 1)
		term.WriteString("e")
		for range 1000 {
			term.WriteString("\u0301")
		}
		if n := len(term.CellAt(0, 0).Content); n > maxCarriedCluster+utf8.UTFMax {
			t.Errorf("cell content: got %d bytes, want at most %d", n, maxCarriedCluster+utf8.UTFMax)
		}
	})
}

func TestPrintingASCIIDoesNotAllocate(t *testing.T) {
	term := newTestTerminal(t, 80, 24)
	line := []byte(strings.Repeat("abcdefgh", 10)[:79] + "\r")
	allocs := testing.AllocsPerRun(100, func() {
		term.Write(line)
	})
	if allocs > 0 {
		t.Errorf("allocations per %d-byte line: got = %v, want = 0", len(line), allocs)
	}
}

func BenchmarkWriteASCII(b *testing.B) {
	term := NewEmulator(80, 24)
	line := []byte(strings.Repeat("abcdefgh", 10)[:79] + "\r\n")
	b.SetBytes(int64(len(line)))
	b.ReportAllocs()
	for b.Loop() {
		term.Write(line)
	}
}
