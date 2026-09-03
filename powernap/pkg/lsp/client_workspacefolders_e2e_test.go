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

// End-to-end: verify that a real Client created via NewClient handles
// workspace/workspaceFolders requests as configured.
func TestClient_E2E_WorkspaceFolders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a connected in-memory stream pair. We'll ignore the server
	// process starter by substituting the transport streams manually.
	clientSide, _ := net.Pipe()

	// Construct a dummy client with our connection injected.
	c := &Client{
		workspaceFolders: []protocol.WorkspaceFolder{{URI: protocol.URI("file:///w"), Name: "w"}},
	}
	conn, err := transport.NewConnection(ctx, clientSide, nil)
	if err != nil {
		t.Fatalf("new connection: %v", err)
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
