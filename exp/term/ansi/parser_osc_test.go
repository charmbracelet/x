package ansi

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOscSequence(t *testing.T) {
	cases := []testCase{
		{
			name:  "parse",
			input: "\x1b]2;charmbracelet: ~/Source/bubbletea\x07",
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						[]byte("2"), []byte("charmbracelet: ~/Source/bubbletea"),
					},
					bell: true,
				},
			},
		},
		{
			name:  "empty",
			input: "\x1b]\x07",
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{{}},
					bell:   true,
				},
			},
		},
		{
			name:  "max_params",
			input: fmt.Sprintf("\x1b]%s\x1b", strings.Repeat(";", DefaultMaxOscParameters+1)),
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
						{},
					},
					bell: false,
				},
			},
		},
		{
			name:  "bell_terminated",
			input: "\x1b]11;ff/00/ff\x07",
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						[]byte("11"), []byte("ff/00/ff"),
					},
					bell: true,
				},
			},
		},
		{
			name:  "esc_st_terminated",
			input: "\x1b]11;ff/00/ff\x1b\\",
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						[]byte("11"), []byte("ff/00/ff"),
					},
					bell: false,
				},
				testEscSequence{
					ignore: false,
					rune:   '\\',
				},
			},
		},
		{
			name: "utf8",
			input: string([]byte{
				0x1b, 0x5d, 0x32, 0x3b, 0x65, 0x63, 0x68, 0x6f, 0x20, 0x27,
				0xc2, 0xaf, 0x5c, 0x5f, 0x28, 0xe3, 0x83, 0x84, 0x29, 0x5f,
				0x2f, 0xc2, 0xaf, 0x27, 0x20, 0x26, 0x26, 0x20, 0x73, 0x6c,
				0x65, 0x65, 0x70, 0x20, 0x31, 0x9c,
			}),
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						[]byte("2"), []byte(`echo '¯\_(ツ)_/¯' && sleep 1`),
					},
					bell: false,
				},
			},
		},
		{
			name:  "string_terminator",
			input: "\x1b]2;\xe6\x9c\xab\x1b\\",
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						[]byte("2"), []byte("\xe6"),
					},
				},
				testEscSequence{
					ignore: false,
					rune:   '\\',
				},
			},
		},
		{
			name:  "exceed_max_buffer_size",
			input: string(append(append([]byte{0x1b, 0x5d, 0x35, 0x32, 0x3b, 0x73}, bytes.Repeat([]byte{'a'}, DefaultMaxOscBytes+100)...), 0x07)),
			expected: []testSequence{
				testOscSequence{
					params: [][]byte{
						[]byte("52"), append([]byte{'s'}, bytes.Repeat([]byte{'a'}, DefaultMaxOscBytes+100)...),
					},
					bell: true,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			New(testHandler(dispatcher)).Parse([]byte(c.input))
			assert.Equal(t, len(c.expected), len(dispatcher.dispatched))
			assert.Equal(t, c.expected, dispatcher.dispatched)
		})
	}
}
