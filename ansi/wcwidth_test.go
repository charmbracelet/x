package ansi

import (
	"testing"
	"unicode"
)

// TestWcClusterWidth pins the rule that separates the two width models: under
// wcwidth a grapheme cluster is worth the sum of its codepoints, however the
// terminal ends up shaping it.
func TestWcClusterWidth(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wc       int
		grapheme int
	}{
		{"ascii", "a", 1, 1},
		{"combining acute", "é", 1, 1},
		{"devanagari conjunct", "स्ते", 2, 1},
		{"devanagari kssa", "क्ष", 2, 1},
		{"cjk", "世", 2, 2},
		{"emoji", "🌈", 2, 2},
		{"vs16", "⚠️", 1, 2},
		{"vs15", "☹︎", 1, 1},
		{"keycap", "1️⃣", 1, 2},
		{"zwj pair", "👨‍💻", 4, 2},
		{"zwj family", "👨‍👩‍👧‍👦", 8, 2},
		{"zwj flag", "🏳️‍🌈", 3, 2},
		{"skin tone", "👍🏽", 4, 2},
		{"regional indicators", "🇺🇸", 2, 2},
		{"halfwidth voiced mark", "ｶﾞ", 2, 1},
		{"c1 string", "\u009bx", 1, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := wcClusterWidth(tc.in); got != tc.wc {
				t.Errorf("wcClusterWidth(%q) = %d, want %d", tc.in, got, tc.wc)
			}
			if got := wcClusterWidth([]byte(tc.in)); got != tc.wc {
				t.Errorf("wcClusterWidth([]byte(%q)) = %d, want %d", tc.in, got, tc.wc)
			}
			if got := StringWidthWc(tc.in); got != tc.wc {
				t.Errorf("StringWidthWc(%q) = %d, want %d", tc.in, got, tc.wc)
			}
			if got := StringWidth(tc.in); got != tc.grapheme {
				t.Errorf("StringWidth(%q) = %d, want %d", tc.in, got, tc.grapheme)
			}
		})
	}
}

// TestWcRuneWidth covers the classification the Unicode tables are consulted
// for, since [runewidth] alone reports several of these as single width.
func TestWcRuneWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"nul", 0, 0},
		{"bell", 0x07, 0},
		{"del", 0x7f, 0},
		{"c1 pad", 0x80, 0},
		{"c1 csi", 0x9b, 0},
		{"c1 apc", 0x9f, 0},
		{"ascii", 'a', 1},
		{"latin1", 'é', 1},
		{"combining acute", 0x0301, 0},
		{"devanagari virama", 0x094d, 0},
		{"devanagari vowel sign", 0x0947, 0},
		{"hebrew point", 0x05b0, 0},
		{"arabic fatha", 0x064e, 0},
		{"thai vowel", 0x0e34, 0},
		{"zwj", 0x200d, 0},
		{"vs15", 0xfe0e, 0},
		{"vs16", 0xfe0f, 0},
		{"enclosing keycap", 0x20e3, 0},
		{"cjk", '世', 2},
		{"hangul", '한', 2},
		{"halfwidth katakana", 0xff76, 1},
		{"emoji", 0x1f308, 2},
		{"skin tone modifier", 0x1f3fd, 2},
		{"regional indicator", 0x1f1fa, 1},
		{"musical symbol mark", 0x1d167, 0},
		{"variation selector supplement", 0xe0100, 0},
		{"tag character", 0xe0041, 0},
	}

	for _, tc := range tests {
		if got := wcRuneWidth(tc.r); got != tc.want {
			t.Errorf("%s: wcRuneWidth(%U) = %d, want %d", tc.name, tc.r, got, tc.want)
		}
	}
}

// TestWcRuneWidthEastAsian pins that East Asian Ambiguous runes above ASCII
// honor RUNEWIDTH_EASTASIAN: nothing between printable ASCII and the first
// combining mark may short-circuit past runewidth.
func TestWcRuneWidthEastAsian(t *testing.T) {
	wcOptions.EastAsianWidth = true
	defer func() { wcOptions.EastAsianWidth = false }()

	for _, r := range []rune{0xa1, 0xb0, 0xd7} {
		if got := wcRuneWidth(r); got != 2 {
			t.Errorf("wcRuneWidth(%U) = %d, want 2", r, got)
		}
	}
	if got := StringWidthWc("¡°é"); got != 6 {
		t.Errorf("StringWidthWc(%q) = %d, want 6", "¡°é", got)
	}
}

// TestIsZeroWidthHigh cross-checks the fast path against the Unicode tables
// over the whole planes above the BMP. The tables' strided ranges interleave
// when merged, so the lookup must not binary-search them directly.
func TestIsZeroWidthHigh(t *testing.T) {
	for r := rune(0x10000); r <= 0x10FFFF; r++ {
		if want, got := unicode.In(r, zeroWidthTables...), isZeroWidthHigh(r); want != got {
			t.Errorf("isZeroWidthHigh(%U) = %v, want %v", r, got, want)
		}
	}
}

func BenchmarkStringWidthWcUnicode(b *testing.B) {
	for _, s := range []string{"नमस्ते दुनिया", "世界你好", "👨‍👩‍👧‍👦 👨‍💻"} {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = StringWidthWc(s)
			}
		})
	}
}
