// Package config represents configuration management for language servers.
package config

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

//go:embed lsps.json
var lspsJSON []byte

// ServerConfig represents the configuration for a language server.
type ServerConfig struct {
	Command           string            `mapstructure:"command" json:"command"`
	Args              []string          `mapstructure:"args" json:"args,omitempty"`
	FileTypes         []string          `mapstructure:"filetypes" json:"filetypes"`
	RootMarkers       []string          `mapstructure:"root_markers" json:"root_markers"`
	Environment       map[string]string `mapstructure:"environment" json:"environment,omitempty"`
	Settings          map[string]any    `mapstructure:"settings" json:"settings,omitempty"`
	InitOptions       map[string]any    `mapstructure:"init_options" json:"init_options,omitempty"`
	EnableSnippets    bool              `mapstructure:"enable_snippets" json:"-"`
	SingleFileSupport bool              `mapstructure:"single_file_support" json:"-"`
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

// LoadDefaults loads default server configurations from the embedded JSON.
func (m *Manager) LoadDefaults() error {
	servers := make(map[string]*ServerConfig)
	if err := json.Unmarshal(lspsJSON, &servers); err != nil {
		return fmt.Errorf("failed to parse embedded lsps.json: %w", err)
	}

	// Filter to only supported servers and apply overrides
	for name, server := range servers {
		if _, ok := supportedServers[name]; !ok {
			delete(servers, name)
			continue
		}
		if len(server.Command) == 0 {
			delete(servers, name)
			continue
		}

		if _, ok := snippetSupport[name]; ok {
			server.EnableSnippets = true
		}
		if _, ok := singleFileSupport[name]; ok {
			server.SingleFileSupport = true
		}
	}

	m.config.Servers = servers
	m.applyDefaults()
	return nil
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
