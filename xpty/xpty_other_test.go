//go:build !windows

package xpty

import (
	"context"
	"testing"
	"time"
)

func TestWaitProcessCancelIgnored(t *testing.T) {
	// The unix path ignores ctx: it waits for the process to exit on its
	// own. Cancellation there is exec.CommandContext's job.
	cmd := helperCmd("GO_HELPER_SLEEP_SHORT=1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	if err := WaitProcess(ctx, cmd); err != nil {
		t.Fatalf("expected nil error on unix, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Fatalf("returned after %v; unix path should wait for natural exit", elapsed)
	}
}
