//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !zos && !windows && !solaris && !plan9
// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!zos,!windows,!solaris,!plan9

package term

import (
	"fmt"
	"runtime"
)

type state struct{}

func isTerminal(fd int) bool {
	return false
}

func makeRaw(fd int) (*State, error) {
	return nil, fmt.Errorf("terminal: MakeRaw not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func getState(fd int) (*State, error) {
	return nil, fmt.Errorf("terminal: GetState not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func restore(fd int, state *State) error {
	return fmt.Errorf("terminal: Restore not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func getSize(fd int) (width, height int, err error) {
	return 0, 0, fmt.Errorf("terminal: GetSize not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func setState(fd int, state *State) error {
	return fmt.Errorf("terminal: SetState not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func readPassword(fd int) ([]byte, error) {
	return nil, fmt.Errorf("terminal: ReadPassword not implemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}
