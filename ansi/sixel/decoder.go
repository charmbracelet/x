package sixel

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
)

// buildDefaultDecodePalette will build a map that we'll use as the palette during
// the decoding process- it's pre-loaded with the default colors for sixels, in case
// we are decoding a legacy sixel image that doesn't define its own colors (technically
// permitted).
func buildDefaultDecodePalette() map[int]color.Color {

	// Undefined colors in sixel images use a set of default colors: 0-15
	// are sixel-specific, 16-255 are the same as the xterm 256-color values
	return map[int]color.Color{
		// Sixel-specific default colors
		0:  color.RGBA{0, 0, 0, 255},
		1:  color.RGBA{51, 51, 204, 255},
		2:  color.RGBA{204, 36, 36, 255},
		3:  color.RGBA{51, 204, 51, 255},
		4:  color.RGBA{204, 51, 204, 255},
		5:  color.RGBA{51, 204, 204, 255},
		6:  color.RGBA{204, 204, 51, 255},
		7:  color.RGBA{120, 120, 120, 255},
		8:  color.RGBA{69, 69, 69, 255},
		9:  color.RGBA{87, 87, 153, 255},
		10: color.RGBA{153, 69, 69, 255},
		11: color.RGBA{87, 153, 87, 255},
		12: color.RGBA{153, 87, 153, 255},
		13: color.RGBA{87, 153, 153, 255},
		14: color.RGBA{153, 153, 87, 255},
		15: color.RGBA{204, 204, 204, 255},

		// xterm colors
		16:  color.RGBA{0, 0, 0, 255},       // Black1
		17:  color.RGBA{0, 0, 95, 255},      // DarkBlue2
		18:  color.RGBA{0, 0, 135, 255},     // DarkBlue1
		19:  color.RGBA{0, 0, 175, 255},     // DarkBlue
		20:  color.RGBA{0, 0, 215, 255},     // Blue3
		21:  color.RGBA{0, 0, 255, 255},     // Blue2
		22:  color.RGBA{0, 95, 0, 255},      // DarkGreen4
		23:  color.RGBA{0, 95, 95, 255},     // DarkGreenBlue5
		24:  color.RGBA{0, 95, 135, 255},    // DarkGreenBlue4
		25:  color.RGBA{0, 95, 175, 255},    // DarkGreenBlue3
		26:  color.RGBA{0, 95, 215, 255},    // GreenBlue8
		27:  color.RGBA{0, 95, 255, 255},    // GreenBlue7
		28:  color.RGBA{0, 135, 0, 255},     // DarkGreen3
		29:  color.RGBA{0, 135, 95, 255},    // DarkGreen2
		30:  color.RGBA{0, 135, 0, 255},     // DarkGreenBlue2
		31:  color.RGBA{0, 135, 175, 255},   // DarkGreenBlue1
		32:  color.RGBA{0, 125, 215, 255},   // GreenBlue6
		33:  color.RGBA{0, 135, 255, 255},   // GreenBlue5
		34:  color.RGBA{0, 175, 0, 255},     // DarkGreen1
		35:  color.RGBA{0, 175, 95, 255},    // DarkGreen
		36:  color.RGBA{0, 175, 135, 255},   // DarkBlueGreen
		37:  color.RGBA{0, 175, 175, 255},   // DarkGreenBlue
		38:  color.RGBA{0, 175, 215, 255},   // GreenBlue4
		39:  color.RGBA{0, 175, 255, 255},   // GreenBlue3
		40:  color.RGBA{0, 215, 0, 255},     // Green7
		41:  color.RGBA{0, 215, 95, 255},    // Green6
		42:  color.RGBA{0, 215, 135, 255},   // Green5
		43:  color.RGBA{0, 215, 175, 255},   // BlueGreen1
		44:  color.RGBA{0, 215, 215, 255},   // GreenBlue2
		45:  color.RGBA{0, 215, 255, 255},   // GreenBlue1
		46:  color.RGBA{0, 255, 0, 255},     // Green4
		47:  color.RGBA{0, 255, 95, 255},    // Green3
		48:  color.RGBA{0, 255, 135, 255},   // Green2
		49:  color.RGBA{0, 255, 175, 255},   // Green1
		50:  color.RGBA{0, 255, 215, 255},   // BlueGreen
		51:  color.RGBA{0, 255, 255, 255},   // GreenBlue
		52:  color.RGBA{95, 0, 0, 255},      // DarkRed2
		53:  color.RGBA{95, 0, 95, 255},     // DarkPurple4
		54:  color.RGBA{95, 0, 135, 255},    // DarkBluePurple2
		55:  color.RGBA{95, 0, 175, 255},    // DarkBluePurple1
		56:  color.RGBA{95, 0, 215, 255},    // PurpleBlue
		57:  color.RGBA{95, 0, 255, 255},    // Blue1
		58:  color.RGBA{95, 95, 0, 255},     // DarkYellow4
		59:  color.RGBA{95, 95, 95, 255},    // Gray3
		60:  color.RGBA{95, 95, 135, 255},   // PlueBlue8
		61:  color.RGBA{95, 95, 175, 255},   // PaleBlue7
		62:  color.RGBA{95, 95, 215, 255},   // PaleBlue6
		63:  color.RGBA{95, 95, 255, 255},   // PaleBlue5
		64:  color.RGBA{95, 135, 0, 255},    // DarkYellow3
		65:  color.RGBA{95, 135, 95, 255},   // PaleGreen12
		66:  color.RGBA{95, 135, 135, 255},  // PaleGreen11
		67:  color.RGBA{95, 135, 175, 255},  // PaleGreenBlue10
		68:  color.RGBA{95, 135, 215, 255},  // PaleGreenBlue9
		69:  color.RGBA{95, 135, 255, 255},  // PaleBlue4
		70:  color.RGBA{95, 175, 0, 255},    // DarkGreenYellow
		71:  color.RGBA{95, 175, 95, 255},   // PaleGreen11
		72:  color.RGBA{95, 175, 135, 255},  // PaleGreen10
		73:  color.RGBA{95, 175, 175, 255},  // PaleGreenBlue8
		74:  color.RGBA{95, 175, 215, 255},  // PaleGreenBlue7
		75:  color.RGBA{95, 175, 255, 255},  // PaleGreenBlue6
		76:  color.RGBA{95, 215, 0, 255},    // YellowGreen1
		77:  color.RGBA{95, 215, 95, 255},   // PaleGreen9
		78:  color.RGBA{95, 215, 135, 255},  // PaleGreen8
		79:  color.RGBA{95, 215, 175, 255},  // PaleGreen7
		80:  color.RGBA{95, 215, 215, 255},  // PaleGreenBlue5
		81:  color.RGBA{95, 215, 255, 255},  // PaleGreenBlue4
		82:  color.RGBA{95, 255, 0, 255},    // YellowGreen
		83:  color.RGBA{95, 255, 95, 255},   // PaleGreen6
		84:  color.RGBA{95, 255, 135, 255},  // PaleGreen5
		85:  color.RGBA{95, 255, 175, 255},  // PaleGreen4
		86:  color.RGBA{95, 255, 215, 255},  // PaleGreen3
		87:  color.RGBA{95, 255, 255, 255},  // PaleGreenBlue3
		88:  color.RGBA{135, 0, 0, 255},     // DarkRed1
		89:  color.RGBA{135, 0, 95, 255},    // DarkPurple3
		90:  color.RGBA{135, 0, 135, 255},   // DarkPurple2
		91:  color.RGBA{135, 0, 175, 255},   // DarkBluePurple
		92:  color.RGBA{135, 0, 215, 255},   // BluePurple4
		93:  color.RGBA{135, 0, 255, 255},   // BluePurple3
		94:  color.RGBA{135, 95, 0, 255},    // DarkOrange1
		95:  color.RGBA{135, 95, 95, 255},   // PaleRed5
		96:  color.RGBA{135, 95, 135, 255},  // PalePurple7
		97:  color.RGBA{135, 95, 175, 255},  // PalePurpleBlue
		98:  color.RGBA{135, 95, 215, 255},  // PaleBlue3
		99:  color.RGBA{135, 95, 255, 255},  // PaleBlue2
		100: color.RGBA{135, 135, 0, 255},   // DarkYellow2
		101: color.RGBA{135, 135, 95, 255},  // PaleYellow7
		102: color.RGBA{135, 135, 135, 255}, // Gray2
		103: color.RGBA{135, 135, 175, 255}, // PaleBlue1
		104: color.RGBA{135, 135, 215, 255}, // PaleBlue
		105: color.RGBA{135, 135, 255, 255}, // LightPaleBlue4
		106: color.RGBA{135, 175, 0, 255},   // DarkYellow1
		107: color.RGBA{135, 175, 95, 255},  // PaleYellowGreen3
		108: color.RGBA{135, 175, 135, 255}, // PaleGreen2
		109: color.RGBA{135, 175, 175, 255}, // PaleGreenBlue2
		110: color.RGBA{135, 175, 215, 255}, // PaleGreenBlue1
		111: color.RGBA{135, 175, 255, 255}, // LightPaleGreenBlue6
		112: color.RGBA{135, 215, 0, 255},   // Yellow6
		113: color.RGBA{135, 215, 95, 255},  // PaleYellowGreen2
		114: color.RGBA{135, 215, 135, 255}, // PaleGreen1
		115: color.RGBA{135, 215, 175, 255}, // PaleGreen
		116: color.RGBA{135, 215, 215, 255}, // PaleGreenBlue
		117: color.RGBA{135, 215, 255, 255}, // LightPaleGreenBlue5
		118: color.RGBA{135, 255, 0, 255},   // GreenYellow
		119: color.RGBA{135, 255, 95, 255},  // PaleYellowGreen1
		120: color.RGBA{135, 255, 135, 255}, // LightPaleGreen6
		121: color.RGBA{135, 255, 175, 255}, // LightPaleGreen5
		122: color.RGBA{135, 255, 215, 255}, // LightPaleGreen4
		123: color.RGBA{135, 255, 255, 255}, // LightPaleGreenBlue4
		124: color.RGBA{175, 0, 0, 255},     // DarkRed
		125: color.RGBA{175, 0, 95, 255},    // DarkRedPurple
		126: color.RGBA{175, 0, 135, 255},   // DarkPurple1
		127: color.RGBA{175, 0, 175, 255},   // DarkPurple
		128: color.RGBA{175, 0, 215, 255},   // BluePurple2
		129: color.RGBA{175, 0, 255, 255},   // BluePurple1
		130: color.RGBA{175, 95, 0, 255},    // DarkOrange
		131: color.RGBA{175, 95, 95, 255},   // PaleRed4
		132: color.RGBA{175, 95, 135, 255},  // PalePurpleRed3
		133: color.RGBA{175, 95, 175, 255},  // PalePurple6
		134: color.RGBA{175, 95, 215, 255},  // PaleBluePurple3
		135: color.RGBA{175, 95, 255, 255},  // PaleBluePurple2
		136: color.RGBA{175, 135, 0, 255},   // DarkYellowOrange
		137: color.RGBA{175, 135, 95, 255},  // PaleRedOrange3
		138: color.RGBA{175, 135, 135, 255}, // PaleRed3
		139: color.RGBA{175, 135, 175, 255}, // PalePurple5
		140: color.RGBA{175, 135, 215, 255}, // PaleBluePurple1
		141: color.RGBA{175, 135, 255, 255}, // LightPaleBlue3
		142: color.RGBA{175, 175, 0, 255},   // DarkYellow
		143: color.RGBA{175, 175, 95, 255},  // PaleYellow6
		144: color.RGBA{175, 175, 135, 255}, // PaleYellow5
		145: color.RGBA{175, 175, 175, 255}, // Gray1
		146: color.RGBA{175, 175, 215, 255}, // LightPaleBlue2
		147: color.RGBA{175, 175, 255, 255}, // LightPaleBlue1
		148: color.RGBA{175, 215, 0, 255},   // Yellow5
		149: color.RGBA{175, 215, 95, 255},  // PaleYellow4
		150: color.RGBA{175, 215, 135, 255}, // PaleGreenYellow
		151: color.RGBA{175, 215, 175, 255}, // LightPaleGreen3
		152: color.RGBA{175, 215, 215, 255}, // LightPaleGreenBlue3
		153: color.RGBA{175, 215, 255, 255}, // LightPaleGreenBlue2
		154: color.RGBA{175, 255, 0, 255},   // Yellow4
		155: color.RGBA{175, 255, 95, 255},  // PaleYellowGreen
		156: color.RGBA{175, 255, 135, 255}, // LightPaleYellowGreen1
		157: color.RGBA{175, 255, 215, 255}, // LightPaleGreen2
		158: color.RGBA{175, 255, 215, 255}, // LightPaleGreen1
		159: color.RGBA{175, 255, 255, 255}, // LightPaleGreenBlue1
		160: color.RGBA{215, 0, 0, 255},     // Red2
		161: color.RGBA{215, 0, 95, 255},    // PurpleRed1
		162: color.RGBA{215, 0, 135, 255},   // Purple6
		163: color.RGBA{215, 0, 175, 255},   // Purple5
		164: color.RGBA{215, 0, 215, 255},   // Purple4
		165: color.RGBA{215, 0, 255, 255},   // BluePurple
		166: color.RGBA{215, 95, 0, 255},    // RedOrange1
		167: color.RGBA{215, 95, 95, 255},   // PaleRed2
		168: color.RGBA{215, 95, 135, 255},  // PalePurpleRed2
		169: color.RGBA{215, 95, 175, 255},  // PalePurple4
		170: color.RGBA{215, 95, 215, 255},  // PalePurple3
		171: color.RGBA{215, 95, 255, 255},  // PaleBluePurple
		172: color.RGBA{215, 135, 0, 255},   // Orange2
		173: color.RGBA{215, 135, 95, 255},  // PaleRedOrange2
		174: color.RGBA{215, 135, 135, 255}, // PaleRed1
		175: color.RGBA{215, 135, 175, 255}, // PaleRedPurple
		176: color.RGBA{215, 135, 215, 255}, // PalePurple2
		177: color.RGBA{215, 135, 255, 255}, // LightPaleBluePurple
		178: color.RGBA{215, 175, 0, 255},   // OrangeYellow1
		179: color.RGBA{215, 175, 95, 255},  // PaleOrange1
		180: color.RGBA{215, 175, 135, 255}, // PaleRedOrange1
		181: color.RGBA{215, 175, 175, 255}, // LightPaleRed3
		182: color.RGBA{215, 175, 215, 255}, // LightPalePurple4
		183: color.RGBA{215, 175, 255, 255}, // LightPalePurpleBlue
		184: color.RGBA{215, 215, 0, 255},   // Yellow3
		185: color.RGBA{215, 215, 95, 255},  // PaleYellow3
		186: color.RGBA{215, 215, 135, 255}, // PaleYellow2
		187: color.RGBA{215, 215, 175, 255}, // LightPaleYellow4
		188: color.RGBA{215, 215, 215, 255}, // LightGray
		189: color.RGBA{215, 215, 255, 255}, // LightPaleBlue
		190: color.RGBA{215, 255, 0, 255},   // Yellow2
		191: color.RGBA{215, 255, 95, 255},  // PaleYellow1
		192: color.RGBA{215, 255, 135, 255}, // LightPaleYellow3
		193: color.RGBA{215, 255, 175, 255}, // LightPaleYellowGreen
		194: color.RGBA{215, 255, 215, 255}, // LightPaleGreen
		195: color.RGBA{215, 255, 255, 255}, // LightPaleGreenBlue
		196: color.RGBA{255, 0, 0, 255},     // Red1
		197: color.RGBA{255, 0, 95, 255},    // PurpleRed
		198: color.RGBA{255, 0, 135, 255},   // RedPurple
		199: color.RGBA{255, 0, 175, 255},   // Purple3
		200: color.RGBA{255, 0, 215, 255},   // Purple2
		201: color.RGBA{255, 0, 255, 255},   // Purple1
		202: color.RGBA{255, 95, 0, 255},    // RedOrange
		203: color.RGBA{255, 95, 95, 255},   // PaleRed
		204: color.RGBA{255, 95, 135, 255},  // PalePurpleRed1
		205: color.RGBA{255, 95, 175, 255},  // PalePurpleRed
		206: color.RGBA{255, 95, 215, 255},  // PalePurple1
		207: color.RGBA{255, 95, 255, 255},  // PalePurple
		208: color.RGBA{255, 135, 0, 255},   // Orange1
		209: color.RGBA{255, 135, 95, 255},  // PaleOrangeRed
		210: color.RGBA{255, 135, 135, 255}, // LightPaleRed2
		211: color.RGBA{255, 135, 175, 255}, // LightPalePurpleRed1
		212: color.RGBA{255, 135, 215, 255}, // LightPalePurple3
		213: color.RGBA{255, 135, 255, 255}, // LightPalePurple2
		214: color.RGBA{255, 175, 0, 255},   // Orange
		215: color.RGBA{255, 175, 95, 255},  // PaleRedOrange
		216: color.RGBA{255, 175, 135, 255}, // LightPaleRedOrange1
		217: color.RGBA{255, 175, 175, 255}, // LightPaleRed1
		218: color.RGBA{255, 175, 215, 255}, // LightPalePurpleRed
		219: color.RGBA{255, 175, 255, 255}, // LightPalePurple1
		220: color.RGBA{255, 215, 0, 255},   // OrangeYellow
		221: color.RGBA{255, 215, 95, 255},  // PaleOrange
		222: color.RGBA{255, 215, 135, 255}, // LightPaleOrange
		223: color.RGBA{255, 215, 175, 255}, // LightPaleRedOrange
		224: color.RGBA{255, 215, 215, 255}, // LightPaleRed
		225: color.RGBA{255, 215, 255, 255}, // LightPalePurple
		226: color.RGBA{255, 255, 0, 255},   // Yellow1
		227: color.RGBA{255, 255, 95, 255},  // PaleYellow
		228: color.RGBA{255, 255, 135, 255}, // LightPaleYellow2
		229: color.RGBA{255, 255, 175, 255}, // LightPaleYellow1
		230: color.RGBA{255, 255, 215, 255}, // LightPaleYellow
		231: color.RGBA{255, 255, 255, 255}, // White1
		232: color.RGBA{8, 8, 8, 255},       // Gray4
		233: color.RGBA{18, 18, 18, 255},    // Gray8
		234: color.RGBA{28, 28, 28, 255},    // Gray11
		235: color.RGBA{38, 38, 38, 255},    // Gray15
		236: color.RGBA{48, 48, 48, 255},    // Gray19
		237: color.RGBA{58, 58, 58, 255},    // Gray23
		238: color.RGBA{68, 68, 68, 255},    // Gray27
		239: color.RGBA{78, 78, 78, 255},    // Gray31
		240: color.RGBA{88, 88, 88, 255},    // Gray35
		241: color.RGBA{98, 98, 98, 255},    // Gray39
		242: color.RGBA{108, 108, 108, 255}, // Gray43
		243: color.RGBA{118, 118, 118, 255}, // Gray47
		244: color.RGBA{128, 128, 128, 255}, // Gray51
		245: color.RGBA{138, 138, 138, 255}, // Gray55
		246: color.RGBA{148, 148, 148, 255}, // Gray59
		247: color.RGBA{158, 158, 158, 255}, // Gray62
		248: color.RGBA{168, 168, 168, 255}, // Gray66
		249: color.RGBA{178, 178, 178, 255}, // Gray70
		250: color.RGBA{188, 188, 188, 255}, // Gray74
		251: color.RGBA{198, 198, 198, 255}, // Gray78
		252: color.RGBA{208, 208, 208, 255}, // Gray82
		253: color.RGBA{218, 218, 218, 255}, // Gray86
		254: color.RGBA{228, 228, 228, 255}, // Gray90
		255: color.RGBA{238, 238, 238, 255}, // Gray94
	}
}

