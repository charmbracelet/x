package pony

import (
	"image/color"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// Text represents a text element.
type Text struct {
	BaseElement
	content   string
	style     uv.Style
	wrap      bool
	alignment string // leading, center, trailing
}

var _ Element = (*Text)(nil)

// NewText creates a new text element.
func NewText(content string) *Text {
	return &Text{content: content}
}

// Bold makes the text bold and returns the text for chaining.
func (t *Text) Bold() *Text {
	t.style.Attrs |= uv.AttrBold
	return t
}

// Italic makes the text italic and returns the text for chaining.
func (t *Text) Italic() *Text {
	t.style.Attrs |= uv.AttrItalic
	return t
}

// Underline underlines the text and returns the text for chaining.
func (t *Text) Underline() *Text {
	t.style.Underline = uv.UnderlineSingle
	return t
}

// Strikethrough adds strikethrough and returns the text for chaining.
func (t *Text) Strikethrough() *Text {
	t.style.Attrs |= uv.AttrStrikethrough
	return t
}

// Faint makes the text faint/dim and returns the text for chaining.
func (t *Text) Faint() *Text {
	t.style.Attrs |= uv.AttrFaint
	return t
}

// ForegroundColor sets the foreground color and returns the text for chaining.
func (t *Text) ForegroundColor(c color.Color) *Text {
	t.style.Fg = c
	return t
}

// BackgroundColor sets the background color and returns the text for chaining.
func (t *Text) BackgroundColor(c color.Color) *Text {
	t.style.Bg = c
	return t
}

// Alignment sets the alignment and returns the text for chaining.
func (t *Text) Alignment(alignment string) *Text {
	t.alignment = alignment
	return t
}

// Wrap enables wrapping and returns the text for chaining.
func (t *Text) Wrap(wrap bool) *Text {
	t.wrap = wrap
	return t
}

// Content returns the text content (for external access).
func (t *Text) Content() string {
	return t.content
}

// Draw renders the text to the screen.
func (t *Text) Draw(scr uv.Screen, area uv.Rectangle) {
	t.SetBounds(area)

	if t.content == "" {
		return
	}

	// Apply style to content if specified
	content := t.content
	if !t.style.IsZero() {
		content = t.style.Styled(content)
	}

	// Handle alignment
	if t.alignment != "" && t.alignment != AlignmentLeading {
		content = t.alignText(content, area.Dx())
	}

	// Create styled string
	styled := uv.NewStyledString(content)
	styled.Wrap = t.wrap

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

		switch t.alignment {
		case AlignmentCenter:
			leftPad := padding / 2
			rightPad := padding - leftPad
			aligned := strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
			result = append(result, aligned)

		case AlignmentTrailing:
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
	if t.content == "" {
		return Size{Width: 0, Height: 0}
	}

	// Calculate dimensions
	lines := strings.Split(t.content, "\n")
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
	if t.wrap && width > constraints.MaxWidth {
		width = constraints.MaxWidth
		totalChars := len(t.content)
		height = (totalChars + width - 1) / width
	}

	return constraints.Constrain(Size{Width: width, Height: height})
}

// Children returns nil for text elements.
func (t *Text) Children() []Element {
	return nil
}
