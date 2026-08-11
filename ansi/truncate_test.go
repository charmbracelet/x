package ansi

import (
	"testing"
)

// nolint
var tcases = []struct {
	name        string
	input       string
	extra       string
	width       int
	expectRight string
	expectLeft  string
}{
	{
		"empty",
		"",
		"",
		0,
		"",
		"",
	},
	{
		"truncate_length_0",
		"foo",
		"",
		0,
		"",
		"foo",
	},
	{
		"equalascii",
		"one",
		".",
		3,
		"one",
		"",
	},
	{
		"equalemoji",
		"on👋",
		".",
		3,
		"on.",
		".👋",
	},
	{
		"simple multiple words",
		"a couple of words",
		"",
		6,
		"a coup",
		"le of words",
	},
	{
		"equalcontrolemoji",
		"one\x1b[0m",
		".",
		3,
		"one\x1b[0m",
		"\x1b[0m",
	},
	{
		"truncate_tail_greater",
		"foo",
		"...",
		5,
		"foo",
		"",
	},
	{
		"simple",
		"foobar",
		"",
		3,
		"foo",
		"bar",
	},
	{
		"passthrough",
		"foobar",
		"",
		10,
		"foobar",
		"",
	},
	{
		"ascii",
		"hello",
		"",
		3,
		"hel",
		"lo",
	},
	{
		"emoji",
		"👋",
		"",
		2,
		"👋",
		"",
	},
	{
		"wideemoji",
		"🫧",
		"",
		2,
		"🫧",
		"",
	},
	{
		"controlemoji",
		"\x1b[31mhello 👋abc\x1b[0m",
		"",
		8,
		"\x1b[31mhello 👋\x1b[0m",
		"\x1b[31mabc\x1b[0m",
	},
	{
		"osc8",
		"\x1b]8;;https://charm.sh\x1b\\Charmbracelet 🫧\x1b]8;;\x1b\\",
		"",
		5,
		"\x1b]8;;https://charm.sh\x1b\\Charm\x1b]8;;\x1b\\",
		"\x1b]8;;https://charm.sh\x1b\\bracelet 🫧\x1b]8;;\x1b\\",
	},
	{
		// The 8-bit OSC start (0x9D) is unaffected by this change and is
		// kept as-is — that's what makes this case still meaningfully
		// "_8bit": it exercises the 8-bit OSC start path. What changed is
		// the terminator: 8-bit C1 ST (0x9C) is no longer recognized
		// because it collides with UTF-8 continuation bytes, so the
		// terminator is now the 7-bit form (ESC \\). Both expected
		// truncated forms mirror that.
		"osc8_8bit",
		"\x9d8;;https://charm.sh\x1b\\Charmbracelet 🫧\x9d8;;\x1b\\",
		"",
		5,
		"\x9d8;;https://charm.sh\x1b\\Charm\x9d8;;\x1b\\",
		"\x9d8;;https://charm.sh\x1b\\bracelet 🫧\x9d8;;\x1b\\",
	},
	{
		"style_tail",
		"\x1B[38;5;219mHiya!",
		"…",
		3,
		"\x1B[38;5;219mHi…",
		"\x1B[38;5;219m…a!",
	},
	{
		"double_style_tail",
		"\x1B[38;5;219mHiya!\x1B[38;5;219mHello",
		"…",
		7,
		"\x1B[38;5;219mHiya!\x1B[38;5;219mH…",
		"\x1B[38;5;219m\x1B[38;5;219m…llo",
	},
	{
		"noop",
		"\x1B[7m--",
		"",
		2,
		"\x1B[7m--",
		"\x1b[7m",
	},
	{
		"double_width",
		"\x1B[38;2;249;38;114m你好\x1B[0m",
		"",
		3,
		"\x1B[38;2;249;38;114m你\x1B[0m",
		"\x1B[38;2;249;38;114m好\x1B[0m",
	},
	{
		"double_width_rune",
		"你",
		"",
		1,
		"",
		"你",
	},
	{
		"double_width_runes",
		"你好",
		"",
		2,
		"你",
		"好",
	},
	{
		"spaces_only",
		"    ",
		"…",
		2,
		" …",
		"…  ",
	},
	{
		"longer_tail",
		"foo",
		"...",
		2,
		"",
		"...o",
	},
	{
		"same_tail_width",
		"foo",
		"...",
		3,
		"foo",
		"",
	},
	{
		"same_tail_width_control",
		"\x1b[31mfoo\x1b[0m",
		"...",
		3,
		"\x1b[31mfoo\x1b[0m",
		"\x1b[31m\x1b[0m",
	},
	{
		"same_width",
		"foo",
		"",
		3,
		"foo",
		"",
	},
	{
		"truncate_with_tail",
		"foobar",
		".",
		4,
		"foo.",
		".ar",
	},
	{
		"style",
		"I really \x1B[38;2;249;38;114mlove\x1B[0m Go!",
		"",
		8,
		"I really\x1B[38;2;249;38;114m\x1B[0m",
		" \x1B[38;2;249;38;114mlove\x1B[0m Go!",
	},
	{
		"dcs",
		"\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\foobar",
		"…",
		4,
		"\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\foo…",
		"\x1BPq#0;2;0;0;0#1;2;100;100;0#2;2;0;100;0#1~~@@vv@@~~@@~~$#2??}}GG}}??}}??-#1!14@\x1B\\…ar",
	},
	{
		"emoji_tail",
		"\x1b[36mHello there!\x1b[m",
		"😃",
		8,
		"\x1b[36mHello 😃\x1b[m",
		"\x1b[36m😃ere!\x1b[m",
	},
	{
		"unicode",
		"\x1b[35mClaire‘s Boutique\x1b[0m",
		"",
		8,
		"\x1b[35mClaire‘s\x1b[0m",
		"\x1b[35m Boutique\x1b[0m",
	},
	{
		"wide_chars",
		"こんにちは",
		"…",
		7,
		"こんに…",
		"…ちは",
	},
	{
		"style_wide_chars",
		"\x1b[35mこんにちは\x1b[m",
		"…",
		7,
		"\x1b[35mこんに…\x1b[m",
		"\x1b[35m…ちは\x1b[m",
	},
	{
		"osc8_lf",
		"สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\nสวัสดีสวัสดี\x1b]8;;\x1b\\",
		"…",
		9,
		"สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\\n…\x1b]8;;\x1b\\",
		"\x1b]8;;https://example.com\x1b\\…วัสดีสวัสดี\x1b]8;;\x1b\\",
	},
	{
		"simple japanese text prefix/suffix",
		"耐許ヱヨカハ調出あゆ監",
		"…",
		13,
		"耐許ヱヨカハ…",
		"…調出あゆ監",
	},
	{
		"simple japanese text",
		"耐許ヱヨカハ調出あゆ監",
		"",
		14,
		"耐許ヱヨカハ調",
		"出あゆ監",
	},
	{
		"new line inside and outside range",
		"\n\nsomething\nin\nthe\nway\n\n",
		"-",
		10,
		"\n\nsomething\n-",
		"-n\nthe\nway\n\n",
	},
	{
		"multi-width graphemes with newlines - japanese text",
		`耐許ヱヨカハ調出あゆ監件び理別よン國給災レホチ権輝モエフ会割もフ響3現エツ文時しだびほ経機ムイメフ敗文ヨク現義なさド請情ゆじょて憶主管州けでふく。排ゃわつげ美刊ヱミ出見ツ南者オ抜豆ハトロネ論索モネニイ任償スヲ話破リヤヨ秒止口イセソス止央のさ食周健でてつだ官送ト読聴遊容ひるべ。際ぐドらづ市居ネムヤ研校35岩6繹ごわク報拐イ革深52球ゃレスご究東スラ衝3間ラ録占たス。

禁にンご忘康ざほぎル騰般ねど事超スんいう真表何カモ自浩ヲシミ図客線るふ静王ぱーま写村月掛焼詐面ぞゃ。昇強ごントほ価保キ族85岡モテ恋困ひりこな刊並せご出来ぼぎむう点目ヲウ止環公ニレ事応タス必書タメムノ当84無信升ちひょ。価ーぐ中客テサ告覧ヨトハ極整
ラ得95稿はかラせ江利ス宏丸霊ミ考整ス静将ず業巨職ノラホ収嗅ざな。`,
		"",
		14,
		"耐許ヱヨカハ調",
		`出あゆ監件び理別よン國給災レホチ権輝モエフ会割もフ響3現エツ文時しだびほ経機ムイメフ敗文ヨク現義なさド請情ゆじょて憶主管州けでふく。排ゃわつげ美刊ヱミ出見ツ南者オ抜豆ハトロネ論索モネニイ任償スヲ話破リヤヨ秒止口イセソス止央のさ食周健でてつだ官送ト読聴遊容ひるべ。際ぐドらづ市居ネムヤ研校35岩6繹ごわク報拐イ革深52球ゃレスご究東スラ衝3間ラ録占たス。

禁にンご忘康ざほぎル騰般ねど事超スんいう真表何カモ自浩ヲシミ図客線るふ静王ぱーま写村月掛焼詐面ぞゃ。昇強ごントほ価保キ族85岡モテ恋困ひりこな刊並せご出来ぼぎむう点目ヲウ止環公ニレ事応タス必書タメムノ当84無信升ちひょ。価ーぐ中客テサ告覧ヨトハ極整
ラ得95稿はかラせ江利ス宏丸霊ミ考整ス静将ず業巨職ノラホ収嗅ざな。`,
	},
}

