package ansi

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCsiSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "no_params",
			input: "\x1b[m",
			expected: []testSequence{
				testCsiSequence{
					rune:   'm',
					params: [][]uint{{0}},
					ignore: false,
				},
			},
		},
		{
			name:  "max_params",
			input: "\x1b[" + strings.Repeat("1;", DefaultMaxParameters-1) + "p",
			expected: []testSequence{
				testCsiSequence{
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
						{0},
					},
					ignore: false,
					rune:   'p',
				},
			},
		},
		{
			name:  "ignore_long",
			input: "\x1b[" + strings.Repeat("1;", DefaultMaxParameters+2) + "p",
			expected: []testSequence{
				testCsiSequence{
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
					rune:   'p',
				},
			},
		},
		{
			name:  "trailing_semicolon",
			input: "\x1b[4;m",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint{{4}, {0}},
					ignore: false,
					rune:   'm',
				},
			},
		},
		{
			name:  "leading_semicolon",
			input: "\x1b[;4m",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint{{0}, {4}},
					ignore: false,
					rune:   'm',
				},
			},
		},
		{
			name:  "long_param",
			input: "\x1b[18446744073709551615m",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint{{math.MaxUint}},
					ignore: false,
					rune:   'm',
				},
			},
		},
		{
			name:  "reset",
			input: "\x1b[3;1\x1b[?1049h",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint{{1049}},
					marker: '?',
					ignore: false,
					rune:   'h',
				},
			},
		},
		{
			name:  "subparams",
			input: "\x1b[38:2:255:0:255;1m",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint{
						{38, 2, 255, 0, 255},
						{1},
					},
					ignore: false,
					rune:   'm',
				},
			},
		},
		{
			name:  "params_buffer_filled_with_subparams",
			input: "\x1b[::::::::::::::::::::::::::::::::x\x1b",
			expected: []testSequence{
				testCsiSequence{
					params: [][]uint{
						{
							0, 0, 0, 0, 0, 0, 0, 0,
							0, 0, 0, 0, 0, 0, 0, 0,
							0, 0, 0, 0, 0, 0, 0, 0,
							0, 0, 0, 0, 0, 0, 0, 0,
						},
					},
					ignore: true,
					rune:   'x',
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
