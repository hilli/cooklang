package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/hilli/cooklang"
)

// PrintRenderer renders recipes as print-optimized HTML designed to fit on a single page.
// It includes embedded CSS for clean printing without browser chrome or interactive elements.
type PrintRenderer struct{}

// printCSS contains embedded CSS optimized for single-page recipe printing
const printCSS = `
<style>
  @page {
    size: A4;
    margin: 1.5cm;
  }

  * {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  body {
    font-family: Georgia, 'Times New Roman', serif;
    font-size: 11pt;
    line-height: 1.4;
    color: #222;
    max-width: 100%;
  }

  .recipe-print {
    max-width: 100%;
  }

  .recipe-header {
    border-bottom: 2px solid #333;
    padding-bottom: 0.5em;
    margin-bottom: 0.75em;
    overflow: hidden;
  }

  .recipe-header-content {
    overflow: hidden;
  }

  .recipe-image {
    float: right;
    max-width: 150px;
    max-height: 150px;
    width: auto;
    height: auto;
    margin-left: 1em;
    margin-bottom: 0.5em;
    border-radius: 4px;
    border: 1px solid #ddd;
  }

  .recipe-title {
    font-size: 20pt;
    font-weight: bold;
    margin: 0 0 0.25em 0;
    color: #111;
  }

  .recipe-description {
    font-style: italic;
    color: #444;
    margin-bottom: 0.5em;
  }

  .recipe-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5em 1.5em;
    font-size: 9pt;
    color: #555;
  }

  .recipe-meta-item {
    display: inline;
  }

  .recipe-meta-label {
    font-weight: bold;
  }

  .recipe-body {
    display: flex;
    gap: 1.5em;
  }

  .recipe-ingredients {
    flex: 0 0 35%;
    max-width: 35%;
  }

  .recipe-instructions {
    flex: 1;
  }

  h2 {
    font-size: 12pt;
    font-weight: bold;
    border-bottom: 1px solid #999;
    padding-bottom: 0.25em;
    margin-bottom: 0.5em;
    color: #333;
  }

  .ingredients-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .ingredients-list li {
    padding: 0.2em 0;
    border-bottom: 1px dotted #ccc;
  }

  .ingredients-list li:last-child {
    border-bottom: none;
  }

  .ingredient-qty {
    font-weight: bold;
    display: inline-block;
    min-width: 4em;
  }

  .ingredient-name {
    color: #222;
  }

  .instructions-list {
    list-style: none;
    padding: 0;
    margin: 0;
    counter-reset: step-counter;
  }

  .instructions-list li {
    padding: 0.4em 0 0.4em 2em;
    position: relative;
    text-align: justify;
  }

  .instructions-list li::before {
    counter-increment: step-counter;
    content: counter(step-counter);
    position: absolute;
    left: 0;
    top: 0.4em;
    font-weight: bold;
    font-size: 11pt;
    color: #555;
    background: #f0f0f0;
    width: 1.5em;
    height: 1.5em;
    border-radius: 50%;
    text-align: center;
    line-height: 1.5em;
  }

  .ing {
    font-weight: bold;
  }

  .qty {
    color: #555;
  }

  .cw {
    font-style: italic;
  }

  .tmr {
    background: #f5f5f5;
    padding: 0.1em 0.3em;
    border-radius: 3px;
    font-family: 'Courier New', monospace;
    font-size: 10pt;
  }

  .optional {
    font-style: italic;
    color: #666;
  }

  .optional-marker {
    font-size: 9pt;
    color: #888;
    font-style: italic;
  }

  .recipe-footer {
    margin-top: 1em;
    padding-top: 0.5em;
    border-top: 1px solid #ccc;
    font-size: 8pt;
    color: #888;
    display: flex;
    justify-content: space-between;
  }

  .recipe-tags {
    font-style: italic;
  }

  /* Print-specific adjustments */
  @media print {
    body {
      print-color-adjust: exact;
      -webkit-print-color-adjust: exact;
    }

    .recipe-print {
      page-break-inside: avoid;
    }

    .recipe-image {
      print-color-adjust: exact;
      -webkit-print-color-adjust: exact;
    }

    .instructions-list li {
      page-break-inside: avoid;
    }
  }
</style>
`

