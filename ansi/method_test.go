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
		{"flag emoji wcwidth", WcWidth, "🏳️‍🌈", 1},
		{"flag emoji grapheme width", GraphemeWidth, "🏳️‍🌈", 2},
	}
	for _, tt := range tests {
		if got := tt.m.StringWidth(tt.in); got != tt.want {
			t.Errorf("%s: Method.StringWidth(%q) = %d, want %d", tt.name, tt.in, got, tt.want)
		}
	}
}
