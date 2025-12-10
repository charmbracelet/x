// Package main example.
package main

import (
	"fmt"
	"os"

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

func main() {
	const tmpl = `
<vstack spacing="1">
	<box border="double">
		<text>pony Layout Showcase</text>
	</box>

	<divider />

	<vstack spacing="0">
		<text>Vertical Stack Demo:</text>
		<box border="normal">
			<text>Item 1</text>
		</box>
		<box border="normal">
			<text>Item 2</text>
		</box>
		<box border="normal">
			<text>Item 3</text>
		</box>
	</vstack>

	<divider />

	<vstack spacing="0">
		<text>Horizontal Stack Demo:</text>
		<hstack spacing="2">
			<box border="rounded">
				<text>Left Box</text>
			</box>
			<box border="rounded">
				<text>Middle Box</text>
			</box>
			<box border="rounded">
				<text>Right Box</text>
			</box>
		</hstack>
	</vstack>

	<divider />

	<vstack spacing="0">
		<text>Width/Height Attributes Demo:</text>
		<hstack spacing="1">
			<box border="normal" width="30%">
				<text>30% width</text>
			</box>
			<box border="normal" width="70%">
				<text>70% width</text>
			</box>
		</hstack>
	</vstack>

	<divider />

	<vstack spacing="0">
		<text>Fixed Size Demo:</text>
		<hstack spacing="1">
			<box border="normal" width="20">
				<text>Fixed 20</text>
			</box>
			<box border="normal" width="30">
				<text>Fixed 30 cells wide</text>
			</box>
		</hstack>
	</vstack>

	<divider />

	<vstack spacing="0">
		<text>Nested Layout Demo:</text>
		<box border="thick">
			<vstack spacing="1">
				<hstack spacing="1">
					<box border="normal">
						<text>A</text>
					</box>
					<box border="normal">
						<text>B</text>
					</box>
				</hstack>
				<hstack spacing="1">
					<box border="normal">
						<text>C</text>
					</box>
					<box border="normal">
						<text>D</text>
					</box>
				</hstack>
			</vstack>
		</box>
	</vstack>

	<divider />

	<text>Border Styles:</text>
	<hstack spacing="1">
		<box border="normal">
			<text>Normal</text>
		</box>
		<box border="rounded">
			<text>Rounded</text>
		</box>
		<box border="thick">
			<text>Thick</text>
		</box>
		<box border="double">
			<text>Double</text>
		</box>
	</hstack>
</vstack>
`

	t := pony.MustParse[any](tmpl)
	w, h := getSize()
	output := t.Render(nil, w, h)
	fmt.Print(output)
}
