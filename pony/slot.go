package pony

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Slot represents a placeholder for dynamic content.
// Slots are filled with Elements passed via RenderWithSlots.
type Slot struct {
	BaseElement
	Name    string
	element Element // The actual element to render (filled during render)
}

var _ Element = (*Slot)(nil)

// NewSlot creates a new slot element.
func NewSlot(name string) *Slot {
	return &Slot{Name: name}
}

// Draw renders the slot's element if it exists.
func (s *Slot) Draw(scr uv.Screen, area uv.Rectangle) {
	s.SetBounds(area)

	if s.element != nil {
		s.element.Draw(scr, area)
	}
}

// Layout calculates the slot's element size if it exists.
func (s *Slot) Layout(constraints Constraints) Size {
	if s.element != nil {
		return s.element.Layout(constraints)
	}
	return Size{Width: 0, Height: 0}
}

// Children returns the slot's element children if it exists.
func (s *Slot) Children() []Element {
	if s.element != nil {
		return []Element{s.element}
	}
	return nil
}

// setElement sets the element for this slot (used internally during rendering).
func (s *Slot) setElement(elem Element) {
	s.element = elem
}
