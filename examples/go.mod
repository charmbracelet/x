module examples

go 1.18

require (
	github.com/charmbracelet/bubbletea/v2 v2.0.0-alpha.1.0.20241105155825-ead55032fd81
	github.com/charmbracelet/lipgloss/v2 v2.0.0-20241105145349-c8e32d1b422c
	github.com/charmbracelet/x/ansi v0.4.4
	github.com/charmbracelet/x/cellbuf v0.0.5
	github.com/lucasb-eyer/go-colorful v1.2.0
)

require (
	github.com/charmbracelet/colorprofile v0.1.6 // indirect
	github.com/charmbracelet/x/input v0.2.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
)

require (
	github.com/charmbracelet/x/term v0.2.0 // indirect
	github.com/charmbracelet/x/wcwidth v0.0.0-20241011142426-46044092ad91 // indirect
	github.com/charmbracelet/x/windows v0.2.0 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)

replace github.com/charmbracelet/x/ansi => ../ansi

replace github.com/charmbracelet/x/cellbuf => ../cellbuf

replace github.com/charmbracelet/x/term => ../term

replace github.com/charmbracelet/x/input => ../input

replace github.com/charmbracelet/x/windows => ../windows

replace github.com/charmbracelet/x/exp => ../exp
