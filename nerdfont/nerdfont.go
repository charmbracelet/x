// Package nerdfont detects whether the terminal is expected to render Nerd
// Font glyphs, and picks between a Nerd Font glyph and a portable fallback
// accordingly.
//
// Detection cannot know which font the terminal actually uses. It combines
// the environment (terminals that ship their own symbols font) with a scan
// for installed Nerd Font or Powerline fonts. Set [EnvVar] or [SetOverride]
// to override the result.
package nerdfont

import (
	"fmt"
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
// [Supported] on hot paths. Overrides set through [EnvVar] or [SetOverride]
// are applied first.
func Detect() Result {
	mu.Lock()
	override := programOverride
	mu.Unlock()
	return resolve(os.LookupEnv, override)
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

// SetOverride forces the detection result for subsequent probes. Use it to
// wire an application-level Nerd Font setting, before the first probe.
// [EnvVar] takes precedence when both are set.
func SetOverride(supported bool) {
	mu.Lock()
	programOverride = &supported
	cache = nil
	mu.Unlock()
}

// ClearOverride removes the override set by [SetOverride], restoring
// automatic detection.
func ClearOverride() {
	mu.Lock()
	programOverride = nil
	cache = nil
	mu.Unlock()
}

// Reset clears the memoized probe result and any override so that the next
// call to [Supported] probes again. It exists for tests.
func Reset() {
	ClearOverride()
}

var (
	mu    sync.Mutex
	cache *Result
	// programOverride holds the override set by SetOverride, if any.
	programOverride *bool
)

// cached returns the memoized probe result, resolving it on first use.
func cached() Result {
	mu.Lock()
	defer mu.Unlock()
	if cache == nil {
		result := resolve(os.LookupEnv, programOverride)
		cache = &result
	}
	return *cache
}

// resolve applies the [EnvVar] environment variable, then the programmatic
// override, and only then probes the terminal.
func resolve(getenv lookupEnv, override *bool) Result {
	if getenv == nil {
		getenv = os.LookupEnv
	}
	if value, ok := getenv(EnvVar); ok {
		if supported, parsed := parseOverride(value); parsed {
			return Result{supported, fmt.Sprintf("%s=%s override", EnvVar, value)}
		}
	}
	if override != nil {
		return Result{*override, "override set by the application"}
	}
	return probe(getenv, fontDirs(), fontFamilies)
}
