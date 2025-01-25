package ansi

import (
	"testing"
)

func TestSosPmApcSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "apc7",
			input: "\x1b_Gf=24,s=10,v=20,o=z;aGVsbG8gd29ybGQ=\x1b\\",
			expected: []any{
				[]byte("Gf=24,s=10,v=20,o=z;aGVsbG8gd29ybGQ="),
				Cmd('\\'),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := testParser(dispatcher)
			parser.Parse([]byte(c.input))
			assertEqual(t, len(c.expected), len(dispatcher.dispatched))
			assertEqual(t, c.expected, dispatcher.dispatched)
		})
	}
}
