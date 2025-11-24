package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/pony"
)

func main() {
	const tmpl = `
<vstack gap="1">
	<box border="rounded">
		<text>Hello, World!</text>
	</box>
	<text>Welcome to pony - a declarative markup language for terminal UIs.</text>
	<divider />
	<hstack gap="2">
		<text>Left</text>
		<text>Right</text>
	</hstack>
</vstack>
`

	// Get terminal size
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width, height = 80, 24 // fallback
	}

	t := pony.MustParse[any](tmpl)
	output := t.Render(nil, width, height)
	fmt.Print(output)
}