func TestTruncate(t *testing.T) {
	for i, c := range tcases {
		t.Run(c.name, func(t *testing.T) {
			if result := Truncate(c.input, c.width, c.extra); result != c.expectRight {
				t.Errorf("test case %d failed:\nexpected: %q\n     got: %q", i+1, c.expectRight, result)
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

func TestTruncateLeft(t *testing.T) {
	for i, c := range tcases {
		t.Run(c.name, func(t *testing.T) {
			if result := TruncateLeft(c.input, c.width, c.extra); result != c.expectLeft {
				t.Errorf("test case %d failed:\nexpected: %q\n     got: %q", i+1, c.expectLeft, result)
			}
		})
	}
}

func BenchmarkTruncateLeft(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			TruncateLeft("foo", 2, "")
		}
	})
}

func TestCut(t *testing.T) {
	for i, c := range []struct {
		desc   string
		input  string
		left   int
		right  int
		expect string
	}{
		{
			"simple string",
			"This is a long string", 2, 6,
			"is i",
		},
		{
			"with ansi",
			"I really \x1B[38;2;249;38;114mlove\x1B[0m Go!", 4, 25,
			"ally \x1b[38;2;249;38;114mlove\x1b[0m Go!",
		},
		{
			"left is 0",
			"Foo \x1B[38;2;249;38;114mbar\x1B[0mbaz", 0, 5,
			"Foo \x1B[38;2;249;38;114mb\x1B[0m",
		},
		{
			"right is 0",
			"\x1b[7mHello\x1b[m", 3, 0,
			"",
		},
		{
			"right is less than left",
			"\x1b[7mHello\x1b[m", 3, 2,
			"",
		},
		{
			"cut size is 0",
			"\x1b[7mHello\x1b[m", 2, 2,
			"",
		},
		{
			"maintains open ansi",
			"\x1b[38;5;212;48;5;63mHello, Artichoke!\x1b[m", 7, 16,
			"\x1b[38;5;212;48;5;63mArtichoke\x1b[m",
		},
		{
			"multiline",
			"\n\x1b[38;2;98;98;98m\nif [ -f RE\nADME.md ]; then\x1b[m\n\x1b[38;2;98;98;98m    echo oi\x1b[m\n\x1b[38;2;98;98;98mfi\x1b[m\n", 8, 13,
			"\x1b[38;2;98;98;98mRE\nADM\x1b[m\x1b[38;2;98;98;98m\x1b[m\x1b[38;2;98;98;98m\x1b[m",
		},
	} {
		t.Run(c.desc, func(t *testing.T) {
			got := Cut(c.input, c.left, c.right)
			if got != c.expect {
				t.Errorf("%s (#%d):\nexpected: %q\ngot:      %q", c.desc, i+1, c.expect, got)
			}
		})
	}
}

func TestByteToGraphemeRange(t *testing.T) {
	cases := []struct {
		name   string
		feed   [2]int
		expect [2]int
		input  string
	}{
		{
			name:   "simple",
			input:  "hello world from x/ansi",
			feed:   [2]int{2, 9},
			expect: [2]int{2, 9},
		},
		{
			name:   "with emoji",
			input:  " Downloads",
			feed:   [2]int{4, 7},
			expect: [2]int{2, 5},
		},
		{
			name:   "start out of bounds",
			input:  "some text",
			feed:   [2]int{-1, 5},
			expect: [2]int{0, 5},
		},
		{
			name:   "end out of bounds",
			input:  "some text",
			feed:   [2]int{1, 50},
			expect: [2]int{1, 9},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			charStart, charStop := ByteToGraphemeRange(tt.input, tt.feed[0], tt.feed[1])
			if expect := tt.expect[0]; expect != charStart {
				t.Errorf("expected start to be %d, got %d", expect, charStart)
			}
			if expect := tt.expect[1]; expect != charStop {
				t.Errorf("expected stop to be %d, got %d", expect, charStop)
			}
		})
	}
}
