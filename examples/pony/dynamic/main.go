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
<vstack spacing="1">
	<box border="double" border-color="cyan">
		<text font-weight="bold" foreground-color="yellow">{{ upper .Title }}</text>
	</box>

	<hstack spacing="2">
		<box border="rounded" border-color="green">
			<vstack>
				<text font-weight="bold">User Info:</text>
				<text>Name: {{ .Username }}</text>
				<text>Messages: {{ .Count }}</text>
				<text>Time: {{ .Time }}</text>
			</vstack>
		</box>

		<box border="rounded" border-color="blue">
			<vstack>
				<text font-weight="bold" foreground-color="blue">Recent Messages:</text>
				{{ range .Messages }}
				<text foreground-color="cyan">• {{ . }}</text>
				{{ end }}
			</vstack>
		</box>
	</hstack>

	<divider foreground-color="gray" />

	{{ if eq .Status "online" }}
	<text font-weight="bold" foreground-color="green">Status: ● Online</text>
	{{ else }}
	<text font-weight="bold" foreground-color="red">Status: ○ Offline</text>
	{{ end }}

	<box border="normal" border-color="magenta">
		<text font-style="italic">{{ printf "Generated at %s with pony" .Time }}</text>
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
