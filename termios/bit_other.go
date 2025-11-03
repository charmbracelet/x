//go:build !darwin && !netbsd && !openbsd && !windows
// +build !darwin,!netbsd,!openbsd,!windows

package termios

func speed(b uint32) uint32 { return b }
func bit(b uint32) uint32   { return b }
