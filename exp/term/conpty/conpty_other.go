//go:build !windows
// +build !windows

package conpty

// ConPty represents a Windows Console Pseudo-terminal.
// https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session#preparing-the-communication-channels
type ConPty struct{}

// New creates a new ConPty.
// This function is not supported on non-Windows platforms.
func New(int, int, int) (*ConPty, error) {
	return nil, ErrUnsupported
}

// Size returns the size of the ConPty.
func (*ConPty) Size() (int, int, error) {
	return 0, 0, ErrUnsupported
}

// Close closes the ConPty.
func (*ConPty) Close() error {
	return ErrUnsupported
}

// Fd returns the file descriptor of the ConPty.
func (*ConPty) Fd() uintptr {
	return 0
}

// Read implements io.Reader.
func (*ConPty) Read([]byte) (int, error) {
	return 0, ErrUnsupported
}

// Write implements io.Writer.
func (*ConPty) Write([]byte) (int, error) {
	return 0, ErrUnsupported
}

// Resize resizes the ConPty.
func (*ConPty) Resize(int, int) error {
	return ErrUnsupported
}
