# x

<p>
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="https://user-images.githubusercontent.com/25087/236529178-465e9b98-3401-47dd-8691-ea475d96c3ad.png" height="200" />
    <source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/25087/236529273-6f8c841f-f11b-4ec8-b01d-7e3d9b17c85f.png" height="200" />
    <img src="https://user-images.githubusercontent.com/25087/236529178-465e9b98-3401-47dd-8691-ea475d96c3ad.png" height="200" alt="A 3D rendering of an X"/>
  </picture>
  <br><a href="https://github.com/charmbracelet/x/actions"><img src="https://github.com/charmbracelet/x/workflows/build/badge.svg" alt="Build Status"></a>
</p>

This repository contains experimental packages with no promises of
backwards compatibility. Once they mature here, they might be moved
into other repositories.

Currently the following packages are available:

- [`ansi`](./ansi): ANSI escape sequence parser and definitions • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/ansi)
- [`conpty`](./conpty): Windows Console Pseudo-terminal library • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/conpty)
- [`editor`](./editor): open files in text editors • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/editor)
- [`errors`](./errors): `errors.Join` in older Go versions • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/errors)
- [`golden`](./exp/golden): verify golden file equality • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/golden)
- [`higherorder`](./exp/higherorder): generic higher order functions • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/higherorder)
- [`input`](./input): terminal event input handler and driver • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/input)
- [`json`](./json): JSON parsing using generics • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/json)
- [`open`](./exp/open): open a file/URL using `open`, `xdg-open`, etc • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/open)
- [`ordered`](./exp/ordered): generic `min`, `max`, and `clamp` functions for ordered types • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/ordered)
- [`maps`](./exp/maps): generic maps utilities
- [`slice`](./exp/slice): generic slice utilities • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/slice)
- [`sshkey`](./sshkey): open and parse SSH keys, asks for passphrases when needed • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/sshkey)
- [`strings`](./exp/strings): utilities for working with strings • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/strings)
- [`teatest`](./exp/teatest): a library for testing [Bubble Tea](https://github.com/charmbracelet/bubbletea) programs • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/exp/teatest)
- [`term`](./term): terminal utilities and helpers • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/term)
- [`termios`](./termios): Termios unified API and library • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/termios)
- [`windows`](./windows): Windows API used at Charmbracelet • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/windows)
- [`xpty`](./xpty): cross-platform PTY interface • [Docs](https://pkg.go.dev/github.com/charmbracelet/x/xpty)

## Feedback

We'd love to hear your thoughts on this project. Feel free to drop us a note!

- [Twitter](https://twitter.com/charmcli)
- [The Fediverse](https://mastodon.social/@charmcli)
- [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/charmbracelet/x/raw/main/LICENSE)

---

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source • نحنُ نحب المصادر المفتوحة
