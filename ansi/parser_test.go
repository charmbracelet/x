package ansi

import (
	"os"
	"reflect"
	"slices"
	"testing"
)

type csiSequence struct {
	Cmd    Cmd
	Params Params
}

type dcsSequence struct {
	Cmd    Cmd
	Params Params
	Data   []byte
}

type testCase struct {
	name     string
	input    string
	expected []any
}

type testDispatcher struct {
	dispatched []any
}

func (d *testDispatcher) dispatchRune(r rune) {
	d.dispatched = append(d.dispatched, r)
}

func (d *testDispatcher) dispatchControl(b byte) {
	d.dispatched = append(d.dispatched, b)
}

func (d *testDispatcher) dispatchEsc(cmd Cmd) {
	d.dispatched = append(d.dispatched, cmd)
}

func (d *testDispatcher) dispatchCsi(cmd Cmd, params Params) {
	params = slices.Clone(params)
	d.dispatched = append(d.dispatched, csiSequence{Cmd: cmd, Params: params})
}

func (d *testDispatcher) dispatchDcs(cmd Cmd, params Params, data []byte) {
	params = slices.Clone(params)
	data = slices.Clone(data)
	d.dispatched = append(d.dispatched, dcsSequence{Cmd: cmd, Params: params, Data: data})
}

func (d *testDispatcher) dispatchOsc(cmd int, data []byte) {
	data = slices.Clone(data)
	d.dispatched = append(d.dispatched, data)
}

func (d *testDispatcher) dispatchApc(data []byte) {
	data = slices.Clone(data)
	d.dispatched = append(d.dispatched, data)
}

func testParser(d *testDispatcher) *Parser {
	p := NewParser()
	p.SetHandler(Handler{
		Print:     d.dispatchRune,
		Execute:   d.dispatchControl,
		HandleEsc: d.dispatchEsc,
		HandleCsi: d.dispatchCsi,
		HandleDcs: d.dispatchDcs,
		HandleOsc: d.dispatchOsc,
		HandleApc: d.dispatchApc,
	})
	p.SetParamsSize(16)
	p.SetDataSize(0)
	return p
}

func TestControlSequence(t *testing.T) {
	cases := []testCase{
		{
			name:     "just_esc",
			input:    "\x1b",
			expected: []any{},
		},
		{
			name:  "double_esc",
			input: "\x1b\x1b",
			expected: []any{
				byte(0x1b),
			},
		},
		// {
		// 	name:  "esc_bracket",
		// 	input: "\x1b[",
		// 	expected: []Sequence{
		// 		EscSequence('['),
		// 	},
		// },
		// {
		// 	name:  "csi_rune_esc_bracket",
		// 	input: "\x1b[1;2;3mabc\x1b\x1bP",
		// 	expected: []Sequence{
		// 		CsiSequence{
		// 			Params: []Parameter{1, 2, 3},
		// 			Cmd:    'm',
		// 		},
		// 		Rune('a'),
		// 		Rune('b'),
		// 		Rune('c'),
		// 		ControlCode(0x1b),
		// 		EscSequence('P'),
		// 	},
		// },
		{
			name:  "csi plus text",
			input: "Hello, \x1b[31mWorld!\x1b[0m",
			expected: []any{
				rune('H'),
				rune('e'),
				rune('l'),
				rune('l'),
				rune('o'),
				rune(','),
				rune(' '),
				csiSequence{
					Params: Params{31},
					Cmd:    'm',
				},
				rune('W'),
				rune('o'),
				rune('r'),
				rune('l'),
				rune('d'),
				rune('!'),
				csiSequence{
					Params: Params{0},
					Cmd:    'm',
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := testParser(dispatcher)
			parser.Parse([]byte(c.input))
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
		name: "params",
		parser: func() *Parser {
			p := NewParser()
			p.SetDataSize(0)
			p.SetParamsSize(16)
			return p
		}(),
	},
	{
		name: "params and data",
		parser: func() *Parser {
			p := NewParser()
			p.SetDataSize(1024)
			p.SetParamsSize(16)
			return p
		}(),
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
				p.parser.Parse(bts)
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
				p.parser.Parse(bts)
			}
		})
	}
}

func BenchmarkParserStateChanges(b *testing.B) {
	input := []byte("\x1b]2;X\x1b\\こんにちは\x1b[0m \x1bP0@\x1b\\")

	for _, p := range parsers {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p.parser.Parse(input)
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
