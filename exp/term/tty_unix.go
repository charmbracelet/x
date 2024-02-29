package term

import (
	"fmt"
	"os"
)

// OpenTTY opens a new TTY.
func OpenTTY() (*os.File, error) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return nil, fmt.Errorf("could not open a new TTY: %w", err)
	}
	return f, nil
}
