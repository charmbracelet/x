package pony

import uv "github.com/charmbracelet/ultraviolet"

// Spacer represents empty space that can grow to fill available space.
type Spacer struct {
	BaseElement
	Size int // fixed size, 0 means flexible
}

var _ Element = (*Spacer)(nil)

// NewSpacer creates a new spacer element.
func NewSpacer() *Spacer {
	return &Spacer{}
}

// NewFixedSpacer creates a new spacer with fixed size.
func NewFixedSpacer(size int) *Spacer {
	return &Spacer{Size: size}
}

// WithSize sets the size and returns the spacer for chaining.
func (s *Spacer) WithSize(size int) *Spacer {
	s.Size = size
	return s
}

// Draw renders the spacer (nothing to draw).
func (s *Spacer) Draw(_ uv.Screen, area uv.Rectangle) {
	s.SetBounds(area)
	// Spacers are invisible
}

// Layout calculates the spacer size.
func (s *Spacer) Layout(constraints Constraints) Size {
	if s.Size > 0 {
		return constraints.Constrain(Size{Width: s.Size, Height: s.Size})
	}
	// Flexible spacer - take all available space
	return Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight}
}

// Children returns nil for spacers.
func (s *Spacer) Children() []Element {
	return nil
}
