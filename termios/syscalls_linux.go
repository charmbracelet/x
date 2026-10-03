//go:build linux
// +build linux

package termios

import "syscall"

func init() {
	allCcOpts[SWTCH] = syscall.VSWTC
	allCcOpts[WERASE] = syscall.VWERASE
	allCcOpts[DISCARD] = syscall.VDISCARD
	allInputOpts[IUCLC] = syscall.IUCLC
	allLineOpts[IUTF8] = syscall.IUTF8
	allLineOpts[XCASE] = syscall.XCASE
	allOutputOpts[OLCUC] = syscall.OLCUC
}
