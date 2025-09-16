package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// ServerConfig represents the configuration for a language server.
type ServerConfig struct {
	Command           string            `mapstructure:"command"`
	Args              []string          `mapstructure:"args"`
	FileTypes         []string          `mapstructure:"filetypes"`
	RootMarkers       []string          `mapstructure:"root_markers"`
	Environment       map[string]string `mapstructure:"environment"`
	Settings          map[string]any    `mapstructure:"settings"`
	InitOptions       map[string]any    `mapstructure:"init_options"`
	EnableSnippets    bool              `mapstructure:"enable_snippets"`
	SingleFileSupport bool              `mapstructure:"single_file_support"`
}

// Config represents the overall configuration.
type Config struct {
	Servers map[string]*ServerConfig `mapstructure:"servers"`
}

// Manager manages configuration loading and access.
type Manager struct {
	config *Config
}

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	return &Manager{
		config: &Config{
			Servers: make(map[string]*ServerConfig),
		},
	}
}

// LoadDefaults loads default server configurations.
func (m *Manager) LoadDefaults() {
	m.config.Servers = defaultServers()
	m.applyDefaults()
}

// GetServers returns all server configurations.
func (m *Manager) GetServers() map[string]*ServerConfig {
	return m.config.Servers
}

// GetServer returns a specific server configuration.
func (m *Manager) GetServer(name string) (*ServerConfig, bool) {
	server, exists := m.config.Servers[name]
	return server, exists
}

// AddServer adds or updates a server configuration.
func (m *Manager) AddServer(name string, config *ServerConfig) {
	m.config.Servers[name] = config
}

// RemoveServer removes a server configuration.
func (m *Manager) RemoveServer(name string) {
	delete(m.config.Servers, name)
}

// applyDefaults applies default values to server configurations.
func (m *Manager) applyDefaults() {
	for _, server := range m.config.Servers {
		if server.RootMarkers == nil {
			server.RootMarkers = []string{".git"}
		}

		if server.Environment == nil {
			server.Environment = make(map[string]string)
		}

		if server.Settings == nil {
			server.Settings = make(map[string]any)
		}
	}
}

