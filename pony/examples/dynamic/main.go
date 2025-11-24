package main

import (
	"fmt"
	"os"
	"time"

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

type AppData struct {
	Title    string
	Username string
	Messages []string
	Status   string
	Count    int
	Time     string
}

func main() {
	const tmpl = `
<vstack gap="1">
	<box border="double" border-style="fg:cyan; bold">
		<text style="bold; fg:yellow">{{ upper .Title }}</text>
	</box>

	<hstack gap="2">
		<box border="rounded" border-style="fg:green">
			<vstack>
				<text style="bold">User Info:</text>
				<text>Name: {{ .Username }}</text>
				<text>Messages: {{ .Count }}</text>
				<text>Time: {{ .Time }}</text>
			</vstack>
		</box>

		<box border="rounded" border-style="fg:blue">
			<vstack>
				<text style="bold; fg:blue">Recent Messages:</text>
				{{ range .Messages }}
				<text style="fg:cyan">• {{ . }}</text>
				{{ end }}
			</vstack>
		</box>
	</hstack>

	<divider style="fg:gray" />

	{{ if eq .Status "online" }}
	<text style="fg:green; bold">Status: ● Online</text>
	{{ else }}
	<text style="fg:red; bold">Status: ○ Offline</text>
	{{ end }}

	<box border="normal" border-style="fg:magenta">
		<text style="italic">{{ printf "Generated at %s with pony" .Time }}</text>
	</box>
</vstack>
`

	data := AppData{
		Title:    "Dynamic Dashboard",
		Username: "Alice",
		Messages: []string{
			"Welcome to pony!",
			"Templates are working great",
			"This is powered by UV",
		},
		Status: "online",
		Count:  42,
		Time:   time.Now().Format("15:04:05"),
	}

	t := pony.MustParse[AppData](tmpl)
	w, h := getSize()
	output := t.Render(data, w, h)
	fmt.Print(output)
}
