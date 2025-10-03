package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilli/cooklang/renderers"
	"github.com/spf13/cobra"
)

var (
	renderFormat string
	renderOutput string
)

var renderCmd = &cobra.Command{
	Use:   "render <recipe-file>",
	Short: "Render a recipe in different formats",
	Long: `Render a Cooklang recipe in various output formats.

Supported formats:
  • cooklang - Original Cooklang format
  • markdown - Markdown format (default)
  • html     - HTML format

Examples:
  cook render recipe.cook
  cook render recipe.cook --format=html
  cook render recipe.cook --format=markdown --output=recipe.md
  cook render recipe.cook -f html -o recipe.html`,
	Args: cobra.ExactArgs(1),
	RunE: runRender,
}

func init() {
	renderCmd.Flags().StringVarP(&renderFormat, "format", "f", "markdown", "Output format (cooklang, markdown, html)")
	renderCmd.Flags().StringVarP(&renderOutput, "output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(renderCmd)
}

func runRender(cmd *cobra.Command, args []string) error {
	filename := args[0]
	recipe, err := readRecipeFile(filename)
	if err != nil {
		return err
	}

	var output string

	switch strings.ToLower(renderFormat) {
	case "cooklang", "cook":
		renderer := renderers.NewCooklangRenderer()
		output = renderer.RenderRecipe(recipe)
	case "markdown", "md":
		renderer := renderers.NewMarkdownRenderer()
		output = renderer.RenderRecipe(recipe)
	case "html":
		renderer := renderers.NewHTMLRenderer()
		output = renderer.RenderRecipe(recipe)
	default:
		return fmt.Errorf("unsupported format: %s (supported: cooklang, markdown, html)", renderFormat)
	}

	// Output to file or stdout
	if renderOutput != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(renderOutput)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		if err := os.WriteFile(renderOutput, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printSuccess("Rendered to: %s", renderOutput)
	} else {
		fmt.Println(output)
	}

	return nil
}
