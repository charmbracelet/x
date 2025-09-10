package transport

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/jsonrpc2"
)

// Connection represents a managed connection to a language server.
type Connection struct {
	conn      jsonrpc2.JSONRPC2
	transport *Transport
	router    *Router
	logger    *slog.Logger

	// State management
	closed   atomic.Bool
	closeMu  sync.Mutex
	closeErr error

	// Request tracking
	requestMu sync.Mutex
	requests  map[jsonrpc2.ID]chan *Message
	nextID    int64
}

// NewConnection creates a new managed connection.
func NewConnection(ctx context.Context, stream io.ReadWriteCloser, logger *slog.Logger) (*Connection, error) {
	c := &Connection{
		router:   NewRouter(),
		logger:   logger,
		requests: make(map[jsonrpc2.ID]chan *Message),
	}

	// Create JSON-RPC connection
	conn := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(c.handleRequest),
	)

	c.conn = conn
	c.transport = NewWithConn(conn)

	return c, nil
}

// Call makes a request to the language server and waits for a response.
func (c *Connection) Call(ctx context.Context, method string, params any, result any) error {
	if c.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	return c.conn.Call(ctx, method, params, result)
}

// Notify sends a notification to the language server.
func (c *Connection) Notify(ctx context.Context, method string, params any) error {
	if c.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	return c.conn.Notify(ctx, method, params)
}

// handleRequest handles incoming requests from the language server.
func (c *Connection) handleRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if c.logger != nil {
		c.logger.Debug("Handling request", "method", req.Method)
	}

	return c.router.Route(ctx, req)
}

// RegisterHandler registers a handler for a specific method.
func (c *Connection) RegisterHandler(method string, handler Handler) {
	c.router.Handle(method, handler)
}

// RegisterNotificationHandler registers a notification handler.
func (c *Connection) RegisterNotificationHandler(method string, handler NotificationHandler) {
	c.router.HandleNotification(method, handler)
}

// Close closes the connection.
func (c *Connection) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed.Load() {
		return c.closeErr
	}

	c.closed.Store(true)

	// Close the JSON-RPC connection
	if c.conn != nil {
		c.closeErr = c.conn.Close()
	}

	// Close any pending requests
	c.requestMu.Lock()
	for _, ch := range c.requests {
		close(ch)
	}
	c.requests = nil
	c.requestMu.Unlock()

	return c.closeErr
}

// IsConnected returns true if the connection is still active.
func (c *Connection) IsConnected() bool {
	return !c.closed.Load()
}

// generateID generates a unique request ID.
func (c *Connection) generateID() jsonrpc2.ID {
	id := atomic.AddInt64(&c.nextID, 1)
	return jsonrpc2.ID{
		Num:      uint64(id),
		IsString: false,
	}
}
