package ansi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDcsSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "max_params",
			input: fmt.Sprintf("\x1bP%sp\x1b\\", strings.Repeat("1;", DefaultMaxParameters+1)),
			expected: []testSequence{
				testDcsSequence{
					rune: 'p',
					data: []byte{},
					params: [][]uint{
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
						{1},
					},
					ignore: true,
				},
				testEscSequence{
					rune: '\\',
				},
			},
		},
		{
			name:  "reset",
			input: "\x1b[3;1\x1bP1$tx\x9c",
			expected: []testSequence{
				testDcsSequence{
					params: [][]uint{{1}},
					inter:  '$',
					rune:   't',
					ignore: false,
					data:   []byte{'x'},
				},
			},
		},
		{
			name:  "parse",
			input: "\x1bP0;1|17/ab\x9c",
			expected: []testSequence{
				testDcsSequence{
					params: [][]uint{{0}, {1}},
					rune:   '|',
					ignore: false,
					data:   []byte("17/ab"),
				},
			},
		},
		{
			name:  "intermediate_reset_on_exit",
			input: "\x1bP=1sZZZ\x1b+\x5c",
			expected: []testSequence{
				testDcsSequence{
					params: [][]uint{{1}},
					marker: '=',
					rune:   's',
					ignore: false,
					data:   []byte{'Z', 'Z', 'Z'},
				},
				testEscSequence{
					inter:  '+',
					rune:   0x5c,
					ignore: false,
				},
			},
		},
		{
			name:  "put_utf8",
			input: "\x1bP+r😃\x1b\\",
			expected: []testSequence{
				testDcsSequence{
					rune:   'r',
					params: [][]uint{{0}},
					inter:  '+',
					data:   []byte("😃"),
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
			parser := NewParser()
			parser.Handler = testHandler(dispatcher)
			parser.Parse([]byte(c.input))
			assert.Equal(t, len(c.expected), len(dispatcher.dispatched))
			assert.Equal(t, c.expected, dispatcher.dispatched)
		})
	}
}
