# CLAUDE.md

## Project Overview

This is a fork of [charmbracelet/x](https://github.com/charmbracelet/x) — a Go monorepo of terminal and TUI libraries. The primary focus of this fork is extending the `vt/` package (terminal emulator) with features needed by [amux](https://github.com/weill-labs/amux).

## Architecture

**Go monorepo with per-package modules.** Each top-level directory (`vt/`, `ansi/`, `cellbuf/`, etc.) has its own `go.mod`. Changes to one package don't require updating others unless there's a dependency.

**`vt/` is the terminal emulator.** Key files:
- `vt/emulator.go` — core `Emulator` struct, `Write()` entry point, parser dispatch
- `vt/handlers.go` — CSI/ESC/DCS handler registration
- `vt/csi_mode.go` — DECSET/DECRST mode handling
- `vt/mode.go` — mode definitions and state
- `vt/key.go` — `SendKey()` for input encoding
- `vt/screen.go` — screen buffer management (main + alt)
- `vt/callbacks.go` — callback interface for mode changes, screen updates

**`ansi/` provides escape sequence constants and parsers.** The `ansi/parser/` sub-package implements the VT state machine. Higher-level types like `Mode`, `KittyKeyboard*`, and sequence builders live in `ansi/`.

## Development

### Build And Test

```bash
cd vt && go build ./...          # build vt package
cd vt && go test ./...           # run vt tests
cd vt && go test -run TestFoo -v # run specific test
```

Since this is a monorepo, always `cd` into the package directory before running Go commands.

### TDD Workflow

All development follows red-green-refactor with **separate commits** for each phase:

1. **Red** — Write failing tests. Commit them alone. Confirm they fail for the right reason.
2. **Green** — Minimal production code to make tests pass. Commit separately.
3. **Refactor** — Simplify, extract helpers, remove duplication. Commit separately.

### Testing

Use table-driven tests for unit tests with multiple cases. Call `t.Parallel()` in each subtest. Tests should read like specs — minimize logic in assertions so a human can read the test and immediately understand expected behavior.

### Pre-Push Rebase

Rebase onto `origin/main` before the first push: `git fetch origin main && git rebase origin/main`.

### PR Title And Description

**Title**: Imperative mood, under 70 characters. Example: "Add DEC 2026 synchronized output buffering to vt".

**Description** must include:
1. **Motivation** — Why this change?
2. **Summary** — What changed? Bullet the key changes.
3. **Testing** — How was it verified? Include exact test commands.
4. **Review focus** — What should reviewers look at?

Use matter-of-fact language. Avoid vague qualifiers like "robust" or "comprehensive".

## Patterns To Follow

**Inject dependencies, do not add package-level `var` for test seams.** Pass swappable dependencies as function parameters or struct fields, not mutable package-level variables.

**One concern per file.** When adding a new feature (e.g., synchronized output, kitty keyboard), create a dedicated file for it rather than stuffing logic into emulator.go.

**Mode handling pattern.** New terminal modes follow this pattern:
1. Define the mode constant in `ansi/` (or use an existing one)
2. Register DECSET/DECRST handlers in `vt/csi_mode.go`
3. Add state fields to the `Emulator` struct in `vt/emulator.go`
4. Implement behavior in a dedicated file (e.g., `vt/sync_output.go`)

**CSI handler registration.** New CSI sequences are registered in `vt/handlers.go` via the handler table.

## Upstream Sync

This fork tracks `upstream` (charmbracelet/x). Periodically sync:
```bash
git fetch upstream
git merge upstream/main
```

Avoid modifying files outside `vt/` unless necessary — this minimizes merge conflicts with upstream.

## After Landing Changes

Push to main on weill-labs/x, then update amux's go.mod to pin the new commit:
```bash
cd ~/sync/github/amux/amux
go get github.com/weill-labs/x/vt@COMMIT
go test ./...
```

## Issue Tracking

Track issues in the [amux Linear project](https://linear.app/weill-labs/project/amux-b3a52334f77c) with `LAB-7xx` identifiers.
