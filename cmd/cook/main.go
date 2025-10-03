package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "cook",
	Short: "A CLI tool for working with Cooklang recipes",
	Long: `Cook is a comprehensive CLI tool for parsing, rendering, and managing Cooklang recipes.

It supports:
  • Parsing and validating recipe files
  • Extracting and consolidating ingredients
  • Creating shopping lists from multiple recipes
  • Rendering recipes in various formats
  • Scaling recipes for different serving sizes
  • Converting units between measurement systems

Visit https://cooklang.org for more information about the Cooklang format.`,
	Version: version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("cook version %s (commit: %s, built: %s)\n", version, commit, date))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
