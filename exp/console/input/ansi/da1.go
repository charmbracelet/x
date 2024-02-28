package ansi

import (
	"fmt"

	"github.com/charmbracelet/x/exp/console/input"
)

// PrimaryDeviceAttributesEvent represents a primary device attributes event.
type PrimaryDeviceAttributesEvent []uint

var _ input.Event = PrimaryDeviceAttributesEvent{}

// String implements input.Event.
func (e PrimaryDeviceAttributesEvent) String() string {
	s := "DA1"
	if len(e) > 0 {
		s += fmt.Sprintf(": %v", []uint(e))
	}
	return s
}

// Type implements input.Event.
func (PrimaryDeviceAttributesEvent) Type() string {
	return "PrimaryDeviceAttributesEvent"
}
