package ansi_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestHardwrapClusterWiderThanLimit covers text whose grapheme cluster is wider
// than the limit — a double-width character wrapped to a single column, say.
// The cluster cannot be split and cannot fit, so it takes a line of its own;
// what it must not do is break a line that is still empty, which puts a blank
// line ahead of it.
func TestHardwrapClusterWiderThanLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		limit int
		want  string
	}{
		{name: "double width, one column", in: "中", limit: 1, want: "中"},
		{name: "two double widths, one column", in: "中文", limit: 1, want: "中\n文"},
		{name: "emoji, one column", in: "\U0001F44D", limit: 1, want: "\U0001F44D"},
		{name: "flag, one column", in: "\U0001F1FA\U0001F1F8", limit: 1, want: "\U0001F1FA\U0001F1F8"},
		{name: "double width then narrow", in: "中a", limit: 1, want: "中\na"},
		// A narrow character first means the line is not empty, so the break
		// before the wide one is a real one.
		{name: "narrow then double width", in: "a中", limit: 1, want: "a\n中"},
		{name: "spaces then a flag", in: "  \U0001F1FA\U0001F1F8", limit: 2, want: "  \n\U0001F1FA\U0001F1F8"},
		{name: "fits exactly", in: "中", limit: 2, want: "中"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ansi.Hardwrap(tc.in, tc.limit, false); got != tc.want {
				t.Errorf("Hardwrap(%q, %d) = %q, want %q", tc.in, tc.limit, got, tc.want)
			}
		})
	}
}

// TestWrapClusterWiderThanLimit is the same for word wrapping, over the cases
// it handles. A cluster wider than the limit that begins a longer word is not
// covered: Wrap("中a", 1) and Wrap("  \U0001F1FA\U0001F1F8", 2) still come out
// as "\n中a" and an unbroken "  \U0001F1FA\U0001F1F8". Those run through the
// word and space buffers rather than the branch below, and reworking that is
// its own change.
func TestWrapClusterWiderThanLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		limit int
		want  string
	}{
		{name: "double width, one column", in: "中", limit: 1, want: "中"},
		{name: "two double widths, one column", in: "中文", limit: 1, want: "中\n文"},
		{name: "emoji, one column", in: "\U0001F44D", limit: 1, want: "\U0001F44D"},
		{name: "narrow then double width", in: "a中", limit: 1, want: "a\n中"},
		{name: "separated by a space", in: "中 文", limit: 1, want: "中\n文"},
		// Suppressing the break must still drop the whitespace waiting to be
		// written, which is what the break itself would have done with it.
		// Left pending, it is flushed onto the next line and pushes it over.
		{name: "leading space before a wide word", in: " éé", limit: 2, want: "éé"},
		{name: "indent before a wide word", in: "  中", limit: 2, want: "中"},
		{name: "indent after a newline", in: "ab\n  中文", limit: 4, want: "ab\n中文"},
		{name: "fits exactly", in: "中", limit: 2, want: "中"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ansi.Wrap(tc.in, tc.limit, ""); got != tc.want {
				t.Errorf("Wrap(%q, %d) = %q, want %q", tc.in, tc.limit, got, tc.want)
			}
		})
	}
}

// TestHardwrapKeepsExplicitBlankLines guards the fix from swallowing blank lines
// the input actually asked for: the suppression is about breaks the wrapper
// adds, not newlines it was given.
func TestHardwrapKeepsExplicitBlankLines(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"a\n\nb", "\n\n", "a\n\n\nb"} {
		if got := ansi.Hardwrap(in, 4, false); got != in {
			t.Errorf("Hardwrap(%q, 4) = %q, want it unchanged", in, got)
		}
	}
}

// FuzzHardwrapWidth asserts the property the fix is really about: a wrapped
// line is never wider than the limit unless it holds one cluster that is
// itself wider, which nothing can help.
//
// Hardwrap only. Wrap does not hold this invariant today and did not before
// this change either: Wrap(" bc", 2, "") returns " bc\n", a three-column line,
// with no wide character involved. Leading whitespace escaping the limit that
// way is its own bug and its own fix.
func FuzzHardwrapWidth(f *testing.F) {
	// Built from whole tokens rather than raw bytes: random bytes are mostly
	// unterminated escape sequences, where a "line" is a fragment of a
	// sequence payload and its width in isolation means nothing.
	tokens := []string{
		// No tab: it contributes no width here but is a column stop to a
		// terminal, which is a separate question from this one.
		"a", "bc", "def ", " ", "  ", "中", "文字",
		"\U0001F44D", "\U0001F44D\U0001F3FB", "\U0001F1FA\U0001F1F8", "é",
		"\x1b[31m", "\x1b[0m", "\x1b]8;;http://x\x07",
	}
	build := func(idx []byte) string {
		var b strings.Builder
		for _, i := range idx {
			b.WriteString(tokens[int(i)%len(tokens)])
		}
		return b.String()
	}

	f.Add([]byte{0, 1, 2, 6, 8}, 4)
	f.Add([]byte{8, 8, 8, 8}, 3)
	f.Add([]byte{10, 10, 10}, 1)
	f.Add([]byte{12, 0, 1, 13}, 5)
	f.Add([]byte{3, 3, 6, 7}, 2)

	f.Fuzz(func(t *testing.T, idx []byte, limit int) {
		if len(idx) == 0 || len(idx) > 24 || limit < 1 || limit > 20 {
			t.Skip()
		}

		s := build(idx)
		out := ansi.Hardwrap(s, limit, false)

		for i, line := range strings.Split(out, "\n") {
			if ansi.StringWidth(line) <= limit {
				continue
			}
			// Allowed only when the line holds a single cluster that is itself
			// too wide; the escape sequences around it carry no width, so they
			// have to come off before asking.
			bare := ansi.Strip(line)
			if cl, _ := ansi.FirstGraphemeCluster(bare, ansi.GraphemeWidth); cl == bare && ansi.StringWidth(cl) > limit {
				continue
			}
			t.Fatalf("line %d of %q is %d wide, over the limit of %d:\n%q",
				i, s, ansi.StringWidth(line), limit, line)
		}
	})
}