// defaultServers returns default server configurations.
func defaultServers() map[string]*ServerConfig {
	return map[string]*ServerConfig{
		// Go
		"gopls": {
			Command:     "gopls",
			Args:        []string{},
			FileTypes:   []string{"go", "gomod", "gowork", "gotmpl"},
			RootMarkers: []string{"go.mod", "go.work", ".git"},
			InitOptions: map[string]any{
				"usePlaceholders":         true,
				"completionDocumentation": true,
				"deepCompletion":          true,
				"hoverKind":               "FullDocumentation",
			},
			Settings: map[string]any{
				"gopls": map[string]any{
					"usePlaceholders":         true,
					"completionDocumentation": true,
					"deepCompletion":          true,
					"hoverKind":               "FullDocumentation",
					"analyses": map[string]any{
						"unusedparams": true,
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Go linter - will be used if installed
		"golangci-lint-langserver": {
			Command:           "golangci-lint-langserver",
			Args:              []string{},
			FileTypes:         []string{"go"},
			RootMarkers:       []string{"go.mod", "go.work", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    false,
			SingleFileSupport: false,
		},

		// Rust
		"rust-analyzer": {
			Command:     "rust-analyzer",
			Args:        []string{},
			FileTypes:   []string{"rs"},
			RootMarkers: []string{"Cargo.toml", ".git"},
			Settings: map[string]any{
				"rust-analyzer": map[string]any{
					"cargo": map[string]any{
						"buildScripts": map[string]any{
							"enable": true,
						},
					},
					"procMacro": map[string]any{
						"enable": true,
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: false, // rust-analyzer requires a Cargo.toml
		},

		// TypeScript/JavaScript
		"typescript-language-server": {
			Command:     "typescript-language-server",
			Args:        []string{"--stdio"},
			FileTypes:   []string{"js", "jsx", "ts", "tsx", "mjs", "cjs"},
			RootMarkers: []string{"package.json", "tsconfig.json", "jsconfig.json", ".git"},
			Settings: map[string]any{
				"typescript": map[string]any{
					"inlayHints": map[string]any{
						"includeInlayParameterNameHints":                        "all",
						"includeInlayParameterNameHintsWhenArgumentMatchesName": false,
						"includeInlayFunctionParameterTypeHints":                true,
						"includeInlayVariableTypeHints":                         true,
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Python
		"pylsp": {
			Command:     "pylsp",
			Args:        []string{},
			FileTypes:   []string{"py", "pyi"},
			RootMarkers: []string{"pyproject.toml", "setup.py", "setup.cfg", "requirements.txt", "Pipfile", ".git"},
			Settings: map[string]any{
				"pylsp": map[string]any{
					"plugins": map[string]any{
						"pycodestyle": map[string]any{
							"enabled":       true,
							"maxLineLength": 88,
						},
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		"pyright": {
			Command:     "pyright-langserver",
			Args:        []string{"--stdio"},
			FileTypes:   []string{"py", "pyi"},
			RootMarkers: []string{"pyproject.toml", "setup.py", "setup.cfg", "requirements.txt", "Pipfile", "pyrightconfig.json", ".git"},
			Settings: map[string]any{
				"python": map[string]any{
					"analysis": map[string]any{
						"autoSearchPaths":        true,
						"diagnosticMode":         "openFilesOnly",
						"useLibraryCodeForTypes": true,
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		"ruff": {
			Command:           "ruff",
			Args:              []string{"server"},
			FileTypes:         []string{"py", "pyi"},
			RootMarkers:       []string{"pyproject.toml", "ruff.toml", ".ruff.toml", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// C/C++
		"clangd": {
			Command:        "clangd",
			Args:           []string{"--background-index"},
			FileTypes:      []string{"c", "cpp", "cc", "cxx", "h", "hpp", "hh", "hxx"},
			RootMarkers:    []string{"compile_commands.json", "compile_flags.txt", ".clangd", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Lua
		"lua-language-server": {
			Command:     "lua-language-server",
			Args:        []string{},
			FileTypes:   []string{"lua"},
			RootMarkers: []string{".luarc.json", ".luarc.jsonc", ".luacheckrc", ".stylua.toml", "stylua.toml", "selene.toml", "selene.yml", ".git"},
			Settings: map[string]any{
				"Lua": map[string]any{
					"diagnostics": map[string]any{
						"globals": []string{"vim"},
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Java
		"jdtls": {
			Command:        "jdtls",
			Args:           []string{},
			FileTypes:      []string{"java"},
			RootMarkers:    []string{"pom.xml", "build.gradle", "build.gradle.kts", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Ruby
		"solargraph": {
			Command:           "solargraph",
			Args:              []string{"stdio"},
			FileTypes:         []string{"rb", "ruby"},
			RootMarkers:       []string{"Gemfile", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		"ruby-lsp": {
			Command:           "ruby-lsp",
			Args:              []string{},
			FileTypes:         []string{"rb", "ruby"},
			RootMarkers:       []string{"Gemfile", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// PHP
		"intelephense": {
			Command:           "intelephense",
			Args:              []string{"--stdio"},
			FileTypes:         []string{"php"},
			RootMarkers:       []string{"composer.json", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// HTML/CSS/JSON
		"vscode-html-language-server": {
			Command:     "vscode-html-language-server",
			Args:        []string{"--stdio"},
			FileTypes:   []string{"html", "htm", "xhtml"},
			RootMarkers: []string{".git"},
			Settings: map[string]any{
				"html": map[string]any{
					"format": map[string]any{
						"enable": true,
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		"vscode-css-language-server": {
			Command:     "vscode-css-language-server",
			Args:        []string{"--stdio"},
			FileTypes:   []string{"css", "scss", "less"},
			RootMarkers: []string{".git"},
			Settings: map[string]any{
				"css": map[string]any{
					"validate": map[string]any{
						"enable": true,
					},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		"vscode-json-language-server": {
			Command:     "vscode-json-language-server",
			Args:        []string{"--stdio"},
			FileTypes:   []string{"json", "jsonc"},
			RootMarkers: []string{".git"},
			Settings: map[string]any{
				"json": map[string]any{
					"schemas": []any{},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// YAML
		"yaml-language-server": {
			Command:     "yaml-language-server",
			Args:        []string{"--stdio"},
			FileTypes:   []string{"yaml", "yml"},
			RootMarkers: []string{".git"},
			Settings: map[string]any{
				"yaml": map[string]any{
					"schemas": map[string]any{},
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Docker
		"docker-langserver": {
			Command:           "docker-langserver",
			Args:              []string{"--stdio"},
			FileTypes:         []string{"dockerfile"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Bash
		"bash-language-server": {
			Command:     "bash-language-server",
			Args:        []string{"start"},
			FileTypes:   []string{"sh", "bash", "zsh"},
			RootMarkers: []string{".git"},
			Settings: map[string]any{
				"bashIde": map[string]any{
					"globPattern": "*@(.sh|.inc|.bash|.command)",
				},
			},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Vim
		"vim-language-server": {
			Command:           "vim-language-server",
			Args:              []string{"--stdio"},
			FileTypes:         []string{"vim"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Elixir
		"elixir-ls": {
			Command:     "elixir-ls",
			Args:        []string{},
			FileTypes:   []string{"ex", "exs", "elixir"},
			RootMarkers: []string{"mix.exs", ".git"},
			Settings: map[string]any{
				"elixirLS": map[string]any{
					"dialyzerEnabled": false,
				},
			},
			EnableSnippets: true,
		},

		// Haskell
		"haskell-language-server": {
			Command:        "haskell-language-server-wrapper",
			Args:           []string{"--lsp"},
			FileTypes:      []string{"hs", "lhs"},
			RootMarkers:    []string{"stack.yaml", "cabal.project", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Kotlin
		"kotlin-language-server": {
			Command:        "kotlin-language-server",
			Args:           []string{},
			FileTypes:      []string{"kt", "kts"},
			RootMarkers:    []string{"build.gradle", "build.gradle.kts", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Swift
		"sourcekit-lsp": {
			Command:        "sourcekit-lsp",
			Args:           []string{},
			FileTypes:      []string{"swift"},
			RootMarkers:    []string{"Package.swift", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Zig
		"zls": {
			Command:           "zls",
			Args:              []string{},
			FileTypes:         []string{"zig"},
			RootMarkers:       []string{"build.zig", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Nix
		"nil": {
			Command:           "nil",
			Args:              []string{},
			FileTypes:         []string{"nix"},
			RootMarkers:       []string{"flake.nix", "default.nix", "shell.nix", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		"nixd": {
			Command:           "nixd",
			Args:              []string{},
			FileTypes:         []string{"nix"},
			RootMarkers:       []string{"flake.nix", "default.nix", "shell.nix", ".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// TOML
		"taplo": {
			Command:           "taplo",
			Args:              []string{"lsp", "stdio"},
			FileTypes:         []string{"toml"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Terraform
		"terraform-ls": {
			Command:        "terraform-ls",
			Args:           []string{"serve"},
			FileTypes:      []string{"tf", "tfvars"},
			RootMarkers:    []string{".terraform", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Svelte
		"svelteserver": {
			Command:        "svelteserver",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"svelte"},
			RootMarkers:    []string{"package.json", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Vue
		"volar": {
			Command:        "vue-language-server",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"vue"},
			RootMarkers:    []string{"package.json", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Astro
		"astro-ls": {
			Command:        "astro-ls",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"astro"},
			RootMarkers:    []string{"package.json", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Prisma
		"prisma-language-server": {
			Command:        "prisma-language-server",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"prisma"},
			RootMarkers:    []string{"schema.prisma", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// GraphQL
		"graphql-lsp": {
			Command:        "graphql-lsp",
			Args:           []string{"server", "-m", "stream"},
			FileTypes:      []string{"graphql", "gql"},
			RootMarkers:    []string{".graphqlrc", ".graphqlrc.json", ".graphqlrc.yaml", ".graphqlrc.yml", ".graphqlrc.js", ".graphqlrc.ts", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Tailwind CSS
		"tailwindcss-ls": {
			Command:        "tailwindcss-language-server",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"html", "css", "javascript", "javascriptreact", "typescript", "typescriptreact", "vue", "svelte"},
			RootMarkers:    []string{"tailwind.config.js", "tailwind.config.ts", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Markdown
		"marksman": {
			Command:           "marksman",
			Args:              []string{"server"},
			FileTypes:         []string{"md", "markdown"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// LaTeX
		"texlab": {
			Command:           "texlab",
			Args:              []string{},
			FileTypes:         []string{"tex", "bib"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Dart/Flutter
		"dartls": {
			Command:        "dart",
			Args:           []string{"language-server", "--protocol=lsp"},
			FileTypes:      []string{"dart"},
			RootMarkers:    []string{"pubspec.yaml", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Scala
		"metals": {
			Command:        "metals",
			Args:           []string{},
			FileTypes:      []string{"scala", "sbt"},
			RootMarkers:    []string{"build.sbt", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// OCaml
		"ocamllsp": {
			Command:        "ocamllsp",
			Args:           []string{},
			FileTypes:      []string{"ml", "mli"},
			RootMarkers:    []string{"dune-project", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Erlang
		"erlang-ls": {
			Command:        "erlang_ls",
			Args:           []string{},
			FileTypes:      []string{"erl", "hrl"},
			RootMarkers:    []string{"rebar.config", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Clojure
		"clojure-lsp": {
			Command:        "clojure-lsp",
			Args:           []string{},
			FileTypes:      []string{"clj", "cljs", "cljc", "edn"},
			RootMarkers:    []string{"project.clj", "deps.edn", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// F#
		"fsautocomplete": {
			Command:        "fsautocomplete",
			Args:           []string{"--adaptive-lsp-server-enabled"},
			FileTypes:      []string{"fs", "fsi", "fsx"},
			RootMarkers:    []string{".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Solidity
		"solidity-ls": {
			Command:        "solidity-ls",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"sol"},
			RootMarkers:    []string{"hardhat.config.js", "truffle-config.js", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// R
		"r-languageserver": {
			Command:           "R",
			Args:              []string{"--no-echo", "-e", "languageserver::run()"},
			FileTypes:         []string{"r", "rmd"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// Julia
		"julia-lsp": {
			Command:        "julia",
			Args:           []string{"--startup-file=no", "--history-file=no", "-e", "using LanguageServer; runserver()"},
			FileTypes:      []string{"jl"},
			RootMarkers:    []string{"Project.toml", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Perl
		"perlnavigator": {
			Command:           "perlnavigator",
			Args:              []string{"--stdio"},
			FileTypes:         []string{"pl", "pm"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},

		// CMake
		"cmake-language-server": {
			Command:        "cmake-language-server",
			Args:           []string{},
			FileTypes:      []string{"cmake"},
			RootMarkers:    []string{"CMakeLists.txt", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Ansible
		"ansible-language-server": {
			Command:        "ansible-language-server",
			Args:           []string{"--stdio"},
			FileTypes:      []string{"yaml.ansible", "ansible"},
			RootMarkers:    []string{"ansible.cfg", ".ansible-lint", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Protobuf
		"buf-language-server": {
			Command:        "buf",
			Args:           []string{"beta", "lsp"},
			FileTypes:      []string{"proto"},
			RootMarkers:    []string{"buf.yaml", ".git"},
			Settings:       map[string]any{},
			EnableSnippets: true,
		},

		// Assembly
		"asm-lsp": {
			Command:           "asm-lsp",
			Args:              []string{},
			FileTypes:         []string{"asm", "s", "S"},
			RootMarkers:       []string{".git"},
			Settings:          map[string]any{},
			EnableSnippets:    true,
			SingleFileSupport: true,
		},
	}
}

// LoadFromMap loads configuration from a map (useful for testing).
func (m *Manager) LoadFromMap(data map[string]any) error {
	var config Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &config,
		TagName: "mapstructure",
	})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	m.config = &config
	m.applyDefaults()

	return nil
}
