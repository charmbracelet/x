package teatest

import (
	tea "github.com/charmbracelet/bubbletea"
)

// msgCaptureModel wraps a model to capture messages.
type msgCaptureModel struct {
	model  tea.Model
	buffer *msgBuffer
}

func (m msgCaptureModel) Init() tea.Cmd {
	return m.model.Init()
}

func (m msgCaptureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.buffer.append(msg)
	model, cmd := m.model.Update(msg)
	if wrappedModel, ok := model.(msgCaptureModel); ok {
		return wrappedModel, cmd
	}

	return msgCaptureModel{
		model:  model,
		buffer: m.buffer,
	}, cmd
}

func (m msgCaptureModel) View() string {
	return m.model.View()
}
