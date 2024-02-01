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
		term.Ispeed,
		term.Ospeed,
		map[string]uint8{
			"erase": 1,
		},
		map[string]bool{
			"pendin": true,
			"echoke": false,
		},
	); err != nil {
		t.Error(err)
	}
}
