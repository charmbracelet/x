package termios

import (
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

// SetWinSize sets window size for an fd from a Winsize.
func SetWinSize(fd uintptr, w *unix.Winsize) error {
	return unix.IoctlSetWinsize(int(fd), setWinSize, w)
}

// GetWinSize gets window size for an fd.
func GetWinSize(fd uintptr, w *unix.Winsize) (*unix.Winsize, error) {
	return unix.IoctlGetWinsize(int(fd), getWinSize)
}

// SetTermios sets the termios according to the given ssh.TerminalModes.
func SetTermios(fd int, modes ssh.TerminalModes) error {
	term, err := unix.IoctlGetTermios(fd, gets)
	if err != nil {
		return err
	}
	for c, v := range modes {
		if c == ssh.TTY_OP_ISPEED {
			term.Ispeed = v
			continue
		}
		if c == ssh.TTY_OP_OSPEED {
			term.Ospeed = v
			continue
		}
		ccbit, ok := sshCcOpts[c]
		if ok {
			term.Cc[ccbit.value] = uint8(v)
		}
		bbit, ok := sshBoolOpts[c]
		if ok {
			if v != 0 {
				switch bbit.word {
				case I:
					term.Iflag |= bbit.mask
				case O:
					term.Oflag |= bbit.mask
				case L:
					term.Lflag |= bbit.mask
				case C:
					term.Cflag |= bbit.mask
				}
			} else {
				switch bbit.word {
				case I:
					term.Iflag &= ^bbit.mask
				case O:
					term.Oflag &= ^bbit.mask
				case L:
					term.Lflag &= ^bbit.mask
				case C:
					term.Cflag &= ^bbit.mask
				}
			}
		}
	}
	return unix.IoctlSetTermios(fd, sets, term)
}

type ioclBit struct {
	name string
	word int
	mask uint32
}

type ccBit struct {
	name  string
	value int
}

// https://www.man7.org/linux/man-pages/man3/termios.3.html
var sshCcOpts = map[uint8]*ccBit{
	ssh.VINTR:    {"intr", syscall.VINTR},
	ssh.VQUIT:    {"quit", syscall.VQUIT},
	ssh.VERASE:   {"erase", syscall.VERASE},
	ssh.VKILL:    {"kill", syscall.VQUIT},
	ssh.VEOF:     {"eof", syscall.VEOF},
	ssh.VEOL:     {"eol", syscall.VEOL},
	ssh.VEOL2:    {"eol2", syscall.VEOL2},
	ssh.VSTART:   {"start", syscall.VSTART},
	ssh.VSTOP:    {"stop", syscall.VSTOP},
	ssh.VSUSP:    {"susp", syscall.VSUSP},
	ssh.VWERASE:  {"werase", syscall.VWERASE},
	ssh.VREPRINT: {"rprnt", syscall.VREPRINT},
	ssh.VLNEXT:   {"lnext", syscall.VLNEXT},
	ssh.VDISCARD: {"discard", syscall.VDISCARD},

	// XXX: those syscall don't exist... not sure what to do.
	// ssh.VSTATUS:  {"status", syscall.VSTATUS},
	// ssh.VSWTCH:   {"swtch", syscall.VSWTCH},
	// ssh.VFLUSH:   {"flush", syscall.VFLUSH},
	// ssh.VDSUSP:   {"dsusp", syscall.VDSUSP},
}

// https://www.man7.org/linux/man-pages/man3/termios.3.html
var sshBoolOpts = map[uint8]*ioclBit{
	ssh.IGNPAR:  {"ignpar", I, syscall.IGNPAR},
	ssh.PARMRK:  {"parmrk", I, syscall.PARMRK},
	ssh.INPCK:   {"inpck", I, syscall.INPCK},
	ssh.ISTRIP:  {"istrip", I, syscall.ISTRIP},
	ssh.INLCR:   {"inlcr", I, syscall.INLCR},
	ssh.IGNCR:   {"igncr", I, syscall.IGNCR},
	ssh.ICRNL:   {"icrnl", I, syscall.ICRNL},
	ssh.IUCLC:   {"iuclc", I, syscall.IUCLC},
	ssh.IXON:    {"ixon", I, syscall.IXON},
	ssh.IXANY:   {"ixany", I, syscall.IXANY},
	ssh.IXOFF:   {"ixoff", I, syscall.IXOFF},
	ssh.IMAXBEL: {"imaxbel", I, syscall.IMAXBEL},

	ssh.IUTF8:   {"iutf8", L, syscall.IUTF8}, // XXX
	ssh.ISIG:    {"isig", L, syscall.ISIG},
	ssh.ICANON:  {"icanon", L, syscall.ICANON},
	ssh.ECHO:    {"echo", L, syscall.ECHO},
	ssh.ECHOE:   {"echoe", L, syscall.ECHOE},
	ssh.ECHOK:   {"echok", L, syscall.ECHOK},
	ssh.ECHONL:  {"echonl", L, syscall.ECHONL},
	ssh.NOFLSH:  {"noflsh", L, syscall.NOFLSH},
	ssh.TOSTOP:  {"tostop", L, syscall.TOSTOP},
	ssh.IEXTEN:  {"iexten", L, syscall.IEXTEN},
	ssh.ECHOCTL: {"echoctl", L, syscall.ECHOCTL},
	ssh.ECHOKE:  {"echoke", L, syscall.ECHOKE},
	ssh.PENDIN:  {"pendin", L, syscall.PENDIN},
	ssh.XCASE:   {"xcase", L, syscall.XCASE},

	ssh.OPOST:  {"opost", O, syscall.OPOST},
	ssh.OLCUC:  {"olcuc", O, syscall.OLCUC},
	ssh.ONLCR:  {"onlcr", O, syscall.ONLCR},
	ssh.OCRNL:  {"ocrnl", O, syscall.OCRNL},
	ssh.ONOCR:  {"onocr", O, syscall.ONOCR},
	ssh.ONLRET: {"onlret", O, syscall.ONLRET},

	ssh.CS7:    {"cs7", C, syscall.CS7}, // XXX
	ssh.CS8:    {"cs8", C, syscall.CS8},
	ssh.PARENB: {"parenb", C, syscall.PARENB},
	ssh.PARODD: {"parodd", C, syscall.PARODD},
}

const (
	I = iota // Input control
	O        // Output control
	C        // Control
	L        // Line control
)
