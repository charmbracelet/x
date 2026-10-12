// Package lsp provides a client implementation for the Language Server
// Protocol (LSP).
package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/charmbracelet/x/powernap/pkg/transport"
)

// LSP method constants.
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
	MethodTextDocumentRename                 = "textDocument/rename"
	MethodTextDocumentDocumentSymbol         = "textDocument/documentSymbol"
	MethodTextDocumentPrepareCallHierarchy   = "textDocument/prepareCallHierarchy"
	MethodCallHierarchyIncomingCalls         = "callHierarchy/incomingCalls"
	MethodCallHierarchyOutgoingCalls         = "callHierarchy/outgoingCalls"
	MethodWorkspaceConfiguration             = "workspace/configuration"
	MethodWorkspaceWorkspaceFolders          = "workspace/workspaceFolders"
	MethodWorkspaceDidChangeConfiguration    = "workspace/didChangeConfiguration"
	MethodWorkspaceDidChangeWorkspaceFolders = "workspace/didChangeWorkspaceFolders"
	MethodWorkspaceDidChangeWatchedFiles     = "workspace/didChangeWatchedFiles"
)

// Repeated JSON field names used both as LSP/JSON-RPC keys (in
// map[string]any literals such as makeClientCapabilities) and as
// slog field labels. Centralized here so goconst doesn't flag the
// legitimate repetition; refactoring these literals to typed
// protocol structs would be a much bigger change.
const (
	fieldURI                 = "uri"
	fieldVersion             = "version"
	fieldTextDocument        = "textDocument"
	fieldDynamicRegistration = "dynamicRegistration"
	fieldValueSet            = "valueSet"
	// markupKindMarkdown is the LSP MarkupKind value "markdown",
	// used in documentationFormat / contentFormat arrays and as a
	// nested key under "general".
	markupKindMarkdown = "markdown"
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
	conn, err := transport.NewConnection(ctx, stream, slog.Default())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	client.conn = conn

	// Register handlers for server-initiated requests
	client.setupHandlers()

	return client, nil
}

// Kill forcefully terminates the client by canceling the context and closing
// the connection. This ensures any blocked I/O operations are interrupted.
func (c *Client) Kill() {
	c.cancel()
	_ = c.conn.Close()
}

// Initialize sends the initialize request to the language server.
func (c *Client) Initialize(ctx context.Context, enableSnippets bool) error {
	if c.initialized.Load() {
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
			"name":       "powernap",
			fieldVersion: "0.1.0",
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
	paramsJSON, _ := json.MarshalIndent(initParams, "", "  ")
	slog.Debug("Sending initialize request", "params", string(paramsJSON))

	var result protocol.InitializeResult
	err := c.conn.Call(ctx, MethodInitialize, initParams, &result)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	// Store server capabilities
	c.capabilities = result.Capabilities

	c.offsetEncoding = parseOffsetEncoding(result.Capabilities.PositionEncoding, result.OffsetEncoding)

	// Send initialized notification
	err = c.conn.Notify(ctx, MethodInitialized, map[string]any{})
	if err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	c.initialized.Store(true)

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
						fieldURI: c.rootURI,
						"type":   1, // Created
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
	if c.shutdown.Load() {
		return nil
	}

	err := c.conn.Call(ctx, MethodShutdown, nil, nil)
	if err != nil {
		return fmt.Errorf("shutdown request failed: %w", err)
	}

	c.shutdown.Store(true)
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
	return c.initialized.Load()
}

// IsRunning returns whether the client connection is still active.
func (c *Client) IsRunning() bool {
	return c.conn != nil && c.conn.IsConnected() && c.initialized.Load() && !c.shutdown.Load()
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
	if !c.initialized.Load() {
		return fmt.Errorf("client not initialized")
	}

	params := protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(uri),
			LanguageID: protocol.LanguageKind(languageID),
			Version:    int32(version), //nolint:gosec
			Text:       text,
		},
	}

	// Log what we're sending for debugging
	slog.Debug("Sending textDocument/didOpen",
		fieldURI, uri,
		"languageId", languageID,
		fieldVersion, version,
		"textLength", len(text))

	return c.conn.Notify(ctx, MethodTextDocumentDidOpen, params) //nolint:wrapcheck
}

// NotifyDidCloseTextDocument notifies the server that a document was closeed.
func (c *Client) NotifyDidCloseTextDocument(ctx context.Context, uri string) error {
	if !c.initialized.Load() {
		return fmt.Errorf("client not initialized")
	}

	params := protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: protocol.DocumentURI(uri),
		},
	}

	return c.conn.Notify(ctx, MethodTextDocumentDidClose, params) //nolint:wrapcheck
}

