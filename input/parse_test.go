package input

import (
	"image/color"
	"testing"
)

func TestParseSequence_Events(t *testing.T) {
	input := []byte("\x1b\x1b[Ztest\x00\x1b]10;rgb:1234/1234/1234\x07\x1b[27;2;27~\x1b[?1049;2$y")
	want := []Event{
		KeyDownEvent{Sym: KeyTab, Mod: Shift | Alt},
		KeyDownEvent{Rune: 't'},
		KeyDownEvent{Rune: 'e'},
		KeyDownEvent{Rune: 's'},
		KeyDownEvent{Rune: 't'},
		KeyDownEvent{Rune: ' ', Sym: KeySpace, Mod: Ctrl},
		ForegroundColorEvent{color.RGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xff}},
		KeyDownEvent{Sym: KeyEscape, Mod: Shift},
		ReportModeEvent{Mode: 1049, Value: 2},
	}
	for i := 0; len(input) != 0; i++ {
		if i >= len(want) {
			t.Fatalf("reached end of want events")
		}
		n, got := ParseSequence(input)
		if got != want[i] {
			t.Errorf("got %v (%T), want %v (%T)", got, got, want[i], want[i])
		}
		input = input[n:]
	}
}

func BenchmarkParseSequence(b *testing.B) {
	input := []byte("\x1b\x1b[Ztest\x00\x1b]10;1234/1234/1234\x07\x1b[27;2;27~")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseSequence(input)
	}
}