type Decoder struct {
}

// Decode will parse sixel image data into an image or return an error.  Because
// the sixel image format does not have a predictable size, the end of the sixel
// image data can only be identified when ST, ESC, or BEL has been read from a reader.
// In order to avoid reading bytes from a reader one at a time to avoid missing
// the end, this method simply accepts a byte slice instead of a reader. Callers
// should read the entire escape sequence and pass the Ps..Ps portion of the sequence
// to this method.
func (d *Decoder) Decode(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0)), nil
	}

	buffer := bytes.NewBuffer(data)
	b, err := buffer.ReadByte()
	if err != nil {
		return nil, d.readError(err)
	}

	var bounds image.Rectangle
	if b == '"' {
		var fixedWidth, fixedHeight int
		// We have pixel dimensions
		_, err := fmt.Fscanf(buffer, "1;1;%d;%d", &fixedWidth, &fixedHeight)
		if err != nil {
			return nil, d.readError(err)
		}

		bounds = image.Rect(0, 0, fixedWidth, fixedHeight)
	} else {
		// We're parsing the image with no pixel metrics so unread the byte for the
		// main read loop
		_ = buffer.UnreadByte()

		width, height := d.scanSize(data)
		bounds = image.Rect(0, 0, width, height)
	}

	img := image.NewRGBA(bounds)
	palette := buildDefaultDecodePalette()
	var currentX, currentBandY, currentPaletteIndex int

	for {
		b, err := buffer.ReadByte()
		if err != nil {
			return img, d.readError(err)
		}

		// Palette operation
		if b == sixelUseColor {
			_, err = fmt.Fscan(buffer, &currentPaletteIndex)
			if err != nil {
				return img, d.readError(err)
			}

			b, err = buffer.ReadByte()
			if err != nil {
				return img, d.readError(err)
			}

			if b != ';' {
				// If we're not defining a color, move on
				_ = buffer.UnreadByte()
				continue
			}

			var red, green, blue uint32
			// We only know how to read RGB colors, which is preceded by a 2
			_, err = fmt.Fscanf(buffer, "2;%d;%d;%d", &red, &green, &blue)
			if err != nil {
				return img, d.readError(err)
			}

			if red > 100 || green > 100 || blue > 100 {
				return img, fmt.Errorf("invalid palette color: %d,%d,%d", red, green, blue)
			}

			palette[currentPaletteIndex] = color.RGBA64{
				R: uint16(imageConvertChannel(red)),
				G: uint16(imageConvertChannel(green)),
				B: uint16(imageConvertChannel(blue)),
				A: 65525,
			}

			continue
		}

		// LF
		if b == sixelLineBreak {
			currentBandY++
			currentX = 0
			continue
		}

		// CR
		if b == sixelCarriageReturn {
			currentX = 0
			continue
		}

		// RLE operation
		count := 1
		if b == sixelRepeat {
			_, err = fmt.Fscan(buffer, &count)
			if err != nil {
				return img, d.readError(err)
			}

			// The next byte SHOULD be a pixel
			b, err = buffer.ReadByte()
			if err != nil {
				return img, d.readError(err)
			}
		}

		if b >= '?' && b <= '~' {
			color := palette[currentPaletteIndex]
			for i := 0; i < count; i++ {
				d.writePixel(currentX, currentBandY, b, color, img)
				currentX++
			}
		}
	}
}

