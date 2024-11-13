package teatest

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

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
