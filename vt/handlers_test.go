package vt

import "testing"

func TestShiftOutShiftIn(t *testing.T) {
	e := newTestTerminal(t, 10, 1)
	// G1 is DEC special graphics, where "q" is a horizontal line.
	_, _ = e.WriteString("\x1b)0\x0eq\x0fq")
	if got := rowText(e, 0); got != "─q" {
		t.Errorf("row 0: got = %q, want = %q", got, "─q")
	}
}
