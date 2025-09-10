// Package lsp provides a client implementation for the Language Server
// Protocol (LSP).
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/charmbracelet/x/powernap/pkg/transport"
)

// LSP method constants
const (
	MethodInitialize                         = "initialize"
	MethodInitialized                        = "initialized"
	MethodShutdown                           = "shutdown"
	MethodExit                               = "exit"
	MethodTextDocumentDidOpen                = "textDocument/didOpen"
	MethodTextDocumentDidChange              = "textDocument/didChange"
	MethodTextDocumentDidSave                = "textDocument/didSave"
	MethodTextDocumentDidClose               = "textDocument/didClose"
	MethodTextDocumentCompletion             = "textDocument/completion"
	MethodTextDocumentHover                  = "textDocument/hover"
	MethodTextDocumentDefinition             = "textDocument/definition"
	MethodTextDocumentReferences             = "textDocument/references"
	MethodTextDocumentDiagnostic             = "textDocument/publishDiagnostics"
	MethodWorkspaceConfiguration             = "workspace/configuration"
	MethodWorkspaceDidChangeConfiguration    = "workspace/didChangeConfiguration"
	MethodWorkspaceDidChangeWorkspaceFolders = "workspace/didChangeWorkspaceFolders"
	MethodWorkspaceDidChangeWatchedFiles     = "workspace/didChangeWatchedFiles"
)

// NewClient creates a new LSP client with the given configuration.
func NewClient(config ClientConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ID:               config.Command, // Will be updated after initialization
		Name:             config.Command,
		ctx:              ctx,
		cancel:           cancel,
		rootURI:          config.RootURI,
		workspaceFolders: config.WorkspaceFolders,
		config:           config.Settings,
		initOptions:      config.InitOptions,
		offsetEncoding:   UTF16, // Default to UTF16
	}

	// Start the language server process
	stream, err := startServerProcess(ctx, config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start language server: %w", err)
	}

	// Create transport connection
	conn, err := transport.NewConnection(ctx, stream, log.Default())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	client.conn = conn

	// Register handlers for server-initiated requests
	client.setupHandlers()

	return client, nil
}

// Initialize sends the initialize request to the language server.
func (c *Client) Initialize(ctx context.Context, enableSnippets bool) error {
	if c.initialized {
		return fmt.Errorf("client already initialized")
	}

	// Extract root path from URI
	rootPath := ""
	if c.rootURI != "" {
		rootPath = strings.TrimPrefix(c.rootURI, "file://")
	}

	// Prepare workspace folders - some servers don't like nil
	workspaceFolders := c.workspaceFolders
	if workspaceFolders == nil {
		workspaceFolders = []protocol.WorkspaceFolder{}
	}

	initParams := map[string]any{
		"processId": os.Getpid(),
		"clientInfo": map[string]any{
			"name":    "powernap",
			"version": "0.1.0",
		},
		"locale":                "en-us",
		"rootPath":              rootPath, // Deprecated but some servers still use it
		"rootUri":               c.rootURI,
		"capabilities":          c.makeClientCapabilities(enableSnippets),
		"workspaceFolders":      workspaceFolders,
		"initializationOptions": c.initOptions, // Use the client's init options
		"trace":                 "off",         // Can be "off", "messages", or "verbose"
	}

	// Log the initialization params for debugging
	if log.GetLevel() == log.DebugLevel {
		paramsJSON, _ := json.MarshalIndent(initParams, "", "  ")
		log.Debug("Sending initialize request", "params", string(paramsJSON))
	}

	var result protocol.InitializeResult
	err := c.conn.Call(ctx, MethodInitialize, initParams, &result)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	// Store server capabilities
	c.capabilities = result.Capabilities

	// Handle offset encoding
	if result.OffsetEncoding != "" {
		switch result.OffsetEncoding {
		case "utf-8":
			c.offsetEncoding = UTF8
		case "utf-16":
			c.offsetEncoding = UTF16
		case "utf-32":
			c.offsetEncoding = UTF32
		}
	}

	// Send initialized notification
	err = c.conn.Notify(ctx, MethodInitialized, map[string]any{})
	if err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	c.initialized = true

	// For gopls, send workspace/didChangeConfiguration to ensure it's ready
	// This helps gopls properly set up its workspace views
	if strings.Contains(c.Name, "gopls") {
		configParams := map[string]any{
			"settings": c.config,
		}
		_ = c.conn.Notify(ctx, MethodWorkspaceDidChangeConfiguration, configParams)

		// Also send workspace/didChangeWatchedFiles to trigger gopls to scan the workspace
		// This helps with the "no views" error
		if c.rootURI != "" {
			changesParams := map[string]any{
				"changes": []map[string]any{
					{
						"uri":  c.rootURI,
						"type": 1, // Created
					},
				},
			}
			_ = c.conn.Notify(ctx, "workspace/didChangeWatchedFiles", changesParams)
		}
	}

	return nil
}

