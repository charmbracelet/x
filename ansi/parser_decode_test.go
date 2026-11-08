package ansi

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi/parser"
)

func TestDecodeSequence(t *testing.T) {
	type expectedSequence struct {
		seq    []byte
		n      int
		width  int
		params []int
		data   []byte
		cmd    int
	}
	cases := []struct {
		name     string
		input    []byte
		expected []expectedSequence
	}{
		{
			name:     "single byte",
			input:    []byte{0x1b},
			expected: []expectedSequence{{seq: []byte{0x1b}, n: 1}},
		},
		{
			name:  "single byte 2",
			input: []byte{0x00},
			expected: []expectedSequence{
				{seq: []byte{0x00}, n: 1},
			},
		},
		{
			name:  "ASCII printable",
			input: []byte("a"),
			expected: []expectedSequence{
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "ASCII space",
			input: []byte(" "),
			expected: []expectedSequence{
				{seq: []byte{' '}, n: 1, width: 1},
			},
		},
		{
			name:  "ASCII DEL",
			input: []byte{DEL},
			expected: []expectedSequence{
				{seq: []byte{DEL}, n: 1},
			},
		},
		{
			name:  "DEL in the middle of UTF8 string",
			input: []byte{'a', DEL, 'b'},
			expected: []expectedSequence{
				{seq: []byte{'a'}, n: 1, width: 1},
				{seq: []byte{DEL}, n: 1},
				{seq: []byte{'b'}, n: 1, width: 1},
			},
		},
		{
			name:  "DEL in the middle of DCS",
			input: []byte("\x1bP1;2+xa\x7fb\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1bP1;2+xa\x7fb\x1b\\"), n: 12, params: []int{1, 2}, data: []byte{'a', DEL, 'b'}, cmd: 'x' | '+'<<16},
			},
		},
		{
			name:  "ST in the middle of DCS",
			input: []byte("\x1bP1;2+xa\x9cb\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1bP1;2+xa\x9c"), n: 9, params: []int{1, 2}, data: []byte{'a'}, cmd: 'x' | '+'<<16},
				{seq: []byte{'b'}, n: 1, width: 1},
				{seq: []byte("\x1b\\"), n: 2, cmd: '\\'},
			},
		},
		{
			name:     "CSI style sequence",
			input:    []byte("\x1b[1;2;3m"),
			expected: []expectedSequence{{seq: []byte("\x1b[1;2;3m"), n: 8, params: []int{1, 2, 3}, cmd: 'm'}},
		},
		{
			name:  "invalid unterminated CSI sequence",
			input: []byte("\x1b[1;2;3"),
			expected: []expectedSequence{
				{seq: []byte("\x1b[1;2;3"), n: 7, params: []int{1, 2}}, // last param gets collected during DispatchAction
			},
		},
		{
			name:     "set title OSC sequence",
			input:    []byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"),
			expected: []expectedSequence{{seq: []byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"), n: 38, cmd: 2, data: []byte("2;charmbracelet: ~/Source/bubbletea")}},
		},
		{
			name:     "set background OSC sequence with 7-bit ST terminator",
			input:    []byte("\x1b]11;ff/00/ff\x1b\\"),
			expected: []expectedSequence{{seq: []byte("\x1b]11;ff/00/ff\x1b\\"), n: 15, cmd: 11, data: []byte("11;ff/00/ff")}},
		},
		{
			name:  "set background OSC sequence with ST 8-bit terminator",
			input: []byte("\x1b]11;ff/00/ff\x9c\x1baa\x8fa"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]11;ff/00/ff\x9c"), n: 14, cmd: 11, data: []byte("11;ff/00/ff")},
				{seq: []byte{ESC, 'a'}, n: 2, cmd: 'a'},
				{seq: []byte{'a'}, n: 1, width: 1},
				{seq: []byte{SS3}, n: 1},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "set background OSC sequence followed by ESC sequence",
			input: []byte("\x1b]11;ff/00/ff\x1b[1;2;3m"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]11;ff/00/ff"), n: 13, cmd: 11, data: []byte("11;ff/00/ff")},
				{seq: []byte("\x1b[1;2;3m"), n: 8, params: []int{1, 2, 3}, cmd: 'm'},
			},
		},
		{
			name:  "set background OSC ESC terminated",
			input: []byte("\x1b]11;ff/00/ff\x1b"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]11;ff/00/ff"), n: 13, cmd: 11, data: []byte("11;ff/00/ff")},
				{seq: []byte{ESC}, n: 1},
			},
		},
		{
			name:  "multiple sequences",
			input: []byte("\x1b[1;2;3m\x1b]2;charmbracelet: ~/Source/bubbletea\x07\x1b]11;ff/00/ff\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1b[1;2;3m"), n: 8, params: []int{1, 2, 3}, cmd: 'm'},
				{seq: []byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"), n: 38, cmd: 2, data: []byte("2;charmbracelet: ~/Source/bubbletea")},
				{seq: []byte("\x1b]11;ff/00/ff\x1b\\"), n: 15, cmd: 11, data: []byte("11;ff/00/ff")},
			},
		},
		{
			name:  "double ESC",
			input: []byte("\x1b\x1b"),
			expected: []expectedSequence{
				{seq: []byte{0x1b}, n: 1},
				{seq: []byte{0x1b}, n: 1},
			},
		},
		{
			name:  "double ST",
			input: []byte("\x1b\\\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte{ESC, '\\'}, n: 2, cmd: '\\'},
				{seq: []byte{ESC, '\\'}, n: 2, cmd: '\\'},
			},
		},
		{
			name:  "double ST 8-bit",
			input: []byte("\x9c\x9c"),
			expected: []expectedSequence{
				{seq: []byte{ST}, n: 1},
				{seq: []byte{ST}, n: 1},
			},
		},
		{
			name:  "ASCII printables",
			input: []byte("Hello, World!"),
			expected: []expectedSequence{
				{seq: []byte{'H'}, width: 1, n: 1},
				{seq: []byte{'e'}, width: 1, n: 1},
				{seq: []byte{'l'}, width: 1, n: 1},
				{seq: []byte{'l'}, width: 1, n: 1},
				{seq: []byte{'o'}, width: 1, n: 1},
				{seq: []byte{','}, width: 1, n: 1},
				{seq: []byte{' '}, width: 1, n: 1},
				{seq: []byte{'W'}, width: 1, n: 1},
				{seq: []byte{'o'}, width: 1, n: 1},
				{seq: []byte{'r'}, width: 1, n: 1},
				{seq: []byte{'l'}, width: 1, n: 1},
				{seq: []byte{'d'}, width: 1, n: 1},
				{seq: []byte{'!'}, width: 1, n: 1},
			},
		},
		{
			name:  "rune",
			input: []byte("👋"),
			expected: []expectedSequence{
				{seq: []byte("👋"), n: 4, width: 2},
			},
		},
		{
			name:  "inavlid rune",
			input: []byte{0xc3},
			expected: []expectedSequence{
				{seq: []byte{0xc3}, n: 1, width: 1},
			},
		},
		{
			name:  "multiple sequences with UTF8 and double ESC",
			input: []byte("👨🏿‍🌾\x1b\x1b \x1b[?1:2:3mÄabc\x1b\x1bP+q\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("👨🏿‍🌾"), n: 15, width: 2},
				{seq: []byte{ESC}, n: 1},
				{seq: []byte{ESC, ' '}, n: 2, cmd: 0 | ' '<<16},
				{seq: []byte("\x1b[?1:2:3m"), n: 9, params: []int{1 | parser.HasMoreFlag, 2 | parser.HasMoreFlag, 3}, cmd: 'm' | '?'<<8},
				{seq: []byte("Ä"), n: 2, width: 1},
				{seq: []byte{'a'}, n: 1, width: 1},
				{seq: []byte{'b'}, n: 1, width: 1},
				{seq: []byte{'c'}, n: 1, width: 1},
				{seq: []byte{ESC}, n: 1},
				{seq: []byte("\x1bP+q\x1b\\"), n: 6, cmd: 'q' | '+'<<16},
			},
		},
		{
			name:  "style sequences",
			input: []byte("hello, \x1b[1;2;3mworld\x1b[0m!"),
			expected: []expectedSequence{
				{seq: []byte("h"), n: 1, width: 1},
				{seq: []byte("e"), n: 1, width: 1},
				{seq: []byte("l"), n: 1, width: 1},
				{seq: []byte("l"), n: 1, width: 1},
				{seq: []byte("o"), n: 1, width: 1},
				{seq: []byte(","), n: 1, width: 1},
				{seq: []byte(" "), n: 1, width: 1},
				{seq: []byte("\x1b[1;2;3m"), n: 8, params: []int{1, 2, 3}, cmd: 'm'},
				{seq: []byte("w"), n: 1, width: 1},
				{seq: []byte("o"), n: 1, width: 1},
				{seq: []byte("r"), n: 1, width: 1},
				{seq: []byte("l"), n: 1, width: 1},
				{seq: []byte("d"), n: 1, width: 1},
				{seq: []byte("\x1b[0m"), n: 4, params: []int{0}, cmd: 'm'},
				{seq: []byte("!"), n: 1, width: 1},
			},
		},
		{
			name:  "set background OSC with C1",
			input: []byte("\x1b]11;\x90?\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]11;\x90?\x1b\\"), n: 9, cmd: 11, data: []byte("11;\x90?")},
			},
		},
		{
			name:  "unterminated CSI with escape sequence",
			input: []byte("\x1b[1;2;3\x1bOa"),
			expected: []expectedSequence{
				{seq: []byte("\x1b[1;2;3"), n: 7, params: []int{1, 2, 3}}, // params get reset and ignored when unterminated
				{seq: []byte("\x1bO"), n: 2, cmd: 'O'},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "SS3",
			input: []byte("\x1bOa"),
			expected: []expectedSequence{
				{seq: []byte("\x1bO"), n: 2, cmd: 'O'},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "SS3 8-bit",
			input: []byte("\x8fa"),
			expected: []expectedSequence{
				{seq: []byte{SS3}, n: 1},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "ESC sequence with intermediate",
			input: []byte("\x1b Q"),
			expected: []expectedSequence{
				{seq: []byte("\x1b Q"), n: 3, cmd: 'Q' | ' '<<16},
			},
		},
		{
			name:  "ESC followed by C0",
			input: []byte("\x1b[\x00a"),
			expected: []expectedSequence{
				{seq: []byte("\x1b["), n: 2},
				{seq: []byte{0x00}, n: 1},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "unterminated DCS sequence",
			input: []byte("\x1bP1;2+xa"),
			expected: []expectedSequence{
				{seq: []byte("\x1bP1;2+xa"), n: 8, params: []int{1, 2}, data: []byte{'a'}, cmd: 'x' | '+'<<16},
			},
		},
		{
			name:  "invalid DCS sequence",
			input: []byte("\x1bP\x1b\\ab"),
			expected: []expectedSequence{
				{seq: []byte("\x1bP"), n: 2},
				{seq: []byte("\x1b\\"), n: 2, cmd: '\\'},
				{seq: []byte{'a'}, n: 1, width: 1},
				{seq: []byte{'b'}, n: 1, width: 1},
			},
		},
		{
			name:  "single param osc",
			input: []byte("\x1b]112\x07"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]112\x07"), n: 6, cmd: 112, data: []byte("112")},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser()
			p.SetParamsSize(32)
			p.SetDataSize(1024)

			var state byte
			input := tc.input
			results := make([]expectedSequence, 0)
			for len(input) > 0 {
				seq, width, n, newState := DecodeSequence(input, state, p)
				params := slices.Clone(p.params[:p.paramsLen])
				data := slices.Clone(p.data[:p.dataLen])
				results = append(results, expectedSequence{seq: seq, width: width, n: n, params: params, data: data, cmd: p.cmd})
				state = newState
				input = input[n:]
			}
			if len(results) != len(tc.expected) {
				t.Fatalf("expected %d sequences, got %d\n\n%#v\n\n%#v", len(tc.expected), len(results), tc.expected, results)
			}
			for i, r := range results {
				if r.n != tc.expected[i].n {
					t.Errorf("expected %d bytes, got %d", tc.expected[i].n, r.n)
				}
				if r.width != tc.expected[i].width {
					t.Errorf("expected %d width, got %d", tc.expected[i].width, r.width)
				}
				if string(r.seq) != string(tc.expected[i].seq) {
					t.Errorf("expected %q, got %q", string(tc.expected[i].seq), string(r.seq))
				}
				if r.cmd != tc.expected[i].cmd {
					t.Errorf("expected %d cmd, got %d", tc.expected[i].cmd, r.cmd)
				}
				if string(r.data) != string(tc.expected[i].data) {
					t.Errorf("expected %q data, got %q", string(tc.expected[i].data), string(r.data))
				}
				if len(r.params) != len(tc.expected[i].params) {
					t.Fatalf("expected %d params, got %d", len(tc.expected[i].params), len(r.params))
				}
				if len(tc.expected[i].params) > 0 {
					for j, p := range r.params {
						if p != tc.expected[i].params[j] {
							t.Errorf("expected param[%d] = %d, got %d", j, tc.expected[i].params[j], p)
						}
					}
				}
			}
		})
	}
}

// TestDecodeSequenceTooManyParams ensures decoding a sequence with more
// parameters than the parser buffer holds does not panic. Excess parameters
// are consumed but not stored, matching the streaming [Parser] behavior:
// the first parameters are preserved and [Parser.Params] never returns
// more parameters than the buffer holds. See #1037.
func TestDecodeSequenceTooManyParams(t *testing.T) {
	sgr := func(n int) string {
		return "\x1b[" + strings.Repeat("1;", n-1) + "1m"
	}
	ones := func(n int) []int {
		params := make([]int, n)
		for i := range params {
			params[i] = 1
		}
		return params
	}
	decodeAll := func(t *testing.T, p *Parser, input string) (cmd int, params []int, data []byte) {
		t.Helper()
		var state byte
		for len(input) > 0 {
			var n int
			_, _, n, state = DecodeSequence(input, state, p)
			if n == 0 {
				t.Fatalf("decoder made no progress on %q", input)
			}
			input = input[n:]
		}
		params = make([]int, len(p.Params()))
		for i, v := range p.Params() {
			params[i] = int(v)
		}
		return p.Command(), params, p.Data()
	}

	t.Run("csi params exceed buffer", func(t *testing.T) {
		for _, n := range []int{33, 100} {
			p := NewParser()
			cmd, params, _ := decodeAll(t, p, sgr(n))
			if byte(cmd) != 'm' {
				t.Errorf("n=%d: expected cmd 'm', got %q", n, byte(cmd))
			}
			if len(params) != parser.MaxParamsSize {
				t.Fatalf("n=%d: expected %d params, got %d", n, parser.MaxParamsSize, len(params))
			}
			if !slices.Equal(params, ones(parser.MaxParamsSize)) {
				t.Errorf("n=%d: expected first params preserved, got %v", n, params)
			}
		}
	})

	t.Run("csi params at buffer boundary", func(t *testing.T) {
		// In-capacity behavior is unchanged.
		p := NewParser()
		_, params, _ := decodeAll(t, p, sgr(parser.MaxParamsSize))
		if !slices.Equal(params, ones(parser.MaxParamsSize-1)) {
			t.Errorf("expected %d params, got %v", parser.MaxParamsSize-1, params)
		}
	})

	t.Run("separators exceed buffer", func(t *testing.T) {
		p := NewParser()
		_, params, _ := decodeAll(t, p, "\x1b["+strings.Repeat(";", 40)+"m")
		if len(params) != parser.MaxParamsSize {
			t.Fatalf("expected %d params, got %d", parser.MaxParamsSize, len(params))
		}
		for i, v := range params {
			if v != parser.MissingParam {
				t.Errorf("expected param[%d] to be missing, got %d", i, v)
			}
		}
	})

	t.Run("sub-parameters exceed buffer", func(t *testing.T) {
		p := NewParser()
		_, params, _ := decodeAll(t, p, "\x1b["+strings.Repeat("1:", 40)+"1m")
		if len(params) != parser.MaxParamsSize {
			t.Fatalf("expected %d params, got %d", parser.MaxParamsSize, len(params))
		}
		for i, v := range params {
			if !Param(v).HasMore() || Param(v).Param(0) != 1 {
				t.Errorf("expected param[%d] to be 1 with sub-parameters, got %d", i, v)
			}
		}
	})

	t.Run("dcs params exceed buffer", func(t *testing.T) {
		p := NewParser()
		cmd, params, data := decodeAll(t, p, "\x1bP"+strings.Repeat("1;", 32)+"1qpayload\x1b\\")
		if byte(cmd) != 'q' {
			t.Errorf("expected cmd 'q', got %q", byte(cmd))
		}
		if len(params) != parser.MaxParamsSize {
			t.Errorf("expected %d params, got %d", parser.MaxParamsSize, len(params))
		}
		if string(data) != "payload" {
			t.Errorf("expected data %q, got %q", "payload", data)
		}
	})

	t.Run("text around overflowing sequence", func(t *testing.T) {
		p := NewParser()
		input := "a" + sgr(33) + "b"
		var state byte
		var got []int
		for len(input) > 0 {
			var n int
			_, _, n, state = DecodeSequence(input, state, p)
			if n == 0 {
				t.Fatal("decoder made no progress")
			}
			input = input[n:]
			if byte(p.Command()) == 'm' {
				// Params reflect the last decoded sequence; snapshot
				// them before the trailing text resets the parser.
				got = make([]int, len(p.Params()))
				for i, v := range p.Params() {
					got[i] = int(v)
				}
			}
		}
		if len(got) != parser.MaxParamsSize {
			t.Errorf("expected %d params, got %d", parser.MaxParamsSize, len(got))
		}
	})

	t.Run("sequence split across calls", func(t *testing.T) {
		p := NewParser()
		input := sgr(33)
		half := len(input) / 2
		_, _, n, state := DecodeSequence(input[:half], 0, p)
		if n != half {
			t.Fatalf("expected %d bytes consumed, got %d", half, n)
		}
		rest := input[half:]
		for len(rest) > 0 {
			var m int
			_, _, m, state = DecodeSequence(rest, state, p)
			if m == 0 {
				t.Fatal("decoder made no progress")
			}
			rest = rest[m:]
		}
		if got := len(p.Params()); got != parser.MaxParamsSize {
			t.Errorf("expected %d params, got %d", parser.MaxParamsSize, got)
		}
	})

	t.Run("small params buffer", func(t *testing.T) {
		p := NewParser()
		p.SetParamsSize(2)
		_, params, _ := decodeAll(t, p, "\x1b[1;2;3;4;5m")
		if !slices.Equal(params, []int{1, 2}) {
			t.Errorf("expected params [1 2], got %v", params)
		}
	})

	t.Run("zero params buffer", func(t *testing.T) {
		p := NewParser()
		p.SetParamsSize(0)
		var state byte
		input := "\x1b[1;2m"
		for len(input) > 0 {
			var n int
			_, _, n, state = DecodeSequence(input, state, p)
			if n == 0 {
				t.Fatal("decoder made no progress")
			}
			input = input[n:]
		}
		if byte(p.Command()) != 'm' {
			t.Errorf("expected cmd 'm', got %q", byte(p.Command()))
		}
	})
}

func FuzzDecodeSequence(f *testing.F) {
	var b byte
	for b < 0x80 {
		f.Add([]byte{b})
		b++
	}

	f.Add([]byte("\x1b"))
	f.Add([]byte("\x1b[1;2;3m"))
	f.Add([]byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"))
	f.Add([]byte("\x1b]11;ff/00/ff\x1b\\"))
	f.Add([]byte("\x1b]11;ff/00/ff\x9c\x1baa\x8fa"))
	f.Add([]byte("\x1b]11;ff/00/ff\x1b[1;2;3m"))
	f.Add([]byte("\x1b]11;ff/00/ff\x1b"))
	f.Add([]byte("\x1b[1;2;3m\x1b]2;charmbracelet: ~/Source/bubbletea\x07\x1b]11;ff/00/ff\x1b\\"))
	f.Add([]byte("Hello, World!"))
	f.Add([]byte("👋a"))
	f.Add([]byte("👨🏿‍🌾"))
	f.Fuzz(func(t *testing.T, b []byte) {
		var state byte
		var n int
		for len(b) > 0 {
			_, _, n, state = DecodeSequence(b, state, nil)
			if n == 0 {
				break
			}
			b = b[n:]
		}
	})
}

func BenchmarkDecodeSequence(b *testing.B) {
	var state byte
	var n int
	input := []byte("\x1b[1;2;3màbc\x90?123;456+q\x9c\x7f ")
	p := NewParser()
	p.SetParamsSize(32)
	p.SetDataSize(1024)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		in := input
		for len(in) > 0 {
			_, _, n, state = DecodeSequence(in, state, p)
			in = in[n:]
		}
	}
}

func BenchmarkDecodeParser(b *testing.B) {
	p := NewParser()
	p.SetParamsSize(32)
	p.SetDataSize(1024)
	input := []byte("\x1b[1;2;3màbc\x90?123;456+q\x9c\x7f ")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p.Parse(input)
	}
}

