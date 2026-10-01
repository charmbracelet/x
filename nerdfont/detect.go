package nerdfont

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"
)

// fcListTimeout bounds the fc-list probe. It is only a fallback signal, so a
// slow or missing fc-list is treated as "no fonts found".
const fcListTimeout = 500 * time.Millisecond

// lookupEnv reports the value of an environment variable.
type lookupEnv func(string) (string, bool)

// probe reports Nerd Font support. getenv supplies the environment, dirs are
// the font directories to scan, and families returns the installed font
// family names (or raw fc-list lines). They are injectable so that tests do
// not depend on the host. Overrides are applied by [resolve] before probing.
func probe(getenv lookupEnv, dirs []string, families func() []string) Result {
	if getenv == nil {
		getenv = os.LookupEnv
	}

	if name, ok := terminalWithBuiltinSymbols(getenv); ok {
		return Result{true, name + " ships built-in Nerd Font symbols"}
	}

	if families != nil {
		for _, family := range families() {
			if hasNerdFontName(family) {
				return Result{true, "installed font " + familyName(family)}
			}
		}
	}

	if name, ok := scanFontDirs(dirs); ok {
		return Result{true, "installed font " + name}
	}

	return Result{false, "no Nerd Font detected"}
}

// parseOverride interprets an [EnvVar] value. ok is false for empty, "auto",
// and unrecognized values, which all mean "keep probing".
func parseOverride(value string) (supported, ok bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "yes", "on":
		return true, true
	case "0", "f", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

// terminalWithBuiltinSymbols reports terminals that render Nerd Font glyphs
// without a patched font installed: kitty bundles the Symbols Nerd Font as a
// glyph fallback (0.36+), and Ghostty embeds a symbols-only Nerd Font.
func terminalWithBuiltinSymbols(getenv lookupEnv) (string, bool) {
	if hasEnv(getenv, "KITTY_WINDOW_ID") {
		return "kitty", true
	}
	if hasEnv(getenv, "GHOSTTY_RESOURCES_DIR") {
		return "ghostty", true
	}
	switch {
	case getEnv(getenv, "TERM") == "xterm-kitty":
		return "kitty", true
	case getEnv(getenv, "TERM") == "xterm-ghostty":
		return "ghostty", true
	case strings.EqualFold(getEnv(getenv, "TERM_PROGRAM"), "ghostty"):
		return "ghostty", true
	}
	return "", false
}

// hasEnv reports whether key is set to a non-empty value.
func hasEnv(getenv lookupEnv, key string) bool {
	value, ok := getenv(key)
	return ok && value != ""
}

// getEnv returns the value of key, or the empty string when unset.
func getEnv(getenv lookupEnv, key string) string {
	value, _ := getenv(key)
	return value
}

// hasNerdFontName reports whether a font family, an fc-list line, or a font
// file name looks like a Nerd Font or a Powerline-patched font.
func hasNerdFontName(name string) bool {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "nerd") || strings.Contains(lower, "powerline") {
		return true
	}
	// Powerline-style families are often suffixed with "NF", as in
	// "MesloLGS NF" or "MesloLGS-NF-Regular.ttf".
	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == ':' || r == '/' || r == '-'
	})
	return slices.Contains(tokens, "nf")
}

// familyName shortens an fc-list line to the font file name, which reads
// better in a log than the raw "file: family: style" line.
func familyName(line string) string {
	if path, _, ok := strings.Cut(line, ":"); ok {
		if base := filepath.Base(strings.TrimSpace(path)); base != "" && base != "." && base != string(filepath.Separator) {
			return base
		}
	}
	return strings.TrimSpace(line)
}

// scanFontDirs returns the name of the first Nerd Font file found in dirs.
func scanFontDirs(dirs []string) (string, bool) {
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if hasNerdFontName(entry.Name()) {
				return entry.Name(), true
			}
		}
	}
	return "", false
}

// fontDirs returns the directories fonts are installed in on the current
// platform. Terminals fall back to any installed font for glyphs missing
// from their primary font, so a Nerd Font in one of these usually renders
// even when it is not the configured font.
func fontDirs() []string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return compact([]string{
			underHome(home, "Library", "Fonts"),
			"/Library/Fonts",
			"/System/Library/Fonts",
		})
	case "windows":
		return compact([]string{
			underEnv("WINDIR", "Fonts"),
			underEnv("LOCALAPPDATA", "Microsoft", "Windows", "Fonts"),
		})
	default:
		return compact([]string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			underHome(home, ".fonts"),
			underHome(home, ".local", "share", "fonts"),
		})
	}
}

// underHome joins elems onto home, returning the empty string when home is
// unknown so that missing home directories do not become relative paths.
func underHome(home string, elems ...string) string {
	if home == "" {
		return ""
	}
	return filepath.Join(append([]string{home}, elems...)...)
}

// underEnv joins elems onto the value of the environment variable key,
// returning the empty string when it is unset.
func underEnv(key string, elems ...string) string {
	value := os.Getenv(key)
	if value == "" {
		return ""
	}
	return filepath.Join(append([]string{value}, elems...)...)
}

// compact drops empty entries.
func compact(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if path != "" {
			out = append(out, path)
		}
	}
	return out
}

// fontFamilies lists installed fonts using fc-list when it is available.
// A missing or slow fc-list is not an error: the directory scan covers it.
func fontFamilies() []string {
	path, err := exec.LookPath("fc-list")
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), fcListTimeout)
	defer cancel()
	//nolint:gosec // the command path is resolved from PATH by exec.LookPath.
	out, err := exec.CommandContext(ctx, path).Output()
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}
