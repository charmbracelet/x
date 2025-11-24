package pony

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

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
	const markup = `<text font-weight="bold" foreground-color="red">Styled Text</text>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestRenderBoxWithBorderStyle(t *testing.T) {
	const markup = `<box border="rounded" border-color="cyan"><text>Content</text></box>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
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
