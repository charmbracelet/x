package charmtone

import (
	"image/color"

	"github.com/charmbracelet/lipgloss/v2"
)

// Key is a type for color keys.
type Key string

const (
	Cumin     Key = "cumin"
	Tang      Key = "tang"
	Yam       Key = "yam"
	Paprika   Key = "paprika"
	Pumpkin   Key = "pumpkin"
	Uni       Key = "uni"
	Sriracha  Key = "sriracha"
	Coral     Key = "coral"
	Salmon    Key = "salmon"
	Chili     Key = "chili"
	Cherry    Key = "cherry"
	Tuna      Key = "tuna"
	Macaroon  Key = "macaroon"
	Rose      Key = "rose"
	Cheeky    Key = "cheeky"
	Flamingo  Key = "flamingo"
	Mollie    Key = "mollie"
	Blush     Key = "blush"
	Urchin    Key = "urchin"
	Crystal   Key = "crystal"
	Lilac     Key = "lilac"
	Eggplant  Key = "eggplant"
	Violet    Key = "violet"
	Mauve     Key = "mauve"
	Grape     Key = "grape"
	Plum      Key = "plum"
	Orchid    Key = "orchid"
	Fig       Key = "fig"
	Charple   Key = "charple"
	Hazy      Key = "hazy"
	Blueberry Key = "blueberry"
	Sapphire  Key = "sapphire"
	Guppy     Key = "guppy"
	Neptune   Key = "neptune"
	Electric  Key = "electric"
	Anchovy   Key = "anchovy"
	Damson    Key = "damson"
	Malibu    Key = "malibu"
	Sardine   Key = "sardine"
	Zinc      Key = "zinc"
	Turtle    Key = "turtle"
	Lichen    Key = "lichen"
	Guac      Key = "guac"
	Julep     Key = "julep"
	Bok       Key = "bok"
	Mustard   Key = "mustard"
	Melon     Key = "melon"
	Zest      Key = "zest"
	Pepper    Key = "pepper"
	Charcoal  Key = "charcoal"
	Oyster    Key = "oyster"
	Squid     Key = "squid"
	Smoke     Key = "smoke"
	Butter    Key = "butter"
)

// All colors in the palette.
var colors = map[Key]string{
	Cumin:     "#BF976F",
	Tang:      "#FF985A",
	Yam:       "#FFB587",
	Paprika:   "#D36C64",
	Pumpkin:   "#FF6E63",
	Uni:       "#FF937D",
	Sriracha:  "#EB4268",
	Coral:     "#FF577D",
	Salmon:    "#FF7F90",
	Chili:     "#E23080",
	Cherry:    "#FF388B",
	Tuna:      "#FF6DAA",
	Macaroon:  "#E940B0",
	Rose:      "#FF4FBF",
	Cheeky:    "#FF79D0",
	Flamingo:  "#F947E3",
	Mollie:    "#FF60FF",
	Blush:     "#FF84FF",
	Urchin:    "#C337E0",
	Crystal:   "#EB5DFF",
	Lilac:     "#F379FF",
	Eggplant:  "#9C35E1",
	Violet:    "#C259FF",
	Mauve:     "#D46EFF",
	Grape:     "#7134DD",
	Plum:      "#9953FF",
	Orchid:    "#AD6EFF",
	Fig:       "#4A30D9",
	Charple:   "#6B50FF",
	Hazy:      "#8B75FF",
	Blueberry: "#3331B2",
	Sapphire:  "#4949FF",
	Guppy:     "#7272FF",
	Neptune:   "#2B55B3",
	Electric:  "#4776FF",
	Anchovy:   "#719AFC",
	Damson:    "#007AB8",
	Malibu:    "#00A4FF",
	Sardine:   "#4FBEFE",
	Zinc:      "#10B1AE",
	Turtle:    "#0ADCD9",
	Lichen:    "#5CDFEA",
	Guac:      "#14DD9F",
	Julep:     "#00FFB2",
	Bok:       "#68FFD6",
	Mustard:   "#F5EF34",
	Melon:     "#E8FF27",
	Zest:      "#E8FE96",
	Pepper:    "#201F26",
	Charcoal:  "#3A3943",
	Oyster:    "#605F6B",
	Squid:     "#858392",
	Smoke:     "#BFBCC8",
	Butter:    "#FFFAF1",
}

// Tones is a map of colors from the CharmTone palette.
var Tones map[Key]color.Color

func init() {
	// Add gray aliases.
	colors[Key("gray1")] = colors[Pepper]
	colors[Key("gray2")] = colors[Charcoal]
	colors[Key("gray3")] = colors[Oyster]
	colors[Key("gray4")] = colors[Squid]

	// Make color.Colors.
	Tones = make(map[Key]color.Color, len(colors))
	for name, hex := range colors {
		Tones[name] = lipgloss.Color(hex)
	}
}
