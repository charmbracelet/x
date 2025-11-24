package pony

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/golden"
)

func TestFlexibleSpacerBasic(t *testing.T) {
	const markup = `
<vstack>
	<text>Top</text>
	<spacer />
	<text>Bottom</text>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 20, 10)
	golden.RequireEqual(t, output)
}

func TestMultipleFlexibleSpacers(t *testing.T) {
	const markup = `
<vstack>
	<text>First</text>
	<spacer />
	<text>Middle</text>
	<spacer />
	<text>Last</text>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 20, 15)
	golden.RequireEqual(t, output)
}

func TestFlexibleSpacerInHStack(t *testing.T) {
	const markup = `
<hstack>
	<text>Left</text>
	<spacer />
	<text>Right</text>
</hstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 30, 5)
	golden.RequireEqual(t, output)
}

// Test Spacer constructors and methods.
func TestSpacerConstructors(t *testing.T) {
	// NewSpacer
	spacer := NewSpacer()
	if spacer == nil {
		t.Fatal("NewSpacer returned nil")
	}

	// NewFixedSpacer
	fixedSpacer := NewFixedSpacer(10)
	if fixedSpacer == nil {
		t.Fatal("NewFixedSpacer returned nil")
	}

	// FixedSize method
	spacer.FixedSize(5)

	// Children
	if spacer.Children() != nil {
		t.Error("Spacer Children should return nil")
	}

	// Draw (should be no-op)
	spacer.Draw(nil, uv.Rectangle{})
}

// Test Flex constructors and methods.
func TestFlexConstructors(t *testing.T) {
	// NewFlex
	flex := NewFlex(NewText("test"))
	if flex == nil {
		t.Fatal("NewFlex returned nil")
	}
	if flex.Children() == nil {
		t.Error("NewFlex child not set")
	}

	// Method chaining
	flex.Grow(2)
	flex.Shrink(0)
	flex.Basis(10)

	// Children
	children := flex.Children()
	if len(children) != 1 {
		t.Error("Flex Children should return child")
	}

	// Test with nil child
	flexNil := NewFlex(nil)
	if flexNil.Children() != nil {
		t.Error("Flex Children with nil child should return nil")
	}
}

// Test flex helper functions.
func TestFlexHelpers(t *testing.T) {
	// GetFlexShrink
	flex := NewFlex(NewText("test")).Shrink(2)
	if GetFlexShrink(flex) != 2 {
		t.Error("GetFlexShrink failed")
	}

	text := NewText("test")
	if GetFlexShrink(text) != 1 {
		t.Error("GetFlexShrink should return 1 for non-flex")
	}

	// GetFlexBasis
	flex = NewFlex(NewText("test")).Basis(10)
	if GetFlexBasis(flex) != 10 {
		t.Error("GetFlexBasis failed")
	}

	if GetFlexBasis(text) != 0 {
		t.Error("GetFlexBasis should return 0 for non-flex")
	}

	// IsFlexible
	flex = NewFlex(NewText("test")).Grow(1)
	if !IsFlexible(flex) {
		t.Error("IsFlexible should return true")
	}

	if IsFlexible(text) {
		t.Error("IsFlexible should return false for non-flex")
	}
}

// Test Divider constructors and methods.
func TestDividerConstructors(t *testing.T) {
	// NewVerticalDivider
	div := NewVerticalDivider()
	if div == nil {
		t.Fatal("NewVerticalDivider returned nil")
	}

	// Method chaining
	div = NewDivider()
	div.ForegroundColor(RGB(255, 0, 0))
	div.Char("-")

	// Children
	if div.Children() != nil {
		t.Error("Divider Children should return nil")
	}
}
