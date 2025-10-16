package ansi

import "testing"

func TestNotify(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "basic",
			s:    "Hello, World!",
			want: "\x1b]9;Hello, World!\x07",
		},
		{
			name: "empty",
			s:    "",
			want: "\x1b]9;\x07",
		},
		{
			name: "special characters",
			s:    "Line1\nLine2\tTabbed",
			want: "\x1b]9;Line1\nLine2\tTabbed\x07",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Notify(tt.s); got != tt.want {
				t.Errorf("Notify() = %q, want %q", got, tt.want)
			}
		})
	}
}
