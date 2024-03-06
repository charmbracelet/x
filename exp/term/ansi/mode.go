package ansi

// This file define uses multiple sequences to set (SM), reset (RM), and request
// (DECRQM) different ANSI and DEC modes.
//
// See: https://vt100.net/docs/vt510-rm/SM.html
// See: https://vt100.net/docs/vt510-rm/RM.html
// See: https://vt100.net/docs/vt510-rm/DECRQM.html

// Application Cursor Keys (DECCKM) is a mode that determines whether the
// cursor keys send ANSI cursor sequences or application sequences.
//
// See: https://vt100.net/docs/vt510-rm/DECCKM.html
const (
	EnableCursorKeys  = "\x1b[?1h"
	DisableCursorKeys = "\x1b[?1l"
	RequestCursorKeys = "\x1b[?1$p"
)

// Text Cursor Enable Mode (DECTCEM) is a mode that shows/hides the cursor.
//
// See: https://vt100.net/docs/vt510-rm/DECTCEM.html
const (
	ShowCursor              = "\x1b[?25h"
	HideCursor              = "\x1b[?25l"
	RequestCursorVisibility = "\x1b[?25$p"
)

// VT Mouse Tracking is a mode that determines whether the mouse reports on
// button press and release.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking
const (
	EnableMouseTracking  = "\x1b[?1000h"
	DisableMouseTracking = "\x1b[?1000l"
	RequestMouseTracking = "\x1b[?1000$p"
)

// VT Hilite Mouse Tracking is a mode that determines whether the mouse reports on
// button presses, releases, and highlighted cells.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking
const (
	EnableHiliteMouseTracking  = "\x1b[?1001h"
	DisableHiliteMouseTracking = "\x1b[?1001l"
	RequestHiliteMouseTracking = "\x1b[?1001$p"
)

// Cell Motion Mouse Tracking is a mode that determines whether the mouse
// reports on button press, release, and motion events.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking
const (
	EnableCellMotionMouseTracking  = "\x1b[?1002h"
	DisableCellMotionMouseTracking = "\x1b[?1002l"
	RequestCellMotionMouseTracking = "\x1b[?1002$p"
)

// All Mouse Tracking is a mode that determines whether the mouse reports on
// button press, release, motion, and highlight events.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking
const (
	EnableAllMouseTracking  = "\x1b[?1003h"
	DisableAllMouseTracking = "\x1b[?1003l"
	RequestAllMouseTracking = "\x1b[?1003$p"
)

// SGR Mouse Extension is a mode that determines whether the mouse reports events
// formatted with SGR parameters.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking
const (
	EnableSgrMouseExt  = "\x1b[?1006h"
	DisableSgrMouseExt = "\x1b[?1006l"
	RequestSgrMouseExt = "\x1b[?1006$p"
)

// Alternate Screen Buffer is a mode that determines whether the alternate screen
// buffer is active.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-The-Alternate-Screen-Buffer
const (
	EnableAltScreenBuffer  = "\x1b[?1049h"
	DisableAltScreenBuffer = "\x1b[?1049l"
	RequestAltScreenBuffer = "\x1b[?1049$p"
)

// Bracketed Paste Mode is a mode that determines whether pasted text is
// bracketed with escape sequences.
//
// See: https://cirw.in/blog/bracketed-paste
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Bracketed-Paste-Mode
const (
	EnableBracketedPaste  = "\x1b[?2004h"
	DisableBracketedPaste = "\x1b[?2004l"
	RequestBracketedPaste = "\x1b[?2004$p"
)

// Synchronized Output Mode is a mode that determines whether output is
// synchronized with the terminal.
//
// See: https://gist.github.com/christianparpart/d8a62cc1ab659194337d73e399004036
const (
	EnableSyncdOutput  = "\x1b[?2026h"
	DisableSyncdOutput = "\x1b[?2026l"
	RequestSyncdOutput = "\x1b[?2026$p"
)

// Win32Input is a mode that determines whether input is processed by the
// Win32 console and Conpty.
//
// See: https://github.com/microsoft/terminal/blob/main/doc/specs/%234999%20-%20Improved%20keyboard%20handling%20in%20Conpty.md
const (
	EnableWin32Input  = "\x1b[?9001h"
	DisableWin32Input = "\x1b[?9001l"
	RequestWin32Input = "\x1b[?9001$p"
)
