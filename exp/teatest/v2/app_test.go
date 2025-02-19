package teatest_test

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

func TestApp(t *testing.T) {
	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
	)
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Type("I'm typing things, but it'll be ignored by my program")
	tm.Send("ignored msg")
	tm.Send(tea.KeyPressMsg{
		Code: tea.KeyEnter,
	})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}

	out := teatest.TrimEmptyLines(tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	if !regexp.MustCompile(`This program will exit in \d+ seconds`).MatchString(out) {
		t.Fatalf("output does not match the given regular expression: %q", out)
	}
	teatest.RequireEqualOutput(t, out)

	if tm.FinalModel(t).(model) != 9 {
		t.Errorf("expected model to be 10, was %d", m)
	}
}

func TestAppAltScreen(t *testing.T) {
	t.Skip("needs changes in /vt")
	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
		teatest.WithProgramOptions(tea.WithAltScreen()),
	)
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Type("I'm typing things, but it'll be ignored by my program")
	tm.Send("ignored msg")
	tm.Send(tea.KeyPressMsg{
		Code: tea.KeyEnter,
	})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}

	out := teatest.TrimEmptyLines(tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	if !regexp.MustCompile(`This program will exit in \d+ seconds`).MatchString(out) {
		t.Fatalf("output does not match the given regular expression: %q", out)
	}
	teatest.RequireEqualOutput(t, out)

	if tm.FinalModel(t).(model) != 9 {
		t.Errorf("expected model to be 10, was %d", m)
	}
}

func TestAppInteractive(t *testing.T) {
	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
	)

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Send("ignored msg")

	if s := tm.Output(); !strings.Contains(s, "This program will exit in 9 seconds") {
		t.Fatalf("output does not match: expected %q", string(s))
	}

	teatest.WaitForOutput(t, tm, func(s string) bool {
		return strings.Contains(s, "This program will exit in 7 seconds")
	}, teatest.WithDuration(5*time.Second), teatest.WithCheckInterval(time.Millisecond*10))

	tm.Send(tea.KeyPressMsg{
		Code: tea.KeyEnter,
	})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}

	if tm.FinalModel(t).(model) != 7 {
		t.Errorf("expected model to be 7, was %d", m)
	}
}

func readBts(tb testing.TB, r io.Reader) []byte {
	tb.Helper()
	bts, err := io.ReadAll(r)
	if err != nil {
		tb.Fatal(err)
	}
	return bts
}

// A model can be more or less any type of data. It holds all the data for a
// program, so often it's a struct. For this simple example, however, all
// we'll need is a simple integer.
type model int

// Init optionally returns an initial command we should run. In this case we
// want to start the timer.
func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tick
}

// Update is called when messages are received. The idea is that you inspect the
// message and send back an updated model accordingly. You can also return
// a command, which is a function that performs I/O and returns a message.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m--
		if m <= 0 {
			return m, tea.Quit
		}
		return m, tick
	}
	return m, nil
}

// View returns a string based on data in the model. That string which will be
// rendered to the terminal.
func (m model) View() string {
	return fmt.Sprintf("Hi. This program will exit in %d seconds. To quit sooner press any key.\n", m)
}

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
