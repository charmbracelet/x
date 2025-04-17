package charmtone

import (
	"image/color"

	"github.com/charmbracelet/lipgloss/v2"
)

// All colors in the palette.
var colors = map[string]string{
	"cumin":     "#BF976F",
	"tang":      "#FF985A",
	"yam":       "#FFB587",
	"paprika":   "#D36C64",
	"pumpkin":   "#FF6E63",
	"uni":       "#FF937D",
	"sriracha":  "#EB4268",
	"coral":     "#FF577D",
	"salmon":    "#FF7F90",
	"chili":     "#E23080",
	"cherry":    "#FF388B",
	"tuna":      "#FF6DAA",
	"macaroon":  "#E940B0",
	"rose":      "#FF4FBF",
	"cheeky":    "#FF79D0",
	"flamingo":  "#F947E3",
	"mollie":    "#FF60FF",
	"blush":     "#FF84FF",
	"urchin":    "#C337E0",
	"crystal":   "#EB5DFF",
	"lilac":     "#F379FF",
	"eggplant":  "#9C35E1",
	"violet":    "#C259FF",
	"mauve":     "#D46EFF",
	"grape":     "#7134DD",
	"plum":      "#9953FF",
	"orchid":    "#AD6EFF",
	"fig":       "#4A30D9",
	"charple":   "#6B50FF",
	"hazy":      "#8B75FF",
	"blueberry": "#3331B2",
	"sapphire":  "#4949FF",
	"guppy":     "#7272FF",
	"neptune":   "#2B55B3",
	"electric":  "#4776FF",
	"anchovy":   "#719AFC",
	"damson":    "#007AB8",
	"malibu":    "#00A4FF",
	"sardine":   "#4FBEFE",
	"zinc":      "#10B1AE",
	"turtle":    "#0ADCD9",
	"lichen":    "#5CDFEA",
	"guac":      "#14DD9F",
	"julep":     "#00FFB2",
	"bok":       "#68FFD6",
	"mustard":   "#F5EF34",
	"melon":     "#E8FF27",
	"zest":      "#E8FE96",

	"pepper":   "#201F26",
	"charcoal": "#3A3943",
	"oyster":   "#605F6B",
	"squid":    "#858392",
	"smoke":    "#BFBCC8",
	"butter":   "#FFFAF1",
}

// Tones is a map of colors from the CharmTone palette.
var Tones map[string]color.Color

func init() {
	// Add gray aliases.
	colors["gray1"] = colors["pepper"]
	colors["gray2"] = colors["charcoal"]
	colors["gray3"] = colors["oyster"]
	colors["gray4"] = colors["squid"]

	// Make color.Colors.
	Tones = make(map[string]color.Color, len(colors))
	for name, hex := range colors {
		Tones[name] = lipgloss.Color(hex)
	}
}
