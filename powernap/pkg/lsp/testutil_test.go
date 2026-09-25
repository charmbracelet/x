package lsp

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"

	"github.com/charmbracelet/x/powernap/pkg/transport"
	"github.com/sourcegraph/jsonrpc2"
)

// invokeHandler invokes a registered server->client handler on conn by
// reaching into the transport's router via reflection. Test-only.
func invokeHandler(t *testing.T, conn *transport.Connection, method string, params json.RawMessage, out any) error {
	t.Helper()
	// Extract *transport.Router pointer value from unexported field
	v := reflect.ValueOf(conn).Elem().FieldByName("router")
	pp := (**transport.Router)(unsafe.Pointer(v.UnsafeAddr()))
	r := *pp
	// Build a JSON-RPC request
	id := jsonrpc2.ID{Num: 1}
	req := &jsonrpc2.Request{Method: method, ID: id, Params: &params}
	res, err := r.Route(context.Background(), req)
	if err != nil {
		return err
	}
	if out != nil && res != nil {
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, out)
	}
	return nil
}
