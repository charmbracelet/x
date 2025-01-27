package cellbuf

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

func TestReadStyleColor(t *testing.T) {
	tests := []struct {
		name      string
		params    []ansi.Param
		wantN     int
		wantColor color.Color
		wantNil   bool
	}{
		{
			name:    "invalid - too few parameters",
			params:  []ansi.Param{38},
			wantN:   0,
			wantNil: true,
		},
		{
			name:    "implementation defined",
			params:  []ansi.Param{38, 0},
			wantN:   2,
			wantNil: true,
		},
		{
			name:      "transparent",
			params:    []ansi.Param{38, 1},
			wantN:     2,
			wantColor: color.Transparent,
		},
		{
			name:      "RGB semicolon separated",
			params:    []ansi.Param{38, 2, 100, 150, 200},
			wantN:     5,
			wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 255},
		},
		{
			name: "RGB colon separated",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag,
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200,
			},
			wantN:     5,
			wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 255},
		},
		{
			name: "RGB with color space",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag,
				1 | parser.HasMoreFlag, // color space id
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200,
			},
			wantN:     6,
			wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 255},
		},
		// {
		// 	name:      "CMY semicolon separated",
		// 	params:    []ansi.Parameter{38, 3, 100, 150, 200},
		// 	wantN:     5,
		// 	wantColor: color.CMYK{C: 100, M: 150, Y: 200, K: 0},
		// },
		{
			name: "CMY with color space",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				3 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag, // color space id
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200,
			},
			wantN:     6,
			wantColor: color.CMYK{C: 100, M: 150, Y: 200, K: 0},
		},
		// {
		// 	name: "CMY colon separated",
		// 	params: []ansi.Parameter{
		// 		38 | parser.HasMoreFlag,
		// 		3 | parser.HasMoreFlag,
		// 		100 | parser.HasMoreFlag,
		// 		150 | parser.HasMoreFlag,
		// 		200,
		// 	},
		// 	wantN:     5,
		// 	wantColor: color.CMYK{C: 100, M: 150, Y: 200, K: 0},
		// },
		// {
		// 	name:      "CMYK semicolon separated",
		// 	params:    []ansi.Parameter{38, 4, 100, 150, 200, 50},
		// 	wantN:     6,
		// 	wantColor: color.CMYK{C: 100, M: 150, Y: 200, K: 50},
		// },
		{
			name: "CMYK with color space",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				4 | parser.HasMoreFlag,
				1 | parser.HasMoreFlag, // color space id
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200 | parser.HasMoreFlag,
				50,
			},
			wantN:     7,
			wantColor: color.CMYK{C: 100, M: 150, Y: 200, K: 50},
		},
		// {
		// 	name: "CMYK colon separated",
		// 	params: []ansi.Parameter{
		// 		38 | parser.HasMoreFlag,
		// 		4 | parser.HasMoreFlag,
		// 		100 | parser.HasMoreFlag,
		// 		150 | parser.HasMoreFlag,
		// 		200 | parser.HasMoreFlag,
		// 		50,
		// 	},
		// 	wantN:     6,
		// 	wantColor: color.CMYK{C: 100, M: 150, Y: 200, K: 50},
		// },
		{
			name:      "indexed color semicolon",
			params:    []ansi.Param{38, 5, 123},
			wantN:     3,
			wantColor: ansi.ExtendedColor(123),
		},
		{
			name: "indexed color colon",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				5 | parser.HasMoreFlag,
				123,
			},
			wantN:     3,
			wantColor: ansi.ExtendedColor(123),
		},
		{
			name:    "invalid color type",
			params:  []ansi.Param{38, 99},
			wantN:   0,
			wantNil: true,
		},
		{
			name: "RGB with tolerance and color space",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag,
				1 | parser.HasMoreFlag, // color space id
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200 | parser.HasMoreFlag,
				0 | parser.HasMoreFlag, // tolerance value
				1,                      // tolerance color space
			},
			wantN:     8,
			wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 255},
		},
		// Invalid cases
		{
			name:    "empty params",
			params:  []ansi.Param{},
			wantN:   0,
			wantNil: true,
		},
		{
			name:    "single param",
			params:  []ansi.Param{38},
			wantN:   0,
			wantNil: true,
		},
		{
			name:    "nil params",
			params:  nil,
			wantN:   0,
			wantNil: true,
		},
		// Mixed separator cases (should fail)
		{
			name: "RGB mixed separators",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2,                        // semicolon
				100 | parser.HasMoreFlag, // colon
				150,                      // semicolon
				200,
			},
			wantN:   0,
			wantNil: true,
		},
		{
			name: "CMYK mixed separators",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				4,                        // semicolon
				100 | parser.HasMoreFlag, // colon
				150,                      // semicolon
				200 | parser.HasMoreFlag, // colon
				50,
			},
			wantN:   0,
			wantNil: true,
		},
		// Edge cases
		{
			name: "RGB with max values",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag,
				255 | parser.HasMoreFlag,
				255 | parser.HasMoreFlag,
				255,
			},
			wantN:     5,
			wantColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "RGB with negative values",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag,
				-1 | parser.HasMoreFlag,
				-1 | parser.HasMoreFlag,
				-1,
			},
			wantN:   0,
			wantNil: true,
		},
		{
			name: "indexed color with out of range index",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				5 | parser.HasMoreFlag,
				256, // out of range
			},
			wantN:     3,
			wantColor: ansi.ExtendedColor(0),
		},
		{
			name: "indexed color with negative index",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				5 | parser.HasMoreFlag,
				-1,
			},
			wantN:   0,
			wantNil: true,
		},
		{
			name: "RGB truncated params",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				2 | parser.HasMoreFlag,
				100 | parser.HasMoreFlag,
				150,
			},
			wantN:   0,
			wantNil: true,
		},
		{
			name: "CMYK truncated params",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				4 | parser.HasMoreFlag,
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200,
			},
			wantN:   0,
			wantNil: true,
		},
		// RGBA (type 6) test cases
		// {
		// 	name:      "RGBA semicolon separated",
		// 	params:    []Parameter{38, 6, 100, 150, 200, 128},
		// 	wantN:     6,
		// 	wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 128},
		// },
		// {
		// 	name: "RGBA colon separated",
		// 	params: []ansi.Parameter{
		// 		38 | parser.HasMoreFlag,
		// 		6 | parser.HasMoreFlag,
		// 		100 | parser.HasMoreFlag,
		// 		150 | parser.HasMoreFlag,
		// 		200 | parser.HasMoreFlag,
		// 		128,
		// 	},
		// 	wantN:     6,
		// 	wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 128},
		// },
		{
			name: "RGBA with color space",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				6 | parser.HasMoreFlag,
				1 | parser.HasMoreFlag, // color space id
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200 | parser.HasMoreFlag,
				128,
			},
			wantN:     7,
			wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 128},
		},
		{
			name: "RGBA with tolerance and color space",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				6 | parser.HasMoreFlag,
				1 | parser.HasMoreFlag, // color space id
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200 | parser.HasMoreFlag,
				128 | parser.HasMoreFlag,
				0 | parser.HasMoreFlag, // tolerance value
				1,                      // tolerance color space
			},
			wantN:     9,
			wantColor: color.RGBA{R: 100, G: 150, B: 200, A: 128},
		},
		{
			name: "RGBA with max values",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				6 | parser.HasMoreFlag,
				0 | parser.HasMoreFlag, // color space id
				255 | parser.HasMoreFlag,
				255 | parser.HasMoreFlag,
				255 | parser.HasMoreFlag,
				255,
			},
			wantN:     7,
			wantColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "RGBA truncated params",
			params: []ansi.Param{
				38 | parser.HasMoreFlag,
				6 | parser.HasMoreFlag,
				100 | parser.HasMoreFlag,
				150 | parser.HasMoreFlag,
				200,
			},
			wantN:   0,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotColor color.Color
			gotN := ReadStyleColor(tt.params, &gotColor)
			if gotN != tt.wantN {
				t.Errorf("ReadColor() gotN = %v, want %v", gotN, tt.wantN)
			}
			if tt.wantNil {
				if gotColor != nil {
					t.Errorf("ReadColor() gotColor = %v, want nil", gotColor)
				}
				return
			}
			if gotColor != tt.wantColor {
				t.Errorf("ReadColor() gotColor = %v, want %v", gotColor, tt.wantColor)
			}
		})
	}
}
