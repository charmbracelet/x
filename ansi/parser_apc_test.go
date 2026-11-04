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
		{
			name:  "apc_utf8",
			input: "\x1b_G日本\x1b\\ok",
			expected: []any{
				[]byte("G日本"),
				Cmd('\\'),
				'o',
				'k',
			},
		},
		{
			name:  "apc_utf8_continuation_st",
			input: "\x1b_✳\x9cok",
			expected: []any{
				[]byte("✳"),
				'o',
				'k',
			},
		},
		{
			name:     "sos_utf8",
			input:    "\x1bX日本\x1b\\ok",
			expected: []any{Cmd('\\'), 'o', 'k'},
		},
		{
			name:     "pm_utf8",
			input:    "\x1b^✳\x9cok",
			expected: []any{'o', 'k'},
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
