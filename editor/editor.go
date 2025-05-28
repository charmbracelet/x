// Package editor provides a simple way to open files in a text editor.
package editor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const defaultEditor = "nano"

// Option defines an editor option.
//
// An Option may act differently in some editors, or not be supported in
// some of them.
type Option func(editor, filename string) (args []string, pathInArgs bool)

// OpenAtLine opens the file at the given line number in supported editors.
//
// Deprecated: use LineNumber instead.
func OpenAtLine(n int) Option { return LineNumber(n) }

// LineNumber opens the file at the given line number in supported editors. If
// [number] is less than line 1, the file will be opened at line 1.
func LineNumber(number int) Option {
	if number < 1 {
		number = 1
	}
	plusLineEditors := []string{"vi", "vim", "nvim", "nano", "emacs", "kak", "gedit"}
	return func(editor, filename string) ([]string, bool) {
		if slices.Contains(plusLineEditors, editor) {
			return []string{fmt.Sprintf("+%d", number)}, false
		}
		if editor == "code" {
			return []string{
				"--goto",
				fmt.Sprintf("%s:%d", filename, number),
			}, true
		}
		return nil, false
	}
}

// EndOfLine opens the file at the end of the line in supported editors.
func EndOfLine() Option {
	return func(editor, _ string) (args []string, pathInArgs bool) {
		switch editor {
		case "vim", "nvim":
			return []string{"+norm! $"}, false
		}
		return nil, false
	}
}

// Cmd returns a *exec.Cmd editing the given path with $EDITOR or nano if no
// $EDITOR is set.
// Deprecated: use Command or CommandContext instead.
func Cmd(app, path string, options ...Option) (*exec.Cmd, error) {
	return CommandContext(context.Background(), app, path, options...)
}

// Command returns a *exec.Cmd editing the given path with $EDITOR or nano if
// no $EDITOR is set.
func Command(app, path string, options ...Option) (*exec.Cmd, error) {
	return CommandContext(context.Background(), app, path, options...)
}

// CommandContext returns a *exec.Cmd editing the given path with $EDITOR or nano
// if no $EDITOR is set.
func CommandContext(ctx context.Context, app, path string, options ...Option) (*exec.Cmd, error) {
	if os.Getenv("SNAP_REVISION") != "" {
		//nolint:staticcheck
		return nil, fmt.Errorf("Did you install with Snap? %[1]s is sandboxed and unable to open an editor. Please install %[1]s with Go or another package manager to enable editing.", app) //nolint:revive
	}

	editor, args := getEditor()
	editorName := filepath.Base(editor)

	needsToAppendPath := true
	for _, opt := range options {
		optArgs, pathInArgs := opt(editorName, path)
		if pathInArgs {
			needsToAppendPath = false
		}
		args = append(args, optArgs...)
	}
	if needsToAppendPath {
		args = append(args, path)
	}

	return exec.CommandContext(ctx, editor, args...), nil
}

func getEditor() (string, []string) {
	editor := strings.Fields(os.Getenv("EDITOR"))
	if len(editor) > 1 {
		return editor[0], editor[1:]
	}
	if len(editor) == 1 {
		return editor[0], []string{}
	}
	return defaultEditor, []string{}
}
