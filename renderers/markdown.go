package renderers

import (
	"fmt"
	"strings"

	"github.com/hilli/cooklang"
)

// MarkdownRenderer renders recipes in Markdown format
type MarkdownRenderer struct{}

func (mr MarkdownRenderer) RenderRecipe(recipe *cooklang.Recipe) string {
	var result strings.Builder

	// Title
	if recipe.Title != "" {
		result.WriteString(fmt.Sprintf("# %s\n\n", recipe.Title))
	}

	// Metadata section
	if recipe.Description != "" || recipe.Cuisine != "" || recipe.Difficulty != "" ||
		recipe.PrepTime != "" || recipe.TotalTime != "" || recipe.Author != "" ||
		recipe.Servings > 0 || len(recipe.Tags) > 0 {
		result.WriteString("## Recipe Information\n\n")

		if recipe.Description != "" {
			result.WriteString(fmt.Sprintf("**Description:** %s\n\n", recipe.Description))
		}
		if recipe.Cuisine != "" {
			result.WriteString(fmt.Sprintf("**Cuisine:** %s\n\n", recipe.Cuisine))
		}
		if recipe.Difficulty != "" {
			result.WriteString(fmt.Sprintf("**Difficulty:** %s\n\n", recipe.Difficulty))
		}
		if recipe.PrepTime != "" {
			result.WriteString(fmt.Sprintf("**Prep Time:** %s\n\n", recipe.PrepTime))
		}
		if recipe.TotalTime != "" {
			result.WriteString(fmt.Sprintf("**Total Time:** %s\n\n", recipe.TotalTime))
		}
		if recipe.Author != "" {
			result.WriteString(fmt.Sprintf("**Author:** %s\n\n", recipe.Author))
		}
		if recipe.Servings > 0 {
			result.WriteString(fmt.Sprintf("**Servings:** %g\n\n", recipe.Servings))
		}
		if len(recipe.Tags) > 0 {
			result.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(recipe.Tags, ", ")))
		}
	}

	// Instructions
	result.WriteString("## Instructions\n\n")

	stepNum := 1
	currentStep := recipe.FirstStep
	for currentStep != nil {
		result.WriteString(fmt.Sprintf("%d. ", stepNum))

		// Render components in markdown-friendly format
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			switch comp := currentComponent.(type) {
			case *cooklang.Ingredient:
				if comp.Quantity > 0 {
					result.WriteString(fmt.Sprintf("**%s** (%g %s)", comp.Name, comp.Quantity, comp.Unit))
				} else {
					result.WriteString(fmt.Sprintf("**%s**", comp.Name))
				}
			case *cooklang.Cookware:
				if comp.Quantity > 1 {
					result.WriteString(fmt.Sprintf("*%s* (x%d)", comp.Name, comp.Quantity))
				} else {
					result.WriteString(fmt.Sprintf("*%s*", comp.Name))
				}
			case *cooklang.Timer:
				if comp.Name != "" {
					result.WriteString(fmt.Sprintf("⏲️ %s (%s)", comp.Name, comp.Duration))
				} else {
					result.WriteString(fmt.Sprintf("⏲️ %s", comp.Duration))
				}
			case *cooklang.Instruction:
				result.WriteString(comp.Text)
			}
			currentComponent = currentComponent.GetNext()
		}

		result.WriteString("\n\n")
		currentStep = currentStep.NextStep
		stepNum++
	}

	return result.String()
}

// DefaultMarkdownRenderer is the default instance of MarkdownRenderer
var DefaultMarkdownRenderer = MarkdownRenderer{}
