//go:build !darwin
// +build !darwin

package termios

func bit(b uint32) uint32 {
	return b
}
