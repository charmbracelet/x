package teatest

import (
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func TestWaitForErrorReader(t *testing.T) {
	err := doWaitFor(iotest.ErrReader(fmt.Errorf("fake")), func(bts []byte) bool {
		return true
	}, WithDuration(time.Millisecond), WithCheckInterval(10*time.Microsecond))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err.Error() != "WaitFor: fake" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestWaitForTimeout(t *testing.T) {
	err := doWaitFor(strings.NewReader("nope"), func(bts []byte) bool {
		return false
	}, WithDuration(time.Millisecond), WithCheckInterval(10*time.Microsecond))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err.Error() != "WaitFor: condition not met after 1ms. Last output:\nnope" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

type m string

func (m m) Init() (tea.Model, tea.Cmd)          { return m, nil }
func (m m) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m m) View() string                        { return string(m) }

func TestWaitFinishedWithTimeoutFn(t *testing.T) {
	tm := NewTestModel(t, m("a"))
	var timedOut bool
	tm.WaitFinished(t, WithFinalTimeout(time.Nanosecond), WithTimeoutFn(func(testing.TB) {
		timedOut = true
	}))
	if !timedOut {
		t.Fatal("expected timedOut to be set")
	}
}
