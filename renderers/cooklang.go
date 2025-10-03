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
	var metadata strings.Builder

	// Collect metadata
	if recipe.Title != "" {
		metadata.WriteString(fmt.Sprintf("title: %s\n", recipe.Title))
	}
	if recipe.Cuisine != "" {
		metadata.WriteString(fmt.Sprintf("cuisine: %s\n", recipe.Cuisine))
	}
	if !recipe.Date.IsZero() {
		metadata.WriteString(fmt.Sprintf("date: %s\n", recipe.Date.Format("2006-01-02")))
	}
	if recipe.Description != "" {
		metadata.WriteString(fmt.Sprintf("description: %s\n", recipe.Description))
	}
	if recipe.Difficulty != "" {
		metadata.WriteString(fmt.Sprintf("difficulty: %s\n", recipe.Difficulty))
	}
	if recipe.PrepTime != "" {
		metadata.WriteString(fmt.Sprintf("prep_time: %s\n", recipe.PrepTime))
	}
	if recipe.TotalTime != "" {
		metadata.WriteString(fmt.Sprintf("total_time: %s\n", recipe.TotalTime))
	}
	if recipe.Author != "" {
		metadata.WriteString(fmt.Sprintf("author: %s\n", recipe.Author))
	}
	if recipe.Servings > 0 {
		metadata.WriteString(fmt.Sprintf("servings: %g\n", recipe.Servings))
	}
	if len(recipe.Tags) > 0 {
		metadata.WriteString(fmt.Sprintf("tags: %s\n", strings.Join(recipe.Tags, ", ")))
	}
	if len(recipe.Images) > 0 {
		metadata.WriteString(fmt.Sprintf("images: %s\n", strings.Join(recipe.Images, ", ")))
	}

	// Render metadata in YAML frontmatter block if present
	if metadata.Len() > 0 {
		result.WriteString("---\n")
		result.WriteString(metadata.String())
		result.WriteString("---\n\n")
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
