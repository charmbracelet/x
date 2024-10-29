package ansi

import (
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
				{seq: []byte("\x1b[1;2;3"), n: 7, params: []int{1, 2}}, // params get reset and ignored when unterminated
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
			p := NewParser(32, 1024)

			var state byte
			input := tc.input
			results := make([]expectedSequence, 0)
			for len(input) > 0 {
				seq, width, n, newState := DecodeSequence(input, state, p)
				params := append([]int(nil), p.Params[:p.ParamsLen]...)
				data := append([]byte(nil), p.Data[:p.DataLen]...)
				results = append(results, expectedSequence{seq: seq, width: width, n: n, params: params, data: data, cmd: p.Cmd})
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
					t.Errorf("expected %d params, got %d", len(tc.expected[i].params), len(r.params))
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
	p := NewParser(32, 1024)
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
	p := NewParser(32, 1024)
	input := []byte("\x1b[1;2;3màbc\x90?123;456+q\x9c\x7f ")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p.Parse(func(s Sequence) {
		}, input)
	}
}
