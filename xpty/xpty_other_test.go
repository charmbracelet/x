//go:build !windows

package xpty

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestUnixPtyStartControllingTerminal(t *testing.T) {
	p, err := NewUnixPty(80, 24)
	if err != nil {
		t.Fatalf("new pty: %v", err)
	}
	defer p.Close() //nolint:errcheck

	// Opening /dev/tty only works with a controlling terminal.
	cmd := exec.Command("sh", "-c", "exec 3</dev/tty && echo CTTY-OK || echo CTTY-FAIL")
	if err := p.Start(cmd); err != nil {
		t.Fatalf("start: %v", err)
	}

	out := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		b := make([]byte, 256)
		for !strings.Contains(buf.String(), "CTTY-") {
			n, err := p.Read(b)
			buf.Write(b[:n])
			if err != nil {
				break
			}
		}
		out <- buf.String()
	}()

	select {
	case s := <-out:
		if !strings.Contains(s, "CTTY-OK") {
			t.Fatalf("expected a controlling terminal, got %q", s)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}
	_ = cmd.Wait()
}

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
