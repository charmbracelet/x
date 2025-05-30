package ansi

import (
	"fmt"
	"strings"
	"testing"
)

func TestOscSequence(t *testing.T) {
	const maxBufferSize = 1024
	cases := []testCase{
		{
			name:  "parse",
			input: "\x1b]2;charmbracelet: ~/Source/bubbletea\x07",
			expected: []any{
				[]byte("2;charmbracelet: ~/Source/bubbletea"),
			},
		},
		{
			name:  "empty",
			input: "\x1b]\x07",
			expected: []any{
				[]byte{},
			},
		},
		{
			name:  "max_params",
			input: fmt.Sprintf("\x1b]%s\x1b\\", strings.Repeat(";", 17)),
			expected: []any{
				[]byte(strings.Repeat(";", 17)),
				Cmd('\\'),
			},
		},
		{
			name:  "bell_terminated",
			input: "\x1b]11;ff/00/ff\x07",
			expected: []any{
				[]byte("11;ff/00/ff"),
			},
		},
		{
			name:  "esc_st_terminated",
			input: "\x1b]11;ff/00/ff\x1b\\",
			expected: []any{
				[]byte("11;ff/00/ff"),
				Cmd('\\'),
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
			expected: []any{
				[]byte("2;echo '¯\\_(ツ)_/¯' && sleep 1"),
			},
		},
		{
			name:  "string_terminator",
			input: "\x1b]2;\xe6\x9c\xab\x1b\\",
			expected: []any{
				[]byte("2;\xe6"),
				Cmd('\\'),
			},
		},
		{
			name:  "exceed_max_buffer_size",
			input: fmt.Sprintf("\x1b]52;s%s\x07", strings.Repeat("a", maxBufferSize)),
			expected: []any{
				fmt.Appendf(nil, "52;s%s", strings.Repeat("a", maxBufferSize-4)), // 4 is the len of "52;s"
			},
		},
		{
			name:  "title_empty_params_esc",
			input: "\x1b]0;abc\x1b\\\x1b];;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;\x07",
			expected: []any{
				[]byte("0;abc"),
				Cmd('\\'),
				[]byte(strings.Repeat(";", 45)),
			},
		},
		{
			name:  "just command",
			input: "\x1b]112\x07",
			expected: []any{
				[]byte("112"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &testDispatcher{}
			parser := testParser(dispatcher)
			parser.data = make([]byte, maxBufferSize)
			parser.dataLen = maxBufferSize
			parser.Parse([]byte(c.input))
			assertEqual(t, len(c.expected), len(dispatcher.dispatched))
			assertEqual(t, c.expected, dispatcher.dispatched)
		})
	}
}
