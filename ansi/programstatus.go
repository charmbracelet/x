package ansi

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ProgramState is the state of a program reported with the Program Status
// Protocol (OSC 7501).
//
// See: https://www.superlogical.com/rex/docs/build/program-status
type ProgramState string

// Program states.
const (
	// ProgramStateIdle means the program is at rest, waiting for the user's
	// next instruction.
	ProgramStateIdle ProgramState = "idle"
	// ProgramStateWorking means the program is running.
	ProgramStateWorking ProgramState = "working"
	// ProgramStateDone means the program finished a piece of work and the
	// result is ready to look at.
	ProgramStateDone ProgramState = "done"
	// ProgramStateBlocked means the program cannot continue until the user
	// does something.
	ProgramStateBlocked ProgramState = "blocked"
	// ProgramStateError means the program failed and stopped.
	ProgramStateError ProgramState = "error"
	// ProgramStateClear removes the addressed record and every record beneath
	// it. With no id, it removes every record on the terminal. A clear report
	// carries only its id; other fields are ignored.
	ProgramStateClear ProgramState = "clear"
)

// ProgramStatusKind says what a blocked program is waiting for.
type ProgramStatusKind string

// Program status kinds.
const (
	// ProgramStatusKindPermission means the program waits for approval to do
	// something.
	ProgramStatusKindPermission ProgramStatusKind = "permission"
	// ProgramStatusKindQuestion means the user must type an answer.
	ProgramStatusKindQuestion ProgramStatusKind = "question"
	// ProgramStatusKindAuth means the program waits for a login, token, or
	// credential.
	ProgramStatusKindAuth ProgramStatusKind = "auth"
)

// ProgramProgress is the progress of a program status. The zero value is
// indeterminate: the program is busy but has no percentage to report. Use
// [Percent] for a determinate value.
type ProgramProgress struct {
	percent int8
	set     bool
}

// Percent returns a determinate progress. The percentage is clamped to the
// range 0 to 100.
func Percent(p int) ProgramProgress {
	return ProgramProgress{percent: int8(min(max(p, 0), 100)), set: true}
}

// Value returns the percentage and whether the progress is determinate.
func (p ProgramProgress) Value() (int, bool) {
	return int(p.percent), p.set
}

// ProgramStatus is a report of the Program Status Protocol (OSC 7501).
//
// Each report replaces its record completely, so fields such as App and Title
// should be included in every report. ProgramStatus is comparable, so two
// reports can be compared with ==.
//
// See: https://www.superlogical.com/rex/docs/build/program-status
type ProgramStatus struct {
	// State is the program state. Required.
	State ProgramState
	// ID addresses a record. Empty means the root record. It is a "/"
	// separated path of segments matching [A-Za-z0-9_.+-]{1,32}.
	ID string
	// App is a stable machine-readable program name matching
	// [A-Za-z0-9_.+-]{1,32}. Invalid values are omitted.
	App string
	// Kind says what a blocked program waits for. Only used with
	// [ProgramStateBlocked].
	Kind ProgramStatusKind
	// Progress is only used with [ProgramStateWorking] and
	// [ProgramStateBlocked]. The zero value is indeterminate.
	Progress ProgramProgress
	// Title is a short human-readable label for the record.
	Title string
	// Message is one human-readable line describing the record.
	Message string
}

// Errors reported by [ProgramStatus.Validate].
var (
	// ErrProgramStatusState means the program state is unknown.
	ErrProgramStatusState = errors.New("ansi: unknown program status state")
	// ErrProgramStatusID means the record id does not match the protocol
	// grammar.
	ErrProgramStatusID = errors.New("ansi: invalid program status id")
)

// Validate reports why a terminal would discard the whole report, or nil if
// the report is valid. Problems the protocol treats as absent, such as an
// invalid App or a Kind used with the wrong state, are not errors; the
// encoder omits those fields.
func (s ProgramStatus) Validate() error {
	if !validProgramState(s.State) {
		return ErrProgramStatusState
	}
	if s.ID != "" && !validProgramStatusID(s.ID) {
		return ErrProgramStatusID
	}
	return nil
}

