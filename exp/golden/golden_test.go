package golden

import "testing"

func TestRequireEqualUpdate(t *testing.T) {
	enableUpdate(t)
	RequireEqual(t, []byte("test"))
}

func TestRequireEqualNoUpdate(t *testing.T) {
	RequireEqual(t, []byte("test"))
}

func enableUpdate(tb testing.TB) {
	tb.Helper()
	previous := update
	*update = true
	tb.Cleanup(func() {
		update = previous
	})
}
