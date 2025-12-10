package main

import (
	"fmt"

	"github.com/charmbracelet/x/pony"
)

func main() {
	const tmpl = `
<vstack spacing="1">
	<box border="double" border-color="yellow" padding="1">
		<text font-weight="bold" foreground-color="yellow" alignment="center">🎨 Style Helpers Demo</text>
	</box>

	<divider foreground-color="gray" />

	<text font-weight="bold">Using helpers to build styled elements:</text>
	<slot name="styledContent" />

	<divider foreground-color="gray" />

	<text font-weight="bold">Using layout helpers:</text>
	<slot name="panel" />
	<slot name="sections" />
</vstack>
`

	t := pony.MustParse[any](tmpl)

	// Use granular text modifiers - type-safe and SwiftUI-like!
	slots := map[string]pony.Element{
		"styledContent": pony.NewVStack(
			pony.NewText("Error: Something went wrong").ForegroundColor(pony.Hex("#FF5555")).Bold(),
			pony.NewText("Warning: Check this out").ForegroundColor(pony.Hex("#FFFF55")).Bold(),
			pony.NewText("Success: All good!").ForegroundColor(pony.Hex("#50FA7B")).Bold(),
			pony.NewText("Muted: Less important info").ForegroundColor(pony.RGB(128, 128, 128)).Italic(),
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
