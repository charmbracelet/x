package input

// PasteStartEvent is an event that is emitted when a terminal enters
// bracketed-paste mode.
type PasteStartEvent struct{}

var _ Event = PasteStartEvent{}

// String implements Event.
func (e PasteStartEvent) String() string {
	return "paste start"
}

// Type implements Event.
func (PasteStartEvent) Type() string {
	return "PasteStart"
}

// PasteEvent is an event that is emitted when a terminal receives pasted text.
type PasteEndEvent struct{}

var _ Event = PasteEndEvent{}

// String implements Event.
func (e PasteEndEvent) String() string {
	return "paste end"
}

// Type implements Event.
func (PasteEndEvent) Type() string {
	return "PasteEnd"
}
