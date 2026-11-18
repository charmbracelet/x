package lsp

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/charmbracelet/x/powernap/pkg/transport"
)

// This test verifies that the client registers a handler for
// "workspace/workspaceFolders" via setupHandlers and returns the
// configured workspace folders. It will fail if setupHandlers does
// not register the handler (regression of prior bug).
func TestClient_WorkspaceFoldersHandler_ReturnsConfiguredFolders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	clientSide, _ := net.Pipe()

	conn, err := transport.NewConnection(ctx, clientSide, nil)
	if err != nil {
		t.Fatalf("new connection: %v", err)
	}

	// Build a Client with workspace folders and attach the connection,
	// then invoke setupHandlers to register all handlers under test.
	c := &Client{
		workspaceFolders: []protocol.WorkspaceFolder{{URI: protocol.URI("file:///w"), Name: "w"}},
	}
	c.conn = conn
	c.setupHandlers()

	var got []protocol.WorkspaceFolder
	if err := invokeHandler(t, conn, MethodWorkspaceWorkspaceFolders, json.RawMessage("null"), &got); err != nil {
		t.Fatalf("call error: %v", err)
	}
	if len(got) != 1 || string(got[0].URI) != "file:///w" || got[0].Name != "w" {
		t.Fatalf("unexpected folders: %+v", got)
	}
}
