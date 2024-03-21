package input

import (
	"testing"
)

func BenchmarkParseSequence(b *testing.B) {
	input := []byte("\x1b\x1b[Ztest\x00\x1b]10;1234/1234/1234\x07\x1b[27;2;27~")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseSequence(input)
	}
}
