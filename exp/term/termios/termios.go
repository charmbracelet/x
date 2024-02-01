package termios

import (
	"syscall"

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

// GetTermios gets the termios of the given fd.
func GetTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, gets)
}

// SetTermios sets the given termios over the given fd's current termios.
func SetTermios(
	fd int,
	ispeed, ospeed uint32,
	ccs map[string]uint8,
	bools map[string]bool,
) error {
	term, err := unix.IoctlGetTermios(fd, gets)
	if err != nil {
		return err
	}
	term.Ispeed = ispeed
	term.Ospeed = ospeed

	for name, value := range ccs {
		call, ok := allCcOpts[name]
		if !ok {
			continue
		}
		term.Cc[call] = value
	}

	for name, value := range bools {
		bit, ok := allBoolOpts[name]
		if !ok {
			continue
		}
		if value {
			switch bit.word {
			case I:
				term.Iflag |= bit.mask
			case O:
				term.Oflag |= bit.mask
			case L:
				term.Lflag |= bit.mask
			case C:
				term.Cflag |= bit.mask
			}
		} else {
			switch bit.word {
			case I:
				term.Iflag &= ^bit.mask
			case O:
				term.Oflag &= ^bit.mask
			case L:
				term.Lflag &= ^bit.mask
			case C:
				term.Cflag &= ^bit.mask
			}
		}
	}

	return unix.IoctlSetTermios(fd, sets, term)
}

type ioclBit struct {
	word int
	mask uint32
}

const (
	I = iota // Input control
	O        // Output control
	C        // Control
	L        // Line control
)

// https://www.man7.org/linux/man-pages/man3/termios.3.html
var allCcOpts = map[string]int{
	"intr":    syscall.VINTR,
	"quit":    syscall.VQUIT,
	"erase":   syscall.VERASE,
	"kill":    syscall.VQUIT,
	"eof":     syscall.VEOF,
	"eol":     syscall.VEOL,
	"eol2":    syscall.VEOL2,
	"start":   syscall.VSTART,
	"stop":    syscall.VSTOP,
	"susp":    syscall.VSUSP,
	"werase":  syscall.VWERASE,
	"rprnt":   syscall.VREPRINT,
	"lnext":   syscall.VLNEXT,
	"discard": syscall.VDISCARD,

	// XXX: those syscall don't exist... not sure what to do.
	// "status": syscall.VSTATUS,
	// "swtch":  syscall.VSWTCH,
	// "flush":  syscall.VFLUSH,
	// "dsusp":  syscall.VDSUSP,
}

// https://www.man7.org/linux/man-pages/man3/termios.3.html
var allBoolOpts = map[string]*ioclBit{
	"ignpar":  {I, syscall.IGNPAR},
	"parmrk":  {I, syscall.PARMRK},
	"inpck":   {I, syscall.INPCK},
	"istrip":  {I, syscall.ISTRIP},
	"inlcr":   {I, syscall.INLCR},
	"igncr":   {I, syscall.IGNCR},
	"icrnl":   {I, syscall.ICRNL},
	"iuclc":   {I, syscall.IUCLC},
	"ixon":    {I, syscall.IXON},
	"ixany":   {I, syscall.IXANY},
	"ixoff":   {I, syscall.IXOFF},
	"imaxbel": {I, syscall.IMAXBEL},

	"iutf8":   {L, syscall.IUTF8}, // XXX
	"isig":    {L, syscall.ISIG},
	"icanon":  {L, syscall.ICANON},
	"echo":    {L, syscall.ECHO},
	"echoe":   {L, syscall.ECHOE},
	"echok":   {L, syscall.ECHOK},
	"echonl":  {L, syscall.ECHONL},
	"noflsh":  {L, syscall.NOFLSH},
	"tostop":  {L, syscall.TOSTOP},
	"iexten":  {L, syscall.IEXTEN},
	"echoctl": {L, syscall.ECHOCTL},
	"echoke":  {L, syscall.ECHOKE},
	"pendin":  {L, syscall.PENDIN},
	"xcase":   {L, syscall.XCASE},

	"opost":  {O, syscall.OPOST},
	"olcuc":  {O, syscall.OLCUC},
	"onlcr":  {O, syscall.ONLCR},
	"ocrnl":  {O, syscall.OCRNL},
	"onocr":  {O, syscall.ONOCR},
	"onlret": {O, syscall.ONLRET},

	"cs7":    {C, syscall.CS7}, // XXX
	"cs8":    {C, syscall.CS8},
	"parenb": {C, syscall.PARENB},
	"parodd": {C, syscall.PARODD},
}
