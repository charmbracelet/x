package lsp

import (
	"fmt"
	"path/filepath"
	"strings"
)

// TextDocumentSyncManager manages text document synchronization with the language server.
type TextDocumentSyncManager struct {
	client    *Client
	documents map[string]*Document
	syncKind  TextDocumentSyncKind
}

// Document represents an open text document.
type Document struct {
	URI        string
	LanguageID string
	Version    int
	Content    string
}

// NewTextDocumentSyncManager creates a new text document sync manager.
func NewTextDocumentSyncManager(client *Client) *TextDocumentSyncManager {
	syncKind := TextDocumentSyncFull // Default to full sync

	// Extract sync kind from capabilities
	switch v := client.capabilities.TextDocumentSync.(type) {
	case float64:
		syncKind = TextDocumentSyncKind(int(v))
	case int:
		syncKind = TextDocumentSyncKind(v)
	case map[string]interface{}:
		// It's a TextDocumentSyncOptions object
		if change, ok := v["change"].(float64); ok {
			syncKind = TextDocumentSyncKind(int(change))
		}
	case *TextDocumentSyncOptions:
		syncKind = v.Change
	}

	return &TextDocumentSyncManager{
		client:    client,
		documents: make(map[string]*Document),
		syncKind:  syncKind,
	}
}

// Open opens a new text document.
func (m *TextDocumentSyncManager) Open(uri, languageID, content string) error {
	if _, exists := m.documents[uri]; exists {
		return fmt.Errorf("document already open: %s", uri)
	}

	doc := &Document{
		URI:        uri,
		LanguageID: languageID,
		Version:    1,
		Content:    content,
	}

	m.documents[uri] = doc

	return m.client.NotifyDidOpenTextDocument(m.client.ctx, uri, languageID, doc.Version, content)
}

// Change applies changes to an open document.
func (m *TextDocumentSyncManager) Change(uri string, changes []TextDocumentContentChangeEvent) error {
	doc, exists := m.documents[uri]
	if !exists {
		return fmt.Errorf("document not open: %s", uri)
	}

	// Apply changes based on sync kind
	switch m.syncKind {
	case TextDocumentSyncFull:
		// For full sync, we expect a single change with the full content
		if len(changes) > 0 {
			doc.Content = changes[0].Text
		}
	case TextDocumentSyncIncremental:
		// Apply incremental changes
		for _, change := range changes {
			if err := m.applyIncrementalChange(doc, change); err != nil {
				return err
			}
		}
	}

	doc.Version++

	return m.client.NotifyDidChangeTextDocument(m.client.ctx, uri, doc.Version, changes)
}

// Close closes a text document.
func (m *TextDocumentSyncManager) Close(uri string) error {
	if _, exists := m.documents[uri]; !exists {
		return fmt.Errorf("document not open: %s", uri)
	}

	delete(m.documents, uri)

	// Send didClose notification
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}

	return m.client.conn.Notify(m.client.ctx, MethodTextDocumentDidClose, params)
}

// Save notifies the server that a document was saved.
func (m *TextDocumentSyncManager) Save(uri string, includeText bool) error {
	doc, exists := m.documents[uri]
	if !exists {
		return fmt.Errorf("document not open: %s", uri)
	}

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}

	if includeText {
		params["text"] = doc.Content
	}

	return m.client.conn.Notify(m.client.ctx, MethodTextDocumentDidSave, params)
}

// GetDocument returns the document for the given URI.
func (m *TextDocumentSyncManager) GetDocument(uri string) (*Document, bool) {
	doc, exists := m.documents[uri]
	return doc, exists
}

// applyIncrementalChange applies an incremental change to a document.
func (m *TextDocumentSyncManager) applyIncrementalChange(doc *Document, change TextDocumentContentChangeEvent) error {
	if change.Range == nil {
		// Full document change
		doc.Content = change.Text
		return nil
	}

	// Convert content to lines for easier manipulation
	lines := strings.Split(doc.Content, "\n")

	// Validate range
	if change.Range.Start.Line < 0 || change.Range.Start.Line >= len(lines) {
		return fmt.Errorf("invalid start line: %d", change.Range.Start.Line)
	}
	if change.Range.End.Line < 0 || change.Range.End.Line >= len(lines) {
		return fmt.Errorf("invalid end line: %d", change.Range.End.Line)
	}

	// Calculate the start and end positions in the document
	startPos := 0
	for i := 0; i < change.Range.Start.Line; i++ {
		startPos += len(lines[i]) + 1 // +1 for newline
	}
	startPos += change.Range.Start.Character

	endPos := 0
	for i := 0; i < change.Range.End.Line; i++ {
		endPos += len(lines[i]) + 1
	}
	endPos += change.Range.End.Character

	// Apply the change
	newContent := doc.Content[:startPos] + change.Text + doc.Content[endPos:]
	doc.Content = newContent

	return nil
}