// Program Status Protocol limits.
const (
	programStatusMaxSequence   = 4096
	programStatusMaxKey        = 16
	programStatusMaxApp        = 32
	programStatusMaxTitle      = 192
	programStatusMaxTitleEnc   = 256
	programStatusMaxMessage    = 2048
	programStatusMaxMessageEnc = 2732
)

// ClearProgramStatus is a sequence that removes every program status record
// on the terminal.
//
//	OSC 7501 ; state=clear BEL
//
// See: https://www.superlogical.com/rex/docs/build/program-status
const ClearProgramStatus = "\x1b]7501;state=clear\x07"

// RequestProgramStatusSupport is a sequence that asks the terminal whether it
// supports the Program Status Protocol. A supporting terminal replies with
// OSC 7501 ; ? ST.
//
//	OSC 7501 ; ? BEL
//
// See: https://www.superlogical.com/rex/docs/build/program-status
const RequestProgramStatusSupport = "\x1b]7501;?\x07"

// ClearProgramStatusID returns a sequence that removes the record with the
// given id and every record beneath it. An empty id removes every record. It
// returns an empty string if the id is invalid.
//
//	OSC 7501 ; state=clear:id=Id BEL
//
// See: https://www.superlogical.com/rex/docs/build/program-status
func ClearProgramStatusID(id string) string {
	return SetProgramStatus(ProgramStatus{State: ProgramStateClear, ID: id})
}

// SetProgramStatus returns a sequence that reports a program status using the
// Program Status Protocol (OSC 7501).
//
//	OSC 7501 ; key=value:key=value BEL
//
// Terminals discard a whole report when any part of it is invalid, so the
// report is sanitized: each control character in Title and Message is
// replaced with a space and the text is truncated to the protocol limits, and
// an invalid App, or a Kind or Progress used with the wrong state, is omitted.
// It returns an empty string if [ProgramStatus.Validate] fails.
//
// See: https://www.superlogical.com/rex/docs/build/program-status
func SetProgramStatus(s ProgramStatus) string {
	if s.Validate() != nil {
		return ""
	}

	pairs := []string{"state=" + string(s.State)}
	if s.ID != "" {
		pairs = append(pairs, "id="+s.ID)
	}
	if s.State == ProgramStateClear {
		return programStatusSequence(pairs)
	}
	if s.State == ProgramStateBlocked && validProgramStatusKind(s.Kind) {
		pairs = append(pairs, "kind="+string(s.Kind))
	}
	if p, ok := s.Progress.Value(); ok && programStateHasProgress(s.State) {
		pairs = append(pairs, "progress="+strconv.Itoa(p))
	}
	if validProgramStatusSegment(s.App) {
		pairs = append(pairs, "app="+s.App)
	}
	if t := programStatusText(s.Title, programStatusMaxTitle); t != "" {
		pairs = append(pairs, "title="+t)
	}
	if m := programStatusText(s.Message, programStatusMaxMessage); m != "" {
		pairs = append(pairs, "msg="+m)
	}

	return programStatusSequence(pairs)
}

// ParseProgramStatus parses the body of a Program Status Protocol (OSC 7501)
// report, the part between "OSC 7501 ;" and ST. It reports false when a
// terminal must discard the whole report, including for the feature detection
// body "?".
//
// Following the protocol, malformed pairs and unknown keys are skipped, the
// last of a repeated key wins, and an invalid App, Kind, or Progress is
// treated as absent. Fields that do not apply to the state are dropped, so
// the result encodes back to an equivalent report.
//
// See: https://www.superlogical.com/rex/docs/build/program-status
func ParseProgramStatus(body string) (ProgramStatus, bool) {
	// The terminator is unknown here, so assume the longer one, ESC \.
	if len("\x1b]7501;")+len(body)+len("\x1b\\") > programStatusMaxSequence {
		return ProgramStatus{}, false
	}

	vals := map[string]string{}
	for pair := range strings.SplitSeq(body, ":") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if len(k) > programStatusMaxKey {
			return ProgramStatus{}, false
		}
		if !validProgramStatusKey(k) || !validProgramStatusValue(v) {
			continue
		}
		vals[k] = v
	}

	s := ProgramStatus{State: ProgramState(vals["state"])}
	if !validProgramState(s.State) {
		return ProgramStatus{}, false
	}
	if id, ok := vals["id"]; ok {
		if !validProgramStatusID(id) {
			return ProgramStatus{}, false
		}
		s.ID = id
	}

	// Every limit is checked before anything is applied, even for clear.
	app := vals["app"]
	if len(app) > programStatusMaxApp {
		return ProgramStatus{}, false
	}
	title, ok := parseProgramStatusText(vals["title"], programStatusMaxTitleEnc, programStatusMaxTitle)
	if !ok {
		return ProgramStatus{}, false
	}
	msg, ok := parseProgramStatusText(vals["msg"], programStatusMaxMessageEnc, programStatusMaxMessage)
	if !ok {
		return ProgramStatus{}, false
	}
	if s.State == ProgramStateClear {
		return s, true
	}

	if validProgramStatusSegment(app) {
		s.App = app
	}
	s.Title, s.Message = title, msg
	if k := ProgramStatusKind(vals["kind"]); s.State == ProgramStateBlocked && validProgramStatusKind(k) {
		s.Kind = k
	}
	if p, ok := parseProgramStatusProgress(vals["progress"]); ok && programStateHasProgress(s.State) {
		s.Progress = Percent(p)
	}

	return s, true
}

