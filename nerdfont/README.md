# nerdfont

Detect whether the terminal is expected to render [Nerd Font][nf] glyphs, and
pick between a Nerd Font glyph and a portable fallback accordingly.

```go
import "charm.land/x/nerdfont"

// The Powerline git branch glyph, or nothing on terminals that can't
// render it.
branch := nerdfont.Glyph("\ue0a0", "") + " main"
```

## Overriding detection

Applications that expose their own Nerd Font setting can wire it up with
`SetOverride` and `ClearOverride`. Call them before the first probe:

```go
if cfg.NerdFonts != nil {
    nerdfont.SetOverride(*cfg.NerdFonts)
} else {
    nerdfont.ClearOverride()
}
```

## Detection

The first call to `nerdfont.Supported()` (and `nerdfont.Glyph`) probes once and
memoizes the result, so it is safe to call from render paths. `nerdfont.Detect()`
probes again on every call.

Probes run in order, and the first match wins:

| Signal | Result |
| --- | --- |
| `NERDFONT` set to a true/false value | That value, no probing |
| kitty (`KITTY_WINDOW_ID`, `TERM=xterm-kitty`) | Supported: kitty bundles the Symbols Nerd Font as a glyph fallback (0.36+) |
| Ghostty (`GHOSTTY_RESOURCES_DIR`, `TERM=xterm-ghostty`, `TERM_PROGRAM=ghostty`) | Supported: Ghostty embeds a symbols-only Nerd Font |
| Installed font scan | Supported when a Nerd Font or Powerline-patched font is installed |
| Nothing matched | Not supported |

## Caveats

Detection cannot know which font the terminal actually uses, and the installed
font scan is a heuristic: a Nerd Font may be installed but not reachable from
the terminal, and older kitty versions (< 0.36) do not bundle symbols. Set the
`NERDFONT` environment variable to force the result:

```sh
NERDFONT=0 some-tui   # never use Nerd Font glyphs
NERDFONT=1 some-tui   # always use Nerd Font glyphs
```

## License

[MIT](https://github.com/charmbracelet/x/blob/main/LICENSE)

[nf]: https://www.nerdfonts.com
