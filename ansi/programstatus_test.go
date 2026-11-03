package ansi

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// programStatusBody strips the OSC 7501 introducer and BEL terminator.
func programStatusBody(seq string) string {
	return strings.TrimSuffix(strings.TrimPrefix(seq, "\x1b]7501;"), "\x07")
}

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
				Progress: Percent(40), Message: "Pushing image",
			},
			"\x1b]7501;state=working:id=us-east:progress=40:title=VVMgRWFzdA==:msg=UHVzaGluZyBpbWFnZQ==\x07",
		},
		{"progress clamped", ProgramStatus{State: ProgramStateBlocked, Progress: Percent(300)}, "\x1b]7501;state=blocked:progress=100\x07"},
		{"zero progress", ProgramStatus{State: ProgramStateWorking, Progress: Percent(0)}, "\x1b]7501;state=working:progress=0\x07"},
		{"indeterminate progress", ProgramStatus{State: ProgramStateWorking, Progress: ProgramProgress{}}, "\x1b]7501;state=working\x07"},
		{"progress ignored for done", ProgramStatus{State: ProgramStateDone, Progress: Percent(50)}, "\x1b]7501;state=done\x07"},
		{"kind ignored for working", ProgramStatus{State: ProgramStateWorking, Kind: ProgramStatusKindAuth}, "\x1b]7501;state=working\x07"},
		{"unknown kind omitted", ProgramStatus{State: ProgramStateBlocked, Kind: "nope"}, "\x1b]7501;state=blocked\x07"},
		{"invalid app omitted", ProgramStatus{State: ProgramStateIdle, App: "my app"}, "\x1b]7501;state=idle\x07"},
		{"control chars replaced one to one", ProgramStatus{State: ProgramStateError, Message: "a\n\tb\x1b]2;x\x07 \u009bc"}, "\x1b]7501;state=error:msg=" + b64("a  b ]2;x   c") + "\x07"},
		{"only control chars", ProgramStatus{State: ProgramStateError, Message: "\n\n"}, "\x1b]7501;state=error:msg=" + b64("  ") + "\x07"},
		{"invalid utf-8 replaced", ProgramStatus{State: ProgramStateDone, Title: "a\xffb"}, "\x1b]7501;state=done:title=" + b64("a\uFFFDb") + "\x07"},
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

func TestPercent(t *testing.T) {
	cases := []struct {
		in      ProgramProgress
		want    int
		wantSet bool
	}{
		{ProgramProgress{}, 0, false},
		{Percent(0), 0, true},
		{Percent(42), 42, true},
		{Percent(-5), 0, true},
		{Percent(1000), 100, true},
	}
	for _, c := range cases {
		if got, set := c.in.Value(); got != c.want || set != c.wantSet {
			t.Errorf("%+v: got (%d, %v), want (%d, %v)", c.in, got, set, c.want, c.wantSet)
		}
	}
	if Percent(40) != Percent(40) || Percent(40) == (ProgramProgress{}) {
		t.Error("ProgramProgress must compare by value")
	}
}

func TestProgramStatusValidate(t *testing.T) {
	cases := []struct {
		name string
		in   ProgramStatus
		want error
	}{
		{"valid", ProgramStatus{State: ProgramStateDone, ID: "a/b"}, nil},
		{"clear", ProgramStatus{State: ProgramStateClear}, nil},
		{"absent fields are not errors", ProgramStatus{State: ProgramStateWorking, App: "bad app", Kind: "nope"}, nil},
		{"empty state", ProgramStatus{}, ErrProgramStatusState},
		{"unknown state", ProgramStatus{State: "busy"}, ErrProgramStatusState},
		{"bad id", ProgramStatus{State: ProgramStateIdle, ID: "a//b"}, ErrProgramStatusID},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.in.Validate()
			if !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
			if (err == nil) != (SetProgramStatus(c.in) != "") {
				t.Errorf("Validate and SetProgramStatus disagree: err=%v", err)
			}
		})
	}
}

