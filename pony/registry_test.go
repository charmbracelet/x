package pony

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestRegistry(t *testing.T) {
	// Register a simple component
	Register("test-component", func(_ Props, _ []Element) Element {
		return NewText("Custom")
	})
	defer Unregister("test-component")

	// Check it was registered
	if _, ok := GetComponent("test-component"); !ok {
		t.Error("Component not registered")
	}

	// Check it's in the list
	names := RegisteredComponents()
	found := false
	for _, name := range names {
		if name == "test-component" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Component not in registered list")
	}

	// Unregister
	Unregister("test-component")
	if _, ok := GetComponent("test-component"); ok {
		t.Error("Component should be unregistered")
	}
}

func TestCustomComponentInMarkup(t *testing.T) {
	Register("custom", func(props Props, _ []Element) Element {
		return NewText("Custom: " + props.Get("text"))
	})
	defer Unregister("custom")

	const markup = `<custom text="Hello" />`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestBadgeComponent(t *testing.T) {
	tests := []struct {
		name   string
		markup string
	}{
		{
			name:   "badge with text attribute",
			markup: `<badge text="NEW" />`,
		},
		{
			name:   "badge with style",
			markup: `<badge text="BETA" foreground-color="yellow" />`,
		},
		{
			name:   "badge with child text",
			markup: `<badge><text>ALPHA</text></badge>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse[any](tt.markup)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			output := tmpl.Render(nil, 80, 24)
			golden.RequireEqual(t, output)
		})
	}
}

func TestProgressComponent(t *testing.T) {
	tests := []struct {
		name   string
		markup string
	}{
		{
			name:   "progress with value",
			markup: `<progressview value="50" max="100" width="40" />`,
		},
		{
			name:   "progress with custom width",
			markup: `<progressview value="75" max="100" width="30" />`,
		},
		{
			name:   "progress with style",
			markup: `<progressview value="100" max="100" width="20" foreground-color="green" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse[any](tt.markup)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			output := tmpl.Render(nil, 80, 24)
			golden.RequireEqual(t, output)
		})
	}
}

func TestCustomComponentWithChildren(t *testing.T) {
	Register("wrapper", func(_ Props, children []Element) Element {
		return NewBox(NewVStack(children...)).Border("rounded")
	})
	defer Unregister("wrapper")

	const markup = `
<wrapper>
	<text>Child 1</text>
	<text>Child 2</text>
</wrapper>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

// Test ClearRegistry.
func TestClearRegistry(t *testing.T) {
	Register("test-clear", func(Props, []Element) Element {
		return NewText("test")
	})

	if _, ok := GetComponent("test-clear"); !ok {
		t.Error("Component should be registered")
	}

	ClearRegistry()

	if _, ok := GetComponent("test-clear"); ok {
		t.Error("Component should be cleared")
	}

	// Re-register built-in components since we cleared everything
	Register("badge", NewBadge)
	Register("progressview", NewProgressView)
}

// Test component Children methods.
func TestComponentChildren(t *testing.T) {
	// Badge Children
	badge := &Badge{text: "test"}
	if badge.Children() != nil {
		t.Error("Badge Children should return nil")
	}

	// ProgressView Children
	progress := &ProgressView{value: 50, max: 100}
	if progress.Children() != nil {
		t.Error("ProgressView Children should return nil")
	}
}
