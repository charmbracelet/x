package pony

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Divider represents a horizontal or vertical line.
type Divider struct {
	BaseElement
	Vertical bool
	Char     string
	Style    uv.Style
}

var _ Element = (*Divider)(nil)

// NewDivider creates a new divider element.
func NewDivider() *Divider {
	return &Divider{}
}

// NewVerticalDivider creates a new vertical divider.
func NewVerticalDivider() *Divider {
	return &Divider{Vertical: true}
}

// WithStyle sets the style and returns the divider for chaining.
func (d *Divider) WithStyle(style uv.Style) *Divider {
	d.Style = style
	return d
}

// WithChar sets the character and returns the divider for chaining.
func (d *Divider) WithChar(char string) *Divider {
	d.Char = char
	return d
}

// Draw renders the divider to the screen.
func (d *Divider) Draw(scr uv.Screen, area uv.Rectangle) {
	d.SetBounds(area)

	char := d.Char
	if char == "" {
		if d.Vertical {
			char = "│"
		} else {
			char = "─"
		}
	}

	cell := uv.NewCell(scr.WidthMethod(), char)
	if cell != nil {
		cell.Style = d.Style
	}

	if d.Vertical {
		for y := area.Min.Y; y < area.Max.Y; y++ {
			scr.SetCell(area.Min.X, y, cell)
		}
	} else {
		for x := area.Min.X; x < area.Max.X; x++ {
			scr.SetCell(x, area.Min.Y, cell)
		}
	}
}

// Layout calculates the divider size.
func (d *Divider) Layout(constraints Constraints) Size {
	if d.Vertical {
		return Size{Width: 1, Height: constraints.MaxHeight}
	}
	return Size{Width: constraints.MaxWidth, Height: 1}
}

// Children returns nil for dividers.
func (d *Divider) Children() []Element {
	return nil
}
