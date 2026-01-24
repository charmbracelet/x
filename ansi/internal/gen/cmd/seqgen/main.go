package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"

	"github.com/charmbracelet/x/ansi/internal/gen"
	"gopkg.in/yaml.v3"
)

func main() {
	configFile := flag.String("config", "sequences.yaml", "path to sequences YAML file")
	noFormat := flag.Bool("no-format", false, "skip gofmt formatting (for debugging)")
	flag.Parse()

	if err := run(*configFile, *noFormat); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(configFile string, noFormat bool) error {
	// Read the YAML spec
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// Parse the spec
	var spec gen.Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	// Generate the code to a buffer
	var buf bytes.Buffer
	generator := gen.NewGenerator(&spec, &buf)
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// If no-format flag is set, output raw code
	if noFormat {
		_, err = os.Stdout.Write(buf.Bytes())
		return err
	}

	// Format the generated code with gofmt
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting code: %w", err)
	}

	// Write the formatted code to stdout
	_, err = os.Stdout.Write(formatted)
	return err
}
