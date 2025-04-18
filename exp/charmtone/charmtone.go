package charmtone

import (
	"image/color"
	"slices"

	"github.com/charmbracelet/lipgloss/v2"
)

// Key is a type for color keys.
type Key int

const (
	Cumin Key = iota
	Tang
	Yam
	Paprika
	Bengal
	Uni
	Sriracha
	Coral
	Salmon
	Chili
	Cherry
	Tuna
	Macaron
	Pony
	Cheeky
	Flamingo
	Dolly
	Blush
	Urchin
	Crystal
	Lilac
	Prince
	Violet
	Mauve
	Grape
	Plum
	Orchid
	Jelly
	Charple
	Hazy
	Ox
	Sapphire
	Guppy
	Oceania
	Thunder
	Anchovy
	Damson
	Malibu
	Sardine
	Zinc
	Turtle
	Lichen
	Guac
	Julep
	Bok
	Mustard
	Citron
	Zest
	Pepper
	Charcoal
	Oyster
	Squid
	Smoke
	Ash
	Butter
)

func (k Key) String() string {
	return map[Key]string{
		Cumin:    "Cumin",
		Tang:     "Tang",
		Yam:      "Yam",
		Paprika:  "Paprika",
		Bengal:   "Bengal",
		Uni:      "Uni",
		Sriracha: "Sriracha",
		Coral:    "Coral",
		Salmon:   "Salmon",
		Chili:    "Chili",
		Cherry:   "Cherry",
		Tuna:     "Tuna",
		Macaron:  "Macaron",
		Pony:     "Pony",
		Cheeky:   "Cheeky",
		Flamingo: "Flamingo",
		Dolly:    "Dolly",
		Blush:    "Blush",
		Urchin:   "Urchin",
		Crystal:  "Crystal",
		Lilac:    "Lilac",
		Prince:   "Prince",
		Violet:   "Violet",
		Mauve:    "Mauve",
		Grape:    "Grape",
		Plum:     "Plum",
		Orchid:   "Orchid",
		Jelly:    "Jelly",
		Charple:  "Charple",
		Hazy:     "Hazy",
		Ox:       "Ox",
		Sapphire: "Sapphire",
		Guppy:    "Guppy",
		Oceania:  "Oceania",
		Thunder:  "Thunder",
		Anchovy:  "Anchovy",
		Damson:   "Damson",
		Malibu:   "Malibu",
		Sardine:  "Sardine",
		Zinc:     "Zinc",
		Turtle:   "Turtle",
		Lichen:   "Lichen",
		Guac:     "Guac",
		Julep:    "Julep",
		Bok:      "Bok",
		Mustard:  "Mustard",
		Citron:   "Citron",
		Zest:     "Zest",
		Pepper:   "Pepper",
		Charcoal: "Charcoal",
		Oyster:   "Oyster",
		Squid:    "Squid",
		Smoke:    "Smoke",
		Ash:      "Ash",
		Butter:   "Butter",
	}[k]
}

// All Hexes in the palette.
func Hexes() map[Key]string {
	return map[Key]string{
		Cumin:    "#BF976F",
		Tang:     "#FF985A",
		Yam:      "#FFB587",
		Paprika:  "#D36C64",
		Bengal:   "#FF6E63",
		Uni:      "#FF937D",
		Sriracha: "#EB4268",
		Coral:    "#FF577D",
		Salmon:   "#FF7F90",
		Chili:    "#E23080",
		Cherry:   "#FF388B",
		Tuna:     "#FF6DAA",
		Macaron:  "#E940B0",
		Pony:     "#FF4FBF",
		Cheeky:   "#FF79D0",
		Flamingo: "#F947E3",
		Dolly:    "#FF60FF",
		Blush:    "#FF84FF",
		Urchin:   "#C337E0",
		Crystal:  "#EB5DFF",
		Lilac:    "#F379FF",
		Prince:   "#9C35E1",
		Violet:   "#C259FF",
		Mauve:    "#D46EFF",
		Grape:    "#7134DD",
		Plum:     "#9953FF",
		Orchid:   "#AD6EFF",
		Jelly:    "#4A30D9",
		Charple:  "#6B50FF",
		Hazy:     "#8B75FF",
		Ox:       "#3331B2",
		Sapphire: "#4949FF",
		Guppy:    "#7272FF",
		Oceania:  "#2B55B3",
		Thunder:  "#4776FF",
		Anchovy:  "#719AFC",
		Damson:   "#007AB8",
		Malibu:   "#00A4FF",
		Sardine:  "#4FBEFE",
		Zinc:     "#10B1AE",
		Turtle:   "#0ADCD9",
		Lichen:   "#5CDFEA",
		Guac:     "#12C78F",
		Julep:    "#00FFB2",
		Bok:      "#68FFD6",
		Mustard:  "#F5EF34",
		Citron:   "#E8FF27",
		Zest:     "#E8FE96",
		Pepper:   "#201F26", // Gray 1
		Charcoal: "#3A3943", // Gray 2
		Oyster:   "#605F6B", // Gray 3
		Squid:    "#858392", // Gray 4
		Smoke:    "#BFBCC8", // Gray 5
		Ash:      "#DFDBDD", // Gray 6
		Butter:   "#FFFAF1",
	}
}

func Keys() []Key {
	return []Key{
		Cumin,
		Tang,
		Yam,
		Paprika,
		Bengal,
		Uni,
		Sriracha,
		Coral,
		Salmon,
		Chili,
		Cherry,
		Tuna,
		Macaron,
		Pony,
		Cheeky,
		Flamingo,
		Dolly,
		Blush,
		Urchin,
		Crystal,
		Lilac,
		Prince,
		Violet,
		Mauve,
		Grape,
		Plum,
		Orchid,
		Jelly,
		Charple,
		Hazy,
		Ox,
		Sapphire,
		Guppy,
		Oceania,
		Thunder,
		Anchovy,
		Damson,
		Malibu,
		Sardine,
		Zinc,
		Turtle,
		Lichen,
		Guac,
		Julep,
		Bok,
		Mustard,
		Citron,
		Zest,
		Pepper,
		Charcoal,
		Oyster,
		Squid,
		Smoke,
		Ash,
		Butter,
	}
}

// Tones is a map of colors from the CharmTone palette.
func Tones() map[Key]color.Color {
	h := Hexes()
	t := make(map[Key]color.Color, len(h))
	for name, hex := range h {
		t[name] = lipgloss.Color(hex)
	}
	return t
}

// Core indicates which colors are part of the core palette.
func IsPrimary(k Key) bool {
	return slices.Contains([]Key{
		Charple,
		Dolly,
		Julep,
		Zest,
		Butter,
		Hazy,
		Blush,
		Bok,
	}, k)
}

func IsSecondary(k Key) bool {
	return slices.Contains([]Key{
		Turtle,
		Malibu,
		Violet,
		Tuna,
		Coral,
		Uni,
	}, k)
}
