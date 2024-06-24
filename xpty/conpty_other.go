//go:build !windows
// +build !windows

package xpty

import "os/exec"

func (c *ConPty) start(*exec.Cmd) error {
	return ErrUnsupported
}
