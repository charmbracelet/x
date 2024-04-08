package ansi

import "testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("OscSequence.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
