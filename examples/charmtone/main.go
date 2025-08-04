// Package main is a simple command line tool for rendering the CharmTone color
// palette.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/spf13/cobra"
)

const (
	blackCircle = "●"
	whiteCircle = "○"
	rightArrow  = "→"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "charmtone",
		Short: "CharmTone color palette tool",
		Long:  "A command line tool for rendering the CharmTone color palette in various formats",
		Run: func(_ *cobra.Command, _ []string) {
			renderGuide()
		},
	}

	cssCmd := &cobra.Command{
		Use:   "css",
		Short: "Generate CSS variables",
		Long:  "Generate CSS custom properties (variables) for the CharmTone color palette",
		Run: func(_ *cobra.Command, _ []string) {
			renderCSS()
		},
	}

	scssCmd := &cobra.Command{
		Use:   "scss",
		Short: "Print as SCSS variables",
		Long:  "Print SCSS variables for the CharmTone color palette",
		Run: func(_ *cobra.Command, _ []string) {
			renderSCSS()
		},
	}

	vimCmd := &cobra.Command{
		Use:   "vim",
		Short: "Generate Vim colorscheme",
		Long:  "Generate Vim colorscheme for the CharmTone color palette",
		Run: func(_ *cobra.Command, _ []string) {
			renderVim()
		},
	}

	nixCmd := &cobra.Command{
		Use:   "nix",
		Short: "Generate Nix attributes",
		Long:  "Generate Nix attributes for the CharmTone color palette",
		Run: func(_ *cobra.Command, _ []string) {
			renderNix()
		},
	}

	rootCmd.AddCommand(cssCmd, scssCmd, vimCmd, nixCmd)

	// Use Fang to execute the command with enhanced styling and features
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func renderSCSS() {
	for _, k := range charmtone.Keys() {
		name := strings.ToLower(strings.ReplaceAll(k.String(), " ", "-"))
		fmt.Printf("$%s: %s;\n", name, k.Hex())
	}
}

func renderVim() {
	for _, k := range charmtone.Keys() {
		name := strings.ToLower(strings.ReplaceAll(k.String(), " ", "-"))
		fmt.Printf("let %s = '%s'\n", name, k.Hex())
	}
}

func renderCSS() {
	for _, k := range charmtone.Keys() {
		name := strings.ToLower(strings.ReplaceAll(k.String(), " ", "-"))
		fmt.Printf("--charmtone-%s: %s;\n", name, k.Hex())
	}
}

func renderNix() {
	keys := charmtone.Keys()
	for _, k := range keys {
		name := strings.ToLower(strings.ReplaceAll(k.String(), " ", "-"))
		fmt.Printf("%s = \"%s\";\n", name, k.Hex())
	}
}
