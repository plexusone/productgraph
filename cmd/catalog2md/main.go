// Package main converts a CapabilityCatalog JSON file to Pandoc Markdown.
//
// Usage:
//
//	catalog2md input.json > output.md
//	catalog2md input.json -o output.md
//
// The output is designed for deterministic PDF generation via Pandoc:
//
//	pandoc output.md -o output.pdf --pdf-engine=xelatex
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/plexusone/productgraph/pkg/schema"
)

const mdTemplate = `---
title: "{{.Metadata.Title}}"
{{- if .Metadata.Description}}
subtitle: "{{.Metadata.Description}}"
{{- end}}
{{- if .Metadata.Version}}
version: "{{.Metadata.Version}}"
{{- end}}
{{- if .Metadata.Generated}}
date: "{{.Metadata.Generated}}"
{{- end}}
{{- if .Metadata.Source}}
source: "{{.Metadata.Source}}"
{{- end}}
documentclass: report
toc: true
toc-depth: 3
numbersections: true
colorlinks: true
linkcolor: blue
urlcolor: blue
geometry: margin=1in
header-includes:
  - \usepackage{fancyhdr}
  - \pagestyle{fancy}
  - \fancyhead[L]{\leftmark}
  - \fancyhead[R]{\thepage}
---

{{range $capIdx, $cap := .Capabilities}}
# {{$cap.Name}}

{{if $cap.Description}}{{$cap.Description}}{{end}}

{{range $featIdx, $feat := $cap.Features}}
## {{$feat.Name}}

{{if $feat.Description}}{{$feat.Description}}{{end}}

{{if $feat.AdminPath}}**Admin Path:** ` + "`{{$feat.AdminPath}}`" + `{{end}}

{{if $feat.Notes}}> **Note:** {{$feat.Notes}}{{end}}

{{if $feat.Documentation}}
### Documentation

{{range $feat.Documentation}}
- [{{.Label}}]({{.URL}})
{{end}}
{{end}}
{{end}}
{{end}}
`

func main() {
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: catalog2md <input.json> [-o output.md]")
		os.Exit(1)
	}

	inputFile := flag.Arg(0)

	// Read input JSON
	data, err := os.ReadFile(inputFile) //nolint:gosec // path is from args
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var catalog schema.CapabilityCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Sort capabilities and features for deterministic output
	sortCatalog(&catalog)

	// Determine output destination
	var out io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	// Generate Markdown
	if err := renderMarkdown(out, &catalog); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating markdown: %v\n", err)
		os.Exit(1)
	}
}

// sortCatalog sorts capabilities and features alphabetically for deterministic output.
func sortCatalog(catalog *schema.CapabilityCatalog) {
	// Sort capabilities by name
	sort.Slice(catalog.Capabilities, func(i, j int) bool {
		return catalog.Capabilities[i].Name < catalog.Capabilities[j].Name
	})

	// Sort features within each capability
	for i := range catalog.Capabilities {
		sort.Slice(catalog.Capabilities[i].Features, func(a, b int) bool {
			return catalog.Capabilities[i].Features[a].Name < catalog.Capabilities[i].Features[b].Name
		})

		// Sort documentation links within each feature
		for j := range catalog.Capabilities[i].Features {
			sort.Slice(catalog.Capabilities[i].Features[j].Documentation, func(a, b int) bool {
				return catalog.Capabilities[i].Features[j].Documentation[a].Label <
					catalog.Capabilities[i].Features[j].Documentation[b].Label
			})
		}
	}
}

// renderMarkdown generates Pandoc-compatible Markdown from a catalog.
func renderMarkdown(w io.Writer, catalog *schema.CapabilityCatalog) error {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"trim": strings.TrimSpace,
	}

	tmpl, err := template.New("catalog").Funcs(funcMap).Parse(mdTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return tmpl.Execute(w, catalog)
}
