package pony

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// Text represents a text element.
type Text struct {
	BaseElement
	Content string
	Style   uv.Style
	Wrap    bool
	Align   string // left, center, right
}

var _ Element = (*Text)(nil)

// NewText creates a new text element.
func NewText(content string) *Text {
	return &Text{Content: content}
}

// WithStyle sets the style and returns the text for chaining.
func (t *Text) WithStyle(style uv.Style) *Text {
	t.Style = style
	return t
}

// WithAlign sets the alignment and returns the text for chaining.
func (t *Text) WithAlign(align string) *Text {
	t.Align = align
	return t
}

// WithWrap enables wrapping and returns the text for chaining.
func (t *Text) WithWrap(wrap bool) *Text {
	t.Wrap = wrap
	return t
}

// Draw renders the text to the screen.
func (t *Text) Draw(scr uv.Screen, area uv.Rectangle) {
	t.SetBounds(area)

	if t.Content == "" {
		return
	}

	// Apply style to content if specified
	content := t.Content
	if !t.Style.IsZero() {
		content = t.Style.Styled(content)
	}

	// Handle alignment
	if t.Align != "" && t.Align != AlignLeft {
		content = t.alignText(content, area.Dx())
	}

	// Create styled string
	styled := uv.NewStyledString(content)
	styled.Wrap = t.Wrap

	styled.Draw(scr, area)
}

// alignText aligns text within the given width.
func (t *Text) alignText(content string, width int) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// Strip ANSI codes to get actual text width
		plainText := ansi.Strip(line)
		textWidth := ansi.StringWidth(plainText)

		if textWidth >= width {
			result = append(result, line)
			continue
		}

		padding := width - textWidth

		switch t.Align {
		case AlignCenter:
			leftPad := padding / 2
			rightPad := padding - leftPad
			aligned := strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
			result = append(result, aligned)

		case AlignRight:
			aligned := strings.Repeat(" ", padding) + line
			result = append(result, aligned)

		default:
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// Layout calculates the text size.
func (t *Text) Layout(constraints Constraints) Size {
	if t.Content == "" {
		return Size{Width: 0, Height: 0}
	}

	// Calculate dimensions
	lines := strings.Split(t.Content, "\n")
	height := len(lines)

	width := 0
	for _, line := range lines {
		// Use ANSI-aware width calculation
		lineWidth := ansi.StringWidth(line)
		if lineWidth > width {
			width = lineWidth
		}
	}

	// Apply wrapping if enabled
	if t.Wrap && width > constraints.MaxWidth {
		width = constraints.MaxWidth
		totalChars := len(t.Content)
		height = (totalChars + width - 1) / width
	}

	return constraints.Constrain(Size{Width: width, Height: height})
}

// Children returns nil for text elements.
func (t *Text) Children() []Element {
	return nil
}
