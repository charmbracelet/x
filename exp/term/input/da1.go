package input

// PrimaryDeviceAttributesEvent represents a primary device attributes event.
type PrimaryDeviceAttributesEvent []uint

func parsePrimaryDevAttrs(params [][]uint) Event {
	// Primary Device Attributes
	da1 := make(PrimaryDeviceAttributesEvent, len(params))
	for i, p := range params {
		da1[i] = p[0]
	}
	return da1
}
