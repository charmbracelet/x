package pony

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestTextAlignment(t *testing.T) {
	tests := []struct {
		name   string
		markup string
	}{
		{
			name:   "left aligned",
			markup: `<box width="20" border="normal"><text align="left">Left</text></box>`,
		},
		{
			name:   "center aligned",
			markup: `<box width="20" border="normal"><text align="center">Center</text></box>`,
		},
		{
			name:   "right aligned",
			markup: `<box width="20" border="normal"><text align="right">Right</text></box>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse[any](tt.markup)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			output := tmpl.Render(nil, 80, 10)
			golden.RequireEqual(t, output)
		})
	}
}

func TestBoxPadding(t *testing.T) {
	const markup = `<box border="normal" padding="2"><text>Content</text></box>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestVStackAlignment(t *testing.T) {
	const markup = `
<vstack align="center">
	<text>Short</text>
	<text>A bit longer text</text>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestHStackValign(t *testing.T) {
	const markup = `
<hstack valign="middle">
	<text>A</text>
	<text>B</text>
</hstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestCombinedFeatures(t *testing.T) {
	const markup = `
<hstack gap="1">
	<box border="rounded" width="50%" padding="1">
		<text align="center" style="bold">Centered</text>
	</box>
	<box border="normal" width="50%" padding="2">
		<text align="right" style="italic">Right</text>
	</box>
</hstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 100, 20)
	golden.RequireEqual(t, output)
}
