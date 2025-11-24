package main

import (
	"fmt"
	"os"

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


func main() {
	const tmpl = `
<vstack gap="1">
	<header text="Custom Components Demo" style="fg:cyan; bold" />

	<vstack gap="0">
		<text style="bold">Built-in Custom Components:</text>
		<hstack gap="1">
			<badge text="NEW" style="fg:green; bold" />
			<badge text="BETA" style="fg:yellow; bold" />
			<badge text="DEPRECATED" style="fg:red; bold" />
		</hstack>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Progress Bars:</text>
		<hstack gap="1">
			<text>0%</text>
			<progress value="0" max="100" width="30" />
		</hstack>
		<hstack gap="1">
			<text>25%</text>
			<progress value="25" max="100" width="30" style="fg:red" />
		</hstack>
		<hstack gap="1">
			<text>50%</text>
			<progress value="50" max="100" width="30" style="fg:yellow" />
		</hstack>
		<hstack gap="1">
			<text>75%</text>
			<progress value="75" max="100" width="30" style="fg:blue" />
		</hstack>
		<hstack gap="1">
			<text>100%</text>
			<progress value="100" max="100" width="30" style="fg:green" />
		</hstack>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="1">
		<header text="Header with Border" level="1" style="fg:magenta" />
		<text>This header has an underline</text>
		
		<header text="Header without Border" border="false" style="fg:blue; bold" />
		<text>This header has no underline</text>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Status Indicators:</text>
		<hstack gap="2">
			<badge text="✓ Success" style="fg:green" />
			<badge text="⚠ Warning" style="fg:yellow" />
			<badge text="✗ Error" style="fg:red" />
			<badge text="ℹ Info" style="fg:blue" />
		</hstack>
	</vstack>

	<divider style="fg:gray" />

	<box border="rounded" border-style="fg:cyan" padding="1">
		<vstack gap="0">
			<text style="bold; fg:cyan">About Custom Components:</text>
			<text>• Components are registered via pony.Register()</text>
			<text>• Built-ins: badge, progress, header</text>
			<text>• Easy to create your own!</text>
		</vstack>
	</box>
</vstack>
`

	t := pony.MustParse[any](tmpl)
	w, h := getSize()
	output := t.Render(nil, w, h)
	fmt.Print(output)
}
