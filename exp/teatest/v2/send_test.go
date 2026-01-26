package teatest_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

func TestAppSendToOtherProgram(t *testing.T) {
	m1 := &connectedModel{
		name: "m1",
	}
	m2 := &connectedModel{
		name: "m2",
	}

	tm1 := teatest.NewTestModel(t, m1, teatest.WithInitialTermSize(70, 30))
	t.Cleanup(func() {
		if err := tm1.Quit(); err != nil {
			t.Fatal(err)
		}
	})
	tm2 := teatest.NewTestModel(t, m2, teatest.WithInitialTermSize(70, 30))
	t.Cleanup(func() {
		if err := tm2.Quit(); err != nil {
			t.Fatal(err)
		}
	})
	m1.programs = append(m1.programs, tm2)
	m2.programs = append(m2.programs, tm1)

	tm1.Type("pp")
	tm2.Type("pppp")

	tm1.Type("q")
	tm2.Type("q")

	out1 := readBts(t, tm1.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	out2 := readBts(t, tm2.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))

	if string(out1) != string(out2) {
		t.Errorf("output of both models should be the same, got:\n%v\nand:\n%v\n", string(out1), string(out2))
	}

	teatest.RequireEqualOutput(t, out1)
}

type connectedModel struct {
	name     string
	programs []interface{ Send(tea.Msg) }
	msgs     []string
}

type ping string

func (m *connectedModel) Init() tea.Cmd {
	return nil
}

func (m *connectedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "p":
			send := ping("from " + m.name)
			m.msgs = append(m.msgs, string(send))
			for _, p := range m.programs {
				p.Send(send)
			}
			fmt.Printf("sent ping %q to others\n", send)
		case "q":
			return m, tea.Quit
		}
	case ping:
		fmt.Printf("rcvd ping %q on %s\n", msg, m.name)
		m.msgs = append(m.msgs, string(msg))
	}
	return m, nil
}

func (m *connectedModel) View() tea.View {
	return tea.NewView("All pings:\n" + strings.Join(m.msgs, "\n"))
}
