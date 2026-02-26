package ansi_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
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
	{"tab", "foo\tbar", 3, "foo\n\tbar", true},
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
	{"column", "VERTICAL", 1, "V\nE\nR\nT\nI\nC\nA\nL", false},
}

func TestHardwrap(t *testing.T) {
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ansi.Hardwrap(tt.input, tt.limit, tt.preserveSpace); got != tt.expected {
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
	input := "the quick brown foxxxxxxxxxxxxxxxx jumped over the lazy dog."
	limit := 16
	output := ansi.Wrap(input, limit, "")
	if output != "the quick brown\nfoxxxxxxxxxxxxxx\nxx jumped over\nthe lazy dog." {
		t.Errorf("expected %q, got %q", "the quick brown\nfoxxxxxxxxxxxxxx\nxx jumped over\nthe lazy dog.", output)
	}
}

var wrapCases = []struct {
	name        string
	input       string
	expected    string
	width       int
	breakpoints string
}{
	{
		name:     "simple",
		input:    "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!",
		expected: "I really\n\x1B[38;2;249;38;114mlove\x1B[0m Go!",
		width:    8,
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
		expected: "\x1B[38;2;249;38;114ma really\nlong\nstring\x1B[0m",
		width:    10,
	},
	{
		name:     "long style nbsp",
		input:    "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
		expected: "\x1b[38;2;249;38;114ma\nreally\u00a0lon\ng string\x1b[0m",
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
		input:    "\x1b[91mfoo\x1b[0",
		expected: "\x1b[91mfoo\x1b[0",
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
		expected: "Lorem ipsum dolor \x1b[1msit\x1b[m amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor\nincididunt ut labore et dolore\nmagna aliqua. \x1b[31mUt enim\x1b[m ad minim\nveniam, quis nostrud\nexercitation ullamco laboris\nnisi ut aliquip ex ea \x1b[38;5;200mcommodo\nconsequat\x1b[m. Duis aute irure\ndolor in reprehenderit in\nvoluptate velit esse cillum\ndolore eu fugiat nulla\npariatur. \x1b[1;2;33mExcepteur sint\noccaecat cupidatat non\nproident, sunt in culpa qui\nofficia deserunt mollit anim\nid est laborum.\x1b[m",
		width:    30,
	},
	{
		// \u202f and \u205f - Single width spaces
		// \u3000 - Double width space
		name:  "Multi Byte spaces",
		input: "A\u202fB\u202fC\u202fDA\u205f\u205fB\u205fC\u205fDA\u3000B\u3000C\u3000D",
		expected: "" +
			"A\u202fB\u202fC\n" +
			"DA\u205f\u205fB\u205fC\n" +
			"DA\u3000B\n" +
			"C\u3000D",
		width: 7,
	},
	{"hyphen break", "foo-bar", "foo-\nbar", 5, ""},
	{"double space", "f  bar foobaz", "f  bar\nfoobaz", 6, ""},
	{"passthrough", "foobar\n ", "foobar\n ", 0, ""},
	{"pass", "foo", "foo", 3, ""},
	{"toolong", "foobarfoo", "foob\narfo\no", 4, ""},
	{"white space", "foo bar foo", "foo\nbar\nfoo", 4, ""},
	{"broken_at_spaces", "foo bars foobars", "foo\nbars\nfoob\nars", 4, ""},
	{"hyphen", "foob-foobar", "foob\n-foo\nbar", 4, ""},
	{"wide_emoji_breakpoint", "foo🫧 foobar", "foo\n🫧\nfoob\nar", 4, ""},
	{"space_breakpoint", "foo --bar", "foo --bar", 9, ""},
	{"simple", "foo bars foobars", "foo\nbars\nfoob\nars", 4, ""},
	{"limit", "foo bar", "foo\nbar", 5, ""},
	{"remove white spaces", "foo    \nb   ar   ", "foo\nb\nar", 4, ""},
	{"white space trail width", "foo\nb\t a\n bar", "foo\nb\t a\n bar", 4, ""},
	{"explicit_line_break", "foo bar foo\n", "foo\nbar\nfoo\n", 4, ""},
	{"explicit_breaks", "\nfoo bar\n\n\nfoo\n", "\nfoo\nbar\n\n\nfoo\n", 4, ""},
	{"example", " This is a list: \n\n\t* foo\n\t* bar\n\n\n\t* foo  \nbar    ", " This\nis a\nlist: \n\n\t* foo\n\t* bar\n\n\n\t* foo\nbar", 6, ""},
	{"style_code_dont_affect_length", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", "\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m", 7, ""},
	{"style_code_dont_get_wrapped", "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m", "\x1b[38;2;249;38;114m(\x1b[0m\x1b[38;2;248;248;242mjust\nanother\ntest\x1b[38;2;249;38;114m)\x1b[0m", 7, ""},
	{"osc8_wrap", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\ สวัสดีสวัสดี\x1b]8;;\x1b\\", "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nสวัสดีสวัสดี\x1b]8;;\x1b\\", 8, ""},
	{"tab", "foo\tbar", "foo\nbar", 3, ""},
	{"Narrow NBSP", "0\u202f1\u202f2\u202f3\u202f4", "0\u202f1\u202f2\u202f3\n4", 7, ""},
	// Paragraph Separator usually takes one character wid, "th
	// while printing it on terminal, but ansi considers this zero widt, "h.
	{"Paragraph Separator", "0\u20291\u20292\u20293\u20294", "0\u20291\u20292\u20293\u20294", 7, ""},
	{"Medium Mathematical Space", "0\u205f1\u205f2\u205f3\u205f4", "0\u205f1\u205f2\u205f3\n4", 7, ""},
	{"Ideagraphic space", "0\u30001\u30002\u30003\u3000", "0\u30001\u30002\n3\u3000", 7, ""},
	{
		name: "japanese with white spaces narrow",
		input: `耐許ヱヨカハ調出あゆ監件び理別よン國給災レホチ権輝モエフ会割もフ響3現エツ文時しだびほ経機ムイメフ敗文ヨク現義なさド請情ゆじょて憶主管州けでふく。排ゃわつげ美刊ヱミ出見ツ南者オ抜豆ハトロネ論索モネニイ任償スヲ話破リヤヨ秒止口イセソス止央のさ食周健でてつだ官送ト読聴遊容ひるべ。際ぐドらづ市居ネムヤ研校35岩6繹ごわク報拐イ革深52球ゃレスご究東スラ衝3間ラ録占たス。
禁にンご忘康ざほぎル騰般ねど事超スんいう真表何カモ自浩ヲシミ図客線るふ静王ぱーま写村月掛焼詐面ぞゃ。昇強ごントほ価保キ族85岡モテ恋困ひりこな刊並せご出来ぼぎむう点目ヲウ止環公ニレ事応タス必書タメムノ当84無信升ちひょ。価ーぐ中客テサ告覧ヨトハ極整ラ得95稿はかラせ江利ス宏丸霊ミ考整ス静将ず業巨職ノラホ収嗅ざな。`,
		expected: `耐許ヱヨカハ
調出あゆ監件
び理別よン國
給災レホチ権
輝モエフ会割
もフ響3現エツ
文時しだびほ
経機ムイメフ
敗文ヨク現義
なさド請情ゆ
じょて憶主管
州けでふく。
排ゃわつげ美
刊ヱミ出見ツ
南者オ抜豆ハ
トロネ論索モ
ネニイ任償ス
ヲ話破リヤヨ
秒止口イセソ
ス止央のさ食
周健でてつだ
官送ト読聴遊
容ひるべ。際
ぐドらづ市居
ネムヤ研校35
岩6繹ごわク報
拐イ革深52球
ゃレスご究東
スラ衝3間ラ録
占たス。
禁にンご忘康
ざほぎル騰般
ねど事超スん
いう真表何カ
モ自浩ヲシミ
図客線るふ静
王ぱーま写村
月掛焼詐面ぞ
ゃ。昇強ごン
トほ価保キ族8
5岡モテ恋困ひ
りこな刊並せ
ご出来ぼぎむ
う点目ヲウ止
環公ニレ事応
タス必書タメ
ムノ当84無信
升ちひょ。価
ーぐ中客テサ
告覧ヨトハ極
整ラ得95稿は
かラせ江利ス
宏丸霊ミ考整
ス静将ず業巨
職ノラホ収嗅
ざな。`,
		width: 13,
	},
	{
		name: "japanese with white spaces wide",
		input: `耐許ヱヨカハ調出あゆ監件び理別よン國給災レホチ権輝モエフ会割もフ響3現エツ文時しだびほ経機ムイメフ敗文ヨク現義なさド請情ゆじょて憶主管州けでふく。排ゃわつげ美刊ヱミ出見ツ南者オ抜豆ハトロネ論索モネニイ任償スヲ話破リヤヨ秒止口イセソス止央のさ食周健でてつだ官送ト読聴遊容ひるべ。際ぐドらづ市居ネムヤ研校35岩6繹ごわク報拐イ革深52球ゃレスご究東スラ衝3間ラ録占たス。
禁にンご忘康ざほぎル騰般ねど事超スんいう真表何カモ自浩ヲシミ図客線るふ静王ぱーま写村月掛焼詐面ぞゃ。昇強ごントほ価保キ族85岡モテ恋困ひりこな刊並せご出来ぼぎむう点目ヲウ止環公ニレ事応タス必書タメムノ当84無信升ちひょ。価ーぐ中客テサ告覧ヨトハ極整ラ得95稿はかラせ江利ス宏丸霊ミ考整ス静将ず業巨職ノラホ収嗅ざな。`,
		expected: `耐許ヱヨカハ調出あゆ監件び理別
よン國給災レホチ権輝モエフ会割
もフ響3現エツ文時しだびほ経機
ムイメフ敗文ヨク現義なさド請情
ゆじょて憶主管州けでふく。排ゃ
わつげ美刊ヱミ出見ツ南者オ抜豆
ハトロネ論索モネニイ任償スヲ話
破リヤヨ秒止口イセソス止央のさ
食周健でてつだ官送ト読聴遊容ひ
るべ。際ぐドらづ市居ネムヤ研校
35岩6繹ごわク報拐イ革深52球ゃ
レスご究東スラ衝3間ラ録占たス
。
禁にンご忘康ざほぎル騰般ねど事
超スんいう真表何カモ自浩ヲシミ
図客線るふ静王ぱーま写村月掛焼
詐面ぞゃ。昇強ごントほ価保キ族
85岡モテ恋困ひりこな刊並せご出
来ぼぎむう点目ヲウ止環公ニレ事
応タス必書タメムノ当84無信升ち
ひょ。価ーぐ中客テサ告覧ヨトハ
極整ラ得95稿はかラせ江利ス宏丸
霊ミ考整ス静将ず業巨職ノラホ収
嗅ざな。`,
		width: 30,
	},
}

func TestWrap(t *testing.T) {
	format := "case %d, input:\n%s\n\nexpected:\n%s\n\ngot:\n%s\n\n"
	for i, tc := range wrapCases {
		t.Run(tc.name, func(t *testing.T) {
			format := format
			if strings.ContainsFunc(tc.input, func(r rune) bool {
				return (r < 32 && r != '\n' && r != '\t') || // C0 control characters except LF and TAB
					r == 127 || // DEL
					(r >= 0x80 && r <= 0x9F) // C1 control characters
			}) {
				format = strings.ReplaceAll(format, "%s", "%q")
			}
			output := ansi.Wrap(tc.input, tc.width, tc.breakpoints)
			if output != tc.expected {
				t.Errorf(format, i+1, tc.input, tc.expected, output)
			}
		})
	}
}

func BenchmarkWrap(b *testing.B) {
	input := "the quick brown foxxxxxxxxxxxxxxxx jumped over the lazy dog."
	limit := 16
	for i := 0; i < b.N; i++ {
		_ = ansi.Wrap(input, limit, "")
	}
}
