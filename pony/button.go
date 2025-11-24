package pony

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Button represents a clickable button element.
type Button struct {
	BaseElement
	Text        string
	Style       uv.Style
	HoverStyle  uv.Style
	ActiveStyle uv.Style
	Border      string
	Padding     int
	Width       SizeConstraint
	Height      SizeConstraint
}

var _ Element = (*Button)(nil)

// NewButton creates a new button element.
func NewButton(text string) *Button {
	return &Button{
		Text:    text,
		Border:  BorderRounded,
		Padding: 1,
	}
}

// WithStyle sets the button style and returns the button for chaining.
func (b *Button) WithStyle(style uv.Style) *Button {
	b.Style = style
	return b
}

// WithHoverStyle sets the hover style and returns the button for chaining.
func (b *Button) WithHoverStyle(style uv.Style) *Button {
	b.HoverStyle = style
	return b
}

// WithActiveStyle sets the active (pressed) style and returns the button for chaining.
func (b *Button) WithActiveStyle(style uv.Style) *Button {
	b.ActiveStyle = style
	return b
}

// WithBorder sets the border type and returns the button for chaining.
func (b *Button) WithBorder(border string) *Button {
	b.Border = border
	return b
}

// WithPadding sets the padding and returns the button for chaining.
func (b *Button) WithPadding(padding int) *Button {
	b.Padding = padding
	return b
}

// WithWidth sets the width constraint and returns the button for chaining.
func (b *Button) WithWidth(width SizeConstraint) *Button {
	b.Width = width
	return b
}

// WithHeight sets the height constraint and returns the button for chaining.
func (b *Button) WithHeight(height SizeConstraint) *Button {
	b.Height = height
	return b
}

// Draw renders the button to the screen.
func (b *Button) Draw(scr uv.Screen, area uv.Rectangle) {
	b.SetBounds(area)

	// Create text element
	textElem := NewText(b.Text)
	if !b.Style.IsZero() {
		textElem.Style = b.Style
	}
	textElem.Align = AlignCenter

	// Wrap in box with border and padding
	box := NewBox(textElem).
		WithBorder(b.Border).
		WithPadding(b.Padding)

	if !b.Style.IsZero() {
		box.BorderStyle = b.Style
	}

	box.Draw(scr, area)
}

// Layout calculates button size.
func (b *Button) Layout(constraints Constraints) Size {
	// Create text element for sizing
	textElem := NewText(b.Text)
	textSize := textElem.Layout(Unbounded())

	// Add padding and border
	borderSize := 2
	if b.Border == BorderNone || b.Border == BorderHidden {
		borderSize = 0
	}

	paddingSize := b.Padding * 2

	width := textSize.Width + borderSize + paddingSize
	height := textSize.Height + borderSize + paddingSize

	result := Size{Width: width, Height: height}

	// Apply width constraint if specified
	if !b.Width.IsAuto() {
		result.Width = b.Width.Apply(constraints.MaxWidth, result.Width)
	}

	// Apply height constraint if specified
	if !b.Height.IsAuto() {
		result.Height = b.Height.Apply(constraints.MaxHeight, result.Height)
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
			text = t.Content
		}
	}

	btn := NewButton(text)

	if style := parseStyleAttr(props); !style.IsZero() {
		btn.Style = style
	}

	if border := props.Get("border"); border != "" {
		btn.Border = border
	}

	if padding := parseIntAttr(props, "padding", 0); padding > 0 {
		btn.Padding = padding
	}

	if width := props.Get("width"); width != "" {
		btn.Width = parseSizeConstraint(width)
	}

	if height := props.Get("height"); height != "" {
		btn.Height = parseSizeConstraint(height)
	}

	return btn
}

func init() {
	Register("button", NewButtonFromProps)
}
