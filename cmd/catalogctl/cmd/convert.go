package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/plexusone/productgraph/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	outputFile   string
	sortOutput   bool
	formatType   string
	templateFile string
)

var convertCmd = &cobra.Command{
	Use:   "convert <input.json>",
	Short: "Convert a catalog to Markdown or other formats",
	Long: `Convert a CapabilityCatalog JSON file to Pandoc-compatible Markdown.

The output is designed for deterministic PDF generation via Pandoc:

  pandoc output.md -o output.pdf --pdf-engine=xelatex

Examples:
  # Convert to Markdown file
  catalogctl convert catalog.json -o catalog.md

  # Output to stdout
  catalogctl convert catalog.json

  # Convert with sorting disabled (preserve original order)
  catalogctl convert catalog.json --no-sort

  # Use a custom template
  catalogctl convert catalog.json --template custom.tmpl`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	convertCmd.Flags().BoolVar(&sortOutput, "sort", true, "Sort capabilities and features alphabetically")
	convertCmd.Flags().StringVarP(&formatType, "format", "f", "markdown", "Output format: markdown")
	convertCmd.Flags().StringVarP(&templateFile, "template", "t", "", "Custom Go template file")
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Read input JSON
	data, err := os.ReadFile(inputFile) //nolint:gosec // path is from args
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	var catalog schema.CapabilityCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	// Sort if enabled
	if sortOutput {
		sortCatalog(&catalog)
	}

	// Determine output destination
	var out io.Writer = cmd.OutOrStdout()
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	// Load template
	tmplContent := defaultMarkdownTemplate
	if templateFile != "" {
		customTmpl, err := os.ReadFile(templateFile) //nolint:gosec // path is from args
		if err != nil {
			return fmt.Errorf("reading template file: %w", err)
		}
		tmplContent = string(customTmpl)
	}

	// Generate output
	if err := renderMarkdown(out, &catalog, tmplContent); err != nil {
		return fmt.Errorf("generating output: %w", err)
	}

	if outputFile != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Wrote %s\n", outputFile)
	}

	return nil
}

// sortCatalog sorts capabilities and features alphabetically for deterministic output.
func sortCatalog(catalog *schema.CapabilityCatalog) {
	sort.Slice(catalog.Capabilities, func(i, j int) bool {
		return catalog.Capabilities[i].Name < catalog.Capabilities[j].Name
	})

	for i := range catalog.Capabilities {
		sort.Slice(catalog.Capabilities[i].Features, func(a, b int) bool {
			return catalog.Capabilities[i].Features[a].Name < catalog.Capabilities[i].Features[b].Name
		})

		for j := range catalog.Capabilities[i].Features {
			sort.Slice(catalog.Capabilities[i].Features[j].Documentation, func(a, b int) bool {
				return catalog.Capabilities[i].Features[j].Documentation[a].Label <
					catalog.Capabilities[i].Features[j].Documentation[b].Label
			})
		}
	}
}

// renderMarkdown generates Pandoc-compatible Markdown from a catalog.
func renderMarkdown(w io.Writer, catalog *schema.CapabilityCatalog, tmplContent string) error {
	funcMap := template.FuncMap{
		"trim": strings.TrimSpace,
	}

	tmpl, err := template.New("catalog").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return tmpl.Execute(w, catalog)
}

const defaultMarkdownTemplate = `---
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
toc-depth: 2
numbersections: true
colorlinks: true
linkcolor: blue
urlcolor: blue
geometry: margin=2cm
mainfont: Lato
sansfont: Lato
monofont: Lato
header-includes:
  - \usepackage{fancyhdr}
  - \pagestyle{fancy}
  - \fancyhead[L]{\leftmark}
  - \fancyhead[R]{\thepage}
  - \setcounter{tocdepth}{2}
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
**Documentation:**

{{range $feat.Documentation}}
- [{{.Label}}]({{.URL}})
{{end}}
{{end}}
{{end}}
{{end}}
`
