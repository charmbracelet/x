package main

import (
	"fmt"

	"github.com/charmbracelet/x/pony"
)

func main() {
	const tmpl = `
<vstack gap="1">
	<box border="double" border-style="fg:yellow; bold" padding="1">
		<text style="bold; fg:yellow" align="center">🎨 Style Helpers Demo</text>
	</box>

	<divider style="fg:gray" />

	<text style="bold">Using helpers to build styled elements:</text>
	<slot name="styledContent" />

	<divider style="fg:gray" />

	<text style="bold">Using layout helpers:</text>
	<slot name="panel" />
	<slot name="sections" />
</vstack>
`

	t := pony.MustParse[any](tmpl)

	// Build styles using StyleBuilder - type-safe!
	errorStyle := pony.NewStyle().
		Fg(pony.Hex("#FF5555")).
		Bold().
		Build()
	
	warningStyle := pony.NewStyle().
		Fg(pony.Hex("#FFFF55")).
		Bold().
		Build()
	
	successStyle := pony.NewStyle().
		Fg(pony.Hex("#50FA7B")).
		Bold().
		Build()
	
	mutedStyle := pony.NewStyle().
		Fg(pony.RGB(128, 128, 128)).
		Italic().
		Build()

	// Use helpers to build elements
	slots := map[string]pony.Element{
		"styledContent": pony.NewVStack(
			&pony.Text{Content: "Error: Something went wrong", Style: errorStyle},
			&pony.Text{Content: "Warning: Check this out", Style: warningStyle},
			&pony.Text{Content: "Success: All good!", Style: successStyle},
			&pony.Text{Content: "Muted: Less important info", Style: mutedStyle},
		),
		
		// Use Panel helper
		"panel": pony.Panel(
			pony.NewText("This is inside a panel"),
			"rounded",
			2, // padding
		),
		
		// Use Separated helper
		"sections": pony.Separated(
			pony.NewText("Section 1"),
			pony.NewText("Section 2"),
			pony.NewText("Section 3"),
		),
	}

	output := t.RenderWithSlots(nil, slots, 90, 40)
	fmt.Print(output)
	
	fmt.Println("\n\n✨ Benefits of Helpers:")
	fmt.Println("  • Type-safe (no string parsing)")
	fmt.Println("  • IDE autocomplete")
	fmt.Println("  • Compile-time checking")
	fmt.Println("  • Reusable style objects")
	fmt.Println("  • Cleaner code")
}
