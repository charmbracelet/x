//go:build linux
// +build linux

package termios

import "golang.org/x/sys/unix"

const (
	gets       = unix.TCGETS
	sets       = unix.TCSETS
	getWinSize = unix.TIOCGWINSZ
	setWinSize = unix.TIOCSWINSZ
)
