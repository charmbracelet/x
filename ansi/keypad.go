package ansi

// SetApplicationKeypadMode sets the application keypad mode.
// Application Keypad Mode (DECKPAM) is a mode that determines whether the
// keypad sends application sequences or ANSI sequences.
//
// Use [SetNumericKeypadMode] to set the numeric keypad mode.
//
// See: https://vt100.net/docs/vt510-rm/DECKPAM.html
const SetApplicationKeypadMode = "\x1b="

// SetNumericKeypadMode sets the numeric keypad mode.
// Numeric Keypad Mode (DECKPNM) is a mode that determines whether the keypad
// sends application sequences or ANSI sequences.
//
// Use [SetApplicationKeypadMode] to set the application keypad mode.
//
// See: https://vt100.net/docs/vt510-rm/DECKPNM.html
const SetNumericKeypadMode = "\x1b>"
