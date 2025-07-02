// Package open provides functionality for opening files and URLs.
package open

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotSupported occurs when no ways to open a file are found.
var ErrNotSupported = errors.New("not supported")

// Open the given input.
func Open(ctx context.Context, input string) error {
	return With(ctx, "", input)
}

// With opens the given input using the given app.
func With(ctx context.Context, app, input string) error {
	cmd := buildCmd(ctx, app, input)
	if cmd == nil {
		return ErrNotSupported
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("open: %w: %s", err, string(out))
	}
	return nil
}
