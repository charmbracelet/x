package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi"
)

var cases = []struct {
	name          string
	input         string
	limit         int
	expected      string
	preserveSpace bool
}{
	{"empty string", "", 0, "", true},
	{"passthrough", "foobar\n ", 0, "foobar\n ", true},
	{"pass", "foo", 4, "foo", true},
	{"simple", "foobarfoo", 4, "foob\narfo\no", true},
	{"lf", "f\no\nobar", 3, "f\no\noba\nr", true},
	{"lf_space", "foo bar\n  baz", 3, "foo\n ba\nr\n  b\naz", true},
	{"tab", "foo\tbar", 3, "foo\n\tba\nr", true},
	{"unicode_space", "foo\xc2\xa0bar", 3, "foo\nbar", false},
	{"style_nochange", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", 7, "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", true},
	{"style", "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m", 3, "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mju\nst \nano\nthe\nr t\nest\x1B[38;2;249;38;114m\n)\x1B[0m", true},
	{"style_lf", "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!", 8, "I really\n\x1b[38;2;249;38;114mlove\x1b[0m Go!", false},
	{"style_emoji", "I really \x1B[38;2;249;38;114mlove u🫧\x1B[0m", 8, "I really\n\x1b[38;2;249;38;114mlove u🫧\x1b[0m", false},
	{"hyperlink", "I really \x1B]8;;https://example.com/\x1B\\love\x1B]8;;\x1B\\ Go!", 10, "I really \x1b]8;;https://example.com/\x1b\\l\nove\x1b]8;;\x1b\\ Go!", false},
	{"dcs", "\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\foobar", 3, "\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\foo\nbar", false},
	{"begin_with_space", " foo", 4, " foo", false},
	{"style_dont_affect_wrap", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", 7, "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", false},
	{"preserve_style", "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m", 3, "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mju\nst \nano\nthe\nr t\nest\x1B[38;2;249;38;114m\n)\x1B[0m", false},
	{"emoji", "foo🫧foobar", 4, "foo\n🫧fo\nobar", false},
	{"osc8_wrap", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\สวัสดีสวัสดี\x1b]8;;\x1b\\", 8, "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nสวัสดีสวัสดี\x1b]8;;\x1b\\", false},
}

func TestWrap(t *testing.T) {
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ansi.Wrap(tt.input, tt.limit, tt.preserveSpace); got != tt.expected {
				t.Errorf("case %d, expected %q, got %q", i+1, tt.expected, got)
			}
		})
	}
}

var wwCases = []struct {
	name        string
	input       string
	limit       int
	breakPoints string
	expected    string
}{
	{"empty string", "", 0, "", ""},
	{"passthrough", "foobar\n ", 0, "", "foobar\n "},
	{"pass", "foo", 3, "", "foo"},
	{"toolong", "foobarfoo", 4, "", "foobarfoo"},
	{"white space", "foo bar foo", 4, "", "foo\nbar\nfoo"},
	{"broken_at_spaces", "foo bars foobars", 4, "", "foo\nbars\nfoobars"},
	{"hyphen", "foo-foobar", 4, "-", "foo-\nfoobar"},
	{"emoji_breakpoint", "foo😃 foobar", 4, "😃", "foo😃\nfoobar"},
	{"wide_emoji_breakpoint", "foo🫧 foobar", 4, "🫧", "foo🫧\nfoobar"},
	{"space_breakpoint", "foo --bar", 9, "-", "foo --bar"},
	{"simple", "foo bars foobars", 4, "", "foo\nbars\nfoobars"},
	{"limit", "foo bar", 5, "", "foo\nbar"},
	{"remove white spaces", "foo    \nb   ar   ", 4, "", "foo\nb\nar"},
	{"white space trail width", "foo\nb\t a\n bar", 4, "", "foo\nb\t a\n bar"},
	{"explicit_line_break", "foo bar foo\n", 4, "", "foo\nbar\nfoo\n"},
	{"explicit_breaks", "\nfoo bar\n\n\nfoo\n", 4, "", "\nfoo\nbar\n\n\nfoo\n"},
	{"example", " This is a list: \n\n\t* foo\n\t* bar\n\n\n\t* foo  \nbar    ", 6, "", " This\nis a\nlist: \n\n\t* foo\n\t* bar\n\n\n\t* foo\nbar"},
	{"style_code_dont_affect_length", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", 7, "", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m"},
	{"style_code_dont_get_wrapped", "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m", 3, "", "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust\nanother\ntest\x1B[38;2;249;38;114m)\x1B[0m"},
	{"osc8_wrap", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\ สวัสดีสวัสดี\x1b]8;;\x1b\\", 8, "", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nสวัสดีสวัสดี\x1b]8;;\x1b\\"},
}

func TestWordwrap(t *testing.T) {
	for i, tt := range wwCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ansi.Wordwrap(tt.input, tt.limit, tt.breakPoints); got != tt.expected {
				t.Errorf("case %d, expected %q, got %q", i+1, tt.expected, got)
			}
		})
	}
}

func TestWrapWordwrap(t *testing.T) {
	t.Skip("WIP")
	input := "the quick brown foxxxxxxxxxxxxxxxx jumped over the lazy dog."
	limit := 16
	output := ansi.Wordwrap(input, limit, "")
	t.Logf("output: %q", output)
	output = ansi.Wrap(output, limit, false)
	if output != "the quick brown\nfoxxxxxxxxxxxxx\nxxxx jumped over\nthe lazy dog." {
		t.Errorf("expected %q, got %q", "the quick brown\nfoxxxxxxxxxxxxxx\nxx jumped over\nthe lazy dog.", output)
	}
}

const _ = `
 the quick brown
 foxxxxxxxxxxxxxxxx
 jumped over the
 lazy dog.
`

const _ = `
 the quick brown
 foxxxxxxxxxxxxxx
 xx jumped over t
 he lazy dog.
`
