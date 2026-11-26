package ansi

import (
	"strings"
	"testing"
	"time"
)

// The token, pids, time, and macs come from the examples in the spec.
const programResumeToken = "dummy-token-replace-me"

var programResumeTime = time.Unix(1795543200, 0)

func TestSetProgramResume(t *testing.T) {
	cases := []struct {
		name string
		pid  int
		in   ProgramResume
		want string
	}{
		{
			"neovim", 74877,
			ProgramResume{Args: []string{"nvim", "-S", "/home/user/.local/state/nvim/sessions/8f2c1d.vim"}},
			"\x1b]7577;pid=74877:time=1795543200:argv=bnZpbQ==,LVM=,L2hvbWUvdXNlci8ubG9jYWwvc3RhdGUvbnZpbS9zZXNzaW9ucy84ZjJjMWQudmlt:mac=9dbb99bfb856749a87563839d7ce9062818d78cf25909db93f309e0cd3b1909b\x07",
		},
		{
			"lazygit with dir", 74877,
			ProgramResume{Args: []string{"lazygit"}, Dir: "/home/user/src/project"},
			"\x1b]7577;pid=74877:time=1795543200:argv=bGF6eWdpdA==:cwd=L2hvbWUvdXNlci9zcmMvcHJvamVjdA==:mac=35cb909805c380336efd3689a125acc774a9dec8422a1d21492ebec06ca758e1\x07",
		},
		{
			"wrapper", 90412,
			ProgramResume{Args: []string{"resume-wrap", "ssh", "devbox"}},
			"\x1b]7577;pid=90412:time=1795543200:argv=cmVzdW1lLXdyYXA=,c3No,ZGV2Ym94:mac=e76149d0ba54a83570ffab0583867dacdfa23928ba9d18604f0badea88b62b65\x07",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SetProgramResume(programResumeToken, c.pid, programResumeTime, c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestSetProgramResumeArgs(t *testing.T) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"empty argument":        {[]string{"a", ""}, "argv=YQ==,:"},
		"absolute program":      {[]string{"/usr/bin/vim"}, "argv=L3Vzci9iaW4vdmlt:"},
		"shell text stays text": {[]string{"sh", "$(reboot); a"}, "argv=c2g=,JChyZWJvb3QpOyBh:"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := SetProgramResume(programResumeToken, 1, programResumeTime, ProgramResume{Args: c.args})
			if !strings.Contains(got, ":"+c.want+"mac=") {
				t.Errorf("got %q, want it to contain %q", got, c.want)
			}
		})
	}
}

func TestSetProgramResumeInvalid(t *testing.T) {
	cases := map[string]ProgramResume{
		"no args":                    {Dir: "/"},
		"empty program":              {Args: []string{""}},
		"relative program":           {Args: []string{"./run.sh"}},
		"program in subdirectory":    {Args: []string{"bin/app"}},
		"program starting with dash": {Args: []string{"-sh"}},
		"relative dir":               {Args: []string{"nvim"}, Dir: "project"},
		"newline in argument":        {Args: []string{"nvim", "a\nb"}},
		"C1 control in argument":     {Args: []string{"nvim", "a\u0085b"}},
		"DEL in dir":                 {Args: []string{"nvim"}, Dir: "/a\x7f"},
		"invalid UTF-8":              {Args: []string{"nvim", "\xff"}},
	}
	for name, r := range cases {
		t.Run(name, func(t *testing.T) {
			if got := SetProgramResume(programResumeToken, 1, programResumeTime, r); got != "" {
				t.Errorf("got %q, want empty", got)
			}
		})
	}

	valid := ProgramResume{Args: []string{"nvim"}}
	if got := SetProgramResume("", 1, programResumeTime, valid); got != "" {
		t.Errorf("empty token: got %q, want empty", got)
	}
	if got := SetProgramResume(programResumeToken, 0, programResumeTime, valid); got != "" {
		t.Errorf("zero pid: got %q, want empty", got)
	}
	if got := SetProgramResume(programResumeToken, 1, time.Time{}, valid); got != "" {
		t.Errorf("zero time: got %q, want empty", got)
	}
}

func TestSetProgramResumeOverLimits(t *testing.T) {
	clearSeq := ClearProgramResume(programResumeToken, 1, programResumeTime)
	cases := map[string]struct {
		args       []string
		clearFirst bool
	}{
		"within limits": {append([]string{"nvim"}, make([]string, 63)...), false},
		"too many args": {append([]string{"nvim"}, make([]string, 64)...), true},
		"too long":      {[]string{"nvim", strings.Repeat("a", 2048)}, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := SetProgramResume(programResumeToken, 1, programResumeTime, ProgramResume{Args: c.args})
			report, cleared := strings.CutPrefix(got, clearSeq)
			if cleared != c.clearFirst {
				t.Errorf("cleared first = %v, want %v", cleared, c.clearFirst)
			}
			if !strings.HasPrefix(report, "\x1b]7577;pid=1:time=1795543200:argv=bnZpbQ==,") {
				t.Errorf("missing report in %q", got)
			}
		})
	}
}

func TestClearProgramResume(t *testing.T) {
	want := "\x1b]7577;pid=74877:time=1795543200:clear=1:mac=c30603369f98fef581a4098be889175b0c74f0641e49f8b25f4326cb38f97c93\x07"
	if got := ClearProgramResume(programResumeToken, 74877, programResumeTime); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := ClearProgramResume("", 74877, programResumeTime); got != "" {
		t.Errorf("empty token: got %q, want empty", got)
	}
}
