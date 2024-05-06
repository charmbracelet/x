//go:build !linux && !darwin && !freebsd && !dragonfly && !netbsd && !openbsd && !solaris
// +build !linux,!darwin,!freebsd,!dragonfly,!netbsd,!openbsd,!solaris

package xpty

func (*Pty) resize(int, int) error {
	return ErrUnsupported
}

func (*Pty) size() (int, int, error) {
	return 0, 0, ErrUnsupported
}