func TestCommand(t *testing.T) {
	cases := []struct {
		name     string
		cmd      byte
		prefix   byte
		inter    byte
		expected int
	}{
		{
			name:     "CUU", // Cursor Up
			cmd:      'A',
			expected: 'A',
		},
		{
			name:     "DECAWM", // Auto Wrap Mode
			cmd:      'h',
			prefix:   '?',
			expected: 'h' | '?'<<parser.PrefixShift,
		},
		{
			name:     "DECSCUSR", // Set Cursor Style
			cmd:      'q',
			inter:    ' ',
			expected: 'q' | ' '<<parser.IntermedShift,
		},
		{
			name:     "imaginary cmd with both prefix and intermed",
			cmd:      'x',
			prefix:   '>',
			inter:    '(',
			expected: 'x' | '>'<<parser.PrefixShift | '('<<parser.IntermedShift,
		},
		// {
		// 	name:     "ignore bytes beyond the lower 8 bites",
		// 	cmd:      256 + 'x',
		// 	prefix:     256 + '>',
		// 	inter:    256 + '(',
		// 	expected: 'x' | '>'<<parser.PrefixShift | '('<<parser.IntermedShift,
		// },
		{
			name:     "OSC11", // Set background color
			cmd:      11,
			expected: 11,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if cmd := Command(tc.prefix, tc.inter, tc.cmd); cmd != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, cmd)
			}
		})
	}
}

func TestParameter(t *testing.T) {
	cases := []struct {
		name     string
		param    int
		hasMore  bool
		expected int
	}{
		{
			name:     "single param",
			param:    1,
			expected: 1,
		},
		{
			name:     "single param with hasMore",
			param:    1,
			hasMore:  true,
			expected: 1 | parser.HasMoreFlag,
		},
		{
			name:     "negative param",
			param:    -1,
			expected: parser.ParamMask,
		},
		{
			name:     "negative param has more",
			param:    -1,
			hasMore:  true,
			expected: -1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if p := Parameter(tc.param, tc.hasMore); p != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, p)
			}
		})
	}
}
