package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/sourcegraph/jsonrpc2"
)

// Transport handles the low-level communication with the language server.
type Transport struct {
	conn   jsonrpc2.JSONRPC2
	reader io.Reader
	writer io.Writer
	logger *log.Logger
	mu     sync.Mutex
}

// Message represents a JSON-RPC message.
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *jsonrpc2.ID    `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpc2.Error `json:"error,omitempty"`
}

// New creates a new transport.
func New(reader io.Reader, writer io.Writer, logger *log.Logger) *Transport {
	return &Transport{
		reader: reader,
		writer: writer,
		logger: logger,
	}
}

// NewWithConn creates a new transport with an existing JSON-RPC connection.
func NewWithConn(conn jsonrpc2.JSONRPC2) *Transport {
	return &Transport{
		conn: conn,
	}
}

// Send sends a message to the language server.
func (t *Transport) Send(ctx context.Context, msg *Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn != nil {
		// Use existing connection
		if msg.ID != nil {
			// It's a request
			var result json.RawMessage
			err := t.conn.Call(ctx, msg.Method, msg.Params, &result)
			if err != nil {
				return err
			}
			msg.Result = result
		} else {
			// It's a notification
			return t.conn.Notify(ctx, msg.Method, msg.Params)
		}
		return nil
	}

	// Manual implementation for raw reader/writer
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write Content-Length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := t.writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write message body
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("Sent message", "method", msg.Method, "id", msg.ID)
	}

	return nil
}

// Receive receives a message from the language server.
func (t *Transport) Receive(ctx context.Context) (*Message, error) {
	if t.conn != nil {
		// This is handled by the connection's handler
		return nil, fmt.Errorf("receive not supported with existing connection")
	}

	// Read headers
	headers := make(map[string]string)
	scanner := bufio.NewScanner(t.reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Get content length
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %w", err)
	}

	// Read body
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(t.reader, body); err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Parse message
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("Received message", "method", msg.Method, "id", msg.ID)
	}

	return &msg, nil
}

// Close closes the transport.
func (t *Transport) Close() error {
	if t.conn != nil {
		// Connection will be closed by the client
		return nil
	}

	if closer, ok := t.writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	if closer, ok := t.reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	return nil
}

// StreamTransport provides a bidirectional stream for JSON-RPC communication.
type StreamTransport struct {
	reader io.Reader
	writer io.Writer
	closer io.Closer
}

// NewStreamTransport creates a new stream transport.
func NewStreamTransport(reader io.Reader, writer io.Writer, closer io.Closer) *StreamTransport {
	return &StreamTransport{
		reader: reader,
		writer: writer,
		closer: closer,
	}
}

// Read implements io.Reader.
func (s *StreamTransport) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

// Write implements io.Writer.
func (s *StreamTransport) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

// Close implements io.Closer.
func (s *StreamTransport) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

// ObjectStream creates a jsonrpc2.ObjectStream from the transport.
func (s *StreamTransport) ObjectStream() jsonrpc2.ObjectStream {
	return jsonrpc2.NewBufferedStream(s, jsonrpc2.VSCodeObjectCodec{})
}