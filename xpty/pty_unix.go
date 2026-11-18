//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package xpty

import (
	"os/exec"
	"syscall"

	"github.com/charmbracelet/x/termios"
	"golang.org/x/sys/unix"
)

// setSessionAttrs makes the command a session leader with the PTY slave
// (its stdin) as the controlling terminal, so job control and terminal
// generated signals such as Ctrl+C reach it. Caller-provided attributes are
// left untouched.
func setSessionAttrs(c *exec.Cmd) {
	if c.SysProcAttr != nil {
		return
	}
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
}

// setWinsize sets window size for the PTY.
func (p *UnixPty) setWinsize(width, height, x, y int) error {
	var rErr error
	if err := p.Control(func(fd uintptr) {
		rErr = termios.SetWinsize(int(fd), &unix.Winsize{
			Row:    uint16(height), //nolint:gosec
			Col:    uint16(width),  //nolint:gosec
			Xpixel: uint16(x),      //nolint:gosec
			Ypixel: uint16(y),      //nolint:gosec
		})
	}); err != nil {
		rErr = err
	}
	return rErr
}

// size returns the size of the PTY.
func (p *UnixPty) size() (width, height int, err error) {
	var rErr error
	if err := p.Control(func(fd uintptr) {
		ws, err := termios.GetWinsize(int(fd))
		if err != nil {
			rErr = err
			return
		}
		width = int(ws.Col)
		height = int(ws.Row)
	}); err != nil {
		rErr = err
	}

	return width, height, rErr
}
