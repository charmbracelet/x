package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSosPmApcSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "apc7",
			input: "\x1b_Gf=24,s=10,v=20,o=z;aGVsbG8gd29ybGQ=\x1b\\",
			expected: []testSequence{
				testSosPmApcSequence{
					k:    APC,
					data: []byte("Gf=24,s=10,v=20,o=z;aGVsbG8gd29ybGQ="),
				},
				testEscSequence{
					rune: '\\',
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := testParser(dispatcher)
			parser.Parse([]byte(c.input))
			assert.Equal(t, len(c.expected), len(dispatcher.dispatched))
			assert.Equal(t, c.expected, dispatcher.dispatched)
		})
	}
}
