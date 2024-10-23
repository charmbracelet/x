package ansi

import (
	"testing"
)

// nolint
var tcases = []struct {
	name   string
	input  string
	tail   string
	width  int
	expect string
}{
	{"empty", "", "", 0, ""},
	{"equalascii", "one", ".", 3, "one"},
	{"equalemoji", "on👋", ".", 3, "on."},
	{"equalcontrolemoji", "one\x1b[0m", ".", 3, "one\x1b[0m"},
	{"truncate_tail_greater", "foo", "...", 5, "foo"},
	{"simple", "foobar", "", 3, "foo"},
	{"passthrough", "foobar", "", 10, "foobar"},
	{"ascii", "hello", "", 3, "hel"},
	{"emoji", "👋", "", 2, "👋"},
	{"wideemoji", "🫧", "", 2, "🫧"},
	{"controlemoji", "\x1b[31mhello 👋abc\x1b[0m", "", 8, "\x1b[31mhello 👋\x1b[0m"},
	{"osc8", "\x1b]8;;https://charm.sh\x1b\\Charmbracelet 🫧\x1b]8;;\x1b\\", "", 5, "\x1b]8;;https://charm.sh\x1b\\Charm\x1b]8;;\x1b\\"},
	{"osc8_8bit", "\x9d8;;https://charm.sh\x9cCharmbracelet 🫧\x9d8;;\x9c", "", 5, "\x9d8;;https://charm.sh\x9cCharm\x9d8;;\x9c"},
	{"style_tail", "\x1B[38;5;219mHiya!", "…", 3, "\x1B[38;5;219mHi…"},
	{"double_style_tail", "\x1B[38;5;219mHiya!\x1B[38;5;219mHello", "…", 7, "\x1B[38;5;219mHiya!\x1B[38;5;219mH…"},
	{"noop", "\x1B[7m--", "", 2, "\x1B[7m--"},
	{"double_width", "\x1B[38;2;249;38;114m你好\x1B[0m", "", 3, "\x1B[38;2;249;38;114m你\x1B[0m"},
	{"double_width_rune", "你", "", 1, ""},
	{"double_width_runes", "你好", "", 2, "你"},
	{"spaces_only", "    ", "…", 2, " …"},
	{"longer_tail", "foo", "...", 2, ""},
	{"same_tail_width", "foo", "...", 3, "foo"},
	{"same_tail_width_control", "\x1b[31mfoo\x1b[0m", "...", 3, "\x1b[31mfoo\x1b[0m"},
	{"same_width", "foo", "", 3, "foo"},
	{"truncate_with_tail", "foobar", ".", 4, "foo."},
	{"style", "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!", "", 8, "I really\x1B[38;2;249;38;114m\x1B[0m"},
	{"dcs", "\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\foobar", "…", 4, "\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\foo…"},
	{"emoji_tail", "\x1b[36mHello there!\x1b[m", "😃", 8, "\x1b[36mHello 😃\x1b[m"},
	{"unicode", "\x1b[35mClaire‘s Boutique\x1b[0m", "", 8, "\x1b[35mClaire‘s\x1b[0m"},
	{"wide_chars", "こんにちは", "…", 7, "こんに…"},
	{"style_wide_chars", "\x1b[35mこんにちは\x1b[m", "…", 7, "\x1b[35mこんに…\x1b[m"},
	{"osc8_lf", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nสวัสดีสวัสดี\x1b]8;;\x1b\\", "…", 9, "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\n…\x1b]8;;\x1b\\"},
}

func TestTruncate(t *testing.T) {
	for i, c := range tcases {
		t.Run(c.name, func(t *testing.T) {
			if result := Truncate(c.input, c.width, c.tail); result != c.expect {
				t.Errorf("test case %d failed: expected %q, got %q", i+1, c.expect, result)
			}
		})
	}
}

func BenchmarkTruncateString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			Truncate("foo", 2, "")
		}
	})
}
