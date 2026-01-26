package strings

import (
	"testing"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name string
		str  string
		args []string
		want bool
	}{
		{
			name: "empty string with empty args",
			str:  "",
			args: []string{},
			want: false,
		},
		{
			name: "empty string with non-empty args",
			str:  "",
			args: []string{"a", "b"},
			want: false,
		},
		{
			name: "non-empty string with empty args",
			str:  "hello",
			args: []string{},
			want: false,
		},
		{
			name: "string containing one of the args",
			str:  "hello world",
			args: []string{"foo", "world"},
			want: true,
		},
		{
			name: "string containing multiple args",
			str:  "hello world",
			args: []string{"hello", "world"},
			want: true,
		},
		{
			name: "string not containing any args",
			str:  "hello world",
			args: []string{"foo", "bar"},
			want: false,
		},
		{
			name: "empty substring in args",
			str:  "hello",
			args: []string{"", "world"},
			want: true, // empty string is considered contained in any string
		},
		{
			name: "case sensitive matching",
			str:  "Hello World",
			args: []string{"hello", "world"},
			want: false,
		},
		{
			name: "partial matches",
			str:  "abcde",
			args: []string{"bc", "de"},
			want: true,
		},
		{
			name: "repeated args",
			str:  "hello",
			args: []string{"hello", "hello"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsAnyOf(tt.str, tt.args...); got != tt.want {
				t.Errorf("ContainsAnyOf(%q, %v) = %v, want %v", tt.str, tt.args, got, tt.want)
			}
		})
	}
}
