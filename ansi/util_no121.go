//go:build !go1.21
// +build !go1.21

package ansi

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func max[T ordered](a, b T) T { //nolint:predeclared
	if a > b {
		return a
	}
	return b
}
