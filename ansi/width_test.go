package ansi

import (
	"testing"
)

var cases = []struct {
	name     string
	input    string
	stripped string
	width    int
	wcwidth  int
}{
	{"empty", "", "", 0, 0},
	{"ascii", "hello", "hello", 5, 5},
	{"emoji", "👋", "👋", 2, 2},
	{"wideemoji", "🫧", "🫧", 2, 2},
	{"combining", "a\u0300", "à", 1, 1},
	{"control", "\x1b[31mhello\x1b[0m", "hello", 5, 5},
	{"csi8", "\x9b38;5;1mhello\x9bm", "hello", 5, 5},
	{"osc", "\x9d2;charmbracelet: ~/Source/bubbletea\x9c", "", 0, 0},
	{"controlemoji", "\x1b[31m👋\x1b[0m", "👋", 2, 2},
	{"oscwideemoji", "\x1b]2;title👨‍👩‍👦\x07", "", 0, 0},
	{"oscwideemoji", "\x1b[31m👨‍👩‍👦\x1b[m", "👨\u200d👩\u200d👦", 2, 6},
	{"multiemojicsi", "👨‍👩‍👦\x9b38;5;1mhello\x9bm", "👨‍👩‍👦hello", 7, 11},
	{"osc8eastasianlink", "\x9d8;id=1;https://example.com/\x9c打豆豆\x9d8;id=1;\x07", "打豆豆", 6, 6},
	{"dcsarabic", "\x1bP?123$pسلام\x1b\\اهلا", "اهلا", 4, 4},
	{"newline", "hello\nworld", "hello\nworld", 10, 10},
	{"tab", "hello\tworld", "hello\tworld", 10, 10},
	{"controlnewline", "\x1b[31mhello\x1b[0m\nworld", "hello\nworld", 10, 10},
	{"style", "\x1B[38;2;249;38;114mfoo", "foo", 3, 3},
	{"unicode", "\x1b[35m“box”\x1b[0m", "“box”", 5, 5},
	{"just_unicode", "Claire’s Boutique", "Claire’s Boutique", 17, 17},
	{"unclosed_ansi", "Hey, \x1b[7m\n猴", "Hey, \n猴", 7, 7},
	{"double_asian_runes", " 你\x1b[8m好.", " 你好.", 6, 6},
	// Wcwidth measures per codepoint, so the two regional indicators of a
	// flag take a column each and a ZWJ sequence is as wide as the emoji it
	// joins. Grapheme width measures the cluster the terminal draws.
	{"flag", "🇸🇦", "🇸🇦", 2, 2},
	// The halfwidth voiced sound mark is Extend, so it joins the cluster
	// before it, but it's a spacing character that advances the cursor.
	{"half width and ascii", "(ﾟ", "(ﾟ", 1, 2},
}

func TestStrip(t *testing.T) {
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if res := Strip(c.input); res != c.stripped {
				t.Errorf("test case %d (%s) failed:\nexpected %q, got %q", i, c.name, c.stripped, res)
			}
		})
	}
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

func TestWcStringWidth(t *testing.T) {
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if width := StringWidthWc(c.input); width != c.wcwidth {
				t.Errorf("test case %d failed: expected %d, got %d, value %q", i+1, c.wcwidth, width, c.input)
			}
		})
	}
}

func BenchmarkStringWidth(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			StringWidth("foo")
		}
	})
}
