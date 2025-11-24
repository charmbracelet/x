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
	<box border="double" border-style="fg:cyan; bold">
		<text style="bold; fg:yellow" align="center">Alignment & Padding Demo</text>
	</box>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Text Alignment:</text>
		<box border="normal">
			<text align="left">Left aligned text</text>
		</box>
		<box border="normal">
			<text align="center">Center aligned text</text>
		</box>
		<box border="normal">
			<text align="right">Right aligned text</text>
		</box>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Padding Demo:</text>
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

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">VStack Alignment (children):</text>
		<box border="normal">
			<vstack align="left">
				<text>Left</text>
				<text>Aligned</text>
			</vstack>
		</box>
		<box border="normal">
			<vstack align="center">
				<text>Center</text>
				<text>Aligned</text>
			</vstack>
		</box>
		<box border="normal">
			<vstack align="right">
				<text>Right</text>
				<text>Aligned</text>
			</vstack>
		</box>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">HStack Vertical Alignment:</text>
		<box border="normal" height="5">
			<hstack gap="2" valign="top">
				<text>Top</text>
				<text>Aligned</text>
			</hstack>
		</box>
		<box border="normal" height="5">
			<hstack gap="2" valign="middle">
				<text>Middle</text>
				<text>Aligned</text>
			</hstack>
		</box>
		<box border="normal" height="5">
			<hstack gap="2" valign="bottom">
				<text>Bottom</text>
				<text>Aligned</text>
			</hstack>
		</box>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Combined: Width + Padding + Alignment:</text>
		<hstack gap="1">
			<box border="rounded" border-style="fg:red" width="33%" padding="1">
				<text align="center" style="fg:red; bold">Left</text>
			</box>
			<box border="rounded" border-style="fg:green" width="33%" padding="1">
				<text align="center" style="fg:green; bold">Center</text>
			</box>
			<box border="rounded" border-style="fg:blue" width="33%" padding="1">
				<text align="center" style="fg:blue; bold">Right</text>
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
