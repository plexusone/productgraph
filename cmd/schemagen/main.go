// Package main generates JSON Schema files from Go types.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/plexusone/productgraph/pkg/schema"
)

func main() {
	// Find output directory (schema/ relative to module root)
	outputDir := "schema"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0750); err != nil { //nolint:gosec // outputDir is trusted from args
		fmt.Fprintf(os.Stderr, "failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate CapabilityCatalog schema
	if err := generateSchema(outputDir, "capability.schema.json", &schema.CapabilityCatalog{}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate capability schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Schema generation complete")
}

func generateSchema(outputDir, filename string, v any) error {
	r := jsonschema.Reflector{
		DoNotReference: false,
	}

	s := r.Reflect(v)
	s.ID = jsonschema.ID("https://productgraph.io/schema/" + filename)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}

	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, data, 0600); err != nil { //nolint:gosec // path is trusted
		return fmt.Errorf("write schema: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}
