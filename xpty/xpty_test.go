package xpty

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// helperCmd re-execs the test binary as a helper subprocess, so tests
// don't depend on any external command. GO_HELPER_SLEEP makes the helper
// block until killed; GO_HELPER_EXIT sets its exit code.
func helperCmd(env ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

func TestWaitProcessSuccess(t *testing.T) {
	cmd := helperCmd()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := WaitProcess(context.Background(), cmd); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cmd.ProcessState == nil {
		t.Fatal("expected ProcessState to be set")
	}
}

func TestWaitProcessNonZeroExit(t *testing.T) {
	cmd := helperCmd("GO_HELPER_EXIT=3")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	err := WaitProcess(context.Background(), cmd)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 3 {
		t.Fatalf("expected exit code 3, got %d", exitErr.ExitCode())
	}
}

// TestHelperProcess is not a real test; it runs when the test binary is
// re-executed as a helper subprocess.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("GO_HELPER_SLEEP") == "1" {
		time.Sleep(time.Minute)
	}
	if os.Getenv("GO_HELPER_SLEEP_SHORT") == "1" {
		time.Sleep(time.Second)
	}
	if code := os.Getenv("GO_HELPER_EXIT"); code != "" {
		var n int
		if _, err := fmt.Sscan(code, &n); err == nil {
			os.Exit(n)
		}
	}
	os.Exit(0)
}
