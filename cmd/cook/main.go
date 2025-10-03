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
	
	// Global flags
	canonicalMode bool // When true, use canonical spec mode (no extended features)
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

By default, the parser uses extended mode which supports:
  • Multi-word timer names (~roast time{4%hours})
  • Ingredient annotations (@milk{1%l}(cold))
  • Cookware annotations (#pan{}(for frying))
  • Comments as a component type

Use --canonical to disable extended features and parse in strict canonical mode.

Visit https://cooklang.org for more information about the Cooklang format.`,
	Version: version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("cook version %s (commit: %s, built: %s)\n", version, commit, date))
	
	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&canonicalMode, "canonical", false, "Use canonical spec mode (disable extended features)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
