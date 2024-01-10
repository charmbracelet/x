package term

// State contains platform-specific state of a terminal.
type State struct {
	state
}

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	return isTerminal(fd)
}

// MakeRaw puts the terminal connected to the given file descriptor into raw
// mode and returns the previous state of the terminal so that it can be
// restored.
func MakeRaw(fd int) (*State, error) {
	return makeRaw(fd)
}

// GetState returns the current state of a terminal which may be useful to
// restore the terminal after a signal.
func GetState(fd int) (*State, error) {
	return getState(fd)
}

// Restore restores the terminal connected to the given file descriptor to a
// previous state.
func Restore(fd int, oldState *State) error {
	return restore(fd, oldState)
}

// GetSize returns the visible dimensions of the given terminal.
//
// These dimensions don't include any scrollback buffer height.
func GetSize(fd int) (width, height int, err error) {
	return getSize(fd)
}

// ReadPassword reads a line of input from a terminal without local echo.  This
// is commonly used for inputting passwords and other sensitive data. The slice
// returned does not include the \n.
func ReadPassword(fd int) ([]byte, error) {
	return readPassword(fd)
}
