package pony

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestTextLayout(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		constraints Constraints
		wantWidth   int
		wantHeight  int
	}{
		{
			name:        "simple text",
			text:        "Hello",
			constraints: Unbounded(),
			wantWidth:   5,
			wantHeight:  1,
		},
		{
			name:        "multiline text",
			text:        "Line 1\nLine 2\nLine 3",
			constraints: Unbounded(),
			wantWidth:   6,
			wantHeight:  3,
		},
		{
			name:        "empty text",
			text:        "",
			constraints: Unbounded(),
			wantWidth:   0,
			wantHeight:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := NewText(tt.text)
			size := elem.Layout(tt.constraints)

			if size.Width != tt.wantWidth {
				t.Errorf("Layout().Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Layout().Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

func TestVStackLayout(t *testing.T) {
	tests := []struct {
		name        string
		gap         int
		numChildren int
		constraints Constraints
		wantHeight  int
	}{
		{
			name:        "no children",
			gap:         0,
			numChildren: 0,
			constraints: Unbounded(),
			wantHeight:  0,
		},
		{
			name:        "two children no gap",
			gap:         0,
			numChildren: 2,
			constraints: Unbounded(),
			wantHeight:  2, // Each child is 1 line
		},
		{
			name:        "two children with gap",
			gap:         1,
			numChildren: 2,
			constraints: Unbounded(),
			wantHeight:  3, // 1 + gap + 1
		},
		{
			name:        "three children with gap",
			gap:         2,
			numChildren: 3,
			constraints: Unbounded(),
			wantHeight:  7, // 1 + gap + 1 + gap + 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]Element, tt.numChildren)
			for i := range children {
				children[i] = NewText("X")
			}

			elem := NewVStack(children...).Spacing(tt.gap)
			size := elem.Layout(tt.constraints)

			if size.Height != tt.wantHeight {
				t.Errorf("Layout().Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

func TestHStackLayout(t *testing.T) {
	tests := []struct {
		name        string
		gap         int
		numChildren int
		constraints Constraints
		wantWidth   int
	}{
		{
			name:        "no children",
			gap:         0,
			numChildren: 0,
			constraints: Unbounded(),
			wantWidth:   0,
		},
		{
			name:        "two children no gap",
			gap:         0,
			numChildren: 2,
			constraints: Unbounded(),
			wantWidth:   2, // Each child is 1 char wide
		},
		{
			name:        "two children with gap",
			gap:         1,
			numChildren: 2,
			constraints: Unbounded(),
			wantWidth:   3, // 1 + gap + 1
		},
		{
			name:        "three children with gap",
			gap:         2,
			numChildren: 3,
			constraints: Unbounded(),
			wantWidth:   7, // 1 + gap + 1 + gap + 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]Element, tt.numChildren)
			for i := range children {
				children[i] = NewText("X")
			}

			elem := NewHStack(children...).Spacing(tt.gap)
			size := elem.Layout(tt.constraints)

			if size.Width != tt.wantWidth {
				t.Errorf("Layout().Width = %d, want %d", size.Width, tt.wantWidth)
			}
		})
	}
}

func TestBoxLayout(t *testing.T) {
	tests := []struct {
		name       string
		border     string
		content    string
		wantBorder int // expected border size added
	}{
		{
			name:       "box with normal border",
			border:     "normal",
			content:    "Hello",
			wantBorder: 2, // left + right border
		},
		{
			name:       "box with rounded border",
			border:     "rounded",
			content:    "Test",
			wantBorder: 2,
		},
		{
			name:       "box with no border",
			border:     "none",
			content:    "Hello",
			wantBorder: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := NewText(tt.content)
			childSize := child.Layout(Unbounded())

			elem := NewBox(child).Border(tt.border)
			boxSize := elem.Layout(Unbounded())

			expectedWidth := childSize.Width + tt.wantBorder
			expectedHeight := childSize.Height + tt.wantBorder

			if boxSize.Width != expectedWidth {
				t.Errorf("Layout().Width = %d, want %d (Child: %d + Border: %d)",
					boxSize.Width, expectedWidth, childSize.Width, tt.wantBorder)
			}
			if boxSize.Height != expectedHeight {
				t.Errorf("Layout().Height = %d, want %d (Child: %d + Border: %d)",
					boxSize.Height, expectedHeight, childSize.Height, tt.wantBorder)
			}
		})
	}
}

func TestRender(t *testing.T) {
	const markup = `<text>Hello, World!</text>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestRenderBox(t *testing.T) {
	const markup = `<box border="rounded"><text>Test</text></box>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestRenderVStack(t *testing.T) {
	const markup = `
	<vstack spacing="1">
		<text>Line 1</text>
		<text>Line 2</text>
		<text>Line 3</text>
	</vstack>
	`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

// Helper function to check if string contains substring.
func TestRenderComplexLayout(t *testing.T) {
	const markup = `
<vstack spacing="1">
	<box border="double" border-color="cyan">
		<text font-weight="bold" foreground-color="yellow" alignment="center">Title</text>
	</box>
	<hstack spacing="2">
		<box border="rounded" width="50%">
			<text>Left</text>
		</box>
		<box border="rounded" width="50%">
			<text>Right</text>
		</box>
	</hstack>
	<divider foreground-color="gray" />
	<text foreground-color="green">Footer</text>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

// Test Box chaining methods.
func TestBoxWithMethods(t *testing.T) {
	box := NewBox(NewText("test"))

	result := box.
		Margin(2).
		MarginTop(1).
		MarginRight(2).
		MarginBottom(3).
		MarginLeft(4)

	if result == nil {
		t.Error("Method chaining should return box")
	}
}

// Test Text chaining methods.
func TestTextWithMethods(t *testing.T) {
	text := NewText("test")

	result := text.
		Alignment(AlignmentCenter).
		Wrap(true)

	if result == nil {
		t.Error("Method chaining should return text")
	}
}

// Test Box Children with nil child.
func TestBoxChildrenNil(t *testing.T) {
	box := NewBox(nil)
	if box.Children() != nil {
		t.Error("Box Children with nil child should return nil")
	}
}

// Test Props edge cases.
func TestPropsNil(t *testing.T) {
	var props Props

	// Get on nil props
	if props.Get("key") != "" {
		t.Error("Get on nil props should return empty")
	}

	// GetOr on nil props
	if props.GetOr("key", "default") != "default" {
		t.Error("GetOr on nil props should return default")
	}

	// Has on nil props
	if props.Has("key") {
		t.Error("Has on nil props should return false")
	}
}

// Test Constrain edge cases.
func TestConstrainLimits(t *testing.T) {
	c := Constraints{
		MinWidth:  10,
		MaxWidth:  20,
		MinHeight: 5,
		MaxHeight: 15,
	}

	// Size smaller than min
	size := c.Constrain(Size{Width: 5, Height: 3})
	if size.Width != 10 || size.Height != 5 {
		t.Errorf("Constrain should enforce min, got %v", size)
	}

	// Size larger than max
	size = c.Constrain(Size{Width: 25, Height: 20})
	if size.Width != 20 || size.Height != 15 {
		t.Errorf("Constrain should enforce max, got %v", size)
	}
}

// Test Fixed function.
func TestFixedConstraints(t *testing.T) {
	c := Fixed(10, 20)
	if c.MinWidth != 10 || c.MaxWidth != 10 || c.MinHeight != 20 || c.MaxHeight != 20 {
		t.Error("Fixed() not set correctly")
	}
}

// Test element constraint types.
func TestElementConstraintApply(t *testing.T) {
	// FixedConstraint Apply
	fc := FixedConstraint(10)
	if fc.Apply(100) != 10 {
		t.Error("FixedConstraint Apply failed")
	}

	if fc.Apply(5) != 5 {
		t.Error("FixedConstraint should clamp to available")
	}

	// FixedConstraint negative
	fcNeg := FixedConstraint(-5)
	if fcNeg.Apply(100) != 0 {
		t.Error("FixedConstraint with negative value should return 0")
	}

	// PercentConstraint Apply
	pc := PercentConstraint(50)
	if pc.Apply(100) != 50 {
		t.Error("PercentConstraint Apply failed")
	}

	// PercentConstraint negative
	pcNeg := PercentConstraint(-10)
	if pcNeg.Apply(100) != 0 {
		t.Error("PercentConstraint with negative value should return 0")
	}

	// PercentConstraint > 100
	pcOver := PercentConstraint(150)
	if pcOver.Apply(100) != 100 {
		t.Error("PercentConstraint > 100 should return available")
	}

	// AutoConstraint Apply
	ac := AutoConstraint{}
	if ac.Apply(100) != 100 {
		t.Error("AutoConstraint Apply failed")
	}
}
