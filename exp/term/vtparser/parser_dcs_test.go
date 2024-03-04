package parser

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
			input: fmt.Sprintf("\x1bP%sp", strings.Repeat("1;", DefaultMaxParameters+1)),
			expected: []testSequence{
				testDcsHookSequence{
					intermediates: []byte{},
					rune:          'p',
					params: [][]uint16{
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
					},
					ignore: true,
				},
			},
		},
		{
			name:  "reset",
			input: "\x1b[3;1\x1bP1$tx\x9c",
			expected: []testSequence{
				testDcsHookSequence{
					params:        [][]uint16{{1}},
					intermediates: []byte{'$'},
					rune:          't',
					ignore:        false,
				},
				testDcsPutSequence('x'),
				testDcsUnhookSequence{},
			},
		},
		{
			name:  "parse",
			input: "\x1bP0;1|17/ab\x9c",
			expected: []testSequence{
				testDcsHookSequence{
					params:        [][]uint16{{0}, {1}},
					rune:          '|',
					intermediates: []byte{},
					ignore:        false,
				},
				testDcsPutSequence('1'),
				testDcsPutSequence('7'),
				testDcsPutSequence('/'),
				testDcsPutSequence('a'),
				testDcsPutSequence('b'),
				testDcsUnhookSequence{},
			},
		},
		{
			name:  "intermediate_reset_on_exit",
			input: "\x1bP=1sZZZ\x1b+\x5c",
			expected: []testSequence{
				testDcsHookSequence{
					prefix:        "=",
					params:        [][]uint16{{1}},
					intermediates: []byte{},
					rune:          's',
					ignore:        false,
				},
				testDcsPutSequence('Z'),
				testDcsPutSequence('Z'),
				testDcsPutSequence('Z'),
				testDcsUnhookSequence{},
				testEscSequence{
					intermediates: []byte{'+'},
					rune:          0x5c,
					ignore:        false,
				},
			},
		},
		{
			name:  "put_utf8",
			input: "\x1bP+😃\x1b\\",
			expected: []testSequence{
				testRune('😃'),
				testEscSequence{
					intermediates: []byte{},
					rune:          '\\',
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			if err := New(dispatcher).Parse(strings.NewReader(c.input)); err != nil {
				t.Error(err)
			}

			assert.Equal(t, len(c.expected), len(dispatcher.dispatched))
			assert.Equal(t, c.expected, dispatcher.dispatched)
		})
	}
}
