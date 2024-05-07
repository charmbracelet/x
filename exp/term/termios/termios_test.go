//go:build !windows
// +build !windows

package termios

import (
	"os"
	"runtime"
	"testing"
)

// This test is mostly so ./.github/workflows/termios.yml can build the
// tests for the platforms we want to support, and verify everything
// is available for each of them.
func TestTermios(t *testing.T) {
	if runtime.GOOS != "linux" {
		// the way we open a pty below is the linux way.
		t.Skip()
	}
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	fd := int(p.Fd())
	w, err := GetWinsize(fd)
	if err != nil {
		t.Error(err)
	}
	if err := SetWinsize(fd, w); err != nil {
		t.Error(err)
	}

	term, err := GetTermios(fd)
	if err != nil {
		t.Error(err)
	}

	ispeed, ospeed := getSpeed(term)
	if err := SetTermios(
		fd,
		ispeed,
		ospeed,
		map[CC]uint8{
			ERASE:  1,
			CC(50): 12, // invalid, should be ignored
		},
		map[I]bool{
			IGNCR: true,
			IXOFF: false,
			I(50): true, // invalid, should be ignored
		},
		map[O]bool{
			OCRNL: true,
			ONLCR: false,
			O(50): true, // invalid, should be ignored
		},
		map[C]bool{
			CS7:    true,
			PARODD: false,
			C(50):  true, // invalid, should be ignored
		},
		map[L]bool{
			ECHO:  true,
			ECHOE: false,
			L(50): true, // invalid, should be ignored
		},
	); err != nil {
		t.Error(err)
	}

	term, err = GetTermios(fd)
	if err != nil {
		t.Error(err)
	}
	if v := term.Cc[allCcOpts[ERASE]]; v != 1 {
		t.Errorf("Cc.ERROR should be 1, was %d", v)
	}
	if v := term.Iflag & bit(allInputOpts[IGNCR]); v == 0 {
		t.Errorf("I.IGNCR should be true, was %d", v)
	}
	if v := term.Iflag & bit(allInputOpts[IXOFF]); v != 0 {
		t.Errorf("I.IGNCR should be false, was %d", v)
	}
	if v := term.Oflag & bit(allOutputOpts[OCRNL]); v == 0 {
		t.Errorf("O.OCRNL should be true, was %d", v)
	}
	if v := term.Oflag & bit(allOutputOpts[ONLCR]); v != 0 {
		t.Errorf("O.ONLCR should be false, was %d", v)
	}
	if v := term.Cflag & bit(allControlOpts[CS7]); v == 0 {
		t.Errorf("C.CS7 should be true, was %d", v)
	}
	if v := term.Cflag & bit(allControlOpts[PARODD]); v != 0 {
		t.Errorf("C.PARODD should be false, was %d", v)
	}
	if v := term.Lflag & bit(allLineOpts[ECHO]); v == 0 {
		t.Errorf("L.ECHO should be true, was %d", v)
	}
	if v := term.Lflag & bit(allLineOpts[ECHOE]); v != 0 {
		t.Errorf("L.ECHOE should be false, was %d", v)
	}
}
