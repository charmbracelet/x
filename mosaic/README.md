# Mosaic

Mosaic is a tool that allows you to display images in your terminal programs. It
will break down your image to contain a certain number of pixels per cell, then
render those cells. This works best with monospaced fonts.

> ![NOTE] 
> We will be providing a more full-fledged implementation of image
> support for Bubble Tea, but this package is one step in that direction.

To use Mosaic, you need to...

1. Open an image file e.g. `f, err := os.Open(path)`
2. Decode the image e.g. `img, err := jpeg.Decode(f)`
3. Create a new Mosaic renderer e.g. `m := mosaic.New().Width(80).Height(40)`
4. Render the image with Mosaic! e.g. `m.Render(dogImg)`

Here's a full-blown example:

``` go
package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/mosaic"
)

func main() {
	dogImg, err := loadImage("./pekinas.jpg")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	m := mosaic.New().Width(80).Height(40)

	fmt.Println(lipgloss.JoinVertical(lipgloss.Right, lipgloss.JoinHorizontal(lipgloss.Center, m.Render(dogImg))))
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	defer f.Close() //nolint:errcheck
	if err != nil {
		return nil, err
	}
	return jpeg.Decode(f)
}
```

Check out all of the mosaic [examples](https://github.com/charmbracelet/x/tree/main/examples/mosaic)!

## Feedback

We'd love to hear your thoughts on this project. Feel free to drop us a note!

- [Twitter](https://twitter.com/charmcli)
- [The Fediverse](https://mastodon.social/@charmcli)
- [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/charmbracelet/x/raw/main/LICENSE)

---

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source • نحنُ نحب المصادر المفتوحة
