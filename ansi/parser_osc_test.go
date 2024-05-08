package ansi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi/parser"
)

func TestOscSequence(t *testing.T) {
	const maxBufferSize = 1024
	cases := []testCase{
		{
			name:  "parse",
			input: "\x1b]2;charmbracelet: ~/Source/bubbletea\x07",
			expected: []Sequence{
				OscSequence{
					Data: []byte("2;charmbracelet: ~/Source/bubbletea"),
					Cmd:  2,
				},
			},
		},
		{
			name:  "empty",
			input: "\x1b]\x07",
			expected: []Sequence{
				OscSequence{
					Cmd: parser.MissingCommand,
				},
			},
		},
		{
			name:  "max_params",
			input: fmt.Sprintf("\x1b]%s\x1b\\", strings.Repeat(";", 17)),
			expected: []Sequence{
				OscSequence{
					Data: []byte(strings.Repeat(";", 17)),
					Cmd:  parser.MissingCommand,
				},
				EscSequence('\\'),
			},
		},
		{
			name:  "bell_terminated",
			input: "\x1b]11;ff/00/ff\x07",
			expected: []Sequence{
				OscSequence{
					Data: []byte("11;ff/00/ff"),
					Cmd:  11,
				},
			},
		},
		{
			name:  "esc_st_terminated",
			input: "\x1b]11;ff/00/ff\x1b\\",
			expected: []Sequence{
				OscSequence{
					Data: []byte("11;ff/00/ff"),
					Cmd:  11,
				},
				EscSequence('\\'),
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
			expected: []Sequence{
				OscSequence{
					Data: []byte("2;echo '¯\\_(ツ)_/¯' && sleep 1"),
					Cmd:  2,
				},
			},
		},
		{
			name:  "string_terminator",
			input: "\x1b]2;\xe6\x9c\xab\x1b\\",
			expected: []Sequence{
				OscSequence{
					Data: []byte("2;\xe6"),
					Cmd:  2,
				},
				EscSequence('\\'),
			},
		},
		{
			name:  "exceed_max_buffer_size",
			input: fmt.Sprintf("\x1b]52;s%s\x07", strings.Repeat("a", maxBufferSize)),
			expected: []Sequence{
				OscSequence{
					Data: []byte(fmt.Sprintf("52;s%s", strings.Repeat("a", maxBufferSize-4))), // 4 is the len of "52;s"
					Cmd:  52,
				},
			},
		},
		{
			name:  "title_empty_params_esc",
			input: "\x1b]0;abc\x1b\\\x1b];;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;\x07",
			expected: []Sequence{
				OscSequence{
					Data: []byte("0;abc"),
					Cmd:  0,
				},
				EscSequence('\\'),
				OscSequence{
					Data: []byte(strings.Repeat(";", 45)),
					Cmd:  parser.MissingCommand,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := testParser(dispatcher)
			parser.Data = make([]byte, maxBufferSize)
			parser.DataLen = maxBufferSize
			parser.Parse(dispatcher.Dispatch, []byte(c.input))
			assertEqual(t, len(c.expected), len(dispatcher.dispatched))
			assertEqual(t, c.expected, dispatcher.dispatched)
		})
	}
}
