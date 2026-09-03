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

// This test asserts setupHandlers registers workspace/workspaceFolders and returns configured folders.
func TestSetupHandlers_RegistersWorkspaceFolders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	clientSide, _ := net.Pipe()

	conn, err := transport.NewConnection(ctx, clientSide, nil)
	if err != nil {
		t.Fatalf("new connection: %v", err)
	}

	c := &Client{workspaceFolders: []protocol.WorkspaceFolder{{URI: protocol.URI("file:///w"), Name: "w"}}}
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
