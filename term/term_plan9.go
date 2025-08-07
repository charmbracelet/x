package term

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type state struct {
	termName string
	raw      bool
	ctl      *os.File
}

// termName returns the name of the terminal or os.ErrNotExist if there is no terminal.
func termName(fd uintptr) (string, error) {
	ctl, err := os.ReadFile(filepath.Join("/fd", fmt.Sprintf("%dctl", fd)))
	if err != nil {
		return "", err
	}
	f := strings.Fields(string(ctl))
	if len(f) == 0 {
		return "", os.ErrNotExist
	}
	return f[len(f)-1], nil
}

func isTerminal(fd uintptr) bool {
	ctl, err := os.ReadFile(filepath.Join("/fd", fmt.Sprintf("%dctl", fd)))
	if err != nil {
		return false
	}
	if strings.Contains(string(ctl), "/dev/cons") {
		return true
	}
	return false
}

func makeRaw(fd uintptr) (*State, error) {
	t, err := termName(fd)
	if err != nil {
		return nil, err
	}
	ctl, err := os.OpenFile(t, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	if _, err := ctl.Write([]byte("rawon")); err != nil {
		return nil, err
	}
	return &State{state: state{termName: t, raw: true, ctl: ctl}}, nil
}

func getState(fd uintptr) (*State, error) {
	t, err := termName(fd)
	if err != nil {
		return nil, err
	}
	ctl, err := os.OpenFile(t, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	return &State{state: state{termName: t, raw: false, ctl: ctl}}, nil

}

func restore(_ uintptr, state *State) error {
	if _, err := state.ctl.Write([]byte("rawoff")); err != nil {
		return err
	}
	return nil
}

func getSize(fd uintptr) (width, height int, err error) {
	return 80, 40, nil
}

func setState(_ uintptr, state *State) error {
	raw := "rawoff"
	if state.raw {
		raw = "rawon"
	}
	if _, err := state.ctl.Write([]byte(raw)); err != nil {
		return err
	}
	return nil
}

func readPassword(fd uintptr) ([]byte, error) {
	f := os.NewFile(fd, "cons")
	var b [128]byte
	n, err := f.Read(b[:])
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
