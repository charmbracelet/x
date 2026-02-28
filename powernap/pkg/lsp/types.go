package lsp

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/charmbracelet/x/powernap/pkg/transport"
)

// OffsetEncoding represents the character encoding used for text document offsets.
type OffsetEncoding int

const (
	// UTF8 encoding - bytes.
	UTF8 OffsetEncoding = iota
	// UTF16 encoding - default for LSP.
	UTF16
	// UTF32 encoding - codepoints.
	UTF32
)

// GetOffsetEncoding returns the negotiated offset encoding for this client.
// This is set after initialization based on the server's capabilities.
func (c *Client) GetOffsetEncoding() OffsetEncoding {
	return c.offsetEncoding
}

// PositionToByteOffset converts a UTF-16 character offset to a byte offset
// in the given line text. This is necessary because LSP positions are
// specified in UTF-16 code units by default, but Go strings use UTF-8 bytes.
//
// The function handles:
//   - ASCII characters (1 UTF-16 unit = 1 byte)
//   - BMP characters like CJK (1 UTF-16 unit = 2-3 bytes in UTF-8)
//   - Supplementary characters like emoji (2 UTF-16 units = 4 bytes in UTF-8)
//
// If utf16Char is beyond the end of the line, returns len(lineText).
func PositionToByteOffset(lineText string, utf16Char uint32) int {
	if utf16Char == 0 {
		return 0
	}

	var utf16Count uint32
	for byteOffset, r := range lineText {
		if utf16Count >= utf16Char {
			return byteOffset
		}
		// Characters outside BMP (U+10000+) are represented as
		// surrogate pairs in UTF-16, counting as 2 code units.
		var width uint32 = 1
		if r >= 0x10000 {
			width = 2
		}
		// If the desired UTF-16 offset falls within this rune's UTF-16 width
		// (i.e., in the middle of a surrogate pair), clamp to the start.
		if utf16Char < utf16Count+width {
			return byteOffset
		}
		utf16Count += width
	}
	return len(lineText)
}

// Client represents an LSP client connection to a language server.
type Client struct {
	ID               string
	Name             string
	conn             *transport.Connection
	ctx              context.Context
	cancel           context.CancelFunc
	initialized      atomic.Bool
	shutdown         atomic.Bool
	capabilities     protocol.ServerCapabilities
	offsetEncoding   OffsetEncoding
	rootURI          string
	workspaceFolders []protocol.WorkspaceFolder
	config           map[string]any
	initOptions      map[string]any
}

// ClientConfig represents the configuration for creating a new LSP client.
type ClientConfig struct {
	Command          string
	Args             []string
	RootURI          string
	WorkspaceFolders []protocol.WorkspaceFolder
	InitOptions      map[string]any
	Settings         map[string]any
	Environment      map[string]string
	Timeout          time.Duration
}
