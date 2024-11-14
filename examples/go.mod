module examples

go 1.18

require (
	github.com/charmbracelet/x/ansi v0.4.5
	github.com/charmbracelet/x/cellbuf v0.0.6-0.20241106170917-eb0997d7d743
	github.com/charmbracelet/x/input v0.2.0
)

require (
	github.com/charmbracelet/colorprofile v0.1.7 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
)

require (
	github.com/charmbracelet/x/term v0.2.0
	github.com/charmbracelet/x/wcwidth v0.0.0-20241011142426-46044092ad91 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace github.com/charmbracelet/x/term => ../term

replace github.com/charmbracelet/x/input => ../input

replace github.com/charmbracelet/x/vt => ../vt

replace github.com/charmbracelet/x/windows => ../windows

replace github.com/charmbracelet/x/exp => ../exp

replace github.com/charmbracelet/colorprofile => ../../colorprofile/
