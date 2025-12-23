// Package renderers provides different renderers for Cooklang recipes.
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
	}{
		Cooklang: CooklangRenderer{},
		Markdown: MarkdownRenderer{},
		HTML:     HTMLRenderer{},
		Print:    PrintRenderer{},
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
