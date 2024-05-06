package xpty

import (
	"os"
	"os/exec"
)

// Pty represents a classic Unix PTY (pseudo-terminal).
type Pty struct {
	master, slave *os.File
	closed        bool
}

var _ XPty = &Pty{}

// Close implements XPTY.
func (p *Pty) Close() (err error) {
	if p.closed {
		return
	}

	defer func() {
		serr := p.slave.Close()
		if err == nil {
			err = serr
		}
		p.closed = true
	}()
	if err := p.master.Close(); err != nil {
		return err
	}
	return
}

// Fd implements XPTY.
func (p *Pty) Fd() uintptr {
	return p.master.Fd()
}

// Name implements XPTY.
func (p *Pty) Name() string {
	return p.master.Name()
}

// SName returns the name of the slave PTY.
func (p *Pty) SName() string {
	return p.slave.Name()
}

// Read implements XPTY.
func (p *Pty) Read(b []byte) (n int, err error) {
	return p.master.Read(b)
}

// Resize implements XPTY.
func (p *Pty) Resize(width int, height int) (err error) {
	return p.resize(width, height)
}

// Size returns the size of the PTY.
func (p *Pty) Size() (width, height int, err error) {
	return p.size()
}

// Start implements XPTY.
func (p *Pty) Start(c *exec.Cmd) error {
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
		return err
	}
	return nil
}

// Write implements XPTY.
func (p *Pty) Write(b []byte) (n int, err error) {
	return p.master.Write(b)
}

// Master returns the master file of the PTY.
func (p *Pty) Master() *os.File {
	return p.master
}

// Slave returns the slave file of the PTY.
func (p *Pty) Slave() *os.File {
	return p.slave
}
