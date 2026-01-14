package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	m := NewManager()
	if err := m.LoadDefaults(); err != nil {
		t.Fatalf("LoadDefaults failed: %v", err)
	}

	servers := m.GetServers()
	if len(servers) == 0 {
		t.Fatal("Expected some servers to be loaded")
	}

	// Check a few known servers
	testCases := []struct {
		name      string
		cmd       string
		filetypes []string
	}{
		{"gopls", "gopls", []string{"go", "gomod", "gowork", "gotmpl"}},
		{"clangd", "clangd", []string{"c", "cpp", "objc", "objcpp", "cuda"}},
		{"rust_analyzer", "rust-analyzer", []string{"rust"}},
		{"ts_ls", "typescript-language-server", []string{"javascript", "javascriptreact", "javascript.jsx", "typescript", "typescriptreact", "typescript.tsx"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, ok := m.GetServer(tc.name)
			if !ok {
				t.Fatalf("Server %s not found", tc.name)
			}

			if server.Command != tc.cmd {
				t.Errorf("Expected command %q, got %q", tc.cmd, server.Command)
			}

			if len(server.FileTypes) != len(tc.filetypes) {
				t.Errorf("Expected %d filetypes, got %d", len(tc.filetypes), len(server.FileTypes))
			}
		})
	}
}

func TestGetServer(t *testing.T) {
	m := NewManager()
	if err := m.LoadDefaults(); err != nil {
		t.Fatalf("LoadDefaults failed: %v", err)
	}

	// Test existing server
	server, ok := m.GetServer("gopls")
	if !ok {
		t.Fatal("Expected gopls to exist")
	}
	if server.Command != "gopls" {
		t.Errorf("Expected command gopls, got %s", server.Command)
	}

	// Test non-existing server
	_, ok = m.GetServer("nonexistent")
	if ok {
		t.Fatal("Expected nonexistent server to not exist")
	}
}

func TestSingleFileSupport(t *testing.T) {
	m := NewManager()
	if err := m.LoadDefaults(); err != nil {
		t.Fatalf("LoadDefaults failed: %v", err)
	}

	// These servers should have single file support from upstream
	withSupport := []string{"gopls", "pylsp", "bashls", "lua_ls", "efm", "rust_analyzer"}
	for _, name := range withSupport {
		server, ok := m.GetServer(name)
		if !ok {
			t.Errorf("Server %s not found", name)
			continue
		}
		if !server.SingleFileSupport {
			t.Errorf("Expected %s to have SingleFileSupport=true", name)
		}
	}

	// ada_ls does not have single file support
	ada, ok := m.GetServer("ada_ls")
	if !ok {
		t.Fatal("ada_ls not found")
	}
	if ada.SingleFileSupport {
		t.Error("Expected ada_ls to have SingleFileSupport=false")
	}
}

func TestAddRemoveServer(t *testing.T) {
	m := NewManager()
	if err := m.LoadDefaults(); err != nil {
		t.Fatalf("LoadDefaults failed: %v", err)
	}

	// Add a new server
	m.AddServer("test_server", &ServerConfig{
		Command:   "test-lsp",
		Args:      []string{"--stdio"},
		FileTypes: []string{"test"},
	})

	server, ok := m.GetServer("test_server")
	if !ok {
		t.Fatal("Expected test_server to exist after adding")
	}
	if server.Command != "test-lsp" {
		t.Errorf("Expected command test-lsp, got %s", server.Command)
	}

	// Remove the server
	m.RemoveServer("test_server")
	_, ok = m.GetServer("test_server")
	if ok {
		t.Fatal("Expected test_server to not exist after removing")
	}
}
