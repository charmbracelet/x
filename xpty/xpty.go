// Package xpty provides platform-independent interfaces to interact with
// pseudo-terminals (PTYs) in Go. It abstracts the differences between Unix and
// Windows systems and supports both ConPTY and classic Unix PTYs.
package xpty

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/x/term"
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
//
// The returned PTY will be a Unix PTY on Unix systems and a ConPTY on Windows.
// The width and height parameters specify the initial size of the PTY.
// You can pass additional options to the PTY by passing PtyOptions.
//
//	pty, err := xpty.NewPty(80, 24)
//	if err != nil {
//	   // handle error
//	}
//
//	defer pty.Close() // Make sure to close the PTY when done.
//	switch pty := pty.(type) {
//	case xpty.UnixPty:
//	    // Unix PTY
//	case xpty.ConPty:
//	    // ConPTY
//	}
func NewPty(width, height int, opts ...PtyOption) (Pty, error) {
	if runtime.GOOS == "windows" {
		return NewConPty(width, height, opts...)
	}
	return NewUnixPty(width, height, opts...)
}

// WaitProcess waits for the process to exit.
// This exists because on Windows, cmd.Wait() doesn't work with ConPty.
// When the OS is not windows, it'll simply fall back to cmd.Wait().
//
// If ctx is canceled, the process is killed, WaitProcess waits for it to be
// reaped, and the returned error reports the cancellation: ctx.Err() on a
// successful kill, or an error joining ctx.Err() and the kill error if the
// kill itself failed.
func WaitProcess(ctx context.Context, cmd *exec.Cmd) (err error) {
	if runtime.GOOS != "windows" {
		return cmd.Wait() //nolint:wrapcheck
	}

	if cmd.Process == nil {
		return errors.New("process not started")
	}

	type result struct {
		*os.ProcessState
		error
	}

	donec := make(chan result, 1)
	go func() {
		state, err := cmd.Process.Wait()
		donec <- result{state, err}
	}()

	select {
	case <-ctx.Done():
		killErr := cmd.Process.Kill()
		// Reap the process so it doesn't linger as a zombie.
		r := <-donec
		cmd.ProcessState = r.ProcessState
		if killErr != nil {
			return errors.Join(ctx.Err(), killErr)
		}
		return ctx.Err() //nolint:wrapcheck // sentinel; callers use errors.Is
	case r := <-donec:
		cmd.ProcessState = r.ProcessState
		err = r.error
	}

	// On Windows, os.Process.Wait returns nil even when the process exits
	// with a non-zero code; the *exec.ExitError that exec.Cmd.Wait would
	// return never materializes. Synthesize it so callers get the same
	// error shape as exec.Cmd.Wait on any platform.
	if err == nil && cmd.ProcessState != nil && !cmd.ProcessState.Success() {
		err = &exec.ExitError{ProcessState: cmd.ProcessState}
	}

	return err
}
