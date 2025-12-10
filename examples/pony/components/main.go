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
<vstack spacing="1">
	<text font-weight="bold" foreground-color="cyan">Custom Components Demo</text>
	<divider foreground-color="cyan" />

	<vstack spacing="0">
		<text font-weight="bold">Built-in Custom Components:</text>
		<hstack spacing="1">
			<badge text="NEW" foreground-color="green" />
			<badge text="BETA" foreground-color="yellow" />
			<badge text="DEPRECATED" foreground-color="red" />
		</hstack>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Progress Bars:</text>
		<hstack spacing="1">
			<text>0%</text>
			<progressview value="0" max="100" width="30" />
		</hstack>
		<hstack spacing="1">
			<text>25%</text>
			<progressview value="25" max="100" width="30" foreground-color="red" />
		</hstack>
		<hstack spacing="1">
			<text>50%</text>
			<progressview value="50" max="100" width="30" foreground-color="yellow" />
		</hstack>
		<hstack spacing="1">
			<text>75%</text>
			<progressview value="75" max="100" width="30" foreground-color="blue" />
		</hstack>
		<hstack spacing="1">
			<text>100%</text>
			<progressview value="100" max="100" width="30" foreground-color="green" />
		</hstack>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Status Indicators:</text>
		<hstack spacing="2">
			<badge text="✓ Success" foreground-color="green" />
			<badge text="⚠ Warning" foreground-color="yellow" />
			<badge text="✗ Error" foreground-color="red" />
			<badge text="ℹ Info" foreground-color="blue" />
		</hstack>
	</vstack>

	<divider foreground-color="gray" />

	<box border="rounded" border-color="cyan" padding="1">
		<vstack spacing="0">
			<text font-weight="bold" foreground-color="cyan">About Custom Components:</text>
			<text>• Components are registered via pony.Register()</text>
			<text>• Built-ins: badge, progressview</text>
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
