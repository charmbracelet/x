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
	<box border="double" border-color="cyan">
		<text font-weight="bold" foreground-color="yellow" alignment="center">Alignment & Padding Demo</text>
	</box>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Text Alignment:</text>
		<box border="normal">
			<text alignment="left">Left aligned text</text>
		</box>
		<box border="normal">
			<text alignment="center">Center aligned text</text>
		</box>
		<box border="normal">
			<text alignment="right">Right aligned text</text>
		</box>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Padding Demo:</text>
		<box border="rounded" padding="0">
			<text>No padding</text>
		</box>
		<box border="rounded" padding="1">
			<text>Padding: 1</text>
		</box>
		<box border="rounded" padding="2">
			<text>Padding: 2</text>
		</box>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">VStack Alignment (children):</text>
		<box border="normal">
			<vstack alignment="left">
				<text>Left</text>
				<text>Aligned</text>
			</vstack>
		</box>
		<box border="normal">
			<vstack alignment="center">
				<text>Center</text>
				<text>Aligned</text>
			</vstack>
		</box>
		<box border="normal">
			<vstack alignment="right">
				<text>Right</text>
				<text>Aligned</text>
			</vstack>
		</box>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">HStack Vertical Alignment:</text>
		<box border="normal" height="5">
			<hstack spacing="2" alignment="top">
				<text>Top</text>
				<text>Aligned</text>
			</hstack>
		</box>
		<box border="normal" height="5">
			<hstack spacing="2" alignment="middle">
				<text>Middle</text>
				<text>Aligned</text>
			</hstack>
		</box>
		<box border="normal" height="5">
			<hstack spacing="2" alignment="bottom">
				<text>Bottom</text>
				<text>Aligned</text>
			</hstack>
		</box>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Combined: Width + Padding + Alignment:</text>
		<hstack spacing="1">
			<box border="rounded" border-color="red" width="33%" padding="1">
				<text alignment="center" font-weight="bold" foreground-color="red">Left</text>
			</box>
			<box border="rounded" border-color="green" width="33%" padding="1">
				<text alignment="center" font-weight="bold" foreground-color="green">Center</text>
			</box>
			<box border="rounded" border-color="blue" width="33%" padding="1">
				<text alignment="center" font-weight="bold" foreground-color="blue">Right</text>
			</box>
		</hstack>
	</vstack>
</vstack>
`

	t := pony.MustParse[any](tmpl)
	w, h := getSize()
	output := t.Render(nil, w, h)
	fmt.Print(output)
}
