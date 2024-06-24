package ansi

import (
	"testing"

	"github.com/charmbracelet/x/ansi/parser"
)

func TestEscSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "reset",
			input: "\x1b[3;1\x1b(A",
			expected: []Sequence{
				EscSequence('A' | '('<<parser.IntermedShift),
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
