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

func BenchmarkNext(bm *testing.B) {
	bts, err := os.ReadFile("./fixtures/demo.vte")
	if err != nil {
		bm.Fatalf("Error: %v", err)
	}

	bm.ResetTimer()

	var parser Parser
	parser.Parse(nil, bts)
}

func BenchmarkStateChanges(bm *testing.B) {
	input := "\x1b]2;X\x1b\\ \x1b[0m \x1bP0@\x1b\\"

	for i := 0; i < bm.N; i++ {
		var parser Parser
		for i := 0; i < 1000; i++ {
			parser.Parse(nil, []byte(input))
		}
	}
}

func assertEqual[T any](t *testing.T, expected, got T) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("expected:\n  %#v, got:\n  %#v", expected, got)
	}
}
