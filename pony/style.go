package pony

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/lucasb-eyer/go-colorful"
)

// ParseStyle parses a style string into a UV Style.
// Format: "fg:red; bg:#1a1b26; bold; italic".
func ParseStyle(s string) (uv.Style, error) {
	if s == "" {
		return uv.Style{}, nil
	}

	var style uv.Style

	for part := range strings.SplitSeq(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a key:value pair or just an attribute
		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			if len(kv) != 2 {
				return style, fmt.Errorf("invalid style property: %s", part)
			}

			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			switch key {
			case "fg", "foreground", "color":
				c, err := parseColor(value)
				if err != nil {
					return style, fmt.Errorf("invalid foreground color %q: %w", value, err)
				}
				style.Fg = c

			case "bg", "background":
				c, err := parseColor(value)
				if err != nil {
					return style, fmt.Errorf("invalid background color %q: %w", value, err)
				}
				style.Bg = c

			case "underline-color", "ul-color":
				c, err := parseColor(value)
				if err != nil {
					return style, fmt.Errorf("invalid underline color %q: %w", value, err)
				}
				style.UnderlineColor = c

			case "underline", "ul":
				switch value {
				case UnderlineNone:
					style.Underline = uv.UnderlineNone
				case UnderlineSingle, UnderlineSolid:
					style.Underline = uv.UnderlineSingle
				case UnderlineDouble:
					style.Underline = uv.UnderlineDouble
				case UnderlineCurly:
					style.Underline = uv.UnderlineCurly
				case UnderlineDotted:
					style.Underline = uv.UnderlineDotted
				case UnderlineDashed:
					style.Underline = uv.UnderlineDashed
				default:
					return style, fmt.Errorf("unknown underline style: %s", value)
				}

			default:
				return style, fmt.Errorf("unknown style property: %s", key)
			}
		} else {
			// It's an attribute
			if err := parseAttribute(&style, part); err != nil {
				return style, err
			}
		}
	}

	return style, nil
}

// parseAttribute parses a style attribute like "bold", "italic", etc.
func parseAttribute(style *uv.Style, attr string) error {
	attr = strings.ToLower(attr)

	switch attr {
	case "bold":
		style.Attrs |= uv.AttrBold
	case "faint", "dim":
		style.Attrs |= uv.AttrFaint
	case "italic":
		style.Attrs |= uv.AttrItalic
	case "blink":
		style.Attrs |= uv.AttrBlink
	case "rapid-blink":
		style.Attrs |= uv.AttrRapidBlink
	case "reverse", "invert":
		style.Attrs |= uv.AttrReverse
	case "conceal", "hidden":
		style.Attrs |= uv.AttrConceal
	case "strikethrough", "strike":
		style.Attrs |= uv.AttrStrikethrough
	case "underline":
		style.Underline = uv.UnderlineSingle
	default:
		return fmt.Errorf("unknown style attribute: %s", attr)
	}

	return nil
}

// parseColor parses a color string into a color.Color.
// Supports:
//   - Named colors: red, blue, green, etc.
//   - Hex colors: #FF0000, #f00
//   - RGB: rgb(255, 0, 0)
//   - ANSI colors: 0-255
func parseColor(s string) (color.Color, error) {
	s = strings.TrimSpace(s)

	// Check for hex color
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s)
	}

	// Check for rgb()
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		return parseRGBColor(s)
	}

	// Check for ANSI color code (0-255)
	if num, err := strconv.Atoi(s); err == nil && num >= 0 && num <= 255 {
		return ansiColor(num), nil
	}

	// Named color
	return parseNamedColor(s)
}

// parseHexColor parses hex colors like #FF0000 or #f00.
func parseHexColor(s string) (color.Color, error) {
	c, err := colorful.Hex(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %w", err)
	}
	return c, nil
}

// parseRGBColor parses RGB colors like rgb(255, 0, 0).
func parseRGBColor(s string) (color.Color, error) {
	s = strings.TrimPrefix(s, "rgb(")
	s = strings.TrimSuffix(s, ")")
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("rgb() requires 3 values")
	}

	var rgb [3]uint8
	for i, part := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || val < 0 || val > 255 {
			return nil, fmt.Errorf("invalid rgb value: %s", part)
		}
		rgb[i] = uint8(val)
	}

	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}, nil
}

// parseNamedColor parses named colors.
func parseNamedColor(s string) (color.Color, error) {
	s = strings.ToLower(s)

	// Basic ANSI colors
	switch s {
	case "black":
		return color.RGBA{0, 0, 0, 255}, nil
	case "red":
		return color.RGBA{170, 0, 0, 255}, nil
	case "green":
		return color.RGBA{0, 170, 0, 255}, nil
	case "yellow":
		return color.RGBA{170, 85, 0, 255}, nil
	case "blue":
		return color.RGBA{0, 0, 170, 255}, nil
	case "magenta":
		return color.RGBA{170, 0, 170, 255}, nil
	case "cyan":
		return color.RGBA{0, 170, 170, 255}, nil
	case "white":
		return color.RGBA{170, 170, 170, 255}, nil

	// Bright ANSI colors
	case "bright-black", "gray", "grey":
		return color.RGBA{85, 85, 85, 255}, nil
	case "bright-red":
		return color.RGBA{255, 85, 85, 255}, nil
	case "bright-green":
		return color.RGBA{85, 255, 85, 255}, nil
	case "bright-yellow":
		return color.RGBA{255, 255, 85, 255}, nil
	case "bright-blue":
		return color.RGBA{85, 85, 255, 255}, nil
	case "bright-magenta":
		return color.RGBA{255, 85, 255, 255}, nil
	case "bright-cyan":
		return color.RGBA{85, 255, 255, 255}, nil
	case "bright-white":
		return color.RGBA{255, 255, 255, 255}, nil

	default:
		return nil, fmt.Errorf("unknown color: %s", s)
	}
}

// ansiColor returns a color for ANSI 256 color palette.
func ansiColor(code int) color.Color {
	if code < 0 || code > 255 {
		return nil
	}

	// 0-7: standard colors
	if code < 8 {
		colors := []color.RGBA{
			{0, 0, 0, 255},       // black
			{170, 0, 0, 255},     // red
			{0, 170, 0, 255},     // green
			{170, 85, 0, 255},    // yellow
			{0, 0, 170, 255},     // blue
			{170, 0, 170, 255},   // magenta
			{0, 170, 170, 255},   // cyan
			{170, 170, 170, 255}, // white
		}
		return colors[code]
	}

	// 8-15: bright colors
	if code < 16 {
		colors := []color.RGBA{
			{85, 85, 85, 255},    // bright black
			{255, 85, 85, 255},   // bright red
			{85, 255, 85, 255},   // bright green
			{255, 255, 85, 255},  // bright yellow
			{85, 85, 255, 255},   // bright blue
			{255, 85, 255, 255},  // bright magenta
			{85, 255, 255, 255},  // bright cyan
			{255, 255, 255, 255}, // bright white
		}
		return colors[code-8]
	}

	// 16-231: 216 color cube
	if code < 232 {
		code -= 16
		r := (code / 36) * 51
		g := ((code % 36) / 6) * 51
		b := (code % 6) * 51
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255} //nolint:gosec // values bounded to 0-255
	}

	// 232-255: grayscale
	gray := 8 + (code-232)*10
	return color.RGBA{uint8(gray), uint8(gray), uint8(gray), 255} //nolint:gosec // value bounded to 0-255
}
