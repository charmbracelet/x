//go:build darwin
// +build darwin

package termios

func bit(b uint32) uint64 {
	return uint64(b)
}