func TestParseProgramStatus(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want ProgramStatus
		ok   bool
	}{
		{
			"spec example",
			"state=blocked:kind=permission:app=terraform:msg=QXBwbHkgMyB0byBhZGQsIDEgdG8gY2hhbmdlLCAwIHRvIGRlc3Ryb3k/",
			ProgramStatus{State: ProgramStateBlocked, Kind: ProgramStatusKindPermission, App: "terraform", Message: "Apply 3 to add, 1 to change, 0 to destroy?"},
			true,
		},
		{"progress", "state=working:id=us-east:progress=40", ProgramStatus{State: ProgramStateWorking, ID: "us-east", Progress: Percent(40)}, true},
		{"whitespace trimmed", " state = done : app = cargo ", ProgramStatus{State: ProgramStateDone, App: "cargo"}, true},
		{"unpadded base64", "state=done:msg=" + strings.TrimRight(b64("hi"), "="), ProgramStatus{State: ProgramStateDone, Message: "hi"}, true},
		{"last key wins", "state=idle:state=done", ProgramStatus{State: ProgramStateDone}, true},
		{"unknown key ignored", "state=idle:future=1", ProgramStatus{State: ProgramStateIdle}, true},
		{"malformed pairs skipped", "state=idle:noequals:=x:Bad=1:app=a;b", ProgramStatus{State: ProgramStateIdle}, true},
		{"kind dropped for working", "state=working:kind=auth", ProgramStatus{State: ProgramStateWorking}, true},
		{"unknown kind absent", "state=blocked:kind=nope", ProgramStatus{State: ProgramStateBlocked}, true},
		{"progress dropped for done", "state=done:progress=40", ProgramStatus{State: ProgramStateDone}, true},
		{"progress out of range absent", "state=working:progress=101", ProgramStatus{State: ProgramStateWorking}, true},
		{"signed progress absent", "state=working:progress=+5", ProgramStatus{State: ProgramStateWorking}, true},
		{"invalid app absent", "state=idle:app=a,b", ProgramStatus{State: ProgramStateIdle}, true},
		{"clear keeps only id", "state=clear:id=job:app=x:msg=" + b64("y"), ProgramStatus{State: ProgramStateClear, ID: "job"}, true},
		{"feature detection", "?", ProgramStatus{}, false},
		{"empty", "", ProgramStatus{}, false},
		{"missing state", "app=x", ProgramStatus{}, false},
		{"unknown state", "state=busy", ProgramStatus{}, false},
		{"invalid id", "state=idle:id=a//b", ProgramStatus{}, false},
		{"empty id", "state=idle:id=", ProgramStatus{}, false},
		{"key too long", "state=idle:" + strings.Repeat("k", 17) + "=1", ProgramStatus{}, false},
		{"app too long", "state=idle:app=" + strings.Repeat("a", 33), ProgramStatus{}, false},
		{"bad base64", "state=done:msg=A", ProgramStatus{}, false},
		{"text outside value charset skipped", "state=done:msg=@@@@", ProgramStatus{State: ProgramStateDone}, true},
		{"control char in text", "state=done:msg=" + b64("a\nb"), ProgramStatus{}, false},
		{"c1 control char in text", "state=done:msg=" + b64("a\u0085b"), ProgramStatus{}, false},
		{"invalid utf-8 in text", "state=done:msg=" + b64("a\xffb"), ProgramStatus{}, false},
		{"title too long", "state=done:title=" + b64(strings.Repeat("t", 193)), ProgramStatus{}, false},
		{"message too long", "state=done:msg=" + b64(strings.Repeat("m", 2049)), ProgramStatus{}, false},
		{"limits checked for clear", "state=clear:msg=" + b64(strings.Repeat("m", 2049)), ProgramStatus{}, false},
		{"sequence too long", "state=idle:x=" + strings.Repeat("a", 4096), ProgramStatus{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := ParseProgramStatus(c.in)
			if ok != c.ok || got != c.want {
				t.Errorf("got (%+v, %v), want (%+v, %v)", got, ok, c.want, c.ok)
			}
		})
	}
}

func TestProgramStatusTruncation(t *testing.T) {
	// 'é' is 2 bytes, so 2048 bytes would split the last rune.
	msg := "x" + strings.Repeat("é", 1100)
	seq := SetProgramStatus(ProgramStatus{State: ProgramStateDone, Message: msg})
	if len(seq) > programStatusMaxSequence {
		t.Errorf("sequence too long: %d", len(seq))
	}
	got, ok := ParseProgramStatus(programStatusBody(seq))
	if !ok {
		t.Fatal("truncated report did not parse")
	}
	if len(got.Message) != 2047 || !strings.HasPrefix(msg, got.Message) {
		t.Errorf("decoded length = %d, want a 2047 byte prefix", len(got.Message))
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

// FuzzProgramStatusRoundTrip checks that every sequence SetProgramStatus
// writes parses, and re-encodes to the same bytes.
func FuzzProgramStatusRoundTrip(f *testing.F) {
	f.Add("working", "", "crush", "", 40, true, "", "Thinking")
	f.Add("blocked", "a/b", "x", "permission", 300, true, "T\n", "é\x1b")
	f.Add("clear", "job", "", "", 0, false, "", "")
	f.Add("busy", "", "", "", 0, false, "", "")
	f.Fuzz(func(t *testing.T, state, id, app, kind string, percent int, set bool, title, msg string) {
		s := ProgramStatus{
			State: ProgramState(state), ID: id, App: app, Kind: ProgramStatusKind(kind),
			Title: title, Message: msg,
		}
		if set {
			s.Progress = Percent(percent)
		}
		seq := SetProgramStatus(s)
		if (seq == "") != (s.Validate() != nil) {
			t.Fatalf("Validate and SetProgramStatus disagree for %+v", s)
		}
		if seq == "" {
			return
		}
		got, ok := ParseProgramStatus(programStatusBody(seq))
		if !ok {
			t.Fatalf("encoder output %q does not parse", seq)
		}
		if again := SetProgramStatus(got); again != seq {
			t.Fatalf("re-encoded %q, want %q", again, seq)
		}
	})
}

// FuzzParseProgramStatus checks that a parsed report is a fixed point of
// encoding and parsing.
func FuzzParseProgramStatus(f *testing.F) {
	f.Add("state=blocked:kind=permission:app=terraform:msg=QXBwbHk/")
	f.Add("state=working:id=us-east:title=VVMgRWFzdA==:progress=040")
	f.Add(" state = clear : id = a/b ")
	f.Add("?")
	f.Fuzz(func(t *testing.T, body string) {
		got, ok := ParseProgramStatus(body)
		if !ok {
			return
		}
		seq := SetProgramStatus(got)
		if seq == "" {
			t.Fatalf("parsed %+v from %q does not encode", got, body)
		}
		again, ok := ParseProgramStatus(programStatusBody(seq))
		if !ok || again != got {
			t.Fatalf("round trip of %q gave (%+v, %v), want %+v", body, again, ok, got)
		}
	})
}
