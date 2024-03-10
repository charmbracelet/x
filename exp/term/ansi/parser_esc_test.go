package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "reset",
			input: "\x1b[3;1\x1b(A",
			expected: []testSequence{
				testEscSequence{
					inter:  '(',
					ignore: false,
					rune:   'A',
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := NewParser()
			parser.Handler = testHandler(dispatcher)
			parser.Parse([]byte(c.input))
			assert.Equal(t, len(c.expected), len(dispatcher.dispatched))
			for i := range c.expected {
				assert.Equal(t, c.expected[i], dispatcher.dispatched[i])
			}
		})
	}
}
