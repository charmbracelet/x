package nerdfont

import (
	"os"
	"path/filepath"
	"testing"
)

// envFrom turns a map into a [lookupEnv] function.
func envFrom(env map[string]string) lookupEnv {
	return func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}
}

// writeFont creates an empty font file in a fresh temporary directory and
// returns the directory.
func writeFont(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
		t.Fatalf("could not create font file: %v", err)
	}
	return dir
}

func TestProbe(t *testing.T) {
	t.Parallel()

	nerdDir := writeFont(t, "JetBrainsMonoNerdFont-Regular.ttf")
	powerlineDir := writeFont(t, "Meslo LG S for Powerline.ttf")
	plainDir := writeFont(t, "DejaVuSansMono.ttf")

	tests := []struct {
		name     string
		env      map[string]string
		dirs     []string
		families []string
		want     bool
	}{
		{
			name: "empty environment",
			want: false,
		},
		{
			name: "kitty window id",
			env:  map[string]string{"KITTY_WINDOW_ID": "1"},
			want: true,
		},
		{
			name: "kitty term",
			env:  map[string]string{"TERM": "xterm-kitty"},
			want: true,
		},
		{
			name: "ghostty resources dir",
			env:  map[string]string{"GHOSTTY_RESOURCES_DIR": "/usr/share/ghostty"},
			want: true,
		},
		{
			name: "ghostty term",
			env:  map[string]string{"TERM": "xterm-ghostty"},
			want: true,
		},
		{
			name: "ghostty term program",
			env:  map[string]string{"TERM_PROGRAM": "Ghostty"},
			want: true,
		},
		{
			name: "installed nerd font file",
			dirs: []string{plainDir, nerdDir},
			want: true,
		},
		{
			name: "installed powerline font file",
			dirs: []string{powerlineDir},
			want: true,
		},
		{
			name: "plain fonts only",
			dirs: []string{plainDir},
			want: false,
		},
		{
			name:     "fc-list nerd family",
			families: []string{"/usr/share/fonts/x.ttf: Symbols Nerd Font Mono:style=Regular"},
			want:     true,
		},
		{
			name:     "fc-list nf family",
			families: []string{"/usr/share/fonts/y.ttf: MesloLGS NF:style=Regular"},
			want:     true,
		},
		{
			name:     "fc-list plain family",
			families: []string{"/usr/share/fonts/z.ttf: DejaVu Sans Mono:style=Regular"},
			want:     false,
		},
		{
			name:     "missing font directory",
			dirs:     []string{filepath.Join(t.TempDir(), "missing")},
			families: []string{"DejaVu Sans Mono"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			families := func() []string { return tt.families }
			got := probe(envFrom(tt.env), tt.dirs, families)
			if got.Supported != tt.want {
				t.Errorf("probe() supported = %v, want %v", got.Supported, tt.want)
			}
			if got.Reason == "" {
				t.Error("probe() returned an empty reason")
			}
		})
	}
}

func TestParseOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value     string
		supported bool
		ok        bool
	}{
		{value: "1", supported: true, ok: true},
		{value: "true", supported: true, ok: true},
		{value: "TRUE", supported: true, ok: true},
		{value: "yes", supported: true, ok: true},
		{value: " on ", supported: true, ok: true},
		{value: "0", supported: false, ok: true},
		{value: "false", supported: false, ok: true},
		{value: "No", supported: false, ok: true},
		{value: "off", supported: false, ok: true},
		{value: "", ok: false},
		{value: "auto", ok: false},
		{value: "maybe", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			supported, ok := parseOverride(tt.value)
			if ok != tt.ok {
				t.Fatalf("parseOverride(%q) ok = %v, want %v", tt.value, ok, tt.ok)
			}
			if supported != tt.supported {
				t.Errorf("parseOverride(%q) supported = %v, want %v", tt.value, supported, tt.supported)
			}
		})
	}
}

func TestHasNerdFontName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "Symbols Nerd Font Mono", want: true},
		{name: "JetBrainsMono Nerd Font", want: true},
		{name: "MesloLGS NF", want: true},
		{name: "MesloLGS-NF-Regular.ttf", want: true},
		{name: "Meslo LG S for Powerline.ttf", want: true},
		{name: "/usr/share/fonts/nerd/JetBrainsMonoNerdFontMono-Bold.ttf: JetBrainsMono Nerd Font Mono:style=Bold", want: true},
		{name: "DejaVu Sans Mono", want: false},
		{name: "CascadiaCode.ttf", want: false},
		{name: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := hasNerdFontName(tt.name); got != tt.want {
				t.Errorf("hasNerdFontName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFamilyName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		line string
		want string
	}{
		{line: "/usr/share/fonts/x/SymbolsNerdFontMono-Regular.ttf: Symbols Nerd Font Mono:style=Regular", want: "SymbolsNerdFontMono-Regular.ttf"},
		{line: "Symbols Nerd Font Mono", want: "Symbols Nerd Font Mono"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := familyName(tt.line); got != tt.want {
				t.Errorf("familyName(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestFontDirsAreAbsolute(t *testing.T) {
	t.Parallel()

	for _, dir := range fontDirs() {
		if !filepath.IsAbs(dir) {
			t.Errorf("fontDirs() returned the relative path %q", dir)
		}
	}
}
