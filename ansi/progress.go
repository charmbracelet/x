package ansi

import "strconv"

// ResetProgress is a sequence that resets the progress bar to its default
// state (hidden).
//
// OSC 9 ; 4 ; 0 BEL
//
// See: https://learn.microsoft.com/en-us/windows/terminal/tutorials/progress-bar-sequences
const ResetProgress = "\x1b]9;4;0\x07"

// SetProgress returns a sequence for setting the progress bar to a specific
// percentage (0-100) in the "default" state.
//
// OSC 9 ; 4 ; 1 ; Percentage BEL
//
// See: https://learn.microsoft.com/en-us/windows/terminal/tutorials/progress-bar-sequences
func SetProgress(percentage int) string {
	return "\x1b]9;4;1;" + strconv.Itoa(min(max(0, percentage), 100)) + "\x07"
}

// SetErrorProgress returns a sequence for setting the progress bar to a specific
// percentage (0-100) in the "Error" state..
//
// OSC 9 ; 4 ; 2 ; Percentage BEL
//
// See: https://learn.microsoft.com/en-us/windows/terminal/tutorials/progress-bar-sequences
func SetErrorProgress(percentage int) string {
	return "\x1b]9;4;2;" + strconv.Itoa(min(max(0, percentage), 100)) + "\x07"
}

// SetIndeterminateProgress is a sequence that sets the progress bar to the
// indeterminate state.
//
// OSC 9 ; 4 ; 3 BEL
//
// See: https://learn.microsoft.com/en-us/windows/terminal/tutorials/progress-bar-sequences
const SetIndeterminateProgress = "\x1b]9;4;3\x07"
