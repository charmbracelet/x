package vt

import (
	"slices"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestResizeReflowsNarrowAndWide(t *testing.T) {
	e := NewEmulator(12, 4)
	const text = "one two three four"
	_, _ = e.WriteString(text)

	e.Resize(6, 4)
	if got := strings.ReplaceAll(e.String(), "\n", ""); got != text {
		t.Fatalf("content after shrinking = %q, want %q", got, text)
	}

	e.Resize(24, 4)
	if got := strings.TrimRight(e.String(), "\n"); got != text {
		t.Fatalf("content after expanding = %q, want %q", got, text)
	}
}

func TestResizePreservesHardLineBreaks(t *testing.T) {
	e := NewEmulator(8, 4)
	_, _ = e.WriteString("abcdefghij\r\nsecond")

	e.Resize(20, 4)
	if got, want := strings.TrimRight(e.String(), "\n"), "abcdefghij\nsecond"; got != want {
		t.Fatalf("resized screen = %q, want %q", got, want)
	}
}

func TestResizeReflowsScreenAndScrollback(t *testing.T) {
	e := NewEmulator(5, 2)
	const text = "abcdefghijklmnop"
	_, _ = e.WriteString(text)
	if e.ScrollbackLen() == 0 {
		t.Fatal("expected wrapped content in scrollback before resize")
	}

	e.Resize(20, 2)
	if got := strings.TrimRight(e.String(), "\n"); got != text {
		t.Fatalf("expanded screen = %q, want %q", got, text)
	}
	if got := e.ScrollbackLen(); got != 0 {
		t.Fatalf("scrollback length after expanding = %d, want 0", got)
	}
}

func TestResizePreservesCursorAndSavedCursor(t *testing.T) {
	e := NewEmulator(10, 3)
	_, _ = e.WriteString("abcdefgh")
	e.Resize(4, 3)
	_, _ = e.WriteString("Z")
	e.Resize(20, 3)
	if got, want := strings.TrimRight(e.String(), "\n"), "abcdefghZ"; got != want {
		t.Fatalf("content written at resized cursor = %q, want %q", got, want)
	}

	e = NewEmulator(10, 3)
	_, _ = e.WriteString("abcdefg\x1b7h\x1b[H")
	e.Resize(4, 3)
	if got, want := e.scr.saved.Position, uv.Pos(3, 1); got != want {
		t.Fatalf("saved cursor after resize = %v, want %v", got, want)
	}
	_, _ = e.WriteString("\x1b8Z")
	e.Resize(20, 3)
	if got, want := strings.TrimRight(e.String(), "\n"), "abcdefgZ"; got != want {
		t.Fatalf("content written at restored cursor = %q, want %q", got, want)
	}
}

func TestResizePreservesPendingWrap(t *testing.T) {
	t.Run("ASCII", func(t *testing.T) {
		e := NewEmulator(5, 2)
		_, _ = e.WriteString("abcde")
		e.Resize(10, 2)
		_, _ = e.WriteString("f")
		if got, want := strings.TrimRight(e.String(), "\n"), "abcdef"; got != want {
			t.Fatalf("screen = %q, want %q", got, want)
		}
	})

	t.Run("width 2", func(t *testing.T) {
		e := NewEmulator(4, 2)
		_, _ = e.WriteString("ab你")
		e.Resize(8, 2)
		_, _ = e.WriteString("X")
		if got, want := strings.TrimRight(e.String(), "\n"), "ab你X"; got != want {
			t.Fatalf("screen = %q, want %q", got, want)
		}
	})
}

func TestResizePreservesWideAndCombiningCharacters(t *testing.T) {
	t.Run("wide character wraps early", func(t *testing.T) {
		e := NewEmulator(4, 2)
		_, _ = e.WriteString("abc你")
		e.Resize(8, 2)
		if got, want := strings.TrimRight(e.String(), "\n"), "abc你"; got != want {
			t.Fatalf("screen = %q, want %q", got, want)
		}
	})

	t.Run("combining character", func(t *testing.T) {
		e := NewEmulator(4, 3)
		const text = "你e\u0301好"
		_, _ = e.WriteString(text)
		e.Resize(3, 3)
		e.Resize(8, 3)
		if got, want := strings.TrimRight(e.String(), "\n"), text; got != want {
			t.Fatalf("screen = %q, want %q", got, want)
		}
	})
}

func TestResizePreservesBoundarySpaces(t *testing.T) {
	testCases := []string{"abc def", "ab  cd", "abc  def"}
	for _, text := range testCases {
		t.Run(text, func(t *testing.T) {
			e := NewEmulator(4, 3)
			_, _ = e.WriteString(text)
			e.Resize(12, 3)
			if got := logicalContents(e)[0]; got != text {
				t.Fatalf("logical content = %q, want %q", got, text)
			}
		})
	}
}

func TestResizeAfterEditingWrappedRow(t *testing.T) {
	testCases := []struct {
		name string
		edit string
		want string
	}{
		{name: "overwrite", edit: "\x1b[1;4HX", want: "abcX你"},
		{name: "insert", edit: "\x1b[1;3H\x1b[@", want: "ab c你"},
		{name: "delete", edit: "\x1b[1;2H\x1b[P", want: "acX你"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEmulator(4, 3)
			initial := "abc你"
			if tc.name == "delete" {
				initial = "abcX你"
			}
			_, _ = e.WriteString(initial + tc.edit + "\r")
			e.Resize(8, 3)
			if got := logicalContents(e)[0]; got != tc.want {
				t.Fatalf("logical content = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResizeAfterED2PreservesTrailingSpaces(t *testing.T) {
	e := NewEmulator(4, 2)
	_, _ = e.WriteString("abc     \x1b[2J")

	if got, want := logicalContents(e)[0], "abc     "; got != want {
		t.Fatalf("logical content after ED 2 = %q, want %q", got, want)
	}
	e.Resize(8, 4)
	if got, want := logicalContents(e)[0], "abc     "; got != want {
		t.Fatalf("logical content after resize = %q, want %q", got, want)
	}
}

func TestPhysicalScrollbackLimitCanCutLogicalLine(t *testing.T) {
	e := NewEmulator(4, 2)
	e.SetScrollbackSize(2)
	_, _ = e.WriteString("abcdefghijklmnopq")

	if got, want := scrollbackTexts(e.Scrollback()), []string{"efgh", "ijkl"}; !slices.Equal(got, want) {
		t.Fatalf("scrollback rows = %q, want %q", got, want)
	}
	e.Resize(20, 2)
	if got, want := logicalContents(e)[0], "efghijklmnopq"; got != want {
		t.Fatalf("retained logical fragment = %q, want %q", got, want)
	}
}

func logicalContents(e *Emulator) []string {
	lines := make([]uv.Line, 0, e.ScrollbackLen()+e.scr.Height())
	wrapped := make([]bool, 0, cap(lines))
	widths := make([]int, 0, cap(lines))
	if sb := e.Scrollback(); sb != nil {
		lines = append(lines, sb.lines...)
		wrapped = append(wrapped, sb.wrapped...)
		widths = append(widths, sb.wrapWidth...)
	}
	lines = append(lines, e.scr.buf.Lines...)
	wrapped = append(wrapped, e.scr.wrapped...)
	widths = append(widths, e.scr.wrapWidth...)

	var logical []string
	var current strings.Builder
	for y, line := range lines {
		last := lineContentWidth(line)
		if wrapped[y] || widths[y] > 0 {
			last = min(len(line), max(last, widths[y]))
		}
		for x := 0; x < last; x++ {
			cell := &line[x]
			if cell.IsZero() {
				continue
			}
			if cell.Equal(&uv.EmptyCell) {
				current.WriteByte(' ')
			} else {
				current.WriteString(cell.Content)
			}
		}
		if !wrapped[y] {
			logical = append(logical, current.String())
			current.Reset()
		}
	}
	return slices.DeleteFunc(logical, func(line string) bool { return line == "" })
}

func scrollbackTexts(sb *Scrollback) []string {
	lines := make([]string, sb.Len())
	for i := range lines {
		lines[i] = sb.Line(i).String()
	}
	return lines
}
