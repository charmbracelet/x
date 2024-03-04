package parser

import (
	"strings"
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
					intermediates: []byte{'('},
					ignore:        false,
					rune:          'A',
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := New(dispatcher)
			if err := parser.Parse(strings.NewReader(c.input)); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, len(c.expected), len(dispatcher.dispatched))
			for i := range c.expected {
				assert.Equal(t, c.expected[i], dispatcher.dispatched[i])
			}
		})
	}
}
