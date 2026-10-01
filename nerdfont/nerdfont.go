// Package nerdfont detects whether the terminal is expected to render Nerd
// Font glyphs, and picks between a Nerd Font glyph and a portable fallback
// accordingly.
//
// Detection cannot know which font the terminal actually uses. It combines
// the environment (terminals that ship their own symbols font) with a scan
// for installed Nerd Font or Powerline fonts. Set [EnvVar] to override the
// result.
package nerdfont

import (
	"os"
	"sync"
)

// EnvVar overrides detection. Values understood as true or false (1/0,
// t/f, true/false, yes/no, on/off, case-insensitive) force the result.
// "auto" and unrecognized values fall through to probing.
const EnvVar = "NERDFONT"

// Result is the outcome of one probe.
type Result struct {
	// Supported reports whether Nerd Font glyphs are expected to render.
	Supported bool
	// Reason is a human-readable explanation of Supported, suitable for
	// debug logs.
	Reason string
}

// Glyph returns nerd when the terminal is expected to render Nerd Font
// glyphs, and fallback otherwise. Callers own both strings; an empty
// fallback means no glyph is rendered at all.
func Glyph(nerd, fallback string) string {
	if Supported() {
		return nerd
	}
	return fallback
}

// Detect probes the environment for Nerd Font support. Every call performs a
// fresh probe, which may spawn fc-list or read font directories, so prefer
// [Supported] on hot paths.
func Detect() Result {
	return probe(os.LookupEnv, fontDirs(), fontFamilies)
}

// Supported reports whether the terminal is expected to render Nerd Font
// glyphs. The first probe is memoized, so it is safe to call from render
// paths.
func Supported() bool {
	return cached().Supported
}

// Reason returns a human-readable explanation of the memoized [Supported]
// result.
func Reason() string {
	return cached().Reason
}

// Reset clears the memoized probe result so the next call to [Supported]
// probes again. It exists for tests.
func Reset() {
	mu.Lock()
	cache = nil
	mu.Unlock()
}

var (
	mu    sync.Mutex
	cache *Result
)

// cached returns the memoized probe result, probing on first use.
func cached() Result {
	mu.Lock()
	defer mu.Unlock()
	if cache == nil {
		result := probe(os.LookupEnv, fontDirs(), fontFamilies)
		cache = &result
	}
	return *cache
}