func programStatusSequence(pairs []string) string {
	return "\x1b]7501;" + strings.Join(pairs, ":") + "\x07"
}

func validProgramState(s ProgramState) bool {
	switch s {
	case ProgramStateIdle, ProgramStateWorking, ProgramStateDone,
		ProgramStateBlocked, ProgramStateError, ProgramStateClear:
		return true
	}
	return false
}

func programStateHasProgress(s ProgramState) bool {
	return s == ProgramStateWorking || s == ProgramStateBlocked
}

func validProgramStatusKind(k ProgramStatusKind) bool {
	switch k {
	case ProgramStatusKindPermission, ProgramStatusKindQuestion, ProgramStatusKindAuth:
		return true
	}
	return false
}

// programStatusText sanitizes, truncates, and base64 encodes free text.
func programStatusText(s string, limit int) string {
	s = strings.Map(func(r rune) rune {
		if isProgramStatusControl(r) {
			return ' '
		}
		return r
	}, strings.ToValidUTF8(s, "\uFFFD"))
	if len(s) > limit {
		s = s[:limit]
		for !utf8.ValidString(s) {
			s = s[:len(s)-1]
		}
	}
	if s == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// parseProgramStatusText decodes free text, reporting false when the report
// must be discarded. The encoded size is checked before decoding.
func parseProgramStatusText(v string, encLimit, limit int) (string, bool) {
	if len(v) > encLimit {
		return "", false
	}
	b, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(v, "="))
	if err != nil || len(b) > limit || !utf8.Valid(b) {
		return "", false
	}
	s := string(b)
	if strings.ContainsFunc(s, isProgramStatusControl) {
		return "", false
	}
	return s, true
}

func parseProgramStatusProgress(v string) (int, bool) {
	if v == "" || len(v) > 3 || strings.Trim(v, "0123456789") != "" {
		return 0, false
	}
	p, _ := strconv.Atoi(v)
	return p, p <= 100
}

func isProgramStatusControl(r rune) bool {
	return r < 0x20 || (r >= 0x7f && r <= 0x9f)
}

func validProgramStatusID(id string) bool {
	if len(id) > 128 {
		return false
	}
	segs := strings.Split(id, "/")
	if len(segs) > 8 {
		return false
	}
	for _, seg := range segs {
		if !validProgramStatusSegment(seg) {
			return false
		}
	}
	return true
}

func validProgramStatusSegment(s string) bool {
	if len(s) < 1 || len(s) > 32 {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') &&
			c != '_' && c != '.' && c != '+' && c != '-' {
			return false
		}
	}
	return true
}

func validProgramStatusKey(k string) bool {
	if k == "" {
		return false
	}
	for i := range len(k) {
		if k[i] < 'a' || k[i] > 'z' {
			return false
		}
	}
	return true
}

func validProgramStatusValue(v string) bool {
	for i := range len(v) {
		c := v[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') &&
			c != '_' && c != '.' && c != ',' && c != '+' && c != '/' && c != '=' && c != '-' {
			return false
		}
	}
	return true
}
