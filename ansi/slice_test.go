package ansi

import (
	"testing"
)

func TestSlice(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		start  int
		end    int
		expect string
	}{
		{"empty", "", 0, 0, ""},
		{"simple", "foobar", 0, 3, "foo"},
		{"passthrough", "foobar", 0, 6, "foobar"},
		{"ascii", "hello", 0, 3, "hel"},
		{"emoji", "👋", 0, 2, "👋"},
		{"wideemoji", "🫧", 0, 2, "🫧"},
		{"controlemoji", "\x1b[31mhello 👋abc\x1b[0m", 0, 8, "\x1b[31mhello 👋\x1b[0m"},
		{"osc8", "\x1b]8;;https://charm.sh\x1b\\Charmbracelet 🫧\x1b]8;;\x1b\\", 0, 5, "\x1b]8;;https://charm.sh\x1b\\Charm\x1b]8;;\x1b\\"},
		{"osc8_8bit", "\x9d8;;https://charm.sh\x9cCharmbracelet 🫧\x9d8;;\x9c", 0, 5, "\x9d8;;https://charm.sh\x9cCharm\x9d8;;\x9c"},
		{"noop", "\x1B[7m--", 0, 2, "\x1B[7m--"},
		{"double_width", "\x1B[38;2;249;38;114m你好\x1B[0m", 0, 3, "\x1B[38;2;249;38;114m你 \x1B[0m"},
		{"double_width_rune", "你", 0, 1, " "},
		{"double_width_runes", "你好", 0, 2, "你"},
		{"spaces_only", "    ", 0, 2, "  "},
		{"same_width", "foo", 0, 3, "foo"},
		{"style", "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!", 0, 8, "I really\x1B[38;2;249;38;114m\x1B[0m"},
		{"unicode", "\x1b[35mClaire‘s Boutique\x1b[0m", 0, 8, "\x1b[35mClaire‘s\x1b[0m"},
		{"wide_chars", "こんにちは", 0, 7, "こんに "},
		{"style_wide_chars", "\x1b[35mこんにちは\x1b[m", 0, 7, "\x1b[35mこんに \x1b[m"},
		{"osc8_lf", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nสวัสดีสวัสดี\x1b]8;;\x1b\\", 0, 9, "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nส\x1b]8;;\x1b\\"},
		{"beginning_whitespace", "👋🤭🥳😊👌", 1, 6, " 🤭🥳"},
		{"ending_whitespace", "👋🤭🥳😊👌", 4, 9, "🥳😊 "},
		{"double_whitespace", "👋🤭🥳😊👌", 1, 9, " 🤭🥳😊 "},
		{"width_match", "abc", 0, 5, "abc  "},
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := Slice(c.input, c.start, c.end)
			if result != c.expect {
				t.Errorf("test case %d failed: expected %q, got %q", i+1, c.expect, result)
			}
			originalLen := c.end - c.start
			resultLen := StringWidth(result)
			if originalLen != resultLen {
				t.Errorf("test case %d failed: length does not match, expected %d, got %d", i+1, originalLen, resultLen)
			}
		})
	}
}

func BenchmarkSliceString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			Slice("foo", 1, 2)
		}
	})
}
