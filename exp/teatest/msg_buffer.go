package teatest

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// msgBuffer stores messages for checking in WaitForMsg.
type msgBuffer struct {
	msgs []tea.Msg
	mu   sync.Mutex
}

func (b *msgBuffer) append(msg tea.Msg) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.msgs = append(b.msgs, msg)
}

// forEach executes the given function for each message while holding the lock.
func (b *msgBuffer) forEach(fn func(msg tea.Msg) bool) tea.Msg {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, msg := range b.msgs {
		if fn(msg) {
			return msg
		}
	}
	return nil
}
