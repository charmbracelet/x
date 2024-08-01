package ansi

import (
	"os"
	"reflect"
	"testing"
)

type testCase struct {
	name     string
	input    string
	expected []Sequence
}

type testDispatcher struct {
	dispatched []Sequence
}

func (d *testDispatcher) Dispatch(s Sequence) {
	d.dispatched = append(d.dispatched, s.Clone())
}

func testParser(d *testDispatcher) *Parser {
	p := NewParser(16, 0)
	return p
}

func TestControlSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "just_esc",
			input: "\x1b",
			expected: []Sequence{
				ControlCode(0x1b),
			},
		},
		{
			name:  "double_esc",
			input: "\x1b\x1b",
			expected: []Sequence{
				ControlCode(0x1b),
				ControlCode(0x1b),
			},
		},
		{
			name:  "esc_bracket",
			input: "\x1b[",
			expected: []Sequence{
				EscSequence('['),
			},
		},
		{
			name:  "csi_rune_esc_bracket",
			input: "\x1b[1;2;3mabc\x1b\x1bP",
			expected: []Sequence{
				CsiSequence{
					Params: []int{1, 2, 3},
					Cmd:    'm',
				},
				Rune('a'),
				Rune('b'),
				Rune('c'),
				ControlCode(0x1b),
				EscSequence('P'),
			},
		},
		{
			name:  "csi plus text",
			input: "Hello, \x1b[31mWorld!\x1b[0m",
			expected: []Sequence{
				Rune('H'),
				Rune('e'),
				Rune('l'),
				Rune('l'),
				Rune('o'),
				Rune(','),
				Rune(' '),
				CsiSequence{
					Params: []int{31},
					Cmd:    'm',
				},
				Rune('W'),
				Rune('o'),
				Rune('r'),
				Rune('l'),
				Rune('d'),
				Rune('!'),
				CsiSequence{
					Params: []int{0},
					Cmd:    'm',
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := testParser(dispatcher)
			parser.Parse(dispatcher.Dispatch, []byte(c.input))
			assertEqual(t, len(c.expected), len(dispatcher.dispatched))
			for i := range c.expected {
				assertEqual(t, c.expected[i], dispatcher.dispatched[i])
			}
		})
	}
}

var parsers = []struct {
	name   string
	parser *Parser
}{
	{
		name:   "simple",
		parser: &Parser{},
	},
	{
		name:   "params",
		parser: NewParser(16, 0),
	},
	{
		name:   "params and data",
		parser: NewParser(16, 1024),
	},
}

func BenchmarkParser(b *testing.B) {
	bts, err := os.ReadFile("./fixtures/demo.vte")
	if err != nil {
		b.Fatalf("Error: %v", err)
	}

	for _, p := range parsers {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p.parser.Parse(nil, bts)
			}
		})
	}
}

func BenchmarkParserUTF8(b *testing.B) {
	bts, err := os.ReadFile("./fixtures/UTF-8-demo.txt")
	if err != nil {
		b.Fatalf("Error: %v", err)
	}

	for _, p := range parsers {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p.parser.Parse(nil, bts)
			}
		})
	}
}

func BenchmarkParserStateChanges(b *testing.B) {
	input := []byte("\x1b]2;X\x1b\\こんにちは\x1b[0m \x1bP0@\x1b\\")

	for _, p := range parsers {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p.parser.Parse(nil, input)
			}
		})
	}
}

func assertEqual[T any](t *testing.T, expected, got T) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("expected:\n  %#v, got:\n  %#v", expected, got)
	}
}
