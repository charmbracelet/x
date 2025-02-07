package golden

import "testing"

func TestRequireEqualUpdate(t *testing.T) {
	*update = true
	RequireEqual(t, []byte("test"))
}

func TestRequireEqualNoUpdate(t *testing.T) {
	*update = false
	RequireEqual(t, []byte("test"))
}

func TestRequireWithLineBreaks(t *testing.T) {
	*update = false
	RequireEqual(t, []byte("foo\nbar\nbaz\n"))
}
