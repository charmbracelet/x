package ansi

import (
	"strconv"
	"strings"
	"testing"
)

func TestOscSequence_String(t *testing.T) {
	tests := []struct {
		name string
		s    OscSequence
		want string
	}{
		{
			name: "empty",
			s: OscSequence{
				Cmd:  0,
				Data: []byte("0;"),
			},
			want: "\x1b]0;\a",
		},
		{
			name: "with data",
			s: OscSequence{
				Cmd:  1,
				Data: []byte("1;hello"),
			},
			want: "\x1b]1;hello\x07",
		},
		{
			name: "window name",
			s: OscSequence{
				Cmd:  2,
				Data: []byte("2;[No Name] - - NVIM"),
			},
			want: "\x1b]2;[No Name] - - NVIM\a",
		},
		{
			name: "reset cursor color",
			s: OscSequence{
				Cmd:  112,
				Data: []byte("112"),
			},
			want: "\x1b]112\a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Split(string(tt.s.Data), ";")
			if len(parts) == 0 {
				t.Errorf("OSC sequence data is empty")
			}
			if cmd, err := strconv.Atoi(parts[0]); err != nil || cmd != tt.s.Cmd {
				t.Errorf("OSC sequence command is invalid")
			}
			if got := "\x1b]" + string(tt.s.Data) + "\x07"; got != tt.want {
				t.Errorf("OscSequence.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
