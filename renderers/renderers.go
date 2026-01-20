// Package renderers provides different renderers for Cooklang recipes.
//
// This package includes renderers for various output formats:
//   - CooklangRenderer: Renders recipes back to Cooklang format
//   - MarkdownRenderer: Renders recipes as Markdown
//   - HTMLRenderer: Renders recipes as HTML
//   - PrintRenderer: Renders recipes as print-optimized HTML
//   - JSONLDRenderer: Renders recipes as Schema.org JSON-LD for SEO
//
// Example usage:
//
//	recipe, _ := cooklang.ParseFile("recipe.cook")
//
//	// Use the default renderers
//	html := renderers.Default.HTML.RenderRecipe(recipe)
//	markdown := renderers.Default.Markdown.RenderRecipe(recipe)
//
//	// For JSON-LD (SEO structured data)
//	jsonLD, _ := renderers.Default.JSONLD.RenderRecipeJSON(recipe, nil)
package renderers

import "github.com/hilli/cooklang"

// All default renderer instances for convenience
var (
	// Default renderers that can be used directly
	Default = struct {
		Cooklang CooklangRenderer
		Markdown MarkdownRenderer
		HTML     HTMLRenderer
		Print    PrintRenderer
		JSONLD   JSONLDRenderer
	}{
		Cooklang: CooklangRenderer{},
		Markdown: MarkdownRenderer{},
		HTML:     HTMLRenderer{},
		Print:    PrintRenderer{},
		JSONLD:   JSONLDRenderer{},
	}
)

// NewCooklangRenderer creates a new Cooklang renderer
func NewCooklangRenderer() cooklang.RecipeRenderer {
	return CooklangRenderer{}
}

// NewMarkdownRenderer creates a new Markdown renderer
func NewMarkdownRenderer() cooklang.RecipeRenderer {
	return MarkdownRenderer{}
}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() cooklang.RecipeRenderer {
	return HTMLRenderer{}
}

// NewPrintRenderer creates a new print-optimized HTML renderer
func NewPrintRenderer() cooklang.RecipeRenderer {
	return PrintRenderer{}
}
