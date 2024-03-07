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

func parsePrimaryDevAttrs(params [][]uint) Event {
	// Primary Device Attributes
	da1 := make([]uint, len(params))
	for i, p := range params {
		da1[i] = p[0]
	}
	return PrimaryDeviceAttributesEvent(da1)
}
