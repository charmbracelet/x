package ansi

import "testing"

func TestMethod_StringWidth(t *testing.T) {
	tests := []struct {
		name string
		m    Method
		in   string
		want int
	}{
		{"empty string wcwidth", WcWidth, "", 0},
		{"empty string grapheme width", GraphemeWidth, "", 0},
		{"ascii wcwidth", WcWidth, "hello", 5},
		{"ascii grapheme width", GraphemeWidth, "hello", 5},
		{"ansi wcwidth", WcWidth, "\x1b[31mred\x1b[0m", 3},
		{"ansi grapheme width", GraphemeWidth, "\x1b[31mred\x1b[0m", 3},
		{"wide chars wcwidth", WcWidth, "コンニチハ", 10},
		{"wide chars grapheme width", GraphemeWidth, "コンニチハ", 10},
		{"emoji wcwidth", WcWidth, "😀", 2},
		{"emoji grapheme width", GraphemeWidth, "😀", 2},
		// Wcwidth measures per codepoint, so a ZWJ sequence is as wide as the
		// emoji it joins: a white flag, a zero-width VS16 and ZWJ, and a
		// rainbow. Grapheme width measures the cluster as the one glyph a
		// terminal in Unicode core mode draws.
		{"flag emoji wcwidth", WcWidth, "🏳️‍🌈", 3},
		{"flag emoji grapheme width", GraphemeWidth, "🏳️‍🌈", 2},
		// Unicode 15.1 merged Indic conjuncts into single clusters, but a
		// terminal without Unicode core mode still advances once per
		// consonant.
		{"devanagari wcwidth", WcWidth, "नमस्ते", 4},
		{"devanagari grapheme width", GraphemeWidth, "नमस्ते", 3},
		{"zwj family wcwidth", WcWidth, "👨‍👩‍👧‍👦", 8},
		{"zwj family grapheme width", GraphemeWidth, "👨‍👩‍👧‍👦", 2},
		{"skin tone wcwidth", WcWidth, "👍🏽", 4},
		{"skin tone grapheme width", GraphemeWidth, "👍🏽", 2},
		{"combining mark wcwidth", WcWidth, "é", 1},
		{"combining mark grapheme width", GraphemeWidth, "é", 1},
		{"vs16 wcwidth", WcWidth, "⚠️", 1},
		{"vs16 grapheme width", GraphemeWidth, "⚠️", 2},
	}
	for _, tt := range tests {
		if got := tt.m.StringWidth(tt.in); got != tt.want {
			t.Errorf("%s: Method.StringWidth(%q) = %d, want %d", tt.name, tt.in, got, tt.want)
		}
	}
}
