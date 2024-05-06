package xpty

import (
	"io"
	"os/exec"

	"github.com/charmbracelet/x/exp/term"
	"github.com/creack/pty"
)

// ErrUnsupported is returned when a feature is not supported.
var ErrUnsupported = pty.ErrUnsupported

// XPty represents a PTY (pseudo-terminal) interface.
type XPty interface {
	// Fd returns the file descriptor of the PTY.
	Fd() uintptr

	// Close closes the PTY.
	Close() error

	// Read reads data from the PTY.
	Read(p []byte) (n int, err error)

	// Write writes data to the PTY.
	Write(p []byte) (n int, err error)

	// Resize resizes the PTY.
	Resize(width, height int) error

	// Name returns the name of the PTY.
	Name() string

	// Start starts a command on the PTY.
	// The command started will have its standard input, output, and error
	// connected to the PTY.
	// On Windows, calling Wait won't work since the Go runtime doesn't handle
	// ConPTY processes correctly. See https://github.com/golang/go/pull/62710.
	Start(cmd *exec.Cmd) error
}

var _ io.ReadWriteCloser = XPty(nil)

var _ term.File = XPty(nil)
