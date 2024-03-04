package parser

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCsiSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "max_params",
			input: "\x1b[" + strings.Repeat("1;", DefaultMaxParameters-1) + "p",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint16{
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {0},
					},
					ignore:        false,
					intermediates: []byte{},
					rune:          'p',
				},
			},
		},
		{
			name:  "ignore_long",
			input: "\x1b[" + strings.Repeat("1;", DefaultMaxParameters+2) + "p",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint16{
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
						{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1},
					},
					ignore:        true,
					intermediates: []byte{},
					rune:          'p',
				},
			},
		},
		{
			name:  "trailing_semicolon",
			input: "\x1b[4;m",
			expected: []testSequence{
				testCsiSequence{
					params:        [][]uint16{{4}, {0}},
					intermediates: []byte{},
					ignore:        false,
					rune:          'm',
				},
			},
		},
		{
			name:  "leading_semicolon",
			input: "\x1b[;4m",
			expected: []testSequence{
				testCsiSequence{
					params:        [][]uint16{{0}, {4}},
					intermediates: []byte{},
					ignore:        false,
					rune:          'm',
				},
			},
		},
		{
			name:  "long_param",
			input: "\x1b[9223372036854775808m",
			expected: []testSequence{
				testCsiSequence{
					params:        [][]uint16{{math.MaxUint16}},
					intermediates: []byte{},
					ignore:        false,
					rune:          'm',
				},
			},
		},
		{
			name:  "reset",
			input: "\x1b[3;1\x1b[?1049h",
			expected: []testSequence{
				testCsiSequence{
					prefix:        "?",
					params:        [][]uint16{{1049}},
					intermediates: []byte{},
					ignore:        false,
					rune:          'h',
				},
			},
		},
		{
			name:  "subparams",
			input: "\x1b[38:2:255:0:255;1m",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint16{
						{38, 2, 255, 0, 255},
						{1},
					},
					intermediates: []byte{},
					ignore:        false,
					rune:          'm',
				},
			},
		},
		{
			name:  "params_buffer_filled_with_subparams",
			input: "\x1b[::::::::::::::::::::::::::::::::x\x1b",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint16{
						{
							0, 0, 0, 0, 0, 0, 0, 0,
							0, 0, 0, 0, 0, 0, 0, 0,
							0, 0, 0, 0, 0, 0, 0, 0,
							0, 0, 0, 0, 0, 0, 0, 0,
						},
					},
					intermediates: []byte{},
					ignore:        true,
					rune:          'x',
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
			assert.Equal(t, c.expected, dispatcher.dispatched)
		})
	}
}
