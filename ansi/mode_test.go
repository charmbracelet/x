package ansi

import (
	"testing"
)

func TestModeSetting_Methods(t *testing.T) {
	tests := []struct {
		name     string
		mode     ModeSetting
		notRecog bool
		isSet    bool
		isReset  bool
		permSet  bool
		permRst  bool
	}{
		{
			name:     "ModeNotRecognized",
			mode:     ModeNotRecognized,
			notRecog: true,
			isSet:    false,
			isReset:  false,
			permSet:  false,
			permRst:  false,
		},
		{
			name:     "ModeSet",
			mode:     ModeSet,
			notRecog: false,
			isSet:    true,
			isReset:  false,
			permSet:  false,
			permRst:  false,
		},
		{
			name:     "ModeReset",
			mode:     ModeReset,
			notRecog: false,
			isSet:    false,
			isReset:  true,
			permSet:  false,
			permRst:  false,
		},
		{
			name:     "ModePermanentlySet",
			mode:     ModePermanentlySet,
			notRecog: false,
			isSet:    true,
			isReset:  false,
			permSet:  true,
			permRst:  false,
		},
		{
			name:     "ModePermanentlyReset",
			mode:     ModePermanentlyReset,
			notRecog: false,
			isSet:    false,
			isReset:  true,
			permSet:  false,
			permRst:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsNotRecognized(); got != tt.notRecog {
				t.Errorf("IsNotRecognized() = %v, want %v", got, tt.notRecog)
			}
			if got := tt.mode.IsSet(); got != tt.isSet {
				t.Errorf("IsSet() = %v, want %v", got, tt.isSet)
			}
			if got := tt.mode.IsReset(); got != tt.isReset {
				t.Errorf("IsReset() = %v, want %v", got, tt.isReset)
			}
			if got := tt.mode.IsPermanentlySet(); got != tt.permSet {
				t.Errorf("IsPermanentlySet() = %v, want %v", got, tt.permSet)
			}
			if got := tt.mode.IsPermanentlyReset(); got != tt.permRst {
				t.Errorf("IsPermanentlyReset() = %v, want %v", got, tt.permRst)
			}
		})
	}
}

func TestSetMode(t *testing.T) {
	tests := []struct {
		name     string
		modes    []Mode
		expected string
	}{
		{
			name:     "empty modes",
			modes:    []Mode{},
			expected: "",
		},
		{
			name:     "single ANSI mode",
			modes:    []Mode{KeyboardActionMode},
			expected: "\x1b[2h",
		},
		{
			name:     "single DEC mode",
			modes:    []Mode{CursorKeysMode},
			expected: "\x1b[?1h",
		},
		{
			name:     "multiple ANSI modes",
			modes:    []Mode{KeyboardActionMode, InsertReplaceMode},
			expected: "\x1b[2;4h",
		},
		{
			name:     "multiple DEC modes",
			modes:    []Mode{CursorKeysMode, AutoWrapMode},
			expected: "\x1b[?1;7h",
		},
		{
			name:     "mixed ANSI and DEC modes",
			modes:    []Mode{KeyboardActionMode, CursorKeysMode},
			expected: "\x1b[2h\x1b[?1h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetMode(tt.modes...); got != tt.expected {
				t.Errorf("SetMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResetMode(t *testing.T) {
	tests := []struct {
		name     string
		modes    []Mode
		expected string
	}{
		{
			name:     "empty modes",
			modes:    []Mode{},
			expected: "",
		},
		{
			name:     "single ANSI mode",
			modes:    []Mode{KeyboardActionMode},
			expected: "\x1b[2l",
		},
		{
			name:     "single DEC mode",
			modes:    []Mode{CursorKeysMode},
			expected: "\x1b[?1l",
		},
		{
			name:     "multiple ANSI modes",
			modes:    []Mode{KeyboardActionMode, InsertReplaceMode},
			expected: "\x1b[2;4l",
		},
		{
			name:     "multiple DEC modes",
			modes:    []Mode{CursorKeysMode, AutoWrapMode},
			expected: "\x1b[?1;7l",
		},
		{
			name:     "mixed ANSI and DEC modes",
			modes:    []Mode{KeyboardActionMode, CursorKeysMode},
			expected: "\x1b[2l\x1b[?1l",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResetMode(tt.modes...); got != tt.expected {
				t.Errorf("ResetMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRequestMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		expected string
	}{
		{
			name:     "ANSI mode",
			mode:     KeyboardActionMode,
			expected: "\x1b[2$p",
		},
		{
			name:     "DEC mode",
			mode:     CursorKeysMode,
			expected: "\x1b[?1$p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequestMode(tt.mode); got != tt.expected {
				t.Errorf("RequestMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReportMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		value    ModeSetting
		expected string
	}{
		{
			name:     "ANSI mode not recognized",
			mode:     KeyboardActionMode,
			value:    ModeNotRecognized,
			expected: "\x1b[2;0$y",
		},
		{
			name:     "DEC mode set",
			mode:     CursorKeysMode,
			value:    ModeSet,
			expected: "\x1b[?1;1$y",
		},
		{
			name:     "ANSI mode reset",
			mode:     InsertReplaceMode,
			value:    ModeReset,
			expected: "\x1b[4;2$y",
		},
		{
			name:     "DEC mode permanently set",
			mode:     AutoWrapMode,
			value:    ModePermanentlySet,
			expected: "\x1b[?7;3$y",
		},
		{
			name:     "ANSI mode permanently reset",
			mode:     SendReceiveMode,
			value:    ModePermanentlyReset,
			expected: "\x1b[12;4$y",
		},
		{
			name:     "Invalid mode setting defaults to not recognized",
			mode:     KeyboardActionMode,
			value:    5,
			expected: "\x1b[2;0$y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReportMode(tt.mode, tt.value); got != tt.expected {
				t.Errorf("ReportMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestModeImplementations(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		expected int
	}{
		{
			name:     "ANSIMode",
			mode:     ANSIMode(42),
			expected: 42,
		},
		{
			name:     "DECMode",
			mode:     DECMode(99),
			expected: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.Mode(); got != tt.expected {
				t.Errorf("Mode() = %v, want %v", got, tt.expected)
			}
		})
	}
}