// CreateFullDocumentChange creates a change event for full document sync.
func CreateFullDocumentChange(content string) []TextDocumentContentChangeEvent {
	return []TextDocumentContentChangeEvent{
		{
			Text: content,
		},
	}
}

// OpenFile opens a new text document from a file path.
func (m *TextDocumentSyncManager) OpenFile(filePath, content string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Detect language from file path
	languageID := detectLanguage(absPath)

	// Create file URI
	uri := FilePathToURI(absPath)

	return m.Open(uri, languageID, content)
}

// FilePathToURI converts a file path to a file URI.
func FilePathToURI(filePath string) string {
	// Ensure absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	// Convert to forward slashes for URI
	absPath = filepath.ToSlash(absPath)

	// Add file:// prefix
	if strings.HasPrefix(absPath, "/") {
		return "file://" + absPath
	}
	// Windows path
	return "file:///" + absPath
}

// URIToFilePath converts a file URI to a file path.
func URIToFilePath(uri string) string {
	// Remove file:// prefix
	path := strings.TrimPrefix(uri, "file://")
	path = strings.TrimPrefix(path, "file:///")

	// Convert to native path separators
	return filepath.FromSlash(path)
}

// detectLanguage detects the language ID from file path.
func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	base := filepath.Base(filePath)
	
	// Check specific filenames first
	switch base {
	case "Dockerfile":
		return "dockerfile"
	case "Makefile", "makefile", "GNUmakefile":
		return "makefile"
	case "go.mod", "go.sum":
		return "go.mod"
	case "Cargo.toml", "Cargo.lock":
		return "toml"
	case "package.json", "tsconfig.json", "jsconfig.json":
		return "json"
	case "pyproject.toml":
		return "toml"
	}
	
	// Check extensions
	switch ext {
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescriptreact"
	case ".jsx":
		return "javascriptreact"
	case ".py", ".pyi":
		return "python"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "cpp"
	case ".h":
		// Could be C or C++, default to C
		return "c"
	case ".hpp", ".hh", ".hxx", ".h++":
		return "cpp"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".lua":
		return "lua"
	case ".json", ".jsonc":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".md", ".markdown":
		return "markdown"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".sass":
		return "sass"
	case ".less":
		return "less"
	case ".xml":
		return "xml"
	case ".sh", ".bash":
		return "shellscript"
	case ".zsh":
		return "shellscript"
	case ".fish":
		return "fish"
	case ".vim":
		return "vim"
	case ".tex":
		return "latex"
	case ".r":
		return "r"
	case ".sql":
		return "sql"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".clj", ".cljs", ".cljc":
		return "clojure"
	case ".ex", ".exs":
		return "elixir"
	case ".erl", ".hrl":
		return "erlang"
	case ".dart":
		return "dart"
	case ".hs", ".lhs":
		return "haskell"
	case ".ml", ".mli":
		return "ocaml"
	case ".fs", ".fsi", ".fsx":
		return "fsharp"
	case ".zig":
		return "zig"
	case ".nix":
		return "nix"
	case ".vue":
		return "vue"
	case ".svelte":
		return "svelte"
	case ".astro":
		return "astro"
	case ".proto":
		return "proto"
	case ".graphql", ".gql":
		return "graphql"
	case ".tf", ".tfvars":
		return "terraform"
	case ".prisma":
		return "prisma"
	case ".sol":
		return "solidity"
	case ".jl":
		return "julia"
	case ".pl", ".pm":
		return "perl"
	case ".cmake":
		return "cmake"
	case ".asm", ".s":
		return "asm"
	default:
		// Default to plaintext
		return "plaintext"
	}
}
