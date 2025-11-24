package pony

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestSlotBasic(t *testing.T) {
	const markup = `<slot name="content" />`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	t.Run("without slot", func(t *testing.T) {
		output := tmpl.Render(nil, 80, 24)
		golden.RequireEqual(t, output)
	})

	t.Run("with slot filled", func(t *testing.T) {
		slots := map[string]Element{
			"content": NewText("Slot Content"),
		}

		output := tmpl.RenderWithSlots(nil, slots, 80, 24)
		golden.RequireEqual(t, output)
	})
}

func TestMultipleSlots(t *testing.T) {
	const markup = `
<vstack spacing="1">
	<slot name="header" />
	<slot name="body" />
	<slot name="footer" />
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	slots := map[string]Element{
		"header": NewText("Header Content"),
		"body":   NewText("Body Content"),
		"footer": NewText("Footer Content"),
	}

	output := tmpl.RenderWithSlots(nil, slots, 80, 24)
	golden.RequireEqual(t, output)
}

func TestSlotInBox(t *testing.T) {
	const markup = `<box border="rounded"><slot name="content" /></box>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	slots := map[string]Element{
		"content": NewText("Test"),
	}

	output := tmpl.RenderWithSlots(nil, slots, 20, 5)
	golden.RequireEqual(t, output)
}

func TestSlotWithComplexElement(t *testing.T) {
	const markup = `<slot name="widget" />`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	complexElement := NewBox(
		NewVStack(
			NewText("Title").Bold(),
			NewDivider(),
			NewText("Content"),
		),
	).Border("rounded")

	slots := map[string]Element{
		"widget": complexElement,
	}

	output := tmpl.RenderWithSlots(nil, slots, 80, 24)
	golden.RequireEqual(t, output)
}

func TestSlotMissing(t *testing.T) {
	const markup = `
<vstack>
	<text>Before</text>
	<slot name="missing" />
	<text>After</text>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestSlotsWithTemplateData(t *testing.T) {
	const markup = `
<vstack spacing="1">
	<text>{{ .Title }}</text>
	<slot name="content" />
	<text>{{ .Footer }}</text>
</vstack>
`

	type Data struct {
		Title  string
		Footer string
	}

	tmpl, err := Parse[Data](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	data := Data{
		Title:  "Header from data",
		Footer: "Footer from data",
	}

	slots := map[string]Element{
		"content": NewText("Slot from element"),
	}

	output := tmpl.RenderWithSlots(data, slots, 80, 24)
	golden.RequireEqual(t, output)
}

// Test Slot Children with nil element.
func TestSlotChildrenNil(t *testing.T) {
	slot := NewSlot("test")
	if slot.Children() != nil {
		t.Error("Slot Children with nil element should return nil")
	}
}
