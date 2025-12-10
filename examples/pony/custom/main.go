package main

import (
	"fmt"
	"image/color"
	"os"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/pony"
	"github.com/charmbracelet/x/term"
)

func getSize() (int, int) {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 80, 24
	}
	return width, height
}

// Card is a custom component - a styled container with title and content.
type Card struct {
	pony.BaseElement
	Title   string
	Color   string
	Content []pony.Element
}

var _ pony.Element = (*Card)(nil)

// NewCard creates a card component (this is the factory function).
func NewCard(props pony.Props, children []pony.Element) pony.Element {
	title := props.Get("title")
	color := props.GetOr("color", "blue")

	return &Card{
		Title:   title,
		Color:   color,
		Content: children,
	}
}

// Draw renders the card - we compose from other pony elements!
func (c *Card) Draw(scr uv.Screen, area uv.Rectangle) {
	c.SetBounds(area)

	// Get color from name
	var themeColor color.Color
	switch c.Color {
	case "cyan":
		themeColor = pony.Hex("#00FFFF")
	case "green":
		themeColor = pony.Hex("#00FF00")
	case "red":
		themeColor = pony.Hex("#FF0000")
	case "blue":
		themeColor = pony.Hex("#0000FF")
	default:
		themeColor = pony.Hex("#0000FF")
	}

	// Build card structure using public pony elements - super easy!
	card := pony.NewBox(
		pony.NewVStack(
			// Title
			pony.NewText(c.Title).ForegroundColor(themeColor).Bold(),
			// Divider
			pony.NewDivider(),
			// Content
			pony.NewVStack(c.Content...),
		),
	).Border("rounded").
		BorderColor(themeColor).
		Padding(1)

	// Draw the composed element
	card.Draw(scr, area)
}

// Layout delegates to the composed structure.
func (c *Card) Layout(constraints pony.Constraints) pony.Size {
	var themeColor color.Color
	switch c.Color {
	case "cyan":
		themeColor = pony.Hex("#00FFFF")
	case "green":
		themeColor = pony.Hex("#00FF00")
	case "red":
		themeColor = pony.Hex("#FF0000")
	case "blue":
		themeColor = pony.Hex("#0000FF")
	default:
		themeColor = pony.Hex("#0000FF")
	}

	card := pony.NewBox(
		pony.NewVStack(
			pony.NewText(c.Title).ForegroundColor(themeColor).Bold(),
			pony.NewDivider(),
			pony.NewVStack(c.Content...),
		),
	).Border("rounded").BorderColor(themeColor).Padding(1)

	return card.Layout(constraints)
}

// Children returns the content elements.
func (c *Card) Children() []pony.Element {
	return c.Content
}

func main() {
	// Register our custom Card component
	pony.Register("card", NewCard)

	const tmpl = `
<vstack spacing="1">
	<box border="double" border-color="yellow" padding="1">
		<text font-weight="bold" foreground-color="yellow" alignment="center">🎨 Custom Components Made Easy!</text>
	</box>

	<divider foreground-color="gray" />

	<text font-weight="bold">Using Built-in Components:</text>
	<hstack spacing="2">
		<badge text="NEW" font-weight="bold" foreground-color="green" />
		<badge text="v2.0" font-weight="bold" foreground-color="blue" />
		<progress value="75" max="100" width="30" foreground-color="cyan" />
	</hstack>

	<divider foreground-color="gray" />

	<text font-weight="bold">Using Custom Card Component:</text>
	<hstack spacing="2">
		<card title="Profile" color="cyan">
			<text>Name: Alice</text>
			<text>Status: <badge text="Online" font-weight="bold" foreground-color="green" /></text>
			<text>Level: 42</text>
		</card>

		<card title="Stats" color="green">
			<text>Views: 1,234</text>
			<text>Likes: 567</text>
			<progress value="85" max="100" foreground-color="green" />
		</card>

		<card title="Alerts" color="red">
			<badge text="3" font-weight="bold" foreground-color="red" />
			<text>Warnings</text>
			<text font-style="italic" foreground-color="gray">Action needed</text>
		</card>
	</hstack>

	<divider foreground-color="gray" />

	<box border="rounded" border-color="magenta" padding="2">
		<vstack spacing="0">
			<text font-weight="bold" foreground-color="magenta" alignment="center">Why This Is Awesome:</text>
			<divider />
			<text>✓ Public fields - direct access</text>
			<text>✓ Constructors - NewText(), NewBox(), etc.</text>
			<text>✓ Fluent API - .WithStyle().WithAlign()</text>
			<text>✓ Easy composition - build complex from simple</text>
			<text>✓ No boilerplate - just implement Element interface</text>
		</vstack>
	</box>
</vstack>
`

	t := pony.MustParse[interface{}](tmpl)
	w, h := getSize()
	output := t.Render(nil, w, h)
	fmt.Print(output)
}
