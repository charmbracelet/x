package ansi

import "testing"

func TestSetProgress(t *testing.T) {
	expect := "\x1b]9;4;1;50\x07"
	got := SetProgressBar(50)
	if expect != got {
		t.Errorf("SetProgress(50) = %q, want %q", got, expect)
	}
}

func TestSetProgressNegative(t *testing.T) {
	expect := "\x1b]9;4;1;0\x07"
	got := SetProgressBar(-2)
	if expect != got {
		t.Errorf("SetProgress(-2) = %q, want %q", got, expect)
	}
}

func TestSetProgressAbove100(t *testing.T) {
	expect := "\x1b]9;4;1;100\x07"
	got := SetProgressBar(200)
	if expect != got {
		t.Errorf("SetProgress(200) = %q, want %q", got, expect)
	}
}

func TestSetErrorProgress(t *testing.T) {
	expect := "\x1b]9;4;2;50\x07"
	got := SetErrorProgressBar(50)
	if expect != got {
		t.Errorf("SetProgress(50) = %q, want %q", got, expect)
	}
}

func TestSetErrorProgressNegative(t *testing.T) {
	expect := "\x1b]9;4;2;0\x07"
	got := SetErrorProgressBar(-2)
	if expect != got {
		t.Errorf("SetErrorProgress(-2) = %q, want %q", got, expect)
	}
}

func TestSetErrorProgressAbove100(t *testing.T) {
	expect := "\x1b]9;4;2;100\x07"
	got := SetErrorProgressBar(200)
	if expect != got {
		t.Errorf("SetErrorProgress(200) = %q, want %q", got, expect)
	}
}

func TestSetWarningProgress(t *testing.T) {
	expect := "\x1b]9;4;4;50\x07"
	got := SetWarningProgressBar(50)
	if expect != got {
		t.Errorf("SetProgress(50) = %q, want %q", got, expect)
	}
}

func TestSetWarningProgressNegative(t *testing.T) {
	expect := "\x1b]9;4;4;0\x07"
	got := SetWarningProgressBar(-2)
	if expect != got {
		t.Errorf("SetWarningProgress(-2) = %q, want %q", got, expect)
	}
}

func TestSetWarningProgressAbove100(t *testing.T) {
	expect := "\x1b]9;4;4;100\x07"
	got := SetWarningProgressBar(200)
	if expect != got {
		t.Errorf("SetWarningProgress(200) = %q, want %q", got, expect)
	}
}
