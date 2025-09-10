package lsp

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/superjoy/powernap/pkg/lsp/protocol"
)

// DetectLanguage detects the language of a given file path.
func DetectLanguage(path string) protocol.LanguageKind {
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "dockerfile":
		return protocol.LangDockerfile
	case "go.mod":
		return protocol.LangGoMod
	case "go.sum":
		return protocol.LangGoSum
	case "makefile", "gnumakefile":
		return protocol.LangMakefile
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".abap":
		return protocol.LangABAP
	case ".bat":
		return protocol.LangWindowsBat
	case ".bib", ".bibtex":
		return protocol.LangBibTeX
	case ".clj", ".cljs", ".cljc":
		return protocol.LangClojure
	case ".coffee":
		return protocol.LangCoffeescript
	case ".c", ".h":
		return protocol.LangC
	case ".cpp", ".cxx", ".cc", ".c++", ".hpp", ".hh", ".hxx", ".h++":
		return protocol.LangCPP
	case ".cs":
		return protocol.LangCSharp
	case ".css":
		return protocol.LangCSS
	case ".d":
		return protocol.LangD
	case ".pas", ".pascal":
		return protocol.LangDelphi
	case ".diff", ".patch":
		return protocol.LangDiff
	case ".dart":
		return protocol.LangDart
	case ".dockerfile":
		return protocol.LangDockerfile
	case ".ex", ".exs":
		return protocol.LangElixir
	case ".erl", ".hrl":
		return protocol.LangErlang
	case ".fs", ".fsi", ".fsx", ".fsscript":
		return protocol.LangFSharp
	case ".gitcommit":
		return protocol.LangGitCommit
	case ".gitrebase":
		return protocol.LangGitRebase
	case ".go":
		return protocol.LangGo
	case ".groovy":
		return protocol.LangGroovy
	case ".hbs", ".handlebars":
		return protocol.LangHandlebars
	case ".hs":
		return protocol.LangHaskell
	case ".html", ".htm":
		return protocol.LangHTML
	case ".ini":
		return protocol.LangIni
	case ".java":
		return protocol.LangJava
	case ".js", ".mjs", ".cjs":
		return protocol.LangJavaScript
	case ".jsx":
		return protocol.LangJavaScriptReact
	case ".json", ".jsonc":
		return protocol.LangJSON
	case ".tex", ".latex":
		return protocol.LangLaTeX
	case ".less":
		return protocol.LangLess
	case ".lua":
		return protocol.LangLua
	case ".makefile", "makefile", "gnumakefile":
		return protocol.LangMakefile
	case ".md", ".markdown":
		return protocol.LangMarkdown
	case ".m":
		return protocol.LangObjectiveC
	case ".mm":
		return protocol.LangObjectiveCPP
	case ".pl":
		return protocol.LangPerl
	case ".pm":
		return protocol.LangPerl6
	case ".php":
		return protocol.LangPHP
	case ".ps1", ".psm1":
		return protocol.LangPowershell
	case ".pug", ".jade":
		return protocol.LangPug
	case ".py", ".pyi":
		return protocol.LangPython
	case ".r":
		return protocol.LangR
	case ".cshtml", ".razor":
		return protocol.LangRazor
	case ".rb":
		return protocol.LangRuby
	case ".rs":
		return protocol.LangRust
	case ".scss":
		return protocol.LangSCSS
	case ".sass":
		return protocol.LangSASS
	case ".scala":
		return protocol.LangScala
	case ".shader":
		return protocol.LangShaderLab
	case ".sh", ".bash", ".zsh", ".ksh":
		return protocol.LangShellScript
	case ".sql":
		return protocol.LangSQL
	case ".swift":
		return protocol.LangSwift
	case ".ts":
		return protocol.LangTypeScript
	case ".tsx":
		return protocol.LangTypeScriptReact
	case ".xml":
		return protocol.LangXML
	case ".xsl":
		return protocol.LangXSL
	case ".yaml", ".yml":
		return protocol.LangYAML
	case ".toml":
		return protocol.LangTOML
	case ".fish":
		return protocol.LangFish
	case ".vim":
		return protocol.LangVim
	case ".kt", ".kts":
		return protocol.LangKotlin
	case ".vue":
		return protocol.LangVue
	case ".svelte":
		return protocol.LangSvelte
	default:
		return protocol.LanguageKind("") // Unknown language
	}
}
