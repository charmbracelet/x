package xpty

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// UnixPty represents a classic Unix PTY (pseudo-terminal).
type UnixPty struct {
	master, slave *os.File
}

var _ Pty = &UnixPty{}

// NewUnixPty creates a new Unix PTY.
func NewUnixPty(width, height int, _ ...PtyOption) (*UnixPty, error) {
	ptm, pts, err := pty.Open()
	if err != nil {
		return nil, fmt.Errorf("error creating a unix pty: %w", err)
	}

	p := &UnixPty{
		master: ptm,
		slave:  pts,
	}

	if width >= 0 && height >= 0 {
		if err := p.Resize(width, height); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Close implements XPTY.
func (p *UnixPty) Close() (err error) {
	defer func() {
		serr := p.slave.Close()
		if err == nil {
			err = serr
		}
	}()
	if err := p.master.Close(); err != nil {
		return fmt.Errorf("error closing master end of pty: %w", err)
	}
	return
}

// Fd implements XPTY.
func (p *UnixPty) Fd() uintptr {
	return p.master.Fd()
}

// Name implements XPTY.
func (p *UnixPty) Name() string {
	return p.master.Name()
}

// SlaveName returns the name of the slave PTY.
// This is usually used for remote sessions to identify the running TTY. You
// can find this in SSH sessions defined as $SSH_TTY.
func (p *UnixPty) SlaveName() string {
	return p.slave.Name()
}

// Read implements XPTY.
func (p *UnixPty) Read(b []byte) (n int, err error) {
	n, err = p.master.Read(b)
	if err != nil {
		return n, fmt.Errorf("error reading from master end of pty: %w", err)
	}
	return
}

// Resize implements XPTY.
func (p *UnixPty) Resize(width int, height int) (err error) {
	return p.setWinsize(width, height, 0, 0)
}

// SetWinsize sets window size for the PTY.
func (p *UnixPty) SetWinsize(width, height, x, y int) error {
	return p.setWinsize(width, height, x, y)
}

// Size returns the size of the PTY.
func (p *UnixPty) Size() (width, height int, err error) {
	return p.size()
}

// Start implements XPTY.
func (p *UnixPty) Start(c *exec.Cmd) error {
	if c.Stdout == nil {
		c.Stdout = p.slave
	}
	if c.Stderr == nil {
		c.Stderr = p.slave
	}
	if c.Stdin == nil {
		c.Stdin = p.slave
	}
	if err := c.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}
	return nil
}

// Write implements XPTY.
func (p *UnixPty) Write(b []byte) (n int, err error) {
	n, err = p.master.Write(b)
	if err != nil {
		return n, fmt.Errorf("error writing to master end of pty: %w", err)
	}
	return
}

// Master returns the master end of the PTY.
func (p *UnixPty) Master() *os.File {
	return p.master
}

// Slave returns the slave end of the PTY.
func (p *UnixPty) Slave() *os.File {
	return p.slave
}

// Control runs the given function with the file descriptor of the master PTY.
func (p *UnixPty) Control(fn func(fd uintptr)) error {
	conn, err := p.master.SyscallConn()
	if err != nil {
		return fmt.Errorf("error gaining control of master end of pty: %w", err)
	}

	if err := conn.Control(fn); err != nil {
		return fmt.Errorf("error controlling master end of pty: %w", err)
	}
	return nil
}
