package input

import (
	"testing"
)

func TestFocus(t *testing.T) {
	_, e := ParseSequence([]byte("\x1b[I"))
	switch e.(type) {
	case FocusEvent:
		// ok
	default:
		t.Error("invalid sequence")
	}
}

func TestBlur(t *testing.T) {
	_, e := ParseSequence([]byte("\x1b[O"))
	switch e.(type) {
	case BlurEvent:
		// ok
	default:
		t.Error("invalid sequence")
	}
}
