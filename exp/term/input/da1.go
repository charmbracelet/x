package input

import (
	"fmt"
)

// PrimaryDeviceAttributesEvent represents a primary device attributes event.
type PrimaryDeviceAttributesEvent []uint

// String implements fmt.Stringer.
func (e PrimaryDeviceAttributesEvent) String() string {
	return fmt.Sprintf("%v", []uint(e))
}
