package main

import (
	"fmt"
	"image/color"
	"os"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/pony"
)

func getSize() (int, int) {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 80, 24
	}
	return width, height
}

// Card is a custom component - a styled container with title and content
type Card struct {
	pony.BaseElement
	Title   string
	Color   string
	Content []pony.Element
}

var _ pony.Element = (*Card)(nil)

// NewCard creates a card component (this is the factory function)
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

	// Build styles using helpers - no string parsing!
	titleStyle := pony.NewStyle().Fg(themeColor).Bold().Build()
	borderStyle := pony.NewStyle().Fg(themeColor).Build()

	// Build card structure using public pony elements - super easy!
	card := pony.NewBox(
		pony.NewVStack(
			// Title
			&pony.Text{
				Content: c.Title,
				Style:   titleStyle,
			},
			// Divider
			pony.NewDivider(),
			// Content
			pony.NewVStack(c.Content...),
		),
	).WithBorder("rounded").
		WithBorderStyle(borderStyle).
		WithPadding(1)

	// Draw the composed element
	card.Draw(scr, area)
}

// Layout delegates to the composed structure
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

	titleStyle := pony.NewStyle().Fg(themeColor).Bold().Build()

	card := pony.NewBox(
		pony.NewVStack(
			&pony.Text{Content: c.Title, Style: titleStyle},
			pony.NewDivider(),
			pony.NewVStack(c.Content...),
		),
	).WithBorder("rounded").WithPadding(1)

	return card.Layout(constraints)
}

// Children returns the content elements
func (c *Card) Children() []pony.Element {
	return c.Content
}

func main() {
	// Register our custom Card component
	pony.Register("card", NewCard)

	const tmpl = `
<vstack gap="1">
	<box border="double" border-style="fg:yellow; bold" padding="1">
		<text style="bold; fg:yellow" align="center">🎨 Custom Components Made Easy!</text>
	</box>

	<divider style="fg:gray" />

	<text style="bold">Using Built-in Components:</text>
	<hstack gap="2">
		<badge text="NEW" style="fg:green; bold" />
		<badge text="v2.0" style="fg:blue; bold" />
		<progress value="75" max="100" width="30" style="fg:cyan" />
	</hstack>

	<divider style="fg:gray" />

	<text style="bold">Using Custom Card Component:</text>
	<hstack gap="2">
		<card title="Profile" color="cyan">
			<text>Name: Alice</text>
			<text>Status: <badge text="Online" style="fg:green; bold" /></text>
			<text>Level: 42</text>
		</card>

		<card title="Stats" color="green">
			<text>Views: 1,234</text>
			<text>Likes: 567</text>
			<progress value="85" max="100" style="fg:green" />
		</card>

		<card title="Alerts" color="red">
			<badge text="3" style="fg:red; bold" />
			<text>Warnings</text>
			<text style="italic; fg:gray">Action needed</text>
		</card>
	</hstack>

	<divider style="fg:gray" />

	<box border="rounded" border-style="fg:magenta" padding="2">
		<vstack gap="0">
			<text style="bold; fg:magenta" align="center">Why This Is Awesome:</text>
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
