package pony

import uv "github.com/charmbracelet/ultraviolet"

// Spacer represents empty space that can grow to fill available space.
type Spacer struct {
	BaseElement
	fixedSize int // fixed size, 0 means flexible
}

var _ Element = (*Spacer)(nil)

// NewSpacer creates a new spacer element.
func NewSpacer() *Spacer {
	return &Spacer{}
}

// NewFixedSpacer creates a new spacer with fixed size.
func NewFixedSpacer(size int) *Spacer {
	return &Spacer{fixedSize: size}
}

// FixedSize sets the size and returns the spacer for chaining.
func (s *Spacer) FixedSize(size int) *Spacer {
	s.fixedSize = size
	return s
}

// Draw renders the spacer (nothing to draw).
func (s *Spacer) Draw(_ uv.Screen, area uv.Rectangle) {
	s.SetBounds(area)
	// Spacers are invisible
}

// Layout calculates the spacer size.
func (s *Spacer) Layout(constraints Constraints) Size {
	if s.fixedSize > 0 {
		return constraints.Constrain(Size{Width: s.fixedSize, Height: s.fixedSize})
	}
	// Flexible spacer - take all available space
	return Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight}
}

// Children returns nil for spacers.
func (s *Spacer) Children() []Element {
	return nil
}
