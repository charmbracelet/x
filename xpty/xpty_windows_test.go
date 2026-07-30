//go:build windows

package xpty

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitProcessCancel(t *testing.T) {
	cmd := helperCmd("GO_HELPER_SLEEP=1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := WaitProcess(ctx, cmd)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
		t.Fatal("expected process to be reaped after cancellation")
	}
}
