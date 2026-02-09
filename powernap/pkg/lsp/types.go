package lsp

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/charmbracelet/x/powernap/pkg/transport"
)

// OffsetEncoding represents the character encoding used for text document offsets.
type OffsetEncoding int

const (
	// UTF8 encoding - bytes.
	UTF8 OffsetEncoding = iota
	// UTF16 encoding - default for LSP.
	UTF16
	// UTF32 encoding - codepoints.
	UTF32
)

// Client represents an LSP client connection to a language server.
type Client struct {
	ID               string
	Name             string
	conn             *transport.Connection
	ctx              context.Context
	cancel           context.CancelFunc
	initialized      atomic.Bool
	shutdown         atomic.Bool
	capabilities     protocol.ServerCapabilities
	offsetEncoding   OffsetEncoding
	rootURI          string
	workspaceFolders []protocol.WorkspaceFolder
	config           map[string]any
	initOptions      map[string]any
}

// ClientConfig represents the configuration for creating a new LSP client.
type ClientConfig struct {
	Command          string
	Args             []string
	RootURI          string
	WorkspaceFolders []protocol.WorkspaceFolder
	InitOptions      map[string]any
	Settings         map[string]any
	Environment      map[string]string
	Timeout          time.Duration
}
