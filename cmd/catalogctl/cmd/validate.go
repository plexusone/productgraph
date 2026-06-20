package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plexusone/productgraph/pkg/schema"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <input.json>",
	Short: "Validate a catalog JSON file",
	Long: `Validate a CapabilityCatalog JSON file for structural correctness.

Checks:
  - Valid JSON syntax
  - Required fields (metadata.title, capability.id, capability.name, etc.)
  - No duplicate capability or feature IDs

Examples:
  catalogctl validate catalog.json`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	data, err := os.ReadFile(inputFile) //nolint:gosec // path is from args
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	var catalog schema.CapabilityCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	// Validate structure
	errors := validateCatalog(&catalog)
	if len(errors) > 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "Validation errors:")
		for _, e := range errors {
			fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
		}
		return fmt.Errorf("validation failed with %d error(s)", len(errors))
	}

	// Print summary
	featureCount := 0
	docCount := 0
	for _, cap := range catalog.Capabilities {
		featureCount += len(cap.Features)
		for _, feat := range cap.Features {
			docCount += len(feat.Documentation)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Valid: %s\n", inputFile)
	fmt.Fprintf(cmd.OutOrStdout(), "  Title:        %s\n", catalog.Metadata.Title)
	fmt.Fprintf(cmd.OutOrStdout(), "  Version:      %s\n", catalog.Metadata.Version)
	fmt.Fprintf(cmd.OutOrStdout(), "  Capabilities: %d\n", len(catalog.Capabilities))
	fmt.Fprintf(cmd.OutOrStdout(), "  Features:     %d\n", featureCount)
	fmt.Fprintf(cmd.OutOrStdout(), "  Doc links:    %d\n", docCount)

	return nil
}

func validateCatalog(catalog *schema.CapabilityCatalog) []string {
	var errors []string

	// Check metadata
	if catalog.Metadata.Title == "" {
		errors = append(errors, "metadata.title is required")
	}

	// Track IDs for duplicates
	capIDs := make(map[string]bool)
	featIDs := make(map[string]bool)

	for i, cap := range catalog.Capabilities {
		prefix := fmt.Sprintf("capabilities[%d]", i)

		if cap.ID == "" {
			errors = append(errors, fmt.Sprintf("%s.id is required", prefix))
		} else if capIDs[cap.ID] {
			errors = append(errors, fmt.Sprintf("%s.id '%s' is duplicate", prefix, cap.ID))
		} else {
			capIDs[cap.ID] = true
		}

		if cap.Name == "" {
			errors = append(errors, fmt.Sprintf("%s.name is required", prefix))
		}

		for j, feat := range cap.Features {
			featPrefix := fmt.Sprintf("%s.features[%d]", prefix, j)

			if feat.ID == "" {
				errors = append(errors, fmt.Sprintf("%s.id is required", featPrefix))
			} else if featIDs[feat.ID] {
				errors = append(errors, fmt.Sprintf("%s.id '%s' is duplicate", featPrefix, feat.ID))
			} else {
				featIDs[feat.ID] = true
			}

			if feat.Name == "" {
				errors = append(errors, fmt.Sprintf("%s.name is required", featPrefix))
			}

			for k, doc := range feat.Documentation {
				docPrefix := fmt.Sprintf("%s.documentation[%d]", featPrefix, k)

				if doc.Label == "" {
					errors = append(errors, fmt.Sprintf("%s.label is required", docPrefix))
				}
				if doc.URL == "" {
					errors = append(errors, fmt.Sprintf("%s.url is required", docPrefix))
				}
			}
		}
	}

	return errors
}