// NotifyDidChangeTextDocument notifies the server that a document was changed.
func (c *Client) NotifyDidChangeTextDocument(ctx context.Context, uri string, version int, changes []protocol.TextDocumentContentChangeEvent) error {
	if !c.initialized.Load() {
		return fmt.Errorf("client not initialized")
	}

	params := protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			Version: int32(version), //nolint:gosec
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(uri),
			},
		},
		ContentChanges: changes,
	}

	return c.conn.Notify(ctx, MethodTextDocumentDidChange, params) //nolint:wrapcheck
}

// NotifyDidChangeWatchedFiles notifies the server that watched files have
// changed.
func (c *Client) NotifyDidChangeWatchedFiles(ctx context.Context, changes []protocol.FileEvent) error {
	if !c.initialized.Load() {
		return fmt.Errorf("client not initialized")
	}

	params := protocol.DidChangeWatchedFilesParams{
		Changes: changes,
	}

	return c.conn.Notify(ctx, MethodWorkspaceDidChangeWatchedFiles, params) //nolint:wrapcheck
}

// NotifyWorkspaceDidChangeConfiguration notifies the server that the workspace configuration has changed.
func (c *Client) NotifyWorkspaceDidChangeConfiguration(ctx context.Context, settings any) error {
	if !c.initialized.Load() {
		return fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		"settings": settings,
	}

	return c.conn.Notify(ctx, MethodWorkspaceDidChangeConfiguration, params) //nolint:wrapcheck
}

// RequestCompletion requests completion items at the given position.
func (c *Client) RequestCompletion(ctx context.Context, uri string, position protocol.Position) (*protocol.CompletionList, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := protocol.CompletionParams{
		Context: protocol.CompletionContext{
			TriggerKind: protocol.Invoked,
		},
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(uri),
			},
			Position: position,
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
			return nil, err //nolint:wrapcheck
		}
		if err := json.Unmarshal(data, &completionList); err != nil {
			return nil, err //nolint:wrapcheck
		}
	case []any:
		// It's an array of CompletionItem
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		var items []protocol.CompletionItem
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err //nolint:wrapcheck
		}
		completionList.Items = items
		completionList.IsIncomplete = false
	}

	return &completionList, nil
}

// RequestHover requests hover information at the given position.
func (c *Client) RequestHover(ctx context.Context, uri string, position protocol.Position) (*protocol.Hover, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]any{
		fieldTextDocument: map[string]any{
			fieldURI: uri,
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

// FindReferences finds all references to the symbol at the given position.
func (c *Client) FindReferences(ctx context.Context, filepath string, line, character int, includeDeclaration bool) ([]protocol.Location, error) {
	uri := string(protocol.URIFromPath(filepath))
	params := protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(uri),
			},
			Position: protocol.Position{
				Line:      uint32(line),      //nolint:gosec
				Character: uint32(character), //nolint:gosec
			},
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: includeDeclaration,
		},
	}

	var result []protocol.Location
	err := c.conn.Call(ctx, MethodTextDocumentReferences, params, &result)
	if err != nil {
		return nil, fmt.Errorf("find references request failed: %w", err)
	}
	return result, nil
}

// RequestRename requests a rename of the symbol at the given position.
func (c *Client) RequestRename(ctx context.Context, filepath string, line, character int, newName string) (*protocol.WorkspaceEdit, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	uri := string(protocol.URIFromPath(filepath))
	params := protocol.RenameParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: protocol.DocumentURI(uri),
		},
		Position: protocol.Position{
			Line:      uint32(line),      //nolint:gosec
			Character: uint32(character), //nolint:gosec
		},
		NewName: newName,
	}

	var result protocol.WorkspaceEdit
	err := c.conn.Call(ctx, MethodTextDocumentRename, params, &result)
	if err != nil {
		return nil, fmt.Errorf("rename request failed: %w", err)
	}
	return &result, nil
}

// RequestDocumentSymbols requests the document symbols for the given file.
func (c *Client) RequestDocumentSymbols(ctx context.Context, filepath string) ([]protocol.DocumentSymbolResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	uri := string(protocol.URIFromPath(filepath))
	params := protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: protocol.DocumentURI(uri),
		},
	}

	var result protocol.Or_Result_textDocument_documentSymbol
	err := c.conn.Call(ctx, MethodTextDocumentDocumentSymbol, params, &result)
	if err != nil {
		return nil, fmt.Errorf("document symbol request failed: %w", err)
	}
	return result.Results() //nolint:wrapcheck
}

