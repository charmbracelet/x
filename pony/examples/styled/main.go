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
		<text style="bold; fg:yellow">Styled pony Showcase</text>
	</box>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Color Examples:</text>
		<text style="fg:red">Red text</text>
		<text style="fg:green">Green text</text>
		<text style="fg:blue">Blue text</text>
		<text style="fg:cyan">Cyan text</text>
		<text style="fg:magenta">Magenta text</text>
		<text style="fg:yellow">Yellow text</text>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Hex Colors:</text>
		<text style="fg:#FF5555">Bright Red (#FF5555)</text>
		<text style="fg:#50FA7B">Bright Green (#50FA7B)</text>
		<text style="fg:#8BE9FD">Bright Cyan (#8BE9FD)</text>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Text Attributes:</text>
		<text style="bold">Bold text</text>
		<text style="italic">Italic text</text>
		<text style="underline">Underlined text</text>
		<text style="strikethrough">Strikethrough text</text>
		<text style="bold; italic; fg:magenta">Combined: bold + italic + magenta</text>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Background Colors:</text>
		<text style="bg:red; fg:white">White on Red</text>
		<text style="bg:green; fg:black">Black on Green</text>
		<text style="bg:blue; fg:white">White on Blue</text>
	</vstack>

	<divider style="fg:gray" />

	<hstack gap="2">
		<box border="normal" border-style="fg:red">
			<text style="fg:red; bold">Red Box</text>
		</box>
		<box border="rounded" border-style="fg:green">
			<text style="fg:green; bold">Green Box</text>
		</box>
		<box border="thick" border-style="fg:blue">
			<text style="fg:blue; bold">Blue Box</text>
		</box>
	</hstack>
</vstack>
`

	t := pony.MustParse[any](tmpl)
	w, h := getSize()
	output := t.Render(nil, w, h)
	fmt.Print(output)
}
