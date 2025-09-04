package lsp

import (
	"context"
	"time"

	"github.com/charmbracelet/x/powernap/pkg/transport"
	"github.com/sourcegraph/jsonrpc2"
)

// OffsetEncoding represents the character encoding used for text document offsets.
type OffsetEncoding int

const (
	// UTF8 encoding - bytes
	UTF8 OffsetEncoding = iota
	// UTF16 encoding - default for LSP
	UTF16
	// UTF32 encoding - codepoints
	UTF32
)

// Client represents an LSP client connection to a language server.
type Client struct {
	ID              string
	Name            string
	conn            *transport.Connection
	ctx             context.Context
	cancel          context.CancelFunc
	initialized     bool
	shutdown        bool
	capabilities    ServerCapabilities
	offsetEncoding  OffsetEncoding
	rootURI         string
	workspaceFolders []WorkspaceFolder
	config          interface{}
	initOptions     interface{}
}

// ServerCapabilities represents the capabilities of a language server.
type ServerCapabilities struct {
	TextDocumentSync           interface{}                      `json:"textDocumentSync,omitempty"` // Can be TextDocumentSyncKind or TextDocumentSyncOptions
	CompletionProvider         *CompletionOptions               `json:"completionProvider,omitempty"`
	HoverProvider              interface{}                      `json:"hoverProvider,omitempty"` // Can be bool or object
	DefinitionProvider         interface{}                      `json:"definitionProvider,omitempty"` // Can be bool or object
	ReferencesProvider         interface{}                      `json:"referencesProvider,omitempty"` // Can be bool or object
	DocumentHighlightProvider  interface{}                      `json:"documentHighlightProvider,omitempty"` // Can be bool or object
	DocumentSymbolProvider     interface{}                      `json:"documentSymbolProvider,omitempty"` // Can be bool or object
	WorkspaceSymbolProvider    interface{}                      `json:"workspaceSymbolProvider,omitempty"` // Can be bool or object
	CodeActionProvider         interface{}                      `json:"codeActionProvider,omitempty"`
	CodeLensProvider           *CodeLensOptions                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider interface{}                      `json:"documentFormattingProvider,omitempty"` // Can be bool or object
	DocumentRangeFormattingProvider interface{}                 `json:"documentRangeFormattingProvider,omitempty"` // Can be bool or object
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	RenameProvider             interface{}                      `json:"renameProvider,omitempty"`
	DocumentLinkProvider       *DocumentLinkOptions             `json:"documentLinkProvider,omitempty"`
	ExecuteCommandProvider     *ExecuteCommandOptions           `json:"executeCommandProvider,omitempty"`
	SemanticTokensProvider     interface{}                      `json:"semanticTokensProvider,omitempty"`
	Workspace                  *WorkspaceCapabilities           `json:"workspace,omitempty"`
}

// TextDocumentSyncOptions represents text document sync options.
type TextDocumentSyncOptions struct {
	OpenClose         bool                 `json:"openClose,omitempty"`
	Change            TextDocumentSyncKind `json:"change,omitempty"`
	WillSave          bool                 `json:"willSave,omitempty"`
	WillSaveWaitUntil bool                 `json:"willSaveWaitUntil,omitempty"`
	Save              interface{}          `json:"save,omitempty"` // Can be bool or SaveOptions
}

// TextDocumentSyncKind defines how text documents are synced.
type TextDocumentSyncKind int

const (
	TextDocumentSyncNone TextDocumentSyncKind = iota
	TextDocumentSyncFull
	TextDocumentSyncIncremental
)

// CompletionOptions represents completion provider options.
type CompletionOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	AllCommitCharacters []string `json:"allCommitCharacters,omitempty"`
	ResolveProvider     bool     `json:"resolveProvider,omitempty"`
}

// CodeLensOptions represents code lens provider options.
type CodeLensOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

// DocumentOnTypeFormattingOptions represents on-type formatting options.
type DocumentOnTypeFormattingOptions struct {
	FirstTriggerCharacter string   `json:"firstTriggerCharacter"`
	MoreTriggerCharacter  []string `json:"moreTriggerCharacter,omitempty"`
}

// DocumentLinkOptions represents document link provider options.
type DocumentLinkOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

// ExecuteCommandOptions represents execute command provider options.
type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

// WorkspaceCapabilities represents workspace-specific capabilities.
type WorkspaceCapabilities struct {
	WorkspaceFolders WorkspaceFoldersServerCapabilities `json:"workspaceFolders,omitempty"`
	FileOperations   *FileOperationOptions              `json:"fileOperations,omitempty"`
}

// WorkspaceFoldersServerCapabilities represents workspace folder capabilities.
type WorkspaceFoldersServerCapabilities struct {
	Supported           bool        `json:"supported,omitempty"`
	ChangeNotifications interface{} `json:"changeNotifications,omitempty"`
}

// FileOperationOptions represents file operation capabilities.
type FileOperationOptions struct {
	DidCreate *FileOperationRegistrationOptions
	WillCreate *FileOperationRegistrationOptions
	DidRename *FileOperationRegistrationOptions
	WillRename *FileOperationRegistrationOptions
	DidDelete *FileOperationRegistrationOptions
	WillDelete *FileOperationRegistrationOptions
}

// FileOperationRegistrationOptions represents file operation registration options.
type FileOperationRegistrationOptions struct {
	Filters []FileOperationFilter
}

// FileOperationFilter represents a file operation filter.
type FileOperationFilter struct {
	Scheme  string
	Pattern FileOperationPattern
}

// FileOperationPattern represents a file operation pattern.
type FileOperationPattern struct {
	Glob    string
	Matches string
	Options *FileOperationPatternOptions
}

// FileOperationPatternOptions represents file operation pattern options.
type FileOperationPatternOptions struct {
	IgnoreCase bool
}

// WorkspaceFolder represents a workspace folder.
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// ClientConfig represents the configuration for creating a new LSP client.
type ClientConfig struct {
	Command         string
	Args            []string
	RootURI         string
	WorkspaceFolders []WorkspaceFolder
	InitOptions     interface{}
	Settings        interface{}
	Environment     map[string]string
	Timeout         time.Duration
}

// Notification represents an LSP notification.
type Notification struct {
	Method string
	Params interface{}
}

// MethodCall represents an LSP method call from the server.
type MethodCall struct {
	ID     jsonrpc2.ID
	Method string
	Params interface{}
}