// Shutdown sends a shutdown request to the language server.
func (c *Client) Shutdown(ctx context.Context) error {
	if c.shutdown {
		return nil
	}

	err := c.conn.Call(ctx, MethodShutdown, nil, nil)
	if err != nil {
		return fmt.Errorf("shutdown request failed: %w", err)
	}

	c.shutdown = true
	return nil
}

// Exit sends an exit notification to the language server.
func (c *Client) Exit() error {
	err := c.conn.Notify(c.ctx, MethodExit, nil)
	if err != nil {
		return fmt.Errorf("exit notification failed: %w", err)
	}

	c.cancel()
	return nil
}

// GetCapabilities returns the server capabilities.
func (c *Client) GetCapabilities() protocol.ServerCapabilities {
	return c.capabilities
}

// IsInitialized returns whether the client has been initialized.
func (c *Client) IsInitialized() bool {
	return c.initialized
}

// IsRunning returns whether the client connection is still active.
func (c *Client) IsRunning() bool {
	return c.conn != nil && c.conn.IsConnected() && c.initialized && !c.shutdown
}

// RegisterNotificationHandler registers a handler for server-initiated notifications.
func (c *Client) RegisterNotificationHandler(method string, handler transport.NotificationHandler) {
	if c.conn != nil {
		c.conn.RegisterNotificationHandler(method, handler)
	}
}

// RegisterHandler registers a handler for server-initiated requests.
func (c *Client) RegisterHandler(method string, handler transport.Handler) {
	if c.conn != nil {
		c.conn.RegisterHandler(method, handler)
	}
}

// NotifyDidOpenTextDocument notifies the server that a document was opened.
func (c *Client) NotifyDidOpenTextDocument(ctx context.Context, uri string, languageID string, version int, text string) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": languageID,
			"version":    version,
			"text":       text,
		},
	}

	// Log what we're sending for debugging
	log.Debug("Sending textDocument/didOpen",
		"uri", uri,
		"languageId", languageID,
		"version", version,
		"textLength", len(text))

	return c.conn.Notify(ctx, MethodTextDocumentDidOpen, params)
}

// NotifyDidChangeTextDocument notifies the server that a document was changed.
func (c *Client) NotifyDidChangeTextDocument(ctx context.Context, uri string, version int, changes []protocol.TextDocumentContentChangeEvent) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":     uri,
			"version": version,
		},
		"contentChanges": changes,
	}

	return c.conn.Notify(ctx, MethodTextDocumentDidChange, params)
}

// NotifyDidChangeWatchedFiles notifies the server that watched files have
// changed.
func (c *Client) NotifyDidChangeWatchedFiles(ctx context.Context, changes []protocol.FileEvent) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"changes": changes,
	}

	return c.conn.Notify(ctx, MethodWorkspaceDidChangeWatchedFiles, params)
}

// NotifyWorkspaceDidChangeConfiguration notifies the server that the workspace configuration has changed.
func (c *Client) NotifyWorkspaceDidChangeConfiguration(ctx context.Context, settings any) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"settings": settings,
	}

	return c.conn.Notify(ctx, MethodWorkspaceDidChangeConfiguration, params)
}

// RequestCompletion requests completion items at the given position.
func (c *Client) RequestCompletion(ctx context.Context, uri string, position protocol.Position) (*protocol.CompletionList, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
		"context": map[string]any{
			"triggerKind": 1, // Invoked
		},
	}

	var result any
	err := c.conn.Call(ctx, MethodTextDocumentCompletion, params, &result)
	if err != nil {
		return nil, fmt.Errorf("completion request failed: %w", err)
	}

	// Parse the result - can be CompletionList or []CompletionItem
	var completionList protocol.CompletionList

	switch v := result.(type) {
	case map[string]any:
		// It's a CompletionList
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &completionList); err != nil {
			return nil, err
		}
	case []any:
		// It's an array of CompletionItem
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var items []protocol.CompletionItem
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err
		}
		completionList.Items = items
		completionList.IsIncomplete = false
	}

	return &completionList, nil
}

// RequestHover requests hover information at the given position.
func (c *Client) RequestHover(ctx context.Context, uri string, position protocol.Position) (*protocol.Hover, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	}

	var result protocol.Hover
	err := c.conn.Call(ctx, MethodTextDocumentHover, params, &result)
	if err != nil {
		return nil, fmt.Errorf("hover request failed: %w", err)
	}

	return &result, nil
}

// setupHandlers registers handlers for server-initiated requests.
func (c *Client) setupHandlers() {
	// Handle workspace/configuration requests
	c.conn.RegisterHandler(MethodWorkspaceConfiguration, func(ctx context.Context, method string, params json.RawMessage) (any, error) {
		var configParams protocol.ConfigurationParams
		if err := json.Unmarshal(params, &configParams); err != nil {
			return nil, err
		}

		// Return configuration for each requested item
		result := make([]any, len(configParams.Items))
		for i := range configParams.Items {
			result[i] = c.config
		}

		return result, nil
	})

	// Handle other common server requests
	// Add more handlers as needed
}

