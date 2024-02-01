//go:build darwin || netbsd || freebsd || openbsd || linux || dragonfly
// +build darwin netbsd freebsd openbsd linux dragonfly

package termios

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// SetWinSize sets window size for an fd from a Winsize.
func SetWinSize(fd int, w *unix.Winsize) error {
	return unix.IoctlSetWinsize(fd, ioctlSetWinSize, w)
}

// GetWinSize gets window size for an fd.
func GetWinSize(fd int) (*unix.Winsize, error) {
	return unix.IoctlGetWinsize(fd, ioctlSetWinSize)
}

// GetTermios gets the termios of the given fd.
func GetTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, ioctlGets)
}

// SetTermios sets the given termios over the given fd's current termios.
func SetTermios(
	fd int,
	ispeed, ospeed uint32,
	ccs map[CC]uint8,
	iflag map[I]bool,
	oflag map[O]bool,
	cflag map[C]bool,
	lflag map[L]bool,
) error {
	term, err := unix.IoctlGetTermios(fd, ioctlGets)
	if err != nil {
		return err
	}
	term.Ispeed = speed(ispeed)
	term.Ospeed = speed(ospeed)

	for key, value := range ccs {
		call, ok := allCcOpts[key]
		if !ok {
			continue
		}
		term.Cc[call] = value
	}

	for key, value := range iflag {
		mask, ok := allInputOpts[key]
		if ok {
			if value {
				term.Iflag |= bit(mask)
			} else {
				term.Iflag &= ^bit(mask)
			}
		}
	}
	for key, value := range oflag {
		mask, ok := allOutputOpts[key]
		if ok {
			if value {
				term.Oflag |= bit(mask)
			} else {
				term.Oflag &= ^bit(mask)
			}
		}
	}
	for key, value := range cflag {
		mask, ok := allControlOpts[key]
		if ok {
			if value {
				term.Cflag |= bit(mask)
			} else {
				term.Cflag &= ^bit(mask)
			}
		}
	}
	for key, value := range lflag {
		mask, ok := allLineOpts[key]
		if ok {
			if value {
				term.Lflag |= bit(mask)
			} else {
				term.Lflag &= ^bit(mask)
			}
		}
	}
	return unix.IoctlSetTermios(fd, ioctlSets, term)
}

type CC uint8

const (
	INTR CC = iota
	QUIT
	ERASE
	KILL
	EOF
	EOL
	EOL2
	START
	STOP
	SUSP
	WERASE
	RPRNT
	LNEXT
	DISCARD
)

// https://www.man7.org/linux/man-pages/man3/termios.3.html
var allCcOpts = map[CC]int{
	INTR:    syscall.VINTR,
	QUIT:    syscall.VQUIT,
	ERASE:   syscall.VERASE,
	KILL:    syscall.VQUIT,
	EOF:     syscall.VEOF,
	EOL:     syscall.VEOL,
	EOL2:    syscall.VEOL2,
	START:   syscall.VSTART,
	STOP:    syscall.VSTOP,
	SUSP:    syscall.VSUSP,
	WERASE:  syscall.VWERASE,
	RPRNT:   syscall.VREPRINT,
	LNEXT:   syscall.VLNEXT,
	DISCARD: syscall.VDISCARD,

	// XXX: these syscalls don't exist
	// STATUS: syscall.VSTATUS,
	// SWTCH:  syscall.VSWTCH,
	// FLUSH:  syscall.VFLUSH,
	// DSUSP:  syscall.VDSUSP,
}

// Input Controls
type I uint8

const (
	IGNPAR I = iota
	PARMRK
	INPCK
	ISTRIP
	INLCR
	IGNCR
	ICRNL
	IXON
	IXANY
	IXOFF
	IMAXBEL
	IUCLC
)

var allInputOpts = map[I]uint32{
	IGNPAR:  syscall.IGNPAR,
	PARMRK:  syscall.PARMRK,
	INPCK:   syscall.INPCK,
	ISTRIP:  syscall.ISTRIP,
	INLCR:   syscall.INLCR,
	IGNCR:   syscall.IGNCR,
	ICRNL:   syscall.ICRNL,
	IXON:    syscall.IXON,
	IXANY:   syscall.IXANY,
	IXOFF:   syscall.IXOFF,
	IMAXBEL: syscall.IMAXBEL,
	// XXX:
	// "iuclc":   {I, syscall.IUCLC},
}

// Line Controls.
type L uint8

const (
	ISIG L = iota
	ICANON
	ECHO
	ECHOE
	ECHOK
	ECHONL
	NOFLSH
	TOSTOP
	IEXTEN
	ECHOCTL
	ECHOKE
	PENDIN
	IUTF8
	XCASE
)

var allLineOpts = map[L]uint32{
	ISIG:    syscall.ISIG,
	ICANON:  syscall.ICANON,
	ECHO:    syscall.ECHO,
	ECHOE:   syscall.ECHOE,
	ECHOK:   syscall.ECHOK,
	ECHONL:  syscall.ECHONL,
	NOFLSH:  syscall.NOFLSH,
	TOSTOP:  syscall.TOSTOP,
	IEXTEN:  syscall.IEXTEN,
	ECHOCTL: syscall.ECHOCTL,
	ECHOKE:  syscall.ECHOKE,
	PENDIN:  syscall.PENDIN,
	// XXX:
	// "iutf8":   {L, syscall.IUTF8},
	// "xcase":   {L, syscall.XCASE},
}

// Output Controls
type O uint8

const (
	OPOST O = iota
	ONLCR
	OCRNL
	ONOCR
	ONLRET
	OLCUC
)

var allOutputOpts = map[O]uint32{
	OPOST:  syscall.OPOST,
	ONLCR:  syscall.ONLCR,
	OCRNL:  syscall.OCRNL,
	ONOCR:  syscall.ONOCR,
	ONLRET: syscall.ONLRET,
	// XXX:
	// "olcuc":   {O, syscall.OLCUC},
}

// Control
type C uint8

const (
	CS7 C = iota
	CS8
	PARENB
	PARODD
)

var allControlOpts = map[C]uint32{
	CS7:    syscall.CS7,
	CS8:    syscall.CS8,
	PARENB: syscall.PARENB,
	PARODD: syscall.PARODD,
}
