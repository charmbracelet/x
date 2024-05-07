//go:build !linux && !darwin && !freebsd && !dragonfly && !netbsd && !openbsd && !solaris
// +build !linux,!darwin,!freebsd,!dragonfly,!netbsd,!openbsd,!solaris

package xpty

func (p *UnixPty) setWinsize(int, int, int, int) error {
	return ErrUnsupported
}

func (*UnixPty) size() (int, int, error) {
	return 0, 0, ErrUnsupported
}
