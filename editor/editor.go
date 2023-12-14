package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultEditor = "nano"

// Get editors that have support for the `+[line number]` flag
// to edit the file at the current line that is in the pager
func getEditorsWithLineNumberSupport() []string {
	return []string{
		"vi",
		"vim",
		"nvim",
		"nano",
	}
}

func hasLineNumberSupport(editor string) bool {
	for _, supportedEditor := range getEditorsWithLineNumberSupport() {
		if editor == supportedEditor {
			return true
		}
	}
    return false
}

// Cmd returns a *exec.Cmd editing the given path with $EDITOR or nano if no
// $EDITOR is set.
func Cmd(app, path string, lineNumber_optional ...uint) (*exec.Cmd, error) {
	if os.Getenv("SNAP_REVISION") != "" {
		return nil, fmt.Errorf("Did you install with Snap? %[1]s is sandboxed and unable to open an editor. Please install %[1]s with Go or another package manager to enable editing.", app)
	}

	editor, args := getEditor()

	// Add line number to open the editor at if provided and a supported editor is being used
	if len(lineNumber_optional) == 1 && hasLineNumberSupport(editor) {
		lineNumber := lineNumber_optional[0]

		lineNumberArg := fmt.Sprintf("+%d", lineNumber)
		// Insert line position arg before file name and other flags (required for nano)
		args = append([]string{lineNumberArg}, args...)

	}

	return exec.Command(editor, append(args, path)...), nil
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
