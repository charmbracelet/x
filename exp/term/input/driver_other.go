//go:build !windows
// +build !windows

package input

// ReadInput reads input events from the terminal.
//
// It reads up to len(e) events into e and returns the number of events read
// and an error, if any.
func (d *Driver) ReadInput(e []Event) (n int, err error) {
	return d.readInput(e)
}

// PeekInput peeks at input events from the terminal without consuming them.
//
// If the number of events requested is greater than the number of events
// available in the buffer, the number of available events will be returned.
func (d *Driver) PeekInput(n int) ([]Event, error) {
	return d.peekInput(n)
}
