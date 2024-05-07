//go:build darwin || netbsd || freebsd || openbsd || linux || dragonfly
// +build darwin netbsd freebsd openbsd linux dragonfly

package termios

import "golang.org/x/sys/unix"

func setSpeed(term *unix.Termios, ispeed, ospeed uint32) {
	term.Ispeed = speed(ispeed)
	term.Ospeed = speed(ospeed)
}
