// Package cmd implements the catalogctl CLI commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "catalogctl",
	Short: "Manage product capability catalogs",
	Long: `catalogctl is a CLI for managing product capability catalogs.

It provides commands to convert, validate, and generate documentation
from CapabilityCatalog JSON files.

Examples:
  # Convert a catalog to Markdown
  catalogctl convert catalog.json -o catalog.md

  # Convert and output to stdout
  catalogctl convert catalog.json

  # Validate a catalog against the schema
  catalogctl validate catalog.json`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}
