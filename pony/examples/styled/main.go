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
	<box border="double" border-color="cyan">
		<text font-weight="bold" foreground-color="yellow">Styled pony Showcase</text>
	</box>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Color Examples:</text>
		<text foreground-color="red">Red text</text>
		<text foreground-color="green">Green text</text>
		<text foreground-color="blue">Blue text</text>
		<text foreground-color="cyan">Cyan text</text>
		<text foreground-color="magenta">Magenta text</text>
		<text foreground-color="yellow">Yellow text</text>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Hex Colors:</text>
		<text foreground-color="#FF5555">Bright Red (#FF5555)</text>
		<text foreground-color="#50FA7B">Bright Green (#50FA7B)</text>
		<text foreground-color="#8BE9FD">Bright Cyan (#8BE9FD)</text>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Text Attributes:</text>
		<text font-weight="bold">Bold text</text>
		<text font-style="italic">Italic text</text>
		<text text-decoration="underline">Underlined text</text>
		<text text-decoration="strikethrough">Strikethrough text</text>
		<text font-weight="bold" font-style="italic" foreground-color="magenta">Combined: bold + italic + magenta</text>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="0">
		<text font-weight="bold">Background Colors:</text>
		<text background-color="red" foreground-color="white">White on Red</text>
		<text background-color="green" foreground-color="black">Black on Green</text>
		<text background-color="blue" foreground-color="white">White on Blue</text>
	</vstack>

	<divider foreground-color="gray" />

	<hstack spacing="2">
		<box border="normal" border-color="red">
			<text foreground-color="red" font-weight="bold">Red Box</text>
		</box>
		<box border="rounded" border-color="green">
			<text foreground-color="green" font-weight="bold">Green Box</text>
		</box>
		<box border="thick" border-color="blue">
			<text foreground-color="blue" font-weight="bold">Blue Box</text>
		</box>
	</hstack>
</vstack>
`

	t := pony.MustParse[any](tmpl)
	w, h := getSize()
	output := t.Render(nil, w, h)
	fmt.Print(output)
}
