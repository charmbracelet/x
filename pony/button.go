package pony

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Button represents a clickable button element.
type Button struct {
	BaseElement
	text        string
	style       uv.Style
	hoverStyle  uv.Style
	activeStyle uv.Style
	border      string
	padding     int
	width       SizeConstraint
	height      SizeConstraint
}

var _ Element = (*Button)(nil)

// NewButton creates a new button element.
func NewButton(text string) *Button {
	return &Button{
		text:    text,
		border:  BorderRounded,
		padding: 1,
	}
}

// Style sets the button style and returns the button for chaining.
func (b *Button) Style(style uv.Style) *Button {
	b.style = style
	return b
}

// HoverStyle sets the hover style and returns the button for chaining.
func (b *Button) HoverStyle(style uv.Style) *Button {
	b.hoverStyle = style
	return b
}

// ActiveStyle sets the active (pressed) style and returns the button for chaining.
func (b *Button) ActiveStyle(style uv.Style) *Button {
	b.activeStyle = style
	return b
}

// Border sets the border type and returns the button for chaining.
func (b *Button) Border(border string) *Button {
	b.border = border
	return b
}

// Padding sets the padding and returns the button for chaining.
func (b *Button) Padding(padding int) *Button {
	b.padding = padding
	return b
}

// Width sets the width constraint and returns the button for chaining.
func (b *Button) Width(width SizeConstraint) *Button {
	b.width = width
	return b
}

// Height sets the height constraint and returns the button for chaining.
func (b *Button) Height(height SizeConstraint) *Button {
	b.height = height
	return b
}

// Draw renders the button to the screen.
func (b *Button) Draw(scr uv.Screen, area uv.Rectangle) {
	b.SetBounds(area)

	// Create text element
	textElem := NewText(b.text).Alignment(AlignmentCenter)
	if !b.style.IsZero() {
		// Apply style to text content
		if b.style.Fg != nil {
			textElem = textElem.ForegroundColor(b.style.Fg)
		}
		if b.style.Attrs&uv.AttrBold != 0 {
			textElem = textElem.Bold()
		}
		if b.style.Attrs&uv.AttrItalic != 0 {
			textElem = textElem.Italic()
		}
	}

	// Wrap in box with border and padding
	box := NewBox(textElem).
		Border(b.border).
		Padding(b.padding)

	if !b.style.IsZero() && b.style.Fg != nil {
		box = box.BorderColor(b.style.Fg)
	}

	box.Draw(scr, area)
}

// Layout calculates button size.
func (b *Button) Layout(constraints Constraints) Size {
	// Create text element for sizing
	textElem := NewText(b.text)
	textSize := textElem.Layout(Unbounded())

	// Add padding and border
	borderSize := 2
	if b.border == BorderNone || b.border == BorderHidden {
		borderSize = 0
	}

	paddingSize := b.padding * 2

	width := textSize.Width + borderSize + paddingSize
	height := textSize.Height + borderSize + paddingSize

	result := Size{Width: width, Height: height}

	// Apply width constraint if specified
	if !b.width.IsAuto() {
		result.Width = b.width.Apply(constraints.MaxWidth, result.Width)
	}

	// Apply height constraint if specified
	if !b.height.IsAuto() {
		result.Height = b.height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns nil for buttons.
func (b *Button) Children() []Element {
	return nil
}

// NewButtonFromProps creates a button from props (for parser).
func NewButtonFromProps(props Props, children []Element) Element {
	text := props.Get("text")
	if text == "" && len(children) > 0 {
		if t, ok := children[0].(*Text); ok {
			text = t.Content()
		}
	}

	btn := NewButton(text)

	// Parse foreground color for button text/border
	if fgColor := props.Get("foreground-color"); fgColor != "" {
		if c, err := parseColor(fgColor); err == nil {
			style := uv.Style{Fg: c}
			if props.Get("font-weight") == FontWeightBold {
				style.Attrs |= uv.AttrBold
			}
			btn = btn.Style(style)
		}
	}

	if border := props.Get("border"); border != "" {
		btn = btn.Border(border)
	}

	if padding := parseIntAttr(props, "padding", 0); padding > 0 {
		btn = btn.Padding(padding)
	}

	if width := props.Get("width"); width != "" {
		btn = btn.Width(parseSizeConstraint(width))
	}

	if height := props.Get("height"); height != "" {
		btn = btn.Height(parseSizeConstraint(height))
	}

	return btn
}

func init() {
	Register("button", NewButtonFromProps)
}
