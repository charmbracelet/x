package ansi

import "testing"

func TestSetProgress(t *testing.T) {
	expect := "\x1b]9;4;1;50\x07"
	got := SetProgress(50)
	if expect != got {
		t.Errorf("SetProgress(50) = %q, want %q", got, expect)
	}
}

func TestSetErrorProgress(t *testing.T) {
	expect := "\x1b]9;4;2;50\x07"
	got := SetErrorProgress(50)
	if expect != got {
		t.Errorf("SetProgress(50) = %q, want %q", got, expect)
	}
}
