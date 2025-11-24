package pony

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
)

// Badge represents a badge element (like "NEW", "BETA", status indicators).
type Badge struct {
	BaseElement
	text  string
	color color.Color
}

var _ Element = (*Badge)(nil)

// NewBadge creates a badge component.
func NewBadge(props Props, children []Element) Element {
	text := props.Get("text")
	if text == "" && len(children) > 0 {
		// Use first child as text
		if t, ok := children[0].(*Text); ok {
			text = t.Content()
		}
	}

	var fgColor color.Color
	if colorStr := props.Get("foreground-color"); colorStr != "" {
		if c, err := parseColor(colorStr); err == nil {
			fgColor = c
		}
	}

	return &Badge{
		text:  text,
		color: fgColor,
	}
}

// Draw renders the badge.
func (b *Badge) Draw(scr uv.Screen, area uv.Rectangle) {
	b.SetBounds(area)

	if b.text == "" {
		return
	}

	// Badges are rendered as [TEXT] with styling
	content := "[" + b.text + "]"
	if b.color != nil {
		style := uv.Style{Fg: b.color}
		content = style.Styled(content)
	}

	styled := uv.NewStyledString(content)
	styled.Draw(scr, area)
}

// Layout calculates badge size.
func (b *Badge) Layout(constraints Constraints) Size {
	width := len(b.text) + 2 // text + brackets
	return constraints.Constrain(Size{Width: width, Height: 1})
}

// Children returns nil.
func (b *Badge) Children() []Element {
	return nil
}

// ProgressView represents a progress bar element.
type ProgressView struct {
	BaseElement
	value int
	max   int
	width SizeConstraint
	color color.Color
	char  string
}

var _ Element = (*ProgressView)(nil)

// NewProgressView creates a progress bar component.
func NewProgressView(props Props, _ []Element) Element {
	value := parseIntAttr(props, "value", 0)
	maxValue := parseIntAttr(props, "max", 100)
	width := parseSizeConstraint(props.Get("width"))
	char := props.GetOr("char", "█")

	var fgColor color.Color
	if colorStr := props.Get("foreground-color"); colorStr != "" {
		if c, err := parseColor(colorStr); err == nil {
			fgColor = c
		}
	}

	return &ProgressView{
		value: value,
		max:   maxValue,
		width: width,
		color: fgColor,
		char:  char,
	}
}

// Draw renders the progress bar.
func (p *ProgressView) Draw(scr uv.Screen, area uv.Rectangle) {
	p.SetBounds(area)

	if area.Dx() == 0 {
		return
	}

	// Calculate filled portion
	filled := 0
	if p.max > 0 {
		filled = min((area.Dx()*p.value)/p.max, area.Dx())
	}

	// Create cell for filled portion
	filledCell := uv.NewCell(scr.WidthMethod(), p.char)
	if filledCell != nil && p.color != nil {
		filledCell.Style = uv.Style{Fg: p.color}
	}

	// Create cell for empty portion
	emptyCell := uv.NewCell(scr.WidthMethod(), "░")

	// Draw progress bar
	for x := 0; x < area.Dx(); x++ {
		if x < filled {
			scr.SetCell(area.Min.X+x, area.Min.Y, filledCell)
		} else {
			scr.SetCell(area.Min.X+x, area.Min.Y, emptyCell)
		}
	}
}

// Layout calculates progress bar size.
func (p *ProgressView) Layout(constraints Constraints) Size {
	// Default width if not specified
	width := 20

	// Apply width constraint if specified
	if !p.width.IsAuto() {
		// For fixed width, use the constraint value directly
		width = p.width.Apply(constraints.MaxWidth, width)
	} else {
		// For auto, take available width
		width = constraints.MaxWidth
	}

	return Size{Width: width, Height: 1}
}

// Children returns nil.
func (p *ProgressView) Children() []Element {
	return nil
}

// init registers built-in custom components.
func init() {
	Register("badge", NewBadge)
	Register("progressview", NewProgressView)
}
