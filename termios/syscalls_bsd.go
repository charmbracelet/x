//go:build darwin && netbsd && freebsd && netbsd
// +build darwin,netbsd,freebsd,netbsd

package term

import "syscall"

func init() {
	allCcOpts[STATUS] = syscall.VSTATUS
	allCcOpts[DSUSP] = syscall.VDSUSP
	allCcOpts[WERASE] = syscall.VWERASE
	allCcOpts[DISCARD] = syscall.VDISCARD
}
