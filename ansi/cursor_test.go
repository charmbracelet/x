package ansi

import (
	"io"
	"testing"
)

func BenchmarkCursorPosition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, CursorPosition(10, 20))
	}
}

func BenchmarkWriteCursorPosition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = WriteCursorPosition(io.Discard, 10, 20)
	}
}
