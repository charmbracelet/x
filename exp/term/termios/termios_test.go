//go:build !windows
// +build !windows

package termios

import (
	"os"
	"runtime"
	"testing"
)

func TestTermios(t *testing.T) {
	// this test just ensures the lib is available for the current Os
	if runtime.GOOS != "linux" {
		t.Skip()
	}
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	fd := int(p.Fd())
	w, err := GetWinSize(fd)
	if err != nil {
		t.Error(err)
	}
	if err := SetWinSize(fd, w); err != nil {
		t.Error(err)
	}

	term, err := GetTermios(fd)
	if err != nil {
		t.Error(err)
	}

	if err := SetTermios(
		fd,
		uint32(term.Ispeed),
		uint32(term.Ospeed),
		map[CC]uint8{
			ERASE: 1,
		},
		map[I]bool{
			IGNCR: true,
			IXOFF: false,
		},
		map[O]bool{
			OCRNL: true,
			ONLCR: false,
		},
		map[C]bool{
			CS7: true,
			CS8: false,
		},
		map[L]bool{
			ECHO:  true,
			ECHOE: false,
		},
	); err != nil {
		t.Error(err)
	}
}
