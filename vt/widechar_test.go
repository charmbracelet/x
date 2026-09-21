package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// TestWideCharAtLineEnd covers a double-width character meeting the right edge.
// The pending-wrap rule asks whether the cursor reached the last column, which
// is the right question only for a character one column wide: a two-column one
// ending exactly at the edge left the cursor sitting on its second half, so the
// next character overwrote it, and one with a single column left was written
// past the edge and lost instead of moving to the next line.
func TestWideCharAtLineEnd(t *testing.T) {
	t.Parallel()

	const thumbsUp = "\U0001F44D" // two columns

	tests := []struct {
		name  string
		input string
		// want is the screen as rows of cell contents; "" is the blank that
		// trails a wide cell, " " an untouched cell.
		want [][]string
	}{
		{
			name:  "narrow character fills the line, next one wraps",
			input: "abcd" + "e",
			want:  [][]string{{"a", "b", "c", "d"}, {"e", " ", " ", " "}},
		},
		{
			name:  "wide character ends at the last column, next one wraps",
			input: "ab" + thumbsUp + "c",
			want:  [][]string{{"a", "b", thumbsUp, ""}, {"c", " ", " ", " "}},
		},
		{
			name:  "wide character does not fit, so it wraps whole",
			input: "abc" + thumbsUp,
			want:  [][]string{{"a", "b", "c", " "}, {thumbsUp, "", " ", " "}},
		},
		{
			name:  "wide character does not fit, and text follows it",
			input: "abc" + thumbsUp + "d",
			want:  [][]string{{"a", "b", "c", " "}, {thumbsUp, "", "d", " "}},
		},
		{
			name:  "two wide characters filling a line exactly",
			input: thumbsUp + thumbsUp + "x",
			want:  [][]string{{thumbsUp, "", thumbsUp, ""}, {"x", " ", " ", " "}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := NewEmulator(4, 2)
			if _, err := e.Write([]byte(tc.input)); err != nil {
				t.Fatalf("writing: %v", err)
			}

			for y, row := range tc.want {
				for x, want := range row {
					cell := e.CellAt(x, y)
					if cell == nil {
						t.Errorf("cell (%d,%d) is missing", x, y)
						continue
					}
					if cell.Content != want {
						t.Errorf("cell (%d,%d) = %q, want %q", x, y, cell.Content, want)
					}
				}
			}
		})
	}
}

// TestWideCharOnANarrowScreen covers a character too wide for the screen at
// all. No line could ever hold it, so the wrap must not go looking for one:
// moving down first would spend a line before writing a character that will
// not show either way. Nothing here can be displayed on one column; what is
// being pinned is that the attempt costs nothing extra.
func TestWideCharOnANarrowScreen(t *testing.T) {
	t.Parallel()

	e := NewEmulator(1, 3)
	if _, err := e.Write([]byte("\U0001F44D")); err != nil {
		t.Fatalf("writing: %v", err)
	}

	if got := e.CursorPosition(); got.Y != 0 {
		t.Errorf("cursor = %v, want it still on the first row", got)
	}
	if n := e.ScrollbackLen(); n != 0 {
		t.Errorf("scrollback grew to %d lines writing a single character", n)
	}
}

// TestWideCharPendingWrapCursorColumn pins where the cursor is left when a wide
// character fills the line. Pending wrap keeps the cursor on the character
// rather than past it, and for a two-column one that has to be the column it
// reached: everything that reads the cursor back afterwards — a backspace, a
// cursor report, a resize — works from that column, and leaving it on the
// character's first half puts them all one place out.
func TestWideCharPendingWrapCursorColumn(t *testing.T) {
	t.Parallel()

	const thumbsUp = "\U0001F44D"

	t.Run("cursor sits on the last column the character reached", func(t *testing.T) {
		t.Parallel()

		e := NewEmulator(4, 2)
		if _, err := e.Write([]byte("ab" + thumbsUp)); err != nil {
			t.Fatalf("writing: %v", err)
		}
		if got := e.CursorPosition().X; got != 3 {
			t.Errorf("cursor x = %d, want 3 -- the emoji covers columns 2 and 3", got)
		}
	})

	t.Run("backspace steps onto the character, not past it", func(t *testing.T) {
		t.Parallel()

		e := NewEmulator(4, 2)
		if _, err := e.Write([]byte("ab" + thumbsUp + "\bX")); err != nil {
			t.Fatalf("writing: %v", err)
		}
		// Backspace from the emoji's last column lands on the emoji, so X
		// replaces it; "b" is not what is being backed over.
		if c := e.CellAt(1, 0); c == nil || c.Content != "b" {
			t.Errorf("cell 1 = %q, want %q -- the backspace went one column too far", cellAt(c), "b")
		}
	})
}

func cellAt(c *uv.Cell) string {
	if c == nil {
		return "<nil>"
	}
	return c.Content
}
