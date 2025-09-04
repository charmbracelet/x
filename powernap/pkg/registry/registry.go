package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/powernap/pkg/config"
	"github.com/charmbracelet/x/powernap/pkg/lsp"
)

// Registry manages multiple language server instances.
type Registry struct {
	mu       sync.RWMutex
	clients  map[string]*lsp.Client
	configs  map[string]*config.ServerConfig
	logger   *log.Logger
}

// New creates a new registry.
func New() *Registry {
	return &Registry{
		clients: make(map[string]*lsp.Client),
		configs: make(map[string]*config.ServerConfig),
		logger:  log.Default(),
	}
}

// NewWithLogger creates a new registry with a custom logger.
func NewWithLogger(logger *log.Logger) *Registry {
	return &Registry{
		clients: make(map[string]*lsp.Client),
		configs: make(map[string]*config.ServerConfig),
		logger:  logger,
	}
}

// LoadConfig loads server configurations from a config manager.
func (r *Registry) LoadConfig(cfg *config.Manager) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	servers := cfg.GetServers()
	for name, serverCfg := range servers {
		r.configs[name] = serverCfg
	}
	
	return nil
}

// StartServer starts a language server for the given name and project path.
func (r *Registry) StartServer(ctx context.Context, name string, projectPath string) (*lsp.Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if server is already running
	if client, exists := r.clients[name]; exists {
		return client, nil
	}
	
	// Get server configuration
	serverCfg, exists := r.configs[name]
	if !exists {
		return nil, fmt.Errorf("no configuration found for server: %s", name)
	}
	
	// Find project root
	rootPath := r.findProjectRoot(projectPath, serverCfg.RootMarkers)
	if rootPath == "" {
		// Check if server supports single file mode
		if !serverCfg.SingleFileSupport {
			return nil, fmt.Errorf("language server %s requires a project root with one of: %v", name, serverCfg.RootMarkers)
		}
		rootPath = projectPath
	}
	
	// Create workspace folders
	workspaceFolders := []lsp.WorkspaceFolder{
		{
			URI:  "file://" + rootPath,
			Name: filepath.Base(rootPath),
		},
	}
	
	// Create client configuration
	clientCfg := lsp.ClientConfig{
		Command:          serverCfg.Command,
		Args:             serverCfg.Args,
		RootURI:          "file://" + rootPath,
		WorkspaceFolders: workspaceFolders,
		InitOptions:      serverCfg.InitOptions,
		Settings:         serverCfg.Settings,
		Environment:      serverCfg.Environment,
	}
	
	// Create and initialize client
	client, err := lsp.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	
	// Initialize the client
	if err := client.Initialize(ctx, serverCfg.EnableSnippets); err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}
	
	// Store the client
	r.clients[name] = client
	
	r.logger.Info("Started language server", "name", name, "root", rootPath)
	return client, nil
}

// StopServer stops a running language server.
func (r *Registry) StopServer(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	client, exists := r.clients[name]
	if !exists {
		return fmt.Errorf("server not running: %s", name)
	}
	
	// Shutdown the client
	if err := client.Shutdown(ctx); err != nil {
		r.logger.Error("Failed to shutdown server", "name", name, "error", err)
	}
	
	// Send exit notification
	if err := client.Exit(); err != nil {
		r.logger.Error("Failed to exit server", "name", name, "error", err)
	}
	
	// Remove from registry
	delete(r.clients, name)
	
	r.logger.Info("Stopped language server", "name", name)
	return nil
}

// RestartServer restarts a language server.
func (r *Registry) RestartServer(ctx context.Context, name string, projectPath string) (*lsp.Client, error) {
	// Stop the server if it's running
	if err := r.StopServer(ctx, name); err != nil {
		// Ignore error if server wasn't running
		r.logger.Debug("Server was not running", "name", name)
	}
	
	// Start the server
	return r.StartServer(ctx, name, projectPath)
}

// GetClient returns a running client by name.
func (r *Registry) GetClient(name string) (*lsp.Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	client, exists := r.clients[name]
	return client, exists
}

// GetClientsForFile returns all appropriate clients for the given file.
// This allows multiple language servers to handle the same file type (e.g., gopls and golangci-lint for Go files).
func (r *Registry) GetClientsForFile(ctx context.Context, filePath string) ([]*lsp.Client, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Detect language from file
	language := detectLanguage(absPath)
	if language == "" {
		return nil, fmt.Errorf("unsupported file type: %s", filepath.Ext(absPath))
	}
	
	r.mu.RLock()
	
	// Find all servers that support this language
	var serverNames []string
	for name, cfg := range r.configs {
		for _, ft := range cfg.FileTypes {
			// Match by extension or language ID
			ext := filepath.Ext(absPath)
			if ft == language || ft == strings.TrimPrefix(ext, ".") || "."+ft == ext {
				serverNames = append(serverNames, name)
				break
			}
		}
	}
	
	r.mu.RUnlock()
	
	if len(serverNames) == 0 {
		return nil, fmt.Errorf("no language servers found for language: %s", language)
	}
	
	var clients []*lsp.Client
	projectDir := filepath.Dir(absPath)
	
	// Start or get each server
	for _, serverName := range serverNames {
		// Check if server is already running
		if client, exists := r.GetClient(serverName); exists {
			clients = append(clients, client)
		} else {
			// Start the server
			client, err := r.StartServer(ctx, serverName, projectDir)
			if err != nil {
				r.logger.Warn("Failed to start server", "name", serverName, "error", err)
				continue
			}
			clients = append(clients, client)
		}
	}
	
	if len(clients) == 0 {
		return nil, fmt.Errorf("failed to start any language servers for language: %s", language)
	}
	
	return clients, nil
}

// GetClientForFile returns a single appropriate client for the given file.
// This is a convenience method that returns the first available client.
// For multiple servers support, use GetClientsForFile instead.
func (r *Registry) GetClientForFile(ctx context.Context, filePath string) (*lsp.Client, error) {
	clients, err := r.GetClientsForFile(ctx, filePath)
	if err != nil {
		return nil, err
	}
	
	if len(clients) == 0 {
		return nil, fmt.Errorf("no clients available")
	}
	
	return clients[0], nil
}

// ListClients returns a list of running clients.
func (r *Registry) ListClients() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.clients))
	for name := range r.clients {
		names = append(names, name)
	}
	
	return names
}

// StopAll stops all running language servers.
func (r *Registry) StopAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var errs []error
	
	for name, client := range r.clients {
		if err := client.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown %s: %w", name, err))
		}
		
		if err := client.Exit(); err != nil {
			errs = append(errs, fmt.Errorf("failed to exit %s: %w", name, err))
		}
	}
	
	// Clear all clients
	r.clients = make(map[string]*lsp.Client)
	
	if len(errs) > 0 {
		return fmt.Errorf("errors stopping servers: %v", errs)
	}
	
	return nil
}

// findProjectRoot finds the project root based on root markers.
func (r *Registry) findProjectRoot(startPath string, rootMarkers []string) string {
	currentPath := startPath
	
	for {
		// Check for root markers
		for _, marker := range rootMarkers {
			markerPath := filepath.Join(currentPath, marker)
			if fileExists(markerPath) {
				return currentPath
			}
		}
		
		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached filesystem root
			break
		}
		
		currentPath = parentPath
	}
	
	return ""
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	// Use os.Stat to check if the file/directory exists
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
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
	case ".vue":
		return "vue"
	case ".svelte":
		return "svelte"
	default:
		return ""
	}
}
