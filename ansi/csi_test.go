package ansi

import (
	"testing"

	"github.com/charmbracelet/x/ansi/parser"
)

func TestCsiSequence_Marker(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		want int
	}{
		{
			name: "no marker",
			s:    CsiSequence{},
			want: 0,
		},
		{
			name: "marker",
			s: CsiSequence{
				Cmd: 'u' | '?'<<parser.MarkerShift,
			},
			want: '?',
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Marker(); got != tt.want {
				t.Errorf("CsiSequence.Marker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCsiSequence_Intermediate(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		want int
	}{
		{
			name: "no intermediate",
			s:    CsiSequence{},
			want: 0,
		},
		{
			name: "marker",
			s: CsiSequence{
				Cmd: 'u' | '?'<<parser.MarkerShift,
			},
			want: 0,
		},
		{
			name: "intermediate",
			s: CsiSequence{
				Cmd: 'u' | '$'<<parser.IntermedShift,
			},
			want: '$',
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Intermediate(); got != tt.want {
				t.Errorf("CsiSequence.Intermediate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCsiSequence_Command(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		want int
	}{
		{
			name: "no command",
			s:    CsiSequence{},
			want: 0,
		},
		{
			name: "command",
			s: CsiSequence{
				Cmd: 'u' | '?'<<parser.MarkerShift,
			},
			want: 'u',
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Command(); got != tt.want {
				t.Errorf("CsiSequence.Command() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCsiSequence_Param(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		i    int
		want int
	}{
		{
			name: "no param",
			s:    CsiSequence{},
			i:    0,
			want: -1,
		},
		{
			name: "param",
			s: CsiSequence{
				Params: []int{1, 2, 3},
			},
			i:    1,
			want: 2,
		},
		{
			name: "missing param",
			s: CsiSequence{
				Params: []int{1, parser.MissingParam, 3},
			},
			i:    1,
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Param(tt.i); got != tt.want {
				t.Errorf("CsiSequence.Param() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCsiSequence_HasMore(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		i    int
		want bool
	}{
		{
			name: "no param",
			s:    CsiSequence{},
			i:    0,
			want: false,
		},
		{
			name: "has more",
			s: CsiSequence{
				Params: []int{1 | parser.HasMoreFlag, 2, 3},
			},
			i:    0,
			want: true,
		},
		{
			name: "no more",
			s: CsiSequence{
				Params: []int{1, 2, 3},
			},
			i:    0,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.HasMore(tt.i); got != tt.want {
				t.Errorf("CsiSequence.HasMore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCsiSequence_Len(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		want int
	}{
		{
			name: "no param",
			s:    CsiSequence{},
			want: 0,
		},
		{
			name: "len",
			s: CsiSequence{
				Params: []int{1, 2, 3},
			},
			want: 3,
		},
		{
			name: "len with missing param",
			s: CsiSequence{
				Params: []int{1, parser.MissingParam, 3},
			},
			want: 3,
		},
		{
			name: "len with more flag",
			s: CsiSequence{
				Params: []int{1 | parser.HasMoreFlag, 2, 3},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Len(); got != tt.want {
				t.Errorf("CsiSequence.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCsiSequence_String(t *testing.T) {
	tests := []struct {
		name string
		s    CsiSequence
		want string
	}{
		{
			name: "empty",
			s:    CsiSequence{Cmd: 'R'},
			want: "\x1b[R",
		},
		{
			name: "with data",
			s: CsiSequence{
				Cmd:    'A',
				Params: []int{1, 2, 3},
			},
			want: "\x1b[1;2;3A",
		},
		{
			name: "with more flag",
			s: CsiSequence{
				Cmd:    'A',
				Params: []int{1 | parser.HasMoreFlag, 2, 3},
			},
			want: "\x1b[1:2;3A",
		},
		{
			name: "with intermediate",
			s: CsiSequence{
				Cmd:    'A' | '$'<<parser.IntermedShift,
				Params: []int{1, 2, 3},
			},
			want: "\x1b[1;2;3$A",
		},
		{
			name: "with marker",
			s: CsiSequence{
				Cmd:    'A' | '?'<<parser.MarkerShift,
				Params: []int{1, 2, 3},
			},
			want: "\x1b[?1;2;3A",
		},
		{
			name: "with marker intermediate and more flag",
			s: CsiSequence{
				Cmd:    'A' | '?'<<parser.MarkerShift | '$'<<parser.IntermedShift,
				Params: []int{1, 2 | parser.HasMoreFlag, 3},
			},
			want: "\x1b[?1;2:3$A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("CsiSequence.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
