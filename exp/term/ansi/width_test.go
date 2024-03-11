package ansi

import (
	"testing"
)

var cases = []struct {
	name  string
	input string
	width int
}{
	{"empty", "", 0},
	{"ascii", "hello", 5},
	{"emoji", "👋", 2},
	{"wideemoji", "🫧", 2},
	{"combining", "a\u0300", 1},
	{"control", "\x1b[31mhello\x1b[0m", 5},
	{"csi8", "\x9b38;5;1mhello\x9bm", 5},
	{"osc", "\x9d2;charmbracelet: ~/Source/bubbletea\x9c", 0},
	{"controlemoji", "\x1b[31m👋\x1b[0m", 2},
	{"oscwideemoji", "\x1b]2;title👨‍👩‍👦\x07", 0},
	{"oscwideemoji", "\x1b[31m👨‍👩‍👦\x1b[m", 2},
	{"multiemojicsi", "👨‍👩‍👦\x9b38;5;1mhello\x9bm", 7},
	{"osc8eastasianlink", "\x9d8;id=1;https://example.com/\x9c打豆豆\x9d8;id=1;\x07", 6},
	{"dcsarabic", "\x1bP?123$pسلام\x1b\\اهلا", 4},
}

func TestStringWidth(t *testing.T) {
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if width := StringWidth(c.input); width != c.width {
				t.Errorf("test case %d failed: expected %d, got %d", i+1, c.width, width)
			}
		})
	}
}

func BenchmarkStringWidth(b *testing.B) {
	for _, c := range cases {
		for i := 0; i < b.N; i++ {
			StringWidth(c.input)
		}
	}
}
