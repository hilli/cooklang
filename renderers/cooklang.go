package renderers

import (
	"fmt"
	"strings"

	"github.com/hilli/cooklang"
)

// CooklangRenderer renders recipes in the original Cooklang format
type CooklangRenderer struct{}

func (cr CooklangRenderer) RenderRecipe(recipe *cooklang.Recipe) string {
	var result strings.Builder

	// Render metadata
	if recipe.Title != "" {
		result.WriteString(fmt.Sprintf(">> title: %s\n", recipe.Title))
	}
	if recipe.Cuisine != "" {
		result.WriteString(fmt.Sprintf(">> cuisine: %s\n", recipe.Cuisine))
	}
	if !recipe.Date.IsZero() {
		result.WriteString(fmt.Sprintf(">> date: %s\n", recipe.Date.Format("2006-01-02")))
	}
	if recipe.Description != "" {
		result.WriteString(fmt.Sprintf(">> description: %s\n", recipe.Description))
	}
	if recipe.Difficulty != "" {
		result.WriteString(fmt.Sprintf(">> difficulty: %s\n", recipe.Difficulty))
	}
	if recipe.PrepTime != "" {
		result.WriteString(fmt.Sprintf(">> prep_time: %s\n", recipe.PrepTime))
	}
	if recipe.TotalTime != "" {
		result.WriteString(fmt.Sprintf(">> total_time: %s\n", recipe.TotalTime))
	}
	if recipe.Author != "" {
		result.WriteString(fmt.Sprintf(">> author: %s\n", recipe.Author))
	}
	if recipe.Servings > 0 {
		result.WriteString(fmt.Sprintf(">> servings: %g\n", recipe.Servings))
	}
	if len(recipe.Tags) > 0 {
		result.WriteString(fmt.Sprintf(">> tags: %s\n", strings.Join(recipe.Tags, ", ")))
	}
	if len(recipe.Images) > 0 {
		result.WriteString(fmt.Sprintf(">> images: %s\n", strings.Join(recipe.Images, ", ")))
	}

	// Add blank line after metadata
	if result.Len() > 0 {
		result.WriteString("\n")
	}

	// Render steps
	currentStep := recipe.FirstStep
	for currentStep != nil {
		// Iterate through components in this step
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			result.WriteString(currentComponent.Render())
			currentComponent = currentComponent.GetNext()
		}

		// Add newline after each step
		result.WriteString("\n")

		// Move to next step
		if currentStep.NextStep != nil {
			result.WriteString("\n") // Extra newline between steps
		}
		currentStep = currentStep.NextStep
	}

	return result.String()
}

// DefaultCooklangRenderer is the default instance of CooklangRenderer
var DefaultCooklangRenderer = CooklangRenderer{}
