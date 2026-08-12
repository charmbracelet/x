//go:build darwin
// +build darwin

package termios

import "syscall"

func init() {
	allCcOpts[WERASE] = syscall.VWERASE
	allCcOpts[DISCARD] = syscall.VDISCARD
	allLineOpts[IUTF8] = syscall.IUTF8
}
