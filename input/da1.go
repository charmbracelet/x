package input

import "github.com/charmbracelet/x/ansi"

// PrimaryDeviceAttributesEvent represents a primary device attributes event.
type PrimaryDeviceAttributesEvent []uint

func parsePrimaryDevAttrs(csi *ansi.CsiSequence) Event {
	// Primary Device Attributes
	da1 := make(PrimaryDeviceAttributesEvent, len(csi.Params))
	for i, p := range csi.Params {
		if !ansi.Parameter(p).HasMore() {
			da1[i] = uint(p)
		}
	}
	return da1
}