// RequestDefinition requests the definition of the symbol at the given position.
func (c *Client) RequestDefinition(ctx context.Context, filepath string, line, character int) ([]protocol.Location, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	uri := string(protocol.URIFromPath(filepath))
	params := protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(uri),
			},
			Position: protocol.Position{
				Line:      uint32(line),      //nolint:gosec
				Character: uint32(character), //nolint:gosec
			},
		},
	}

	var result json.RawMessage
	err := c.conn.Call(ctx, MethodTextDocumentDefinition, params, &result)
	if err != nil {
		return nil, fmt.Errorf("definition request failed: %w", err)
	}

	if string(result) == "null" {
		return nil, nil
	}

	// Try []Location first (most common response from gopls).
	var locs []protocol.Location
	if err := json.Unmarshal(result, &locs); err == nil && len(locs) > 0 {
		return locs, nil
	}

	// Try single Location.
	var loc protocol.Location
	if err := json.Unmarshal(result, &loc); err == nil && loc.URI != "" {
		return []protocol.Location{loc}, nil
	}

	return nil, nil
}

// PrepareCallHierarchy prepares a call hierarchy item at the given position.
func (c *Client) PrepareCallHierarchy(ctx context.Context, filepath string, line, character int) ([]protocol.CallHierarchyItem, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	uri := string(protocol.URIFromPath(filepath))
	params := protocol.CallHierarchyPrepareParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(uri),
			},
			Position: protocol.Position{
				Line:      uint32(line),      //nolint:gosec
				Character: uint32(character), //nolint:gosec
			},
		},
	}

	var result []protocol.CallHierarchyItem
	err := c.conn.Call(ctx, MethodTextDocumentPrepareCallHierarchy, params, &result)
	if err != nil {
		return nil, fmt.Errorf("prepare call hierarchy request failed: %w", err)
	}
	return result, nil
}

// IncomingCalls returns all callers of the given call hierarchy item.
func (c *Client) IncomingCalls(ctx context.Context, item protocol.CallHierarchyItem) ([]protocol.CallHierarchyIncomingCall, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := protocol.CallHierarchyIncomingCallsParams{
		Item: item,
	}

	var result []protocol.CallHierarchyIncomingCall
	err := c.conn.Call(ctx, MethodCallHierarchyIncomingCalls, params, &result)
	if err != nil {
		return nil, fmt.Errorf("incoming calls request failed: %w", err)
	}
	return result, nil
}

// OutgoingCalls returns all callees of the given call hierarchy item.
func (c *Client) OutgoingCalls(ctx context.Context, item protocol.CallHierarchyItem) ([]protocol.CallHierarchyOutgoingCall, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := protocol.CallHierarchyOutgoingCallsParams{
		Item: item,
	}

	var result []protocol.CallHierarchyOutgoingCall
	err := c.conn.Call(ctx, MethodCallHierarchyOutgoingCalls, params, &result)
	if err != nil {
		return nil, fmt.Errorf("outgoing calls request failed: %w", err)
	}
	return result, nil
}

func parseEncoding(encoding string) (OffsetEncoding, bool) {
	switch encoding {
	case "utf-8":
		return UTF8, true
	case "utf-16":
		return UTF16, true
	case "utf-32":
		return UTF32, true
	default:
		return UTF16, false
	}
}

func parseOffsetEncoding(positionEncoding *protocol.PositionEncodingKind, offsetEncoding string) OffsetEncoding {
	if positionEncoding != nil {
		if encoding, ok := parseEncoding(string(*positionEncoding)); ok {
			return encoding
		}
		slog.Warn("Unknown positionEncoding from language server; falling back", "positionEncoding", string(*positionEncoding))
	}
	if encoding, ok := parseEncoding(offsetEncoding); ok {
		return encoding
	}
	if offsetEncoding != "" {
		slog.Warn("Unknown offsetEncoding from language server; using UTF-16 default", "offsetEncoding", offsetEncoding)
	}
	return UTF16
}

// setupHandlers registers handlers for server-initiated requests.
func (c *Client) setupHandlers() {
	// Handle workspace/configuration requests
	c.conn.RegisterHandler(MethodWorkspaceConfiguration, func(_ context.Context, _ string, params json.RawMessage) (any, error) {
		var configParams protocol.ConfigurationParams
		if err := json.Unmarshal(params, &configParams); err != nil {
			return nil, err //nolint:wrapcheck
		}

		// Return configuration for each requested item
		result := make([]any, len(configParams.Items))
		for i := range configParams.Items {
			result[i] = c.config
		}

		return result, nil
	})

	// Handle workspace/workspaceFolders requests
	c.conn.RegisterHandler(MethodWorkspaceWorkspaceFolders, func(_ context.Context, _ string, _ json.RawMessage) (any, error) {
		// Return configured workspace folders or empty array
		folders := c.workspaceFolders
		if folders == nil {
			folders = []protocol.WorkspaceFolder{}
		}
		return folders, nil
	})

	// Handle other common server requests
	// Add more handlers as needed
}

