// Package main is the entry point for catalogctl, a CLI for managing
// product capability catalogs.
package main

import (
	"os"

	"github.com/plexusone/productgraph/cmd/catalogctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
