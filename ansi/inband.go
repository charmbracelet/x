package ansi

import (
	"fmt"
	"io"
)

// WriteInBandResize writes an in-band terminal resize event sequence to w.
//
//	CSI 48 ; height_cells ; widht_cells ; height_pixels ; width_pixels t
//
// See https://gist.github.com/rockorager/e695fb2924d36b2bcf1fff4a3704bd83
func WriteInBandResize(w io.Writer, heightCells, widthCells, heightPixels, widthPixels int) (int, error) {
	if heightCells < 0 {
		heightCells = 0
	}
	if widthCells < 0 {
		widthCells = 0
	}
	if heightPixels < 0 {
		heightPixels = 0
	}
	if widthPixels < 0 {
		widthPixels = 0
	}
	return fmt.Fprintf(w, "\x1b[48;%d;%d;%d;%dt", heightCells, widthCells, heightPixels, widthPixels)
}

// InBandResize encodes an in-band terminal resize event sequence.
//
//	CSI 48 ; height_cells ; widht_cells ; height_pixels ; width_pixels t
//
// See https://gist.github.com/rockorager/e695fb2924d36b2bcf1fff4a3704bd83
func InBandResize(heightCells, widthCells, heightPixels, widthPixels int) string {
	if heightCells < 0 {
		heightCells = 0
	}
	if widthCells < 0 {
		widthCells = 0
	}
	if heightPixels < 0 {
		heightPixels = 0
	}
	if widthPixels < 0 {
		widthPixels = 0
	}
	return fmt.Sprintf("\x1b[48;%d;%d;%d;%dt", heightCells, widthCells, heightPixels, widthPixels)
}
