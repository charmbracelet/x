package xpty

import (
	"os/exec"

	"github.com/charmbracelet/x/exp/term/conpty"
)

// ConPty is a Windows console pty.
type ConPty struct {
	*conpty.ConPty
}

var _ Pty = &ConPty{}

// NewConPty creates a new ConPty.
func NewConPty(width, height int, opts ...PtyOption) (*ConPty, error) {
	var opt Options
	for _, o := range opts {
		o(opt)
	}

	c, err := conpty.New(width, height, opt.Flags)
	if err != nil {
		return nil, err
	}

	return &ConPty{c}, nil
}

// Name returns the name of the ConPty.
func (c *ConPty) Name() string {
	return "windows-pty"
}

// Start starts a command on the ConPty.
// This is a wrapper around conpty.Spawn.
func (c *ConPty) Start(cmd *exec.Cmd) error {
	return c.start(cmd)
}