// makeClientCapabilities creates the client capabilities for initialization.
func (c *Client) makeClientCapabilities(enableSnippets bool) map[string]any {
	return map[string]any{
		fieldTextDocument: map[string]any{
			"synchronization": map[string]any{
				fieldDynamicRegistration: true,
				"willSave":               true,
				"willSaveWaitUntil":      true,
				"didSave":                true,
			},
			"completion": map[string]any{
				fieldDynamicRegistration: true,
				"completionItem": map[string]any{
					"snippetSupport":          enableSnippets,
					"commitCharactersSupport": true,
					"documentationFormat":     []string{markupKindMarkdown, "plaintext"},
					"deprecatedSupport":       true,
					"preselectSupport":        true,
					"insertReplaceSupport":    true,
					"tagSupport": map[string]any{
						fieldValueSet: []int{1}, // Deprecated
					},
					"resolveSupport": map[string]any{
						"properties": []string{"documentation", "detail", "additionalTextEdits"},
					},
				},
				"contextSupport": true,
			},
			"hover": map[string]any{
				fieldDynamicRegistration: true,
				"contentFormat":          []string{markupKindMarkdown, "plaintext"},
			},
			"definition": map[string]any{
				fieldDynamicRegistration: true,
				"linkSupport":            true,
			},
			"references": map[string]any{
				fieldDynamicRegistration: true,
			},
			"documentHighlight": map[string]any{
				fieldDynamicRegistration: true,
			},
			"documentSymbol": map[string]any{
				fieldDynamicRegistration:            true,
				"hierarchicalDocumentSymbolSupport": true,
			},
			"formatting": map[string]any{
				fieldDynamicRegistration: true,
			},
			"rangeFormatting": map[string]any{
				fieldDynamicRegistration: true,
			},
			"rename": map[string]any{
				fieldDynamicRegistration: true,
				"prepareSupport":         true,
			},
			"publishDiagnostics": map[string]any{
				"relatedInformation":     true,
				"versionSupport":         true,
				"tagSupport":             map[string]any{fieldValueSet: []int{1, 2}},
				"codeDescriptionSupport": true,
				"dataSupport":            true,
			},
			"codeAction": map[string]any{
				fieldDynamicRegistration: true,
				"codeActionLiteralSupport": map[string]any{
					"codeActionKind": map[string]any{
						fieldValueSet: []string{
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
				fieldDynamicRegistration: true,
			},
			"didChangeWatchedFiles": map[string]any{
				fieldDynamicRegistration: true,
				"relativePatternSupport": true,
			},
			"symbol": map[string]any{
				fieldDynamicRegistration: true,
			},
			"configuration":    true,
			"workspaceFolders": true,
			"fileOperations": map[string]any{
				fieldDynamicRegistration: true,
				"didCreate":              true,
				"willCreate":             true,
				"didRename":              true,
				"willRename":             true,
				"didDelete":              true,
				"willDelete":             true,
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
				"engine":     "ECMAScript",
				fieldVersion: "ES2020",
			},
			markupKindMarkdown: map[string]any{
				"parser":     "marked",
				fieldVersion: "1.1.0",
			},
			"positionEncodings": []string{"utf-8", "utf-16"},
		},
	}
}

// startServerProcess starts the language server process.
func startServerProcess(ctx context.Context, config ClientConfig) (io.ReadWriteCloser, error) {
	cmd := exec.CommandContext(ctx, config.Command, config.Args...) //nolint:gosec

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
					slog.Error("Error reading stderr", "error", err)
				}
				break
			}
			if n > 0 {
				slog.Error("Language server stderr", "command", config.Command, "output", string(buf[:n]))
			}
		}
	}()

	closer := &processCloser{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}

	return transport.NewStreamTransport(stdout, stdin, closer), nil
}

type processCloser struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	closeOnce sync.Once
	closeErr  error
}

func (c *processCloser) Close() error {
	c.closeOnce.Do(func() {
		errs := []error{
			c.stdin.Close(),
			c.stdout.Close(),
			c.stderr.Close(),
		}

		done := make(chan error, 1)
		go func() {
			done <- c.cmd.Wait()
		}()

		timeout := time.After(5 * time.Second)
		select {
		case err := <-done:
			errs = append(errs, err)
		case <-timeout:
			errs = append(errs, c.cmd.Process.Kill())
			<-done
		}

		c.closeErr = errors.Join(errs...)
	})
	return c.closeErr
}
