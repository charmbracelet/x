// Package charmtone contains an API for the CharmTone color palette.
package charmtone

import (
	"fmt"
	"image/color"
	"slices"
	"strconv"
)

var _ color.Color = Key(0)

// Key is a type for color keys.
type Key int

// Spectrum: the main CharmTone palette.
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
	Mochi
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

	// Butter is part of the main palette but is no longer tied to any
	// meta-group.
	Butter

	// Neutrals.
	Pepper
	BBQ
	Char
	Iron
	Oyster
	Squid
	Steam
	Smoke
	Steep
	Sash
	Salt
	White

	// Charples: additional shades for the Charple ramp. The other Charples
	// (Jelly, Charple, Hazy) live in the main spectrum above.
	Darple
	Larple

	// Diffs: additions. The brightest color in this group, Julep, lives in
	// the main spectrum above.
	Pickle
	Gator
	Spinach

	// Diffs: deletions. The brightest color in this group, Cherry, lives in
	// the main spectrum above.
	Pom
	Steak
	Toast
)

// RGBA returns the red, green, blue, and alpha values of the color. It
// satisfies the color.Color interface.
func (k Key) RGBA() (r, g, b, a uint32) {
	c, ok := colors[k]
	if !ok {
		panic("invalid color key " + strconv.Itoa(int(k)))
	}
	return c.RGBA()
}

var names = map[Key]string{
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
	Mochi:    "Mochi",
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
	Butter:   "Butter",

	// Neutrals.
	Pepper: "Pepper",
	BBQ:    "BBQ",
	Char:   "Char",
	Iron:   "Iron",
	Oyster: "Oyster",
	Squid:  "Squid",
	Steam:  "Steam",
	Smoke:  "Smoke",
	Steep:  "Steep",
	Sash:   "Sash",
	Salt:   "Salt",
	White:  "White",

	// Charples.
	Darple: "Darple",
	Larple: "Larple",

	// Diffs: additions.
	Pickle:  "Pickle",
	Gator:   "Gator",
	Spinach: "Spinach",

	// Diffs: deletions.
	Pom:   "Pom",
	Steak: "Steak",
	Toast: "Toast",
}

// String returns the official CharmTone name of the color. It satisfies the
// fmt.Stringer interface.
func (k Key) String() string {
	name, ok := names[k]
	if !ok {
		return ""
	}
	return name
}

