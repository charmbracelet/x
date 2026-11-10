package kitty

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"testing"
)

// BenchmarkPNG measures encoding to a fresh buffer, without buffer pooling.
// Frames are deterministic, opaque 1134x756 NRGBA gradients or RGB noise.
// Frame generation is excluded. Encoder measures PNG only; Graphics also
// includes base64 and APC chunk framing. Neither measures terminal rendering.
func BenchmarkPNG(b *testing.B) {
	for _, content := range []string{"smooth", "noise"} {
		img := image.NewNRGBA(image.Rect(0, 0, 1134, 756))
		random := rand.New(rand.NewPCG(1, 2))
		for y := 0; y < img.Rect.Dy(); y++ {
			for x := 0; x < img.Rect.Dx(); x++ {
				c := color.NRGBA{uint8(x * 255 / 1134), uint8(y * 255 / 756), uint8((x + y) * 255 / 1890), 255}
				if content == "noise" {
					c = color.NRGBA{uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32()), 255}
				}
				img.SetNRGBA(x, y, c)
			}
		}
		for _, level := range []struct {
			name  string
			value png.CompressionLevel
		}{{"default", png.DefaultCompression}, {"best-speed", png.BestSpeed}} {
			for _, graphics := range []bool{false, true} {
				mode := "Encoder"
				if graphics {
					mode = "Graphics"
				}
				b.Run(content+"/"+level.name+"/"+mode, func(b *testing.B) {
					enc := Encoder{Format: PNG, PNGCompressionLevel: level.value}
					opts := Options{Transmission: Direct, Format: PNG, PNGCompressionLevel: level.value, Chunk: true, Quiet: 2}
					// Measure PNG/base64 sizes once outside the timed loop.
					var sample bytes.Buffer
					if err := enc.Encode(&sample, img); err != nil {
						b.Fatal(err)
					}
					pngSize := sample.Len()
					b.ReportAllocs()
					b.SetBytes(int64(len(img.Pix)))
					outputSize := 0
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var out bytes.Buffer
						var err error
						if graphics {
							err = EncodeGraphics(&out, img, &opts)
						} else {
							err = enc.Encode(&out, img)
						}
						if err != nil {
							b.Fatal(err)
						}
						outputSize = out.Len()
					}
					b.StopTimer()
					b.ReportMetric(float64(pngSize), "PNG-bytes/op")
					b.ReportMetric(float64(base64.StdEncoding.EncodedLen(pngSize)), "base64-bytes/op")
					b.ReportMetric(float64(outputSize), "output-bytes/op")
				})
			}
		}
	}
}
