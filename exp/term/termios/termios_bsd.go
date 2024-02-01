//go:build darwin || netbsd || freebsd || openbsd
// +build darwin netbsd freebsd openbsd

package termios

import "golang.org/x/sys/unix"

const (
	gets       = unix.TIOCGETA
	sets       = unix.TIOCSETA
	getWinSize = unix.TIOCGWINSZ
	setWinSize = unix.TIOCSWINSZ
)
