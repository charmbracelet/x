package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWrapClusterWiderThanLimit(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		limit int
		want  string
	}{
		{name: "double width alone", in: "世", limit: 1, want: "世"},
		{name: "emoji alone", in: "🚀", limit: 1, want: "🚀"},
		{name: "double width then narrow", in: "世a", limit: 1, want: "世\na"},
		{name: "emoji then narrow", in: "🚀ab", limit: 1, want: "🚀\na\nb"},
		{name: "two double widths", in: "世界", limit: 1, want: "世\n界"},
		{name: "CJK then ascii stuck together was the residual #958 case", in: "中a", limit: 1, want: "中\na"},
		{name: "narrow then wide", in: "a世", limit: 1, want: "a\n世"},
		{name: "fits exactly", in: "世", limit: 2, want: "世"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ansi.Wrap(tc.in, tc.limit, ""); got != tc.want {
				t.Fatalf("Wrap(%q, %d) = %q, want %q", tc.in, tc.limit, got, tc.want)
			}
		})
	}
}

func TestHardwrapClusterWiderThanLimit(t *testing.T) {
	tests := []struct {
		in    string
		limit int
		want  string
	}{
		{"世", 1, "世"},
		{"世界", 1, "世\n界"},
		{"世a", 1, "世\na"},
		{"a世", 1, "a\n世"},
	}
	for _, tc := range tests {
		if got := ansi.Hardwrap(tc.in, tc.limit, false); got != tc.want {
			t.Fatalf("Hardwrap(%q, %d) = %q, want %q", tc.in, tc.limit, got, tc.want)
		}
	}
}