var colors = map[Key]color.RGBA{
	Cumin:    {R: 0xBF, G: 0x97, B: 0x6F, A: 0xFF}, // "#BF976F"
	Tang:     {R: 0xFF, G: 0x98, B: 0x5A, A: 0xFF}, // "#FF985A"
	Yam:      {R: 0xFF, G: 0xB5, B: 0x87, A: 0xFF}, // "#FFB587"
	Paprika:  {R: 0xD3, G: 0x6C, B: 0x64, A: 0xFF}, // "#D36C64"
	Bengal:   {R: 0xFF, G: 0x6E, B: 0x63, A: 0xFF}, // "#FF6E63"
	Uni:      {R: 0xFF, G: 0x93, B: 0x7D, A: 0xFF}, // "#FF937D"
	Sriracha: {R: 0xEB, G: 0x42, B: 0x68, A: 0xFF}, // "#EB4268"
	Coral:    {R: 0xFF, G: 0x57, B: 0x7D, A: 0xFF}, // "#FF577D"
	Salmon:   {R: 0xFF, G: 0x7F, B: 0x90, A: 0xFF}, // "#FF7F90"
	Chili:    {R: 0xE2, G: 0x30, B: 0x80, A: 0xFF}, // "#E23080"
	Cherry:   {R: 0xFF, G: 0x38, B: 0x8B, A: 0xFF}, // "#FF388B"
	Tuna:     {R: 0xFF, G: 0x6D, B: 0xAA, A: 0xFF}, // "#FF6DAA"
	Macaron:  {R: 0xE9, G: 0x40, B: 0xB0, A: 0xFF}, // "#E940B0"
	Pony:     {R: 0xFF, G: 0x4F, B: 0xBF, A: 0xFF}, // "#FF4FBF"
	Cheeky:   {R: 0xFF, G: 0x79, B: 0xD0, A: 0xFF}, // "#FF79D0"
	Flamingo: {R: 0xF9, G: 0x47, B: 0xE3, A: 0xFF}, // "#F947E3"
	Dolly:    {R: 0xFF, G: 0x60, B: 0xFF, A: 0xFF}, // "#FF60FF"
	Blush:    {R: 0xFF, G: 0x84, B: 0xFF, A: 0xFF}, // "#FF84FF"
	Urchin:   {R: 0xC3, G: 0x37, B: 0xE0, A: 0xFF}, // "#C337E0"
	Mochi:    {R: 0xEB, G: 0x5D, B: 0xFF, A: 0xFF}, // "#EB5DFF"
	Lilac:    {R: 0xF3, G: 0x79, B: 0xFF, A: 0xFF}, // "#F379FF"
	Prince:   {R: 0x9C, G: 0x35, B: 0xE1, A: 0xFF}, // "#9C35E1"
	Violet:   {R: 0xC2, G: 0x59, B: 0xFF, A: 0xFF}, // "#C259FF"
	Mauve:    {R: 0xD4, G: 0x6E, B: 0xFF, A: 0xFF}, // "#D46EFF"
	Grape:    {R: 0x71, G: 0x34, B: 0xDD, A: 0xFF}, // "#7134DD"
	Plum:     {R: 0x99, G: 0x53, B: 0xFF, A: 0xFF}, // "#9953FF"
	Orchid:   {R: 0xAD, G: 0x6E, B: 0xFF, A: 0xFF}, // "#AD6EFF"
	Jelly:    {R: 0x4A, G: 0x30, B: 0xD9, A: 0xFF}, // "#4A30D9"
	Charple:  {R: 0x6B, G: 0x50, B: 0xFF, A: 0xFF}, // "#6B50FF"
	Hazy:     {R: 0x8B, G: 0x75, B: 0xFF, A: 0xFF}, // "#8B75FF"
	Ox:       {R: 0x33, G: 0x31, B: 0xB2, A: 0xFF}, // "#3331B2"
	Sapphire: {R: 0x49, G: 0x49, B: 0xFF, A: 0xFF}, // "#4949FF"
	Guppy:    {R: 0x72, G: 0x72, B: 0xFF, A: 0xFF}, // "#7272FF"
	Oceania:  {R: 0x2B, G: 0x55, B: 0xB3, A: 0xFF}, // "#2B55B3"
	Thunder:  {R: 0x47, G: 0x76, B: 0xFF, A: 0xFF}, // "#4776FF"
	Anchovy:  {R: 0x71, G: 0x9A, B: 0xFC, A: 0xFF}, // "#719AFC"
	Damson:   {R: 0x00, G: 0x7A, B: 0xB8, A: 0xFF}, // "#007AB8"
	Malibu:   {R: 0x00, G: 0xA4, B: 0xFF, A: 0xFF}, // "#00A4FF"
	Sardine:  {R: 0x4F, G: 0xBE, B: 0xFE, A: 0xFF}, // "#4FBEFE"
	Zinc:     {R: 0x10, G: 0xB1, B: 0xAE, A: 0xFF}, // "#10B1AE"
	Turtle:   {R: 0x0A, G: 0xDC, B: 0xD9, A: 0xFF}, // "#0ADCD9"
	Lichen:   {R: 0x5C, G: 0xDF, B: 0xEA, A: 0xFF}, // "#5CDFEA"
	Guac:     {R: 0x12, G: 0xC7, B: 0x8F, A: 0xFF}, // "#12C78F"
	Julep:    {R: 0x00, G: 0xFF, B: 0xB2, A: 0xFF}, // "#00FFB2"
	Bok:      {R: 0x68, G: 0xFF, B: 0xD6, A: 0xFF}, // "#68FFD6"
	Mustard:  {R: 0xF5, G: 0xEF, B: 0x34, A: 0xFF}, // "#F5EF34"
	Citron:   {R: 0xE8, G: 0xFF, B: 0x27, A: 0xFF}, // "#E8FF27"
	Zest:     {R: 0xE8, G: 0xFE, B: 0x96, A: 0xFF}, // "#E8FE96"
	Butter:   {R: 0xFF, G: 0xFA, B: 0xF1, A: 0xFF}, // "#FFFAF1"

	// Neutrals.
	Pepper: {R: 0x20, G: 0x1F, B: 0x26, A: 0xFF}, // "#201F26"
	BBQ:    {R: 0x2D, G: 0x2C, B: 0x36, A: 0xFF}, // "#2D2C36"
	Char:   {R: 0x3A, G: 0x39, B: 0x43, A: 0xFF}, // "#3A3943"
	Iron:   {R: 0x4D, G: 0x4C, B: 0x57, A: 0xFF}, // "#4D4C57"
	Oyster: {R: 0x60, G: 0x5F, B: 0x6B, A: 0xFF}, // "#605F6B"
	Squid:  {R: 0x85, G: 0x83, B: 0x92, A: 0xFF}, // "#858392"
	Steam:  {R: 0xA2, G: 0xA0, B: 0xAD, A: 0xFF}, // "#A2A0AD"
	Smoke:  {R: 0xBF, G: 0xBC, B: 0xC8, A: 0xFF}, // "#BFBCC8"
	Steep:  {R: 0xD6, G: 0xD3, B: 0xDC, A: 0xFF}, // "#D6D3DC"
	Sash:   {R: 0xEC, G: 0xEB, B: 0xF0, A: 0xFF}, // "#ECEBF0"
	Salt:   {R: 0xF7, G: 0xF6, B: 0xFB, A: 0xFF}, // "#F7F6FB"
	White:  {R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, // "#FFFFFF"

	// Charples.
	Darple: {R: 0x5B, G: 0x40, B: 0xEC, A: 0xFF}, // "#5B40EC"
	Larple: {R: 0x7B, G: 0x62, B: 0xFF, A: 0xFF}, // "#7B62FF"

	// Diffs: additions.
	Pickle:  {R: 0x00, G: 0xA4, B: 0x75, A: 0xFF}, // "#00A475"
	Gator:   {R: 0x18, G: 0x46, B: 0x3D, A: 0xFF}, // "#18463D"
	Spinach: {R: 0x1C, G: 0x36, B: 0x34, A: 0xFF}, // "#1C3634"

	// Diffs: deletions.
	Pom:   {R: 0xAB, G: 0x24, B: 0x54, A: 0xFF}, // "#AB2454"
	Steak: {R: 0x58, G: 0x22, B: 0x38, A: 0xFF}, // "#582238"
	Toast: {R: 0x41, G: 0x21, B: 0x30, A: 0xFF}, // "#412130"
}

// Hex returns the hex value of the color.
func (k Key) Hex() string {
	c, ok := colors[k]
	if !ok {
		panic("invalid color key " + strconv.Itoa(int(k)))
	}
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// Keys returns a slice of all CharmTone color keys, in iota order: the main
// spectrum, Butter, then neutrals, charples, additions, and deletions.
func Keys() []Key {
	keys := make([]Key, 0, len(colors))
	for k := Cumin; k <= Toast; k++ {
		keys = append(keys, k)
	}
	return keys
}

// Spectrum returns the main CharmTone palette, from Cumin through Zest.
// Butter, neutrals, and meta-group colors (charples, additions, deletions)
// are not included.
func Spectrum() []Key {
	keys := make([]Key, 0, int(Zest-Cumin)+1)
	for k := Cumin; k <= Zest; k++ {
		keys = append(keys, k)
	}
	return keys
}

// Neutrals returns the neutral colors, from Pepper through White.
func Neutrals() []Key {
	return []Key{
		Pepper,
		BBQ,
		Char,
		Iron,
		Oyster,
		Squid,
		Steam,
		Smoke,
		Steep,
		Sash,
		Salt,
		White,
	}
}

// Charples returns the Charple ramp: shades from darkest to lightest. This
// includes colors that are also part of the main spectrum (Jelly, Charple,
// Hazy) alongside the Charple-only Darple and Larple.
func Charples() []Key {
	return []Key{
		Jelly,
		Darple,
		Charple,
		Larple,
		Hazy,
	}
}

// Additions returns the colors used to indicate diff additions, from darkest
// to brightest. The brightest, Julep, is also part of the main spectrum.
func Additions() []Key {
	return []Key{
		Spinach,
		Gator,
		Pickle,
		Julep,
	}
}

// Deletions returns the colors used to indicate diff deletions, from darkest
// to brightest. The brightest, Cherry, is also part of the main spectrum.
func Deletions() []Key {
	return []Key{
		Toast,
		Steak,
		Pom,
		Cherry,
	}
}

// IsSpectrum reports whether k is part of the main spectrum palette
// (Cumin through Zest). Butter is intentionally excluded.
func (k Key) IsSpectrum() bool {
	return k >= Cumin && k <= Zest
}

// IsNeutral reports whether k is one of the neutral colors.
func (k Key) IsNeutral() bool {
	return k >= Pepper && k <= White
}

// IsCharple reports whether k is part of the Charple ramp.
func (k Key) IsCharple() bool {
	return slices.Contains(Charples(), k)
}

// IsAddition reports whether k is one of the diff-addition colors.
func (k Key) IsAddition() bool {
	return slices.Contains(Additions(), k)
}

// IsDeletion reports whether k is one of the diff-deletion colors.
func (k Key) IsDeletion() bool {
	return slices.Contains(Deletions(), k)
}

// IsPrimary indicates which colors are part of the core palette.
func (k Key) IsPrimary() bool {
	return slices.Contains([]Key{
		Charple,
		Dolly,
		Julep,
		Zest,
		Hazy,
		Blush,
		Bok,
		Butter,
	}, k)
}

// IsSecondary indicates which colors are part of the secondary palette.
func (k Key) IsSecondary() bool {
	return slices.Contains([]Key{
		Uni,
		Coral,
		Tuna,
		Violet,
		Malibu,
		Turtle,
	}, k)
}

// IsTertiary indicates which colors are part of the tertiary palette.
//
// Deprecated: the CharmTone formula guide no longer defines a tertiary
// palette. Use [Key.IsSecondary] instead.
func (k Key) IsTertiary() bool {
	return k.IsSecondary()
}

// Deprecated: Charcoal has been renamed to [Char].
const Charcoal = Char

// Deprecated: Ash has been renamed to [Sash].
const Ash = Sash
