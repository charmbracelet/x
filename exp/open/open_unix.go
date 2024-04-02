//go:build !windows
// +build !windows

package open

import (
	"context"
	"os/exec"
)

func buildCmd(ctx context.Context, app, path string) *exec.Cmd {
	if _, err := exec.LookPath("open"); err == nil {
		var arg []string
		if app != "" {
			arg = append(arg, "-a", app)
		}
		arg = append(arg, path)
		return exec.CommandContext(ctx, "open", arg...)
	}

	if _, err := exec.LookPath("xdg-open"); err == nil {
		if app == "" {
			return exec.CommandContext(ctx, app, path)
		}
		return exec.CommandContext(ctx, "xdg-open", path)
	}
	return nil
}
