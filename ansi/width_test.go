package ansi

import (
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi/parser"
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
	{"keycap", "1️⃣", "1️⃣", 2, 1},
	{"combining after ascii", "a\u0301b", "a\u0301b", 2, 2},
	{"variation selector after ascii", "a\ufe0fb", "a\ufe0fb", 3, 2},
	{"zwj after ascii", "a\u200d👩", "a\u200d👩", 3, 3},
	{"ascii at end", "界x", "界x", 3, 3},
	{"mixed sgr and ascii", "ab\x1b[31mcd\x1b[0mef", "abcdef", 6, 6},
	{"tabs and controls", "a\tb\x00c\n", "a\tb\x00c\n", 3, 3},
	{"c1 bytes", "\x9b31mA\x9bm", "A", 1, 1},
	{"invalid utf8", "A\xffB", "AB", 2, 2},
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

func stringWidthRef(m Method, s string) int {
	if s == "" {
		return 0
	}

	var (
		pstate = parser.GroundState // initial state
		width  int
	)

	for i := 0; i < len(s); i++ {
		state, action := parser.Table.Transition(pstate, s[i])
		if action == parser.PrintAction || state == parser.Utf8State {
			if m == WcWidth {
				r, size := utf8.DecodeRuneInString(s[i:])
				width += wcRuneWidth(r)
				i += size - 1
				pstate = parser.GroundState
				continue
			}

			cluster, w := FirstGraphemeCluster(s[i:], m)
			width += w

			i += len(cluster) - 1
			pstate = parser.GroundState
			continue
		}

		if action == parser.PrintAction {
			width++
		}

		pstate = state
	}

	return width
}

func FuzzStringWidthFastPath(f *testing.F) {
	for _, seed := range []string{"1️⃣", "a\u0301", "a\ufe0f", "a\u200d👩", "\x1b[31mhello\x1b[0m", "界x", "A\xffB", "\x9b31mA\x9bm"} {
		f.Add([]byte(seed))
	}
	rng := rand.New(rand.NewSource(1))
	parts := []string{"hello", "1️⃣", "a\u0301", "a\ufe0f", "a\u200d👩", "\x1b[31m", "\x1b[0m", "界", "\x9b31m", "\xff"}
	for range 32 {
		var s strings.Builder
		for range 12 {
			s.WriteString(parts[rng.Intn(len(parts))])
		}
		f.Add([]byte(s.String()))
	}
	for range 32 {
		b := make([]byte, rng.Intn(64))
		for i := range b {
			b[i] = byte(rng.Intn(256))
		}
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		s := string(b)
		for _, m := range []Method{GraphemeWidth, WcWidth} {
			if got, want := stringWidth(m, s), stringWidthRef(m, s); got != want {
				t.Fatalf("stringWidth(%v, %q) = %d, want %d", m, s, got, want)
			}
		}
	})
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

var widthBenchResult int

func BenchmarkStringWidthCases(b *testing.B) {
	cases := []struct{ name, input string }{
		{"ASCII900", strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)},
		{"SGRHeavy", strings.Repeat("\x1b[31mhello\x1b[0m", 60)},
		{"CJK", strings.Repeat("世界你好", 60)},
		{"Emoji", strings.Repeat("👨‍👩‍👦", 60)},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				widthBenchResult = StringWidth(tc.input)
			}
		})
	}
}

func BenchmarkStringWidthWcCases(b *testing.B) {
	cases := []struct{ name, input string }{
		{"ASCII900", strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)},
		{"SGRHeavy", strings.Repeat("\x1b[31mhello\x1b[0m", 60)},
		{"CJK", strings.Repeat("世界你好", 60)},
		{"Emoji", strings.Repeat("👨‍👩‍👦", 60)},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				widthBenchResult = StringWidthWc(tc.input)
			}
		})
	}
}
