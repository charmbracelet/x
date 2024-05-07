package xpty

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/x/exp/term"
	"github.com/creack/pty"
)

// ErrUnsupported is returned when a feature is not supported.
var ErrUnsupported = pty.ErrUnsupported

// Pty represents a PTY (pseudo-terminal) interface.
type Pty interface {
	term.File
	io.ReadWriteCloser

	// Resize resizes the PTY.
	Resize(width, height int) error

	// Size returns the size of the PTY.
	Size() (width, height int, err error)

	// Name returns the name of the PTY.
	Name() string

	// Start starts a command on the PTY.
	// The command started will have its standard input, output, and error
	// connected to the PTY.
	// On Windows, calling Wait won't work since the Go runtime doesn't handle
	// ConPTY processes correctly. See https://github.com/golang/go/pull/62710.
	Start(cmd *exec.Cmd) error
}

// Options represents PTY options.
type Options struct {
	Flags int
}

// PtyOption is a PTY option.
type PtyOption func(o Options)

// NewPty creates a new PTY.
func NewPty(width, height int, opts ...PtyOption) (Pty, error) {
	if runtime.GOOS == "windows" {
		return NewConPty(width, height, opts...)
	}
	return NewUnixPty(width, height, opts...)
}

// WaitProcess waits for the process to exit.
// This exists because on Windows, cmd.Wait() doesn't work with ConPty.
func WaitProcess(ctx context.Context, c *exec.Cmd) (err error) {
	if c.Process == nil {
		return errors.New("process not started")
	}

	type result struct {
		*os.ProcessState
		error
	}

	donec := make(chan result, 1)
	go func() {
		state, err := c.Process.Wait()
		donec <- result{state, err}
	}()

	select {
	case <-ctx.Done():
		err = c.Process.Kill()
	case r := <-donec:
		c.ProcessState = r.ProcessState
		err = r.error
	}

	return
}
