package lsp

import "encoding/json"

// Position represents a position in a text document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           DiagnosticSeverity             `json:"severity,omitempty"`
	Code               any                            `json:"code,omitempty"`
	CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	Tags               []DiagnosticTag                `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
	Data               any                            `json:"data,omitempty"`
}

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// DiagnosticTag represents additional metadata about the diagnostic.
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// CodeDescription represents a description for an error code.
type CodeDescription struct {
	Href string `json:"href"`
}

// DiagnosticRelatedInformation represents related diagnostic information.
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// TextDocumentContentChangeEvent represents a change to a text document.
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// CompletionItem represents a completion item.
type CompletionItem struct {
	Label               string              `json:"label"`
	Kind                CompletionItemKind  `json:"kind,omitempty"`
	Tags                []CompletionItemTag `json:"tags,omitempty"`
	Detail              string              `json:"detail,omitempty"`
	Documentation       any                 `json:"documentation,omitempty"`
	Deprecated          bool                `json:"deprecated,omitempty"`
	Preselect           bool                `json:"preselect,omitempty"`
	SortText            string              `json:"sortText,omitempty"`
	FilterText          string              `json:"filterText,omitempty"`
	InsertText          string              `json:"insertText,omitempty"`
	InsertTextFormat    InsertTextFormat    `json:"insertTextFormat,omitempty"`
	InsertTextMode      InsertTextMode      `json:"insertTextMode,omitempty"`
	TextEdit            *TextEdit           `json:"textEdit,omitempty"`
	TextEditText        string              `json:"textEditText,omitempty"`
	AdditionalTextEdits []TextEdit          `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string            `json:"commitCharacters,omitempty"`
	Command             *Command            `json:"command,omitempty"`
	Data                any                 `json:"data,omitempty"`
}

// CompletionItemKind represents the kind of a completion item.
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// CompletionItemTag represents additional metadata about a completion item.
type CompletionItemTag int

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// InsertTextFormat represents how the insert text should be interpreted.
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// InsertTextMode represents how whitespace and indentation is handled during completion.
type InsertTextMode int

const (
	InsertTextModeAsIs              InsertTextMode = 1
	InsertTextModeAdjustIndentation InsertTextMode = 2
)

// TextEdit represents a textual edit.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Command represents a reference to a command.
type Command struct {
	Title     string `json:"title"`
	Command   string `json:"command"`
	Arguments []any  `json:"arguments,omitempty"`
}

// CompletionList represents a collection of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// Hover represents hover information.
type Hover struct {
	Contents any    `json:"contents"`
	Range    *Range `json:"range,omitempty"`
}

// MarkupContent represents content with markup.
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// MarkupKind represents the kind of markup.
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// InitializeResult represents the result of an initialize request.
type InitializeResult struct {
	Capabilities   ServerCapabilities `json:"capabilities"`
	ServerInfo     *ServerInfo        `json:"serverInfo,omitempty"`
	OffsetEncoding string             `json:"offsetEncoding,omitempty"`
}

// ServerInfo represents information about the server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ConfigurationParams represents parameters for workspace/configuration request.
type ConfigurationParams struct {
	Items []ConfigurationItem `json:"items"`
}

// ConfigurationItem represents a configuration section to retrieve.
type ConfigurationItem struct {
	ScopeURI string `json:"scopeUri,omitempty"`
	Section  string `json:"section,omitempty"`
}

// PublishDiagnosticsParams represents parameters for textDocument/publishDiagnostics notification.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     int          `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// ShowMessageParams represents parameters for window/showMessage notification.
type ShowMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

// MessageType represents the type of a message.
type MessageType int

const (
	MessageTypeError   MessageType = 1
	MessageTypeWarning MessageType = 2
	MessageTypeInfo    MessageType = 3
	MessageTypeLog     MessageType = 4
)

// LogMessageParams represents parameters for window/logMessage notification.
type LogMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

// ProgressParams represents parameters for $/progress notification.
type ProgressParams struct {
	Token ProgressToken   `json:"token"`
	Value json.RawMessage `json:"value"`
}

// ProgressToken represents a progress token.
type ProgressToken any

// WorkDoneProgressBegin represents the beginning of a work done progress.
type WorkDoneProgressBegin struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Cancellable bool   `json:"cancellable,omitempty"`
	Message     string `json:"message,omitempty"`
	Percentage  int    `json:"percentage,omitempty"`
}

// WorkDoneProgressReport represents a progress report.
type WorkDoneProgressReport struct {
	Kind        string `json:"kind"`
	Cancellable bool   `json:"cancellable,omitempty"`
	Message     string `json:"message,omitempty"`
	Percentage  int    `json:"percentage,omitempty"`
}

// WorkDoneProgressEnd represents the end of a work done progress.
type WorkDoneProgressEnd struct {
	Kind    string `json:"kind"`
	Message string `json:"message,omitempty"`
}

// FileChangeType represents the type of a file change.
type FileChangeType uint32

const (
	// Created indicates the file got created.
	Created FileChangeType = 1
	// Change indicates the file got changed.
	Changed FileChangeType = 2
	// Deleted indicates the file got deleted.
	Deleted FileChangeType = 3
)

// FileEvent represents a file change event.
type FileEvent struct {
	URI  string `json:"uri"`
	Type uint32 `json:"type"`
}
