package ansi

import (
	"io"
	"testing"
)

func BenchmarkInBandResize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, InBandResize(80, 24, 1920, 1080))
	}
}

func BenchmarkWriteInBandResize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = WriteInBandResize(io.Discard, 80, 24, 1920, 1080)
	}
}