// WritePixel will accept a sixel byte (from ? to ~) that defines 6 vertical pixels
// and write any filled pixels to the image
func (d *Decoder) writePixel(x int, bandY int, sixel byte, color color.Color, img *image.RGBA) {
	maskedSixel := (sixel - '?') & 63
	yOffset := 0
	for maskedSixel != 0 {
		if maskedSixel&1 != 0 {
			img.Set(x, bandY*6+yOffset, color)
		}

		yOffset++
		maskedSixel >>= 1
	}
}

// scanSize is only used for legacy sixel images that do not define pixel metrics
// near the header (technically permitted). In this case, we need to quickly scan
// the image to figure out what the height and width are. Different terminals
// treat unfilled pixels around the border of the image diffently, but in our case
// we will treat all pixels, even empty ones, as part of the image.  However,
// we will allow the image to end with an LF code without increasing the size
// of the image.
//
// In the interest of speed, this method doesn't really parse the image in any
// meaningful way: pixel codes (? to ~), and the RLE, CR, and LF indicators
// (!, $, -) cannot appear within a sixel image except as themselves, so we
// just ignore everything else.  The only thing we actually take the time to parse
// is the number after the RLE indicator to know how much width to add to the current
// line.
func (d *Decoder) scanSize(data []byte) (int, int) {
	var maxWidth, bandCount int
	buffer := bytes.NewBuffer(data)

	// Pixel values are ? to ~. Each one of these encountered increases the max width.
	// a - is a LF and increases the max band count by one.  a $ is a CR and resets
	// current width.  (char - '?') will get a 6-bit number and the highest bit is
	// the lowest y value, which we should use to increment maxBandPixels.
	//
	// a ! is a RLE indicator, and we should add the numeral to the current width
	var currentWidth int
	var newBand bool
	for {
		b, err := buffer.ReadByte()
		if err != nil {
			return maxWidth, bandCount * 6
		}

		if b == '-' {
			// LF
			currentWidth = 0
			// The image may end with an LF, so we shouldn't increment the band
			// count until we encounter a pixel
			newBand = true
		} else if b == '$' {
			// CR
			currentWidth = 0
		} else if b == '!' || (b >= '?' && b <= '~') {
			// Either an RLE operation or a single pixel

			var count int
			if b == '!' {
				// Get the run length for the RLE operation
				_, err := fmt.Fscan(buffer, &count)
				if err != nil {
					return maxWidth, bandCount * 6
				}
				// Decrement the RLE because the pixel code will follow the
				// RLE and that will count as pixel
				count--
			} else {
				count = 1
			}

			currentWidth += count

			if newBand {
				newBand = false
				bandCount++
			}
			if currentWidth > maxWidth {
				maxWidth = currentWidth
			}
		}
	}
}

// readError will take any error returned from a read method (ReadByte,
// FScanF, etc.) and either wrap or ignore the error. An encountered EOF
// indicates that it's time to return the completed image so we just
// return it.
func (d *Decoder) readError(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}

	return fmt.Errorf("failed to read sixel data: %w", err)
}
