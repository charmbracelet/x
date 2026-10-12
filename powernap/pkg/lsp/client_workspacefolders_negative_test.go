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

// This test simulates the previous bug: if no handler is registered,
// the server's request should fail. This ensures our positive test is
// meaningful and would have failed before the fix.
func TestClient_WorkspaceFoldersHandler_MissingHandlerReturnsError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	clientSide, _ := net.Pipe()

	conn, err := transport.NewConnection(ctx, clientSide, nil)
	if err != nil {
		t.Fatalf("new connection: %v", err)
	}

	// Intentionally do NOT register the handler on conn to simulate the bug.
	_ = conn

	var got []protocol.WorkspaceFolder
	err = invokeHandler(t, conn, MethodWorkspaceWorkspaceFolders, json.RawMessage("null"), &got)
	if err == nil {
		t.Fatalf("expected error due to missing handler, got: %+v", got)
	}
}
