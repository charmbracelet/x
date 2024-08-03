module examples

go 1.18

require (
	github.com/charmbracelet/x/ansi v0.1.4
	github.com/charmbracelet/x/exp/term v0.0.0-20240515162549-69ee4f765313
	github.com/charmbracelet/x/input v0.1.3
	github.com/charmbracelet/x/term v0.1.1
	github.com/rivo/uniseg v0.4.7
)

require (
	github.com/charmbracelet/x/windows v0.1.2 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.22.0 // indirect
)

replace github.com/charmbracelet/x/ansi => ../ansi

replace github.com/charmbracelet/x/term => ../term

replace github.com/charmbracelet/x/input => ../input

replace github.com/charmbracelet/x/windows => ../windows

replace github.com/charmbracelet/x/exp => ../exp