// makeClientCapabilities creates the client capabilities for initialization.
func (c *Client) makeClientCapabilities(enableSnippets bool) map[string]any {
	return map[string]any{
		"textDocument": map[string]any{
			"synchronization": map[string]any{
				"dynamicRegistration": true,
				"willSave":            true,
				"willSaveWaitUntil":   true,
				"didSave":             true,
			},
			"completion": map[string]any{
				"dynamicRegistration": true,
				"completionItem": map[string]any{
					"snippetSupport":          enableSnippets,
					"commitCharactersSupport": true,
					"documentationFormat":     []string{"markdown", "plaintext"},
					"deprecatedSupport":       true,
					"preselectSupport":        true,
					"insertReplaceSupport":    true,
					"tagSupport": map[string]any{
						"valueSet": []int{1}, // Deprecated
					},
					"resolveSupport": map[string]any{
						"properties": []string{"documentation", "detail", "additionalTextEdits"},
					},
				},
				"contextSupport": true,
			},
			"hover": map[string]any{
				"dynamicRegistration": true,
				"contentFormat":       []string{"markdown", "plaintext"},
			},
			"definition": map[string]any{
				"dynamicRegistration": true,
				"linkSupport":         true,
			},
			"references": map[string]any{
				"dynamicRegistration": true,
			},
			"documentHighlight": map[string]any{
				"dynamicRegistration": true,
			},
			"documentSymbol": map[string]any{
				"dynamicRegistration":               true,
				"hierarchicalDocumentSymbolSupport": true,
			},
			"formatting": map[string]any{
				"dynamicRegistration": true,
			},
			"rangeFormatting": map[string]any{
				"dynamicRegistration": true,
			},
			"rename": map[string]any{
				"dynamicRegistration": true,
				"prepareSupport":      true,
			},
			"publishDiagnostics": map[string]any{
				"relatedInformation":     true,
				"versionSupport":         true,
				"tagSupport":             map[string]any{"valueSet": []int{1, 2}},
				"codeDescriptionSupport": true,
				"dataSupport":            true,
			},
			"codeAction": map[string]any{
				"dynamicRegistration": true,
				"codeActionLiteralSupport": map[string]any{
					"codeActionKind": map[string]any{
						"valueSet": []string{
							"quickfix",
							"refactor",
							"refactor.extract",
							"refactor.inline",
							"refactor.rewrite",
							"source",
							"source.organizeImports",
						},
					},
				},
				"isPreferredSupport": true,
				"dataSupport":        true,
				"resolveSupport": map[string]any{
					"properties": []string{"edit"},
				},
			},
		},
		"workspace": map[string]any{
			"applyEdit": true,
			"workspaceEdit": map[string]any{
				"documentChanges":       true,
				"resourceOperations":    []string{"create", "rename", "delete"},
				"failureHandling":       "textOnlyTransactional",
				"normalizesLineEndings": true,
			},
			"didChangeConfiguration": map[string]any{
				"dynamicRegistration": true,
			},
			"didChangeWatchedFiles": map[string]any{
				"dynamicRegistration":    true,
				"relativePatternSupport": true,
			},
			"symbol": map[string]any{
				"dynamicRegistration": true,
			},
			"configuration":    true,
			"workspaceFolders": true,
			"fileOperations": map[string]any{
				"dynamicRegistration": true,
				"didCreate":           true,
				"willCreate":          true,
				"didRename":           true,
				"willRename":          true,
				"didDelete":           true,
				"willDelete":          true,
			},
		},
		"window": map[string]any{
			"workDoneProgress": true,
			"showMessage": map[string]any{
				"messageActionItem": map[string]any{
					"additionalPropertiesSupport": true,
				},
			},
			"showDocument": map[string]any{
				"support": true,
			},
		},
		"general": map[string]any{
			"regularExpressions": map[string]any{
				"engine":  "ECMAScript",
				"version": "ES2020",
			},
			"markdown": map[string]any{
				"parser":  "marked",
				"version": "1.1.0",
			},
			"positionEncodings": []string{"utf-16"},
		},
	}
}

// startServerProcess starts the language server process.
func startServerProcess(ctx context.Context, config ClientConfig) (io.ReadWriteCloser, error) {
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)

	// Set environment variables
	if config.Environment != nil {
		cmd.Env = os.Environ()
		for k, v := range config.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Create stderr pipe to capture error messages
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	// Monitor stderr
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Error("Error reading stderr", "error", err)
				}
				break
			}
			if n > 0 {
				log.Error("Language server stderr", "command", config.Command, "output", string(buf[:n]))
			}
		}
	}()

	// Monitor process exit
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error("Language server process exited with error", "command", config.Command, "error", err)
		} else {
			log.Info("Language server process exited normally", "command", config.Command)
		}
	}()

	// Create a stream transport
	stream := transport.NewStreamTransport(stdout, stdin, &processCloser{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	})

	return stream, nil
}

type processCloser struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
}

func (c *processCloser) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if err := c.stdin.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := c.stdout.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := c.stderr.Close(); err != nil {
		errs = append(errs, err)
	}

	// Give the process time to exit gracefully
	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited
	case <-time.After(5 * time.Second):
		// Timeout, kill the process
		if err := c.cmd.Process.Kill(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing process: %v", errs)
	}

	return nil
}
