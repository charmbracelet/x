//go:build windows
// +build windows

package xpty

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func (c *ConPty) start(cmd *exec.Cmd) error {
	pid, proc, err := c.Spawn(cmd.Path, cmd.Args, &syscall.ProcAttr{
		Dir: cmd.Dir,
		Env: cmd.Env,
		Sys: cmd.SysProcAttr,
	})
	if err != nil {
		return err //nolint:wrapcheck
	}

	cmd.Process, err = os.FindProcess(pid)
	if err != nil {
		// If we can't find the process via os.FindProcess, terminate the
		// process as that's what we rely on for all further operations on the
		// object.
		if tErr := windows.TerminateProcess(windows.Handle(proc), 1); tErr != nil {
			return fmt.Errorf("failed to terminate process after process not found: %w", tErr)
		}
		return fmt.Errorf("failed to find process after starting: %w", err)
	}

	return nil
}
