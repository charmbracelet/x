package ansi

import (
	"testing"
)

func TestDecodeSequence(t *testing.T) {
	type expectedSequence struct {
		seq   []byte
		n     int
		width int
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
			name:  "ascii printable",
			input: []byte("a"),
			expected: []expectedSequence{
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "ascii space",
			input: []byte(" "),
			expected: []expectedSequence{
				{seq: []byte{' '}, n: 1, width: 1},
			},
		},
		{
			name:  "ascii del",
			input: []byte{DEL},
			expected: []expectedSequence{
				{seq: []byte{DEL}, n: 1},
			},
		},
		{
			name:  "del in the middle of utf8 string",
			input: []byte{'a', DEL, 'b'},
			expected: []expectedSequence{
				{seq: []byte{'a'}, n: 1, width: 1},
				{seq: []byte{DEL}, n: 1},
				{seq: []byte{'b'}, n: 1, width: 1},
			},
		},
		{
			name:  "del in the middle of dcs",
			input: []byte("\x1bP1;2+xa\x7fb\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1bP1;2+xa\x7fb\x1b\\"), n: 12},
			},
		},
		{
			name:  "st in the middle of dcs",
			input: []byte("\x1bP1;2+xa\x9cb\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1bP1;2+xa\x9c"), n: 9},
				{seq: []byte{'b'}, n: 1, width: 1},
				{seq: []byte("\x1b\\"), n: 2},
			},
		},
		{
			name:     "csi",
			input:    []byte("\x1b[1;2;3m"),
			expected: []expectedSequence{{seq: []byte("\x1b[1;2;3m"), n: 8}},
		},
		{
			name:  "csi not terminated",
			input: []byte("\x1b[1;2;3"),
			expected: []expectedSequence{
				{seq: []byte("\x1b[1;2;3"), n: 7},
			},
		},
		{
			name:     "osc",
			input:    []byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"),
			expected: []expectedSequence{{seq: []byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"), n: 38}},
		},
		{
			name:     "osc st terminated",
			input:    []byte("\x1b]11;ff/00/ff\x1b\\"),
			expected: []expectedSequence{{seq: []byte("\x1b]11;ff/00/ff\x1b\\"), n: 15}},
		},
		{
			name:     "osc st 8-bit terminated",
			input:    []byte("\x1b]11;ff/00/ff\x9c\x1baa\x8fa"),
			expected: []expectedSequence{{seq: []byte("\x1b]11;ff/00/ff\x9c"), n: 14}, {seq: []byte{ESC, 'a'}, n: 2}, {seq: []byte{'a'}, n: 1, width: 1}, {seq: []byte{SS3}, n: 1}, {seq: []byte{'a'}, n: 1, width: 1}},
		},
		{
			name:  "osc followed by esc sequence",
			input: []byte("\x1b]11;ff/00/ff\x1b[1;2;3m"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]11;ff/00/ff"), n: 13},
				{seq: []byte("\x1b[1;2;3m"), n: 8},
			},
		},
		{
			name:     "osc esc terminated",
			input:    []byte("\x1b]11;ff/00/ff\x1b"),
			expected: []expectedSequence{{seq: []byte("\x1b]11;ff/00/ff"), n: 13}, {seq: []byte{ESC}, n: 1}},
		},
		{
			name:  "multiple sequences",
			input: []byte("\x1b[1;2;3m\x1b]2;charmbracelet: ~/Source/bubbletea\x07\x1b]11;ff/00/ff\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1b[1;2;3m"), n: 8},
				{seq: []byte("\x1b]2;charmbracelet: ~/Source/bubbletea\x07"), n: 38},
				{seq: []byte("\x1b]11;ff/00/ff\x1b\\"), n: 15},
			},
		},
		{
			name:  "double esc",
			input: []byte("\x1b\x1b"),
			expected: []expectedSequence{
				{seq: []byte{0x1b}, n: 1},
				{seq: []byte{0x1b}, n: 1},
			},
		},
		{
			name:  "double st",
			input: []byte("\x1b\\\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte{ESC, '\\'}, n: 2},
				{seq: []byte{ESC, '\\'}, n: 2},
			},
		},
		{
			name:  "double st 8-bit",
			input: []byte("\x9c\x9c"),
			expected: []expectedSequence{
				{seq: []byte{ST}, n: 1},
				{seq: []byte{ST}, n: 1},
			},
		},
		{
			name:  "ascii printables",
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
			name:  "multiple sequences with utf8 and double esc",
			input: []byte("👨🏿‍🌾\x1b\x1b \x1b[?1:2:3mÄabc\x1b\x1bP+q\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("👨🏿‍🌾"), n: 15, width: 2},
				{seq: []byte{ESC}, n: 1},
				{seq: []byte{ESC, ' '}, n: 2},
				{seq: []byte("\x1b[?1:2:3m"), n: 9},
				{seq: []byte("Ä"), n: 2, width: 1},
				{seq: []byte{'a'}, n: 1, width: 1},
				{seq: []byte{'b'}, n: 1, width: 1},
				{seq: []byte{'c'}, n: 1, width: 1},
				{seq: []byte{ESC}, n: 1},
				{seq: []byte("\x1bP+q\x1b\\"), n: 6},
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
				{seq: []byte("\x1b[1;2;3m"), n: 8},
				{seq: []byte("w"), n: 1, width: 1},
				{seq: []byte("o"), n: 1, width: 1},
				{seq: []byte("r"), n: 1, width: 1},
				{seq: []byte("l"), n: 1, width: 1},
				{seq: []byte("d"), n: 1, width: 1},
				{seq: []byte("\x1b[0m"), n: 4},
				{seq: []byte("!"), n: 1, width: 1},
			},
		},
		{
			name:  "osc with c1",
			input: []byte("\x1b]11;\x90?\x1b\\"),
			expected: []expectedSequence{
				{seq: []byte("\x1b]11;\x90?\x1b\\"), n: 9},
			},
		},
		{
			name:  "unterminated csi with escape sequence",
			input: []byte("\x1b[1;2;3\x1bOa"),
			expected: []expectedSequence{
				{seq: []byte("\x1b[1;2;3"), n: 7},
				{seq: []byte("\x1bO"), n: 2},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "ss3",
			input: []byte("\x1bOa"),
			expected: []expectedSequence{
				{seq: []byte("\x1bO"), n: 2},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "ss3 8-bit",
			input: []byte("\x8fa"),
			expected: []expectedSequence{
				{seq: []byte{SS3}, n: 1},
				{seq: []byte{'a'}, n: 1, width: 1},
			},
		},
		{
			name:  "esc sequence with intermediate",
			input: []byte("\x1b Q"),
			expected: []expectedSequence{
				{seq: []byte("\x1b Q"), n: 3},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var state byte
			input := tc.input
			results := make([]expectedSequence, 0)
			for len(input) > 0 {
				seq, width, n, newState := DecodeSequence(input, state, nil)
				state = newState
				input = input[n:]
				results = append(results, expectedSequence{seq: seq, width: width, n: n})
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
