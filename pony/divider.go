package pony

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
)

// Divider represents a horizontal or vertical line.
type Divider struct {
	BaseElement
	vertical bool
	char     string
	color    color.Color
}

var _ Element = (*Divider)(nil)

// NewDivider creates a new divider element.
func NewDivider() *Divider {
	return &Divider{}
}

// NewVerticalDivider creates a new vertical divider.
func NewVerticalDivider() *Divider {
	return &Divider{vertical: true}
}

// ForegroundColor sets the color and returns the divider for chaining.
func (d *Divider) ForegroundColor(c color.Color) *Divider {
	d.color = c
	return d
}

// Char sets the character and returns the divider for chaining.
func (d *Divider) Char(char string) *Divider {
	d.char = char
	return d
}

// Draw renders the divider to the screen.
func (d *Divider) Draw(scr uv.Screen, area uv.Rectangle) {
	d.SetBounds(area)

	char := d.char
	if char == "" {
		if d.vertical {
			char = "│"
		} else {
			char = "─"
		}
	}

	cell := uv.NewCell(scr.WidthMethod(), char)
	if cell != nil && d.color != nil {
		cell.Style = uv.Style{Fg: d.color}
	}

	if d.vertical {
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
	if d.vertical {
		return Size{Width: 1, Height: constraints.MaxHeight}
	}
	return Size{Width: constraints.MaxWidth, Height: 1}
}

// Children returns nil for dividers.
func (d *Divider) Children() []Element {
	return nil
}
