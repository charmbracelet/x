package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/x/powernap/pkg/config"
	"github.com/charmbracelet/x/powernap/pkg/lsp"
	"github.com/charmbracelet/x/powernap/pkg/registry"
)

// Example showing how to use LoadFromMap for custom configuration
func customConfigExample() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: example <file>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Create configuration manager
	cfgManager := config.NewManager()

	// Load configuration from a map instead of TOML file
	configMap := map[string]any{
		"servers": map[string]any{
			"gopls": map[string]any{
				"command":             "gopls",
				"args":                []string{"-remote=auto"},
				"filetypes":           []string{"go", "gomod", "gowork", "gotmpl"},
				"root_markers":        []string{"go.mod", "go.work", ".git"},
				"enable_snippets":     true,
				"single_file_support": true,
				"settings": map[string]any{
					"gopls": map[string]any{
						"usePlaceholders": true,
						"analyses": map[string]any{
							"unusedparams": true,
							"unusedwrite":  true,
						},
					},
				},
			},
			"rust-analyzer": map[string]any{
				"command":         "rust-analyzer",
				"filetypes":       []string{"rs"},
				"root_markers":    []string{"Cargo.toml", ".git"},
				"enable_snippets": true,
				"settings": map[string]any{
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
			},
		},
	}

	// Load the configuration from the map
	if err := cfgManager.LoadFromMap(configMap); err != nil {
		log.Fatalf("Failed to load config from map: %v", err)
	}

	// Create registry
	reg := registry.New()
	if err := reg.LoadConfig(cfgManager); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Get client for file (language detection happens internally)
	client, err := reg.GetClientForFile(ctx, filePath)
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	fmt.Printf("Started language server for %s\n", filePath)

	// Create sync manager
	syncMgr := lsp.NewTextDocumentSyncManager(client)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Open document (file path to URI conversion happens internally)
	if err := syncMgr.OpenFile(filePath, string(content)); err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}

	// Request hover at position 0,0
	fileURI := lsp.FilePathToURI(filePath)
	hover, err := client.RequestHover(ctx, fileURI, lsp.Position{Line: 0, Character: 0})
	if err != nil {
		log.Printf("Hover request failed: %v", err)
	} else if hover != nil {
		fmt.Printf("Hover: %+v\n", hover.Contents)
	}

	// Request completion at position 1,0
	completions, err := client.RequestCompletion(ctx, fileURI, lsp.Position{Line: 1, Character: 0})
	if err != nil {
		log.Printf("Completion request failed: %v", err)
	} else if completions != nil {
		fmt.Printf("Found %d completions\n", len(completions.Items))
	}

	// Keep running for a bit
	time.Sleep(5 * time.Second)

	// Cleanup
	if err := reg.StopAll(ctx); err != nil {
		log.Printf("Error stopping servers: %v", err)
	}
}
