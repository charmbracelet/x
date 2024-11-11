module examples

go 1.18

require (
	github.com/charmbracelet/x/ansi v0.4.5
	github.com/charmbracelet/x/cellbuf v0.0.6-0.20241106170917-eb0997d7d743
	github.com/charmbracelet/x/input v0.2.0
	github.com/charmbracelet/x/termios v0.1.0
	github.com/charmbracelet/x/vt v0.0.0-00010101000000-000000000000
	github.com/creack/pty v1.1.24
	github.com/faiface/pixel v0.10.0
	github.com/fogleman/gg v1.3.0
	golang.org/x/image v0.22.0
)

require (
	github.com/charmbracelet/colorprofile v0.1.7 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/faiface/glhf v0.0.0-20181018222622-82a6317ac380 // indirect
	github.com/faiface/mainthread v0.0.0-20171120011319-8b78f0a41ae3 // indirect
	github.com/go-gl/gl v0.0.0-20190320180904-bf2b1f2f34d7 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20191125211704-12ad95a8df72 // indirect
	github.com/go-gl/mathgl v0.0.0-20190416160123-c4601bc793c7 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
)

require (
	github.com/charmbracelet/x/term v0.2.0
	github.com/charmbracelet/x/wcwidth v0.0.0-20241011142426-46044092ad91 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.27.0
	golang.org/x/text v0.20.0 // indirect
)

replace github.com/charmbracelet/x/ansi => ../ansi

replace github.com/charmbracelet/x/cellbuf => ../cellbuf

replace github.com/charmbracelet/x/term => ../term

replace github.com/charmbracelet/x/input => ../input

replace github.com/charmbracelet/x/vt => ../vt

replace github.com/charmbracelet/x/windows => ../windows

replace github.com/charmbracelet/x/exp => ../exp

replace github.com/charmbracelet/colorprofile => ../../colorprofile
