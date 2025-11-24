package pony

import (
	"image/color"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/golden"
)

func TestParseStyle(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(uv.Style) bool
	}{
		{
			name:    "empty style",
			input:   "",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.IsZero()
			},
		},
		{
			name:    "bold attribute",
			input:   "bold",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Attrs&uv.AttrBold != 0
			},
		},
		{
			name:    "italic attribute",
			input:   "italic",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Attrs&uv.AttrItalic != 0
			},
		},
		{
			name:    "multiple attributes",
			input:   "bold; italic",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Attrs&uv.AttrBold != 0 && s.Attrs&uv.AttrItalic != 0
			},
		},
		{
			name:    "foreground named color",
			input:   "fg:red",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Fg != nil
			},
		},
		{
			name:    "background named color",
			input:   "bg:blue",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Bg != nil
			},
		},
		{
			name:    "foreground hex color",
			input:   "fg:#FF0000",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				if s.Fg == nil {
					return false
				}
				r, g, b, _ := s.Fg.RGBA()
				// Check if it's red (allowing for color conversion)
				return r > g && r > b
			},
		},
		{
			name:    "combined style",
			input:   "fg:cyan; bg:black; bold; italic",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Fg != nil && s.Bg != nil &&
					s.Attrs&uv.AttrBold != 0 &&
					s.Attrs&uv.AttrItalic != 0
			},
		},
		{
			name:    "underline single",
			input:   "underline:single",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Underline == uv.UnderlineSingle
			},
		},
		{
			name:    "underline curly",
			input:   "underline:curly",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Underline == uv.UnderlineCurly
			},
		},
		{
			name:    "underline as attribute",
			input:   "underline",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Underline == uv.UnderlineSingle
			},
		},
		{
			name:    "strikethrough",
			input:   "strikethrough",
			wantErr: false,
			checkFunc: func(s uv.Style) bool {
				return s.Attrs&uv.AttrStrikethrough != 0
			},
		},
		{
			name:    "invalid property",
			input:   "invalid:value",
			wantErr: true,
		},
		{
			name:    "invalid color",
			input:   "fg:notacolor",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := ParseStyle(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStyle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(style) {
				t.Errorf("ParseStyle() style check failed for input %q", tt.input)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(color.Color) bool
	}{
		{
			name:    "named color red",
			input:   "red",
			wantErr: false,
			check: func(c color.Color) bool {
				return c != nil
			},
		},
		{
			name:    "hex color",
			input:   "#FF0000",
			wantErr: false,
			check: func(c color.Color) bool {
				return c != nil
			},
		},
		{
			name:    "short hex color",
			input:   "#f00",
			wantErr: false,
			check: func(c color.Color) bool {
				return c != nil
			},
		},
		{
			name:    "rgb color",
			input:   "rgb(255, 0, 0)",
			wantErr: false,
			check: func(c color.Color) bool {
				if c == nil {
					return false
				}
				r, g, b, _ := c.RGBA()
				return r > g && r > b
			},
		},
		{
			name:    "ansi color code",
			input:   "196",
			wantErr: false,
			check: func(c color.Color) bool {
				return c != nil
			},
		},
		{
			name:    "bright colors",
			input:   "bright-red",
			wantErr: false,
			check: func(c color.Color) bool {
				return c != nil
			},
		},
		{
			name:    "invalid hex",
			input:   "#GGGGGG",
			wantErr: true,
		},
		{
			name:    "invalid rgb",
			input:   "rgb(300, 0, 0)",
			wantErr: true,
		},
		{
			name:    "unknown named color",
			input:   "notacolor",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseColor(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseColor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(c) {
				t.Errorf("parseColor() color check failed for input %q", tt.input)
			}
		})
	}
}

func TestRenderWithStyle(t *testing.T) {
	const markup = `<text style="bold; fg:red">Styled Text</text>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestRenderBoxWithBorderStyle(t *testing.T) {
	const markup = `<box border="rounded" border-style="fg:cyan"><text>Content</text></box>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

// Test additional style attributes.
func TestStyleAttributesCoverage(t *testing.T) {
	tests := []struct {
		input     string
		checkFunc func(uv.Style) bool
	}{
		{"faint", func(s uv.Style) bool { return s.Attrs&uv.AttrFaint != 0 }},
		{"dim", func(s uv.Style) bool { return s.Attrs&uv.AttrFaint != 0 }},
		{"rapid-blink", func(s uv.Style) bool { return s.Attrs&uv.AttrRapidBlink != 0 }},
		{"invert", func(s uv.Style) bool { return s.Attrs&uv.AttrReverse != 0 }},
		{"conceal", func(s uv.Style) bool { return s.Attrs&uv.AttrConceal != 0 }},
		{"hidden", func(s uv.Style) bool { return s.Attrs&uv.AttrConceal != 0 }},
		{"strike", func(s uv.Style) bool { return s.Attrs&uv.AttrStrikethrough != 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			style, err := ParseStyle(tt.input)
			if err != nil {
				t.Errorf("ParseStyle(%q) error = %v", tt.input, err)
				return
			}
			if !tt.checkFunc(style) {
				t.Errorf("ParseStyle(%q) check failed", tt.input)
			}
		})
	}
}

// Test named colors coverage.
func TestNamedColorsCoverage(t *testing.T) {
	colors := []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"gray", "grey", "bright-black",
		"bright-red", "bright-green", "bright-yellow",
		"bright-blue", "bright-magenta", "bright-cyan", "bright-white",
	}

	for _, colorName := range colors {
		t.Run(colorName, func(t *testing.T) {
			c, err := parseColor(colorName)
			if err != nil {
				t.Errorf("parseColor(%q) error = %v", colorName, err)
			}
			if c == nil {
				t.Errorf("parseColor(%q) returned nil", colorName)
			}
		})
	}
}

// Test ANSI colors coverage.
func TestAnsiColorsCoverage(t *testing.T) {
	// Test different ANSI color ranges
	testCodes := []int{0, 7, 8, 15, 16, 100, 231, 232, 240, 255}

	for _, code := range testCodes {
		c := ansiColor(code)
		if c == nil {
			t.Errorf("ansiColor(%d) returned nil", code)
		}
	}

	// Test out of range
	if ansiColor(-1) != nil {
		t.Error("ansiColor(-1) should return nil")
	}
	if ansiColor(256) != nil {
		t.Error("ansiColor(256) should return nil")
	}
}
