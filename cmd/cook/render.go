package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilli/cooklang"
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
  • print    - Print-optimized HTML (single page, embedded CSS)

Examples:
  cook render recipe.cook
  cook render recipe.cook --format=html
  cook render recipe.cook --format=print --output=recipe.html
  cook render recipe.cook --format=markdown --output=recipe.md
  cook render recipe.cook -f html -o recipe.html`,
	Args:              cobra.ExactArgs(1),
	RunE:              runRender,
	ValidArgsFunction: completeCookFiles,
}

func init() {
	renderCmd.Flags().StringVarP(&renderFormat, "format", "f", "markdown", "Output format (cooklang, markdown, html, print)")
	renderCmd.Flags().StringVarP(&renderOutput, "output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(renderCmd)

	// Register flag completion
	_ = renderCmd.RegisterFlagCompletionFunc("format", completeFormatFlag)
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
		output = wrapHTMLDocument(renderer.RenderRecipe(recipe), recipe)
	case "print":
		renderer := renderers.NewPrintRenderer()
		output = renderer.RenderRecipe(recipe)
	default:
		return fmt.Errorf("unsupported format: %s (supported: cooklang, markdown, html, print)", renderFormat)
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

// wrapHTMLDocument wraps an HTML fragment in a complete HTML document with proper charset
func wrapHTMLDocument(content string, recipe *cooklang.Recipe) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString("<html lang=\"en\">\n")
	sb.WriteString("<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n")
	sb.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	if recipe.Title != "" {
		sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", html.EscapeString(recipe.Title)))
	} else {
		sb.WriteString("  <title>Recipe</title>\n")
	}
	sb.WriteString("</head>\n")
	sb.WriteString("<body>\n")
	sb.WriteString(content)
	sb.WriteString("</body>\n")
	sb.WriteString("</html>\n")
	return sb.String()
}
