package transport

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sourcegraph/jsonrpc2"
)

// Ensure Router.Route does not panic when req.Params is nil for requests.
func TestRouter_RequestWithNilParams_DoesNotPanic(t *testing.T) {
	r := NewRouter()
	// Register a no-op handler; it should be invoked even if Params is nil
	r.Handle("test/method", func(ctx context.Context, method string, params json.RawMessage) (any, error) {
		return nil, nil
	})

	// Build a request with nil Params
	id := jsonrpc2.ID{Num: 1}
	req := &jsonrpc2.Request{Method: "test/method", ID: id, Params: nil}

	// If the implementation blindly dereferences *req.Params, this will panic.
	// The correct behavior is to pass an empty RawMessage to the handler.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Route panicked on nil Params: %v", r)
		}
	}()
	if _, err := r.Route(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Ensure Router.Route does not panic when req.Params is nil for notifications.
func TestRouter_NotificationWithNilParams_DoesNotPanic(t *testing.T) {
	r := NewRouter()
	called := false
	r.HandleNotification("test/notify", func(ctx context.Context, method string, params json.RawMessage) {
		called = true
	})

	// Notification has zero-value ID and nil Params
	req := &jsonrpc2.Request{Method: "test/notify", Params: nil}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Route panicked on nil Params (notification): %v", r)
		}
	}()
	if _, err := r.Route(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("notification handler was not called")
	}
}
