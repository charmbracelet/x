//go:build windows
// +build windows

package open

import (
	"context"
	"os/exec"
)

func buildCmd(ctx context.Context, app, path string) *exec.Cmd {
	if app != "" {
		return exec.Command("cmd", "/C", "start", "", app, path)
	}
	return exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", path)
}