func (pr PrintRenderer) RenderRecipe(recipe *cooklang.Recipe) string {
	var result strings.Builder

	// HTML document structure
	result.WriteString("<!DOCTYPE html>\n")
	result.WriteString("<html lang=\"en\">\n")
	result.WriteString("<head>\n")
	result.WriteString("  <meta charset=\"UTF-8\">\n")
	result.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	if recipe.Title != "" {
		result.WriteString(fmt.Sprintf("  <title>%s</title>\n", html.EscapeString(recipe.Title)))
	} else {
		result.WriteString("  <title>Recipe</title>\n")
	}
	result.WriteString(printCSS)
	result.WriteString("</head>\n")
	result.WriteString("<body>\n")
	result.WriteString("<div class=\"recipe-print\">\n")

	// Header section
	result.WriteString("  <div class=\"recipe-header\">\n")

	// Add image if available (use first image)
	if len(recipe.Images) > 0 {
		result.WriteString(fmt.Sprintf("    <img class=\"recipe-image\" src=\"%s\" alt=\"%s\">\n",
			html.EscapeString(recipe.Images[0]),
			html.EscapeString(recipe.Title)))
	}

	result.WriteString("    <div class=\"recipe-header-content\">\n")
	if recipe.Title != "" {
		result.WriteString(fmt.Sprintf("      <h1 class=\"recipe-title\">%s</h1>\n", html.EscapeString(recipe.Title)))
	}
	if recipe.Description != "" {
		result.WriteString(fmt.Sprintf("      <p class=\"recipe-description\">%s</p>\n", html.EscapeString(recipe.Description)))
	}

	// Metadata line
	var metaItems []string
	if recipe.Servings > 0 {
		metaItems = append(metaItems, fmt.Sprintf("<span class=\"recipe-meta-item\"><span class=\"recipe-meta-label\">Servings:</span> %g</span>", recipe.Servings))
	}
	if recipe.PrepTime != "" {
		metaItems = append(metaItems, fmt.Sprintf("<span class=\"recipe-meta-item\"><span class=\"recipe-meta-label\">Prep:</span> %s</span>", html.EscapeString(recipe.PrepTime)))
	}
	if recipe.TotalTime != "" {
		metaItems = append(metaItems, fmt.Sprintf("<span class=\"recipe-meta-item\"><span class=\"recipe-meta-label\">Total:</span> %s</span>", html.EscapeString(recipe.TotalTime)))
	}
	if recipe.Difficulty != "" {
		metaItems = append(metaItems, fmt.Sprintf("<span class=\"recipe-meta-item\"><span class=\"recipe-meta-label\">Difficulty:</span> %s</span>", html.EscapeString(recipe.Difficulty)))
	}
	if recipe.Cuisine != "" {
		metaItems = append(metaItems, fmt.Sprintf("<span class=\"recipe-meta-item\"><span class=\"recipe-meta-label\">Cuisine:</span> %s</span>", html.EscapeString(recipe.Cuisine)))
	}
	if recipe.Author != "" {
		metaItems = append(metaItems, fmt.Sprintf("<span class=\"recipe-meta-item\"><span class=\"recipe-meta-label\">By:</span> %s</span>", html.EscapeString(recipe.Author)))
	}

	if len(metaItems) > 0 {
		result.WriteString("      <div class=\"recipe-meta\">\n")
		result.WriteString("        " + strings.Join(metaItems, "\n        ") + "\n")
		result.WriteString("      </div>\n")
	}
	result.WriteString("    </div>\n")
	result.WriteString("  </div>\n\n")

	// Body with two columns
	result.WriteString("  <div class=\"recipe-body\">\n")

	// Ingredients column
	ingredients := recipe.GetIngredients()
	result.WriteString("    <div class=\"recipe-ingredients\">\n")
	result.WriteString("      <h2>Ingredients</h2>\n")
	if len(ingredients.Ingredients) > 0 {
		result.WriteString("      <ul class=\"ingredients-list\">\n")
		for _, ingredient := range ingredients.Ingredients {
			optionalClass := ""
			if ingredient.Optional {
				optionalClass = " optional"
			}
			result.WriteString(fmt.Sprintf("        <li class=\"%s\">", optionalClass))
			qtyStr := pr.formatQuantity(ingredient.Quantity, ingredient.Unit)
			if qtyStr != "" {
				result.WriteString(fmt.Sprintf("<span class=\"ingredient-qty\">%s</span> ", qtyStr))
			}
			result.WriteString(fmt.Sprintf("<span class=\"ingredient-name\">%s</span>", html.EscapeString(ingredient.Name)))
			if ingredient.Optional {
				result.WriteString(" <span class=\"optional-marker\">(optional)</span>")
			}
			result.WriteString("</li>\n")
		}
		result.WriteString("      </ul>\n")
	}
	result.WriteString("    </div>\n\n")

	// Instructions column
	result.WriteString("    <div class=\"recipe-instructions\">\n")
	result.WriteString("      <h2>Instructions</h2>\n")
	result.WriteString("      <ol class=\"instructions-list\">\n")

	currentStep := recipe.FirstStep
	for currentStep != nil {
		result.WriteString("        <li>")
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			switch comp := currentComponent.(type) {
			case *cooklang.Ingredient:
				optionalClass := ""
				if comp.Optional {
					optionalClass = " optional"
				}
				result.WriteString(fmt.Sprintf("<span class=\"ing%s\">%s</span>", optionalClass, html.EscapeString(comp.Name)))
				if comp.Quantity > 0 {
					qtyStr := pr.formatQuantity(comp.Quantity, comp.Unit)
					result.WriteString(fmt.Sprintf(" <span class=\"qty\">(%s)</span>", qtyStr))
				}
				if comp.Optional {
					result.WriteString(" <span class=\"optional-marker\">(optional)</span>")
				}
			case *cooklang.Cookware:
				result.WriteString(fmt.Sprintf("<span class=\"cw\">%s</span>", html.EscapeString(comp.Name)))
			case *cooklang.Timer:
				if comp.Name != "" {
					result.WriteString(fmt.Sprintf("<span class=\"tmr\">%s: %s</span>", html.EscapeString(comp.Name), html.EscapeString(comp.Duration)))
				} else {
					result.WriteString(fmt.Sprintf("<span class=\"tmr\">%s</span>", html.EscapeString(comp.Duration)))
				}
			case *cooklang.Instruction:
				result.WriteString(html.EscapeString(comp.Text))
			}
			currentComponent = currentComponent.GetNext()
		}
		result.WriteString("</li>\n")
		currentStep = currentStep.NextStep
	}

	result.WriteString("      </ol>\n")
	result.WriteString("    </div>\n")
	result.WriteString("  </div>\n\n")

	// Footer with tags and date
	var footerLeft, footerRight string
	if len(recipe.Tags) > 0 {
		footerLeft = fmt.Sprintf("<span class=\"recipe-tags\">Tags: %s</span>", html.EscapeString(strings.Join(recipe.Tags, ", ")))
	}
	if !recipe.Date.IsZero() {
		footerRight = recipe.Date.Format("2006-01-02")
	}

	if footerLeft != "" || footerRight != "" {
		result.WriteString("  <div class=\"recipe-footer\">\n")
		result.WriteString(fmt.Sprintf("    <span>%s</span>\n", footerLeft))
		result.WriteString(fmt.Sprintf("    <span>%s</span>\n", footerRight))
		result.WriteString("  </div>\n")
	}

	result.WriteString("</div>\n")
	result.WriteString("</body>\n")
	result.WriteString("</html>\n")

	return result.String()
}

// formatQuantity formats a quantity and unit for display
func (pr PrintRenderer) formatQuantity(qty float32, unit string) string {
	if qty <= 0 {
		if qty == -1 {
			if unit != "" {
				return fmt.Sprintf("some %s", unit)
			}
			return "some"
		}
		return ""
	}

	// Format quantity nicely (avoid .0 for whole numbers)
	var qtyStr string
	if qty == float32(int(qty)) {
		qtyStr = fmt.Sprintf("%d", int(qty))
	} else {
		qtyStr = fmt.Sprintf("%g", qty)
	}

	if unit != "" {
		return fmt.Sprintf("%s %s", qtyStr, unit)
	}
	return qtyStr
}

// DefaultPrintRenderer is the default instance of PrintRenderer
var DefaultPrintRenderer = PrintRenderer{}
