//go:build aix
// +build aix

package termios

import "golang.org/x/sys/unix"

const (
	gets       = unix.TCGETA
	sets       = unix.TCSETA
	getWinSize = unix.TIOCGWINSZ
	setWinSize = unix.TIOCSWINSZ
)
