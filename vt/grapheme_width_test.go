package vt

import (
	"testing"
	"time"
	"unicode/utf8"
)

// technologist is one grapheme cluster of three runes: 🧑, ZWJ and 💻.
const technologist = "\U0001F9D1‍\U0001F4BB"

// perRune splits one rune at a time and gives every rune one cell, the way a
// host terminal that never joins clusters does.
func perRune(s string) (string, int) {
	_, size := utf8.DecodeRuneInString(s)
	return s[:size], 1
}

func TestGraphemeWidthFunc(t *testing.T) {
	t.Run("default keeps grapheme clusters", func(t *testing.T) {
		term := newTestTerminal(t, 20, 2)
		term.WriteString(technologist + "|")
		if cell := term.CellAt(2, 0); cell == nil || cell.Content != "|" {
			t.Errorf("cell 2: got %v, want %q", cell, "|")
		}
	})

	t.Run("custom func splits and sizes the pending runes", func(t *testing.T) {
		term := newTestTerminal(t, 20, 2)
		term.SetGraphemeWidthFunc(perRune)
		term.WriteString(technologist + "|")
		if cell := term.CellAt(3, 0); cell == nil || cell.Content != "|" {
			t.Errorf("cell 3: got %v, want %q", cell, "|")
		}
	})

	t.Run("custom func stores each piece as its own cell", func(t *testing.T) {
		term := newTestTerminal(t, 20, 2)
		term.SetGraphemeWidthFunc(perRune)
		term.WriteString(technologist)
		if cell := term.CellAt(1, 0); cell == nil || cell.Content != "‍" {
			t.Errorf("cell 1: got %v, want ZWJ", cell)
		}
	})

	t.Run("nil restores the default", func(t *testing.T) {
		term := newTestTerminal(t, 20, 2)
		term.SetGraphemeWidthFunc(perRune)
		term.SetGraphemeWidthFunc(nil)
		term.WriteString(technologist + "|")
		if cell := term.CellAt(2, 0); cell == nil || cell.Content != "|" {
			t.Errorf("cell 2: got %v, want %q", cell, "|")
		}
	})

	t.Run("safe emulator applies the func", func(t *testing.T) {
		term := NewSafeEmulator(20, 2)
		term.SetGraphemeWidthFunc(perRune)
		term.WriteString(technologist + "|")
		if cell := term.CellAt(3, 0); cell == nil || cell.Content != "|" {
			t.Errorf("cell 3: got %v, want %q", cell, "|")
		}
	})

	t.Run("custom func sees the non-ASCII runs", func(t *testing.T) {
		var seen []string
		term := newTestTerminal(t, 20, 2)
		term.SetGraphemeWidthFunc(func(s string) (string, int) {
			seen = append(seen, s)
			return perRune(s)
		})
		term.WriteString("ab" + technologist + "cdé\r\n")
		want := []string{
			"b" + technologist, technologist, "‍\U0001F4BB", "\U0001F4BB",
			"dé", "é",
		}
		if len(seen) != len(want) {
			t.Fatalf("func saw %q, want %q", seen, want)
		}
		for i := range want {
			if seen[i] != want[i] {
				t.Errorf("call %d: got %q, want %q", i, seen[i], want[i])
			}
		}
	})

	t.Run("printable ASCII is one cell whatever the func says", func(t *testing.T) {
		term := newTestTerminal(t, 20, 2)
		term.SetGraphemeWidthFunc(func(s string) (string, int) {
			t.Errorf("func called with %q", s)
			return perRune(s)
		})
		term.WriteString("ab")
		if cell := term.CellAt(1, 0); cell == nil || cell.Content != "b" {
			t.Errorf("cell 1: got %v, want %q", cell, "b")
		}
	})

	for _, tc := range []struct {
		name string
		f    GraphemeWidthFunc
	}{
		{"an empty piece still consumes a rune", func(string) (string, int) { return "", 1 }},
		{"an overlong piece still consumes a rune", func(s string) (string, int) { return s + "x", 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := newTestTerminal(t, 20, 2)
			term.SetGraphemeWidthFunc(tc.f)
			done := make(chan struct{})
			go func() {
				defer close(done)
				term.WriteString("éè|")
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("write did not return")
			}
			for x, want := range []string{"é", "è", "|"} {
				if cell := term.CellAt(x, 0); cell == nil || cell.Content != want {
					t.Errorf("cell %d: got %v, want %q", x, cell, want)
				}
			}
		})
	}
}
