package ansi

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestSetProgramStatus(t *testing.T) {
	cases := []struct {
		name string
		in   ProgramStatus
		want string
	}{
		{"minimal", ProgramStatus{State: ProgramStateWorking}, "\x1b]7501;state=working\x07"},
		{
			"spec example",
			ProgramStatus{
				State: ProgramStateBlocked, Kind: ProgramStatusKindPermission, App: "terraform",
				Message: "Apply 3 to add, 1 to change, 0 to destroy?",
			},
			"\x1b]7501;state=blocked:kind=permission:app=terraform:msg=QXBwbHkgMyB0byBhZGQsIDEgdG8gY2hhbmdlLCAwIHRvIGRlc3Ryb3k/\x07",
		},
		{
			"all keys",
			ProgramStatus{
				State: ProgramStateWorking, ID: "us-east", Title: "US East",
				Progress: 40, HasProgress: true, Message: "Pushing image",
			},
			"\x1b]7501;state=working:id=us-east:progress=40:title=VVMgRWFzdA==:msg=UHVzaGluZyBpbWFnZQ==\x07",
		},
		{"progress clamped", ProgramStatus{State: ProgramStateBlocked, Progress: 300, HasProgress: true}, "\x1b]7501;state=blocked:progress=100\x07"},
		{"zero progress", ProgramStatus{State: ProgramStateWorking, HasProgress: true}, "\x1b]7501;state=working:progress=0\x07"},
		{"progress ignored for done", ProgramStatus{State: ProgramStateDone, Progress: 50, HasProgress: true}, "\x1b]7501;state=done\x07"},
		{"kind ignored for working", ProgramStatus{State: ProgramStateWorking, Kind: ProgramStatusKindAuth}, "\x1b]7501;state=working\x07"},
		{"unknown kind omitted", ProgramStatus{State: ProgramStateBlocked, Kind: "nope"}, "\x1b]7501;state=blocked\x07"},
		{"invalid app omitted", ProgramStatus{State: ProgramStateIdle, App: "my app"}, "\x1b]7501;state=idle\x07"},
		{"control chars stripped", ProgramStatus{State: ProgramStateError, Message: "  a\n\tb\x1b]2;x\x07 \u009bc  "}, "\x1b]7501;state=error:msg=" + base64.StdEncoding.EncodeToString([]byte("a b ]2;x   c")) + "\x07"},
		{"only control chars", ProgramStatus{State: ProgramStateError, Message: "\n\n"}, "\x1b]7501;state=error\x07"},
		{"clear ignores fields", ProgramStatus{State: ProgramStateClear, App: "x", Message: "y"}, ClearProgramStatus},
		{"unknown state", ProgramStatus{State: "busy"}, ""},
		{"invalid id char", ProgramStatus{State: ProgramStateIdle, ID: "a b"}, ""},
		{"invalid id empty segment", ProgramStatus{State: ProgramStateIdle, ID: "a//b"}, ""},
		{"invalid id depth", ProgramStatus{State: ProgramStateIdle, ID: "a/b/c/d/e/f/g/h/i"}, ""},
		{"invalid id segment length", ProgramStatus{State: ProgramStateIdle, ID: strings.Repeat("a", 33)}, ""},
		{"nested id", ProgramStatus{State: ProgramStateIdle, ID: "build/test"}, "\x1b]7501;state=idle:id=build/test\x07"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SetProgramStatus(c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestProgramStatusTruncation(t *testing.T) {
	// 'é' is 2 bytes, so 2048 bytes would split the last rune.
	msg := "x" + strings.Repeat("é", 1100)
	seq := SetProgramStatus(ProgramStatus{State: ProgramStateDone, Message: msg})
	enc := strings.TrimSuffix(strings.TrimPrefix(seq, "\x1b]7501;state=done:msg="), "\x07")
	dec, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatal(err)
	}
	if len(dec) != 2047 {
		t.Errorf("decoded length = %d, want 2047", len(dec))
	}
	if len(seq) > 4096 {
		t.Errorf("sequence too long: %d", len(seq))
	}
}

func TestClearProgramStatusID(t *testing.T) {
	if got, want := ClearProgramStatusID(""), ClearProgramStatus; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := ClearProgramStatusID("eu-west"), "\x1b]7501;state=clear:id=eu-west\x07"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := ClearProgramStatusID("bad id"); got != "" {
		t.Errorf("got %q, want empty", got)
	}
	if RequestProgramStatusSupport != "\x1b]7501;?\x07" {
		t.Errorf("unexpected query %q", RequestProgramStatusSupport)
	}
}
