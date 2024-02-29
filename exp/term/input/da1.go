package input

import (
	"fmt"
)

// PrimaryDeviceAttributesEvent represents a primary device attributes event.
type PrimaryDeviceAttributesEvent []uint

var _ Event = PrimaryDeviceAttributesEvent{}

// String implements Event.
func (e PrimaryDeviceAttributesEvent) String() string {
	s := "DA1"
	if len(e) > 0 {
		s += fmt.Sprintf(": %v", []uint(e))
	}
	return s
}

// Type implements Event.
func (PrimaryDeviceAttributesEvent) Type() string {
	return "PrimaryDeviceAttributesEvent"
}
