package cellbuf

import (
	"sync"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

var parserPool = sync.Pool{
	New: func() any {
		return ansi.NewParser(parser.MaxParamsSize, 1024*4) // 4MB data buffer
	},
}

// GetParser returns a parser from the pool.
func GetParser() *ansi.Parser {
	return parserPool.Get().(*ansi.Parser)
}

// PutParser returns a parser to the pool.
func PutParser(p *ansi.Parser) {
	p.Reset()
	p.DataLen = 0
	parserPool.Put(p)
}
