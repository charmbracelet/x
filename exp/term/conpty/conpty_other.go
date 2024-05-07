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
func (c *ConPty) Size() (int, int, error) {
	return 0, 0, ErrUnsupported
}
