package teatest

import (
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestRequireEqualOutputUpdate(t *testing.T) {
	enableUpdate(t)
	RequireEqualOutput(t, []byte("test"))
}

func TestWaitForErrorReader(t *testing.T) {
	err := doWaitFor(iotest.ErrReader(fmt.Errorf("fake")), func(bts []byte) bool {
		return true
	}, WithDuration(time.Millisecond), WithCheckInterval(10*time.Microsecond))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err.Error() != "WaitFor: fake" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestWaitForTimeout(t *testing.T) {
	err := doWaitFor(strings.NewReader("nope"), func(bts []byte) bool {
		return false
	}, WithDuration(time.Millisecond), WithCheckInterval(10*time.Microsecond))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err.Error() != "WaitFor: condition not met after 1ms" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func enableUpdate(tb testing.TB) {
	tb.Helper()
	previous := update
	*update = true
	tb.Cleanup(func() {
		update = previous
	})
}
