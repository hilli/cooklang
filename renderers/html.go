package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/hilli/cooklang"
)

// HTMLRenderer renders recipes in HTML format
type HTMLRenderer struct{}

func (hr HTMLRenderer) RenderRecipe(recipe *cooklang.Recipe) string {
	var result strings.Builder

	result.WriteString("<div class=\"recipe\">\n")

	// Title
	if recipe.Title != "" {
		result.WriteString(fmt.Sprintf("  <h1 class=\"recipe-title\">%s</h1>\n", html.EscapeString(recipe.Title)))
	}

	// Metadata section

	result.WriteString("  <div class=\"recipe-info\">\n")
	result.WriteString("    <h2>Recipe Information</h2>\n")
	result.WriteString("    <dl>\n")

	if recipe.Description != "" {
		result.WriteString(fmt.Sprintf("      <dt>Description</dt><dd>%s</dd>\n", html.EscapeString(recipe.Description)))
	}
	if recipe.Cuisine != "" {
		result.WriteString(fmt.Sprintf("      <dt>Cuisine</dt><dd>%s</dd>\n", html.EscapeString(recipe.Cuisine)))
	}
	if recipe.Difficulty != "" {
		result.WriteString(fmt.Sprintf("      <dt>Difficulty</dt><dd>%s</dd>\n", html.EscapeString(recipe.Difficulty)))
	}
	if recipe.PrepTime != "" {
		result.WriteString(fmt.Sprintf("      <dt>Prep Time</dt><dd>%s</dd>\n", html.EscapeString(recipe.PrepTime)))
	}
	if recipe.TotalTime != "" {
		result.WriteString(fmt.Sprintf("      <dt>Total Time</dt><dd>%s</dd>\n", html.EscapeString(recipe.TotalTime)))
	}
	if recipe.Author != "" {
		result.WriteString(fmt.Sprintf("      <dt>Author</dt><dd>%s</dd>\n", html.EscapeString(recipe.Author)))
	}
	if recipe.Servings > 0 {
		result.WriteString(fmt.Sprintf("      <dt>Servings</dt><dd>%g</dd>\n", recipe.Servings))
	}
	if len(recipe.Tags) > 0 {
		result.WriteString(fmt.Sprintf("      <dt>Tags</dt><dd>%s</dd>\n", html.EscapeString(strings.Join(recipe.Tags, ", "))))
	}

	result.WriteString("    </dl>\n")
	result.WriteString("  </div>\n")

	// Ingredients list
	ingredients := recipe.GetIngredients()
	if len(ingredients.Ingredients) > 0 {
		result.WriteString("  <div class=\"recipe-ingredients\">\n")
		result.WriteString("    <h2>Ingredients</h2>\n")
		result.WriteString("    <ul>\n")

		for _, ingredient := range ingredients.Ingredients {
			result.WriteString("      <li>")
			if ingredient.Quantity > 0 {
				if ingredient.Unit != "" {
					result.WriteString(fmt.Sprintf("<span class=\"quantity\">%g %s</span> <span class=\"ingredient\">%s</span>",
						ingredient.Quantity, html.EscapeString(ingredient.Unit), html.EscapeString(ingredient.Name)))
				} else {
					result.WriteString(fmt.Sprintf("<span class=\"quantity\">%g</span> <span class=\"ingredient\">%s</span>",
						ingredient.Quantity, html.EscapeString(ingredient.Name)))
				}
			} else if ingredient.Quantity == -1 {
				// "some" quantity
				if ingredient.Unit != "" {
					result.WriteString(fmt.Sprintf("<span class=\"quantity\">some %s</span> <span class=\"ingredient\">%s</span>",
						html.EscapeString(ingredient.Unit), html.EscapeString(ingredient.Name)))
				} else {
					result.WriteString(fmt.Sprintf("<span class=\"quantity\">some</span> <span class=\"ingredient\">%s</span>",
						html.EscapeString(ingredient.Name)))
				}
			} else {
				result.WriteString(fmt.Sprintf("<span class=\"ingredient\">%s</span>", html.EscapeString(ingredient.Name)))
			}
			result.WriteString("</li>\n")
		}

		result.WriteString("    </ul>\n")
		result.WriteString("  </div>\n")
	}

	// Instructions
	result.WriteString("  <div class=\"recipe-instructions\">\n")
	result.WriteString("    <h2>Instructions</h2>\n")
	result.WriteString("    <ol>\n")

	stepNum := 1
	currentStep := recipe.FirstStep
	for currentStep != nil {
		result.WriteString("      <li class=\"recipe-step\">\n        ")

		// Render components in HTML format
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			switch comp := currentComponent.(type) {
			case *cooklang.Ingredient:
				if comp.Quantity > 0 {
					result.WriteString(fmt.Sprintf("<span class=\"ingredient\">%s</span> <span class=\"quantity\">(%g %s)</span>",
						html.EscapeString(comp.Name), comp.Quantity, html.EscapeString(comp.Unit)))
				} else {
					result.WriteString(fmt.Sprintf("<span class=\"ingredient\">%s</span>", html.EscapeString(comp.Name)))
				}
				if comp.Annotation != "" {
					result.WriteString(fmt.Sprintf(" <span class=\"annotation\">(%s)</span>", html.EscapeString(comp.Annotation)))
				}
			case *cooklang.Cookware:
				if comp.Quantity > 1 {
					result.WriteString(fmt.Sprintf("<span class=\"cookware\">%s</span> <span class=\"quantity\">(x%d)</span>",
						html.EscapeString(comp.Name), comp.Quantity))
				} else {
					result.WriteString(fmt.Sprintf("<span class=\"cookware\">%s</span>", html.EscapeString(comp.Name)))
				}
				if comp.Annotation != "" {
					result.WriteString(fmt.Sprintf(" <span class=\"annotation\">(%s)</span>", html.EscapeString(comp.Annotation)))
				}
			case *cooklang.Timer:
				if comp.Name != "" {
					result.WriteString(fmt.Sprintf("<span class=\"timer\">⏲️ %s (%s)</span>",
						html.EscapeString(comp.Name), html.EscapeString(comp.Duration)))
				} else {
					result.WriteString(fmt.Sprintf("<span class=\"timer\">⏲️ %s</span>", html.EscapeString(comp.Duration)))
				}
				if comp.Annotation != "" {
					result.WriteString(fmt.Sprintf(" <span class=\"annotation\">(%s)</span>", html.EscapeString(comp.Annotation)))
				}
			case *cooklang.Instruction:
				result.WriteString(html.EscapeString(comp.Text))
			}
			currentComponent = currentComponent.GetNext()
		}

		result.WriteString("\n      </li>\n")
		currentStep = currentStep.NextStep
		stepNum++
	}

	result.WriteString("    </ol>\n")
	result.WriteString("  </div>\n")
	result.WriteString("</div>\n")

	return result.String()
}

// DefaultHTMLRenderer is the default instance of HTMLRenderer
var DefaultHTMLRenderer = HTMLRenderer{}
