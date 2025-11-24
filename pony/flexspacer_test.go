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
	if spacer.Size != 0 {
		t.Error("NewSpacer should have size 0")
	}

	// NewFixedSpacer
	fixedSpacer := NewFixedSpacer(10)
	if fixedSpacer == nil {
		t.Fatal("NewFixedSpacer returned nil")
	}
	if fixedSpacer.Size != 10 {
		t.Error("NewFixedSpacer size not set correctly")
	}

	// WithSize
	spacer.WithSize(5)
	if spacer.Size != 5 {
		t.Error("WithSize not set")
	}

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
	if flex.Child == nil {
		t.Error("NewFlex child not set")
	}

	// WithGrow
	flex.WithGrow(2)
	if flex.Grow != 2 {
		t.Error("WithGrow not set")
	}

	// WithShrink
	flex.WithShrink(0)
	if flex.Shrink != 0 {
		t.Error("WithShrink not set")
	}

	// WithBasis
	flex.WithBasis(10)
	if flex.Basis != 10 {
		t.Error("WithBasis not set")
	}

	// Children
	children := flex.Children()
	if len(children) != 1 {
		t.Error("Flex Children should return child")
	}

	// Test with nil child
	flexNil := &Flex{Child: nil}
	if flexNil.Children() != nil {
		t.Error("Flex Children with nil child should return nil")
	}
}

// Test flex helper functions.
func TestFlexHelpers(t *testing.T) {
	// GetFlexShrink
	flex := &Flex{Child: NewText("test"), Shrink: 2}
	if GetFlexShrink(flex) != 2 {
		t.Error("GetFlexShrink failed")
	}

	text := NewText("test")
	if GetFlexShrink(text) != 1 {
		t.Error("GetFlexShrink should return 1 for non-flex")
	}

	// GetFlexBasis
	flex.Basis = 10
	if GetFlexBasis(flex) != 10 {
		t.Error("GetFlexBasis failed")
	}

	if GetFlexBasis(text) != 0 {
		t.Error("GetFlexBasis should return 0 for non-flex")
	}

	// IsFlexible
	flex.Grow = 1
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
	if !div.Vertical {
		t.Error("NewVerticalDivider should be vertical")
	}

	// WithStyle
	div = NewDivider()
	div.WithStyle(uv.Style{Attrs: uv.AttrBold})
	if div.Style.Attrs != uv.AttrBold {
		t.Error("WithStyle not set")
	}

	// WithChar
	div.WithChar("-")
	if div.Char != "-" {
		t.Error("WithChar not set")
	}

	// Children
	if div.Children() != nil {
		t.Error("Divider Children should return nil")
	}
}
