package bt

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vttest"
	"github.com/charmbracelet/x/vttest/snapshot"
)

type testModel struct {
	t     testing.TB
	count int
}

type testTickMsg struct{}

func testTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return testTickMsg{}
	})
}

func (m testModel) Init() tea.Cmd {
	return testTickCmd()
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case testTickMsg:
		m.t.Logf("tick %d", m.count+1)
		m.count++
		if m.count >= 5 {
			return m, tea.Quit
		}

		return m, testTickCmd()

	case tea.WindowSizeMsg:
		m.t.Logf("window size changed: %dx%d", msg.Width, msg.Height)
		return m, tea.Printf("Window size changed: %dx%d", msg.Width, msg.Height)

	case tea.KeyPressMsg:
		m.t.Logf("key pressed: %q", msg.String())
		return m, tea.Printf("You pressed \"%s\"", msg.String())
	}

	return m, nil
}

func (m testModel) View() tea.View {
	var v tea.View
	str := "\x1b[3;31mPress any key to see it echoed back. \x1b[1mWaiting for 5 ticks...\x1b[m\n\n"
	str += "\n"
	str += fmt.Sprintf("\x1b[4mTick count: \x1b[1;%dm%d\x1b[m\n", ansi.White+30, m.count)
	v.SetContent(str)
	v.BackgroundColor = color.Black
	v.WindowTitle = "TestModel"
	return v
}

func TestBubbleTeaInlineProgram(t *testing.T) {
	if helper := os.Getenv("GO_TEST_HELPER_PROCESS"); helper == "1" {
		ttIn, ttOut := os.Stdin, os.Stdout
		m := testModel{t: t}
		p := tea.NewProgram(m,
			tea.WithInput(ttIn),
			tea.WithOutput(ttOut),
		)

		if _, err := p.Run(); err != nil {
			t.Fatalf("failed to run program: %v", err)
		}

		return
	}

	tt, err := vttest.NewTerminal(t, 80, 24)
	if err != nil {
		t.Fatalf("failed to create terminal: %v", err)
	}

	defer tt.Close()

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=^%s$", t.Name()))
	cmd.Env = append(os.Environ(), "GO_TEST_HELPER_PROCESS=1")
	if err := tt.Start(cmd); err != nil {
		t.Fatalf("failed to start command: %v", err)
	}

	var counter atomic.Int32

	go func() {
		time.Sleep(time.Second)
		takeSnapshot(t, tt, int(counter.Add(1)))
		tt.SendText("a")
		time.Sleep(time.Second)
		takeSnapshot(t, tt, int(counter.Add(1)))
		tt.SendText("b")
		time.Sleep(time.Second)
		takeSnapshot(t, tt, int(counter.Add(1)))
		tt.SendText("c")
		time.Sleep(time.Second)
		takeSnapshot(t, tt, int(counter.Add(1)))
	}()

	if err := tt.Wait(cmd); err != nil {
		t.Fatalf("command failed: %v", err)
	}
}

func takeSnapshot(t testing.TB, term *vttest.Terminal, num int) {
	t.Helper()
	snapshot.TestdataEqualf(t, fmt.Sprintf("%d", num), term, "snapshot %d does not match expected", num)
}
