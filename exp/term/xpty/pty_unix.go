//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package xpty

import (
	"github.com/charmbracelet/x/exp/term/termios"
	"golang.org/x/sys/unix"
)

// setWinsize sets window size for the PTY.
func (p *UnixPty) setWinsize(width, height, x, y int) error {
	var rErr error
	if err := p.Control(func(fd uintptr) {
		rErr = termios.SetWinsize(int(fd), &unix.Winsize{
			Row:    uint16(height),
			Col:    uint16(width),
			Xpixel: uint16(x),
			Ypixel: uint16(y),
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
