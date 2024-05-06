//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package xpty

import (
	"github.com/charmbracelet/x/exp/term/termios"
	"golang.org/x/sys/unix"
)

// resize implements XPTY.
func (p *Pty) resize(width int, height int) (err error) {
	conn, err := p.master.SyscallConn()
	if err != nil {
		return err
	}

	return conn.Control(func(fd uintptr) {
		err = termios.SetWinsize(int(fd), &unix.Winsize{
			Row: uint16(height),
			Col: uint16(width),
		})
	})
}

// size returns the size of the PTY.
func (p *Pty) size() (width, height int, err error) {
	conn, err := p.master.SyscallConn()
	if err != nil {
		return 0, 0, err
	}

	err = conn.Control(func(fd uintptr) {
		ws, err := termios.GetWinsize(int(fd))
		if err != nil {
			return
		}
		width = int(ws.Col)
		height = int(ws.Row)
	})

	return
}
