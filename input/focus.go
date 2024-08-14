package input

// FocusEvent represents a focus event.
type FocusEvent struct{}

// String implements fmt.Stringer.
func (FocusEvent) String() string { return "focus" }

// BlurEvent represents a blur event.
type BlurEvent struct{}

// String implements fmt.Stringer.
func (BlurEvent) String() string { return "blur" }
