package cellbuf

import (
	"fmt"
	"testing"
)

var wrapCases = []struct {
	name     string
	input    string
	expected string
	width    int
}{
	{
		name:     "simple",
		input:    "I really \x1B[38;2;249;38;114mlove the\x1B[0m Go language!",
		expected: "I really \x1B[38;2;249;38;114mlove\x1b[m\n\x1B[38;2;249;38;114mthe\x1B[0m Go\nlanguage!",
		width:    14,
	},
	{
		name:     "passthrough",
		input:    "hello world",
		expected: "hello world",
		width:    11,
	},
	{
		name:     "asian",
		input:    "こんにち",
		expected: "こんに\nち",
		width:    7,
	},
	{
		name:     "emoji",
		input:    "😃👰🏻‍♀️🫧",
		expected: "😃\n👰🏻‍♀️\n🫧",
		width:    2,
	},
	{
		name:     "long style",
		input:    "\x1B[38;2;249;38;114ma really long string\x1B[0m",
		expected: "\x1B[38;2;249;38;114ma really\x1b[m\n\x1B[38;2;249;38;114mlong\x1b[m\n\x1B[38;2;249;38;114mstring\x1B[0m",
		width:    10,
	},
	{
		name:     "long style nbsp",
		input:    "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
		expected: "\x1b[38;2;249;38;114ma\x1b[m\n\x1b[38;2;249;38;114mreally\u00a0lon\x1b[m\n\x1b[38;2;249;38;114mg string\x1b[0m",
		width:    10,
	},
	{
		name:     "longer",
		input:    "the quick brown foxxxxxxxxxxxxxxxx jumped over the lazy dog.",
		expected: "the quick brown\nfoxxxxxxxxxxxxxx\nxx jumped over\nthe lazy dog.",
		width:    16,
	},
	{
		name:     "longer asian",
		input:    "猴 猴 猴猴 猴猴猴猴猴猴猴猴猴 猴猴猴 猴猴 猴’ 猴猴 猴.",
		expected: "猴 猴 猴猴\n猴猴猴猴猴猴猴猴\n猴 猴猴猴 猴猴\n猴’ 猴猴 猴.",
		width:    16,
	},
	{
		name:     "long input",
		input:    "Rotated keys for a-good-offensive-cheat-code-incorporated/animal-like-law-on-the-rocks.",
		expected: "Rotated keys for a-good-offensive-cheat-code-incorporated/animal-like-law-\non-the-rocks.",
		width:    76,
	},
	{
		name:     "long input2",
		input:    "Rotated keys for a-good-offensive-cheat-code-incorporated/crypto-line-operating-system.",
		expected: "Rotated keys for a-good-offensive-cheat-code-incorporated/crypto-line-\noperating-system.",
		width:    76,
	},
	{
		name:     "hyphen breakpoint",
		input:    "a-good-offensive-cheat-code",
		expected: "a-good-\noffensive-\ncheat-code",
		width:    10,
	},
	{
		name:     "exact",
		input:    "\x1b[91mfoo\x1b[0m",
		expected: "\x1b[91mfoo\x1b[0m",
		width:    3,
	},
	{
		// XXX: Should we preserve spaces on text wrapping?
		name:     "extra space",
		input:    "foo ",
		expected: "foo",
		width:    3,
	},
	{
		name:     "extra space style",
		input:    "\x1b[mfoo \x1b[m",
		expected: "\x1b[mfoo\x1b[m",
		width:    3,
	},
	{
		name:     "paragraph with styles",
		input:    "Lorem ipsum dolor \x1b[1msit\x1b[m amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. \x1b[31mUt enim\x1b[m ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea \x1b[38;5;200mcommodo consequat\x1b[m. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. \x1b[1;2;33mExcepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\x1b[m",
		expected: "Lorem ipsum dolor \x1b[1msit\x1b[m amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor\nincididunt ut labore et dolore\nmagna aliqua. \x1b[31mUt enim\x1b[m ad minim\nveniam, quis nostrud\nexercitation ullamco laboris\nnisi ut aliquip ex ea \x1b[38;5;200mcommodo\x1b[m\n\x1b[38;5;200mconsequat\x1b[m. Duis aute irure\ndolor in reprehenderit in\nvoluptate velit esse cillum\ndolore eu fugiat nulla\npariatur. \x1b[1;2;33mExcepteur sint\x1b[m\n\x1b[1;2;33moccaecat cupidatat non\x1b[m\n\x1b[1;2;33mproident, sunt in culpa qui\x1b[m\n\x1b[1;2;33mofficia deserunt mollit anim\x1b[m\n\x1b[1;2;33mid est laborum.\x1b[m",
		width:    30,
	},
	{"hyphen break", "foo-bar", "foo-\nbar", 5},
	{"double space", "f  bar foobaz", "f  bar\nfoobaz", 6},
	{"passthrough", "foobar\n ", "foobar\n ", 0},
	{"pass", "foo", "foo", 3},
	{"toolong", "foobarfoo", "foob\narfo\no", 4},
	{"white space", "foo bar foo", "foo\nbar\nfoo", 4},
	{"broken_at_spaces", "foo bars foobars", "foo\nbars\nfoob\nars", 4},
	{"hyphen", "foob-foobar", "foob\n-foo\nbar", 4},
	{"wide_emoji_breakpoint", "foo🫧 foobar", "foo\n🫧\nfoob\nar", 4},
	{"space_breakpoint", "foo --bar", "foo --bar", 9},
	{"simple", "foo bars foobars", "foo\nbars\nfoob\nars", 4},
	{"limit", "foo bar", "foo\nbar", 5},
	{"remove white spaces", "foo    \nb   ar   ", "foo\nb\nar", 4},
	{"white space trail width", "foo\nb\t a\n bar", "foo\nb\t a\n bar", 4},
	{"explicit_line_break", "foo bar foo\n", "foo\nbar\nfoo\n", 4},
	{"explicit_breaks", "\nfoo bar\n\n\nfoo\n", "\nfoo\nbar\n\n\nfoo\n", 4},
	{"example", " This is a list: \n\n\t* foo\n\t* bar\n\n\n\t* foo  \nbar    ", " This\nis a\nlist: \n\n\t* foo\n\t* bar\n\n\n\t* foo\nbar", 6},
	{"style_code_dont_affect_length", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", 7},
	{"style_code_dont_get_wrapped", "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m", "\x1b[38;2;249;38;114m(\x1b[0m\x1b[38;2;248;248;242mjust\x1b[m\n\x1b[38;2;248;248;242manother\x1b[m\n\x1b[38;2;248;248;242mtest\x1b[38;2;249;38;114m)\x1b[0m", 7},
	{"osc8_wrap", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\ สวัสดีสวัสดี\x1b]8;;\x1b\\", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\x1b]8;;\x07\n\x1b]8;;https://example.com\x07สวัสดีสวัสดี\x1b]8;;\x1b\\", 8},
	{"tab", "foo\tbar", "foo\nbar", 3},
	{"wrapped styles example", "", "", 10},
	{
		name:     "punctuation after formatted word with space",
		input:    "\x1b[38;5;203;48;5;236m arm64 \x1b[0m, \x1b[38;5;203;48;5;236m amd64 \x1b[0m, \x1b[38;5;203;48;5;236m i386 \x1b[0m",
		expected: "\x1b[38;5;203;48;5;236m arm64 \x1b[0m,\n\x1b[38;5;203;48;5;236m amd64 \x1b[0m, \x1b[38;5;203;48;5;236m i386 \x1b[0m",
		width:    15,
	},
}

