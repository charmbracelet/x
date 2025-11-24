package pony

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
)

// StyleBuilder provides a fluent API for building styles.
type StyleBuilder struct {
	style uv.Style
}

// NewStyle creates a new style builder.
func NewStyle() *StyleBuilder {
	return &StyleBuilder{}
}

// Fg sets the foreground color.
func (sb *StyleBuilder) Fg(c color.Color) *StyleBuilder {
	sb.style.Fg = c
	return sb
}

// Bg sets the background color.
func (sb *StyleBuilder) Bg(c color.Color) *StyleBuilder {
	sb.style.Bg = c
	return sb
}

// UnderlineColor sets the underline color.
func (sb *StyleBuilder) UnderlineColor(c color.Color) *StyleBuilder {
	sb.style.UnderlineColor = c
	return sb
}

// Bold makes the text bold.
func (sb *StyleBuilder) Bold() *StyleBuilder {
	sb.style.Attrs |= uv.AttrBold
	return sb
}

// Faint makes the text faint/dim.
func (sb *StyleBuilder) Faint() *StyleBuilder {
	sb.style.Attrs |= uv.AttrFaint
	return sb
}

// Italic makes the text italic.
func (sb *StyleBuilder) Italic() *StyleBuilder {
	sb.style.Attrs |= uv.AttrItalic
	return sb
}

// Underline sets single underline.
func (sb *StyleBuilder) Underline() *StyleBuilder {
	sb.style.Underline = uv.UnderlineSingle
	return sb
}

// UnderlineStyle sets the underline style.
func (sb *StyleBuilder) UnderlineStyle(style uv.Underline) *StyleBuilder {
	sb.style.Underline = style
	return sb
}

// Blink makes the text blink.
func (sb *StyleBuilder) Blink() *StyleBuilder {
	sb.style.Attrs |= uv.AttrBlink
	return sb
}

// Reverse reverses foreground and background.
func (sb *StyleBuilder) Reverse() *StyleBuilder {
	sb.style.Attrs |= uv.AttrReverse
	return sb
}

// Strikethrough adds strikethrough.
func (sb *StyleBuilder) Strikethrough() *StyleBuilder {
	sb.style.Attrs |= uv.AttrStrikethrough
	return sb
}

// Build returns the built style.
func (sb *StyleBuilder) Build() uv.Style {
	return sb.style
}

// Color helpers

// Hex creates a color from a hex string.
// Panics if invalid - use HexSafe for error handling.
func Hex(s string) color.Color {
	c, err := parseHexColor(s)
	if err != nil {
		panic(err)
	}
	return c
}

// HexSafe creates a color from a hex string with error handling.
func HexSafe(s string) (color.Color, error) {
	return parseHexColor(s)
}

// RGB creates a color from RGB values.
func RGB(r, g, b uint8) color.Color {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// Common layout helpers

// Panel creates a box with border and padding.
func Panel(child Element, border string, padding int) *Box {
	return NewBox(child).
		Border(border).
		Padding(padding)
}

// PanelWithMargin creates a box with border, padding, and margin.
func PanelWithMargin(child Element, border string, padding, margin int) *Box {
	return NewBox(child).
		Border(border).
		Padding(padding).
		Margin(margin)
}

// Card creates a titled card with content.
func Card(title string, titleColor, borderColor color.Color, children ...Element) Element {
	titleText := NewText(title)
	if titleColor != nil {
		titleText = titleText.ForegroundColor(titleColor).Bold()
	}

	box := NewBox(
		NewVStack(
			titleText,
			NewDivider(),
			NewVStack(children...),
		),
	).Border("rounded").Padding(1)

	if borderColor != nil {
		box = box.BorderColor(borderColor)
	}

	return box
}

// Section creates a section with a header and content.
func Section(header string, headerColor color.Color, children ...Element) Element {
	headerText := NewText(header)
	if headerColor != nil {
		headerText = headerText.ForegroundColor(headerColor).Bold()
	}

	items := []Element{headerText}
	items = append(items, children...)
	return NewVStack(items...)
}

// Separated adds a divider between each child.
func Separated(children ...Element) Element {
	if len(children) == 0 {
		return NewVStack()
	}

	items := make([]Element, 0, len(children)*2-1)
	for i, child := range children {
		items = append(items, child)
		if i < len(children)-1 {
			items = append(items, NewDivider())
		}
	}

	return NewVStack(items...)
}

// Overlay creates a ZStack with children layered on top of each other.
func Overlay(children ...Element) Element {
	return NewZStack(children...)
}

// FlexGrow creates a flex wrapper with the specified grow value.
func FlexGrow(child Element, grow int) *Flex {
	return NewFlex(child).Grow(grow)
}

// Position creates an absolutely positioned element.
func Position(child Element, x, y int) *Positioned {
	return NewPositioned(child, x, y)
}

// PositionRight creates an element positioned relative to the right edge.
func PositionRight(child Element, right, y int) *Positioned {
	return NewPositioned(child, 0, y).Right(right)
}

// PositionBottom creates an element positioned relative to the bottom edge.
func PositionBottom(child Element, x, bottom int) *Positioned {
	return NewPositioned(child, x, 0).Bottom(bottom)
}

// PositionCorner creates an element positioned at a corner.
func PositionCorner(child Element, right, bottom int) *Positioned {
	return NewPositioned(child, 0, 0).Right(right).Bottom(bottom)
}
