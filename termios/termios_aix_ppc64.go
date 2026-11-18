//go:build ppc64 && aix
// +build ppc64,aix

package termios

import "golang.org/x/sys/unix"

const (
	ioctlGets       = unix.TCGETS
	ioctlSets       = unix.TCSETS
	ioctlGetWinSize = unix.TIOCGWINSZ
	ioctlSetWinSize = unix.TIOCSWINSZ
)

func setSpeed(*unix.Termios, uint32, uint32) {
}

func getSpeed(*unix.Termios) (uint32, uint32) {
	return 0, 0
}