func TestWrap(t *testing.T) {
	for i, tc := range wrapCases {
		t.Run(tc.name, func(t *testing.T) {
			output := Wrap(tc.input, tc.width, "")
			if output != tc.expected {
				t.Errorf("case %d, input:\n%q\nexpected:\n%q\n%s\n\ngot:\n%q\n%s", i+1, tc.input, tc.expected, tc.expected, output, output)
			}
		})
	}
}

func TestWrapWhitespaceBeforeBreakpoint(t *testing.T) {
	for _, tc := range []struct {
		name, input, breakpoints, want string
		width                          int
	}{
		{"hyphen", "ab  -", "", "ab\n-", 2},
		{"custom breakpoint", "a  /", "/", "a\n/", 2},
		{"whitespace fits", "ab -", "", "ab \n-", 3},
		{"breakpoint fits", "a -", "", "a -", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Wrap(tc.input, tc.width, tc.breakpoints); got != tc.want {
				t.Errorf("Wrap(%q, %d, %q) = %q, want %q", tc.input, tc.width, tc.breakpoints, got, tc.want)
			}
		})
	}
}

func ExampleWrap() {
	fmt.Println(Wrap("The quick brown fox jumped over the lazy dog.", 20, ""))
	// Output:
	// The quick brown fox
	// jumped over the lazy
	// dog.
}
