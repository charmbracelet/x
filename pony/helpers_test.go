package pony

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/golden"
)

func TestStyleBuilder(t *testing.T) {
	// Test fluent API
	style := NewStyle().
		Fg(RGB(255, 0, 0)).
		Bg(RGB(0, 0, 0)).
		Bold().
		Italic().
		Build()

	if style.Fg == nil {
		t.Error("Foreground color not set")
	}
	if style.Bg == nil {
		t.Error("Background color not set")
	}
	if style.Attrs&uv.AttrBold == 0 {
		t.Error("Bold attribute not set")
	}
	if style.Attrs&uv.AttrItalic == 0 {
		t.Error("Italic attribute not set")
	}
}

func TestStyleBuilderUnderline(t *testing.T) {
	style := NewStyle().Underline().Build()
	if style.Underline != uv.UnderlineSingle {
		t.Error("Underline not set")
	}

	style2 := NewStyle().UnderlineStyle(uv.UnderlineCurly).Build()
	if style2.Underline != uv.UnderlineCurly {
		t.Error("Curly underline not set")
	}
}

func TestColorHelpers(t *testing.T) {
	// Test Hex
	c := Hex("#FF0000")
	if c == nil {
		t.Error("Hex() returned nil")
	}

	// Test HexSafe
	c2, err := HexSafe("#00FF00")
	if err != nil || c2 == nil {
		t.Error("HexSafe() failed")
	}

	_, err = HexSafe("invalid")
	if err == nil {
		t.Error("HexSafe() should error on invalid hex")
	}

	// Test RGB
	c3 := RGB(255, 128, 0)
	if c3 == nil {
		t.Error("RGB() returned nil")
	}
}

func TestLayoutHelpers(t *testing.T) {
	// Test Panel
	content := NewText("Content")
	panel := Panel(content, BorderRounded, 1)

	// Panel creates a Box - verify it works
	if panel == nil {
		t.Error("Panel should not be nil")
	}

	// Test Separated
	sep := Separated(
		NewText("A"),
		NewText("B"),
		NewText("C"),
	)

	vstack, ok := sep.(*VStack)
	if !ok {
		t.Error("Separated should return VStack")
	}

	// Should have 5 items: A, divider, B, divider, C
	if len(vstack.Children()) != 5 {
		t.Errorf("Separated should have 5 items (3 + 2 dividers), got %d", len(vstack.Children()))
	}
}

func TestSeparatedEmpty(t *testing.T) {
	sep := Separated()
	if sep == nil {
		t.Error("Separated() with no children should return empty VStack")
	}
}

func TestCard(t *testing.T) {
	titleColor := RGB(255, 255, 0)  // yellow
	borderColor := RGB(0, 255, 255) // cyan

	card := Card("Title", titleColor, borderColor,
		NewText("Line 1"),
		NewText("Line 2"),
	)

	if card == nil {
		t.Error("Card() returned nil")
	}

	// Card should be a Box
	box, ok := card.(*Box)
	if !ok {
		t.Error("Card should return a Box")
	}

	// Card creates proper box - verified by golden tests
	if box == nil {
		t.Error("Card should not return nil")
	}
}

func TestSection(t *testing.T) {
	headerColor := RGB(0, 255, 255) // cyan
	section := Section("Header", headerColor,
		NewText("Content 1"),
		NewText("Content 2"),
	)

	vstack, ok := section.(*VStack)
	if !ok {
		t.Error("Section should return VStack")
	}

	if len(vstack.Children()) < 2 {
		t.Error("Section should have header + content")
	}
}

func TestHelperRoundTrip(t *testing.T) {
	const markup = `<slot name="content" />`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	elem := NewText("Test").
		ForegroundColor(Hex("#FF5555")).
		Bold().
		Italic().
		Alignment("center")

	slots := map[string]Element{
		"content": elem,
	}

	output := tmpl.RenderWithSlots(nil, slots, 80, 24)
	golden.RequireEqual(t, output)
}

// Test additional StyleBuilder methods.
func TestStyleBuilderAdditionalMethods(t *testing.T) {
	style := NewStyle().
		UnderlineColor(RGB(255, 0, 0)).
		Faint().
		Blink().
		Reverse().
		Strikethrough().
		Build()

	if style.UnderlineColor == nil {
		t.Error("UnderlineColor not set")
	}
	if style.Attrs&uv.AttrFaint == 0 {
		t.Error("Faint not set")
	}
	if style.Attrs&uv.AttrBlink == 0 {
		t.Error("Blink not set")
	}
	if style.Attrs&uv.AttrReverse == 0 {
		t.Error("Reverse not set")
	}
	if style.Attrs&uv.AttrStrikethrough == 0 {
		t.Error("Strikethrough not set")
	}
}

// Test Hex panic.
func TestHexPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Hex should panic on invalid hex")
		}
	}()

	Hex("invalid")
}

// Test additional helper functions.
func TestAdditionalHelpers(t *testing.T) {
	// PanelWithMargin
	panel := PanelWithMargin(NewText("test"), BorderRounded, 1, 2)
	if panel == nil {
		t.Error("PanelWithMargin should not be nil")
	}

	// Overlay
	overlay := Overlay(NewText("a"), NewText("b"))
	if zstack, ok := overlay.(*ZStack); !ok || len(zstack.Children()) != 2 {
		t.Error("Overlay should return ZStack with items")
	}

	// FlexGrow
	flex := FlexGrow(NewText("test"), 2)
	if flex == nil {
		t.Error("FlexGrow should not be nil")
	}

	// Position
	pos := Position(NewText("test"), 5, 10)
	if pos.x != 5 || pos.y != 10 {
		t.Error("Position not set correctly")
	}

	// PositionRight
	pos = PositionRight(NewText("test"), 5, 10)
	if pos.right != 5 || pos.y != 10 {
		t.Error("PositionRight not set correctly")
	}

	// PositionBottom
	pos = PositionBottom(NewText("test"), 5, 10)
	if pos.x != 5 || pos.bottom != 10 {
		t.Error("PositionBottom not set correctly")
	}

	// PositionCorner
	pos = PositionCorner(NewText("test"), 5, 10)
	if pos.right != 5 || pos.bottom != 10 {
		t.Error("PositionCorner not set correctly")
	}
}
