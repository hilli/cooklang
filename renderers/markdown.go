package renderers

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/hilli/cooklang"
)

// simpleTitle capitalizes the first letter of each word in a string
func simpleTitle(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

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
		recipe.Servings > 0 || len(recipe.Tags) > 0 || len(recipe.Images) > 0 ||
		!recipe.Date.IsZero() || len(recipe.Metadata) > 0 {
		result.WriteString("## Recipe Information\n\n")

		if recipe.Description != "" {
			result.WriteString(fmt.Sprintf("**Description:** %s\n\n", recipe.Description))
		}
		if recipe.Cuisine != "" {
			result.WriteString(fmt.Sprintf("**Cuisine:** %s\n\n", recipe.Cuisine))
		}
		if !recipe.Date.IsZero() {
			result.WriteString(fmt.Sprintf("**Date:** %s\n\n", recipe.Date.Format("2006-01-02")))
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
			result.WriteString("**Tags:**\n")
			for _, tag := range recipe.Tags {
				result.WriteString(fmt.Sprintf("  - %s\n", tag))
			}
			result.WriteString("\n")
		}
		if len(recipe.Images) > 0 {
			result.WriteString("**Images:**\n")
			for _, img := range recipe.Images {
				result.WriteString(fmt.Sprintf("  - %s\n", img))
			}
			result.WriteString("\n")
		}
		// Include any additional metadata fields
		if len(recipe.Metadata) > 0 {
			for key, value := range recipe.Metadata {
				// Skip fields already displayed above
				if key != "title" && key != "cuisine" && key != "date" && key != "description" &&
					key != "difficulty" && key != "prep_time" && key != "total_time" &&
					key != "author" && key != "servings" && key != "tags" && key != "images" && key != "image" {
					result.WriteString(fmt.Sprintf("**%s:** %s\n\n", simpleTitle(strings.ReplaceAll(key, "_", " ")), value))
				}
			}
		}
	}

	// Ingredients list
	ingredients := recipe.GetIngredients()
	if len(ingredients.Ingredients) > 0 {
		result.WriteString("## Ingredients\n\n")

		for _, ingredient := range ingredients.Ingredients {
			result.WriteString("- ")
			optionalSuffix := ""
			if ingredient.Optional {
				optionalSuffix = " *(optional)*"
			}
			if ingredient.Quantity > 0 {
				if ingredient.Unit != "" {
					result.WriteString(fmt.Sprintf("**%g %s** %s%s\n", ingredient.Quantity, ingredient.Unit, ingredient.Name, optionalSuffix))
				} else {
					result.WriteString(fmt.Sprintf("**%g** %s%s\n", ingredient.Quantity, ingredient.Name, optionalSuffix))
				}
			} else if ingredient.Quantity == -1 {
				// "some" quantity
				if ingredient.Unit != "" {
					result.WriteString(fmt.Sprintf("**some %s** %s%s\n", ingredient.Unit, ingredient.Name, optionalSuffix))
				} else {
					result.WriteString(fmt.Sprintf("**some** %s%s\n", ingredient.Name, optionalSuffix))
				}
			} else {
				result.WriteString(fmt.Sprintf("%s%s\n", ingredient.Name, optionalSuffix))
			}
		}
		result.WriteString("\n")
	}

	// Instructions
	result.WriteString("## Instructions\n\n")

	stepNum := 1
	currentStep := recipe.FirstStep
	for currentStep != nil {
		// Check if the first component is a section - render it specially
		firstComp := currentStep.FirstComponent
		if section, ok := firstComp.(*cooklang.Section); ok {
			// Render section as a heading
			if section.Name != "" {
				result.WriteString(fmt.Sprintf("### %s\n\n", section.Name))
			}
			stepNum = 1 // Reset step numbering for new section
			// Move to the next component after the section
			currentComponent := section.GetNext()
			if currentComponent != nil {
				result.WriteString(fmt.Sprintf("%d. ", stepNum))
				for currentComponent != nil {
					mr.renderComponent(&result, currentComponent)
					currentComponent = currentComponent.GetNext()
				}
				result.WriteString("\n\n")
				stepNum++
			}
		} else if note, ok := firstComp.(*cooklang.Note); ok {
			// Render notes as blockquotes without step numbers
			result.WriteString(fmt.Sprintf("> %s\n\n", note.Text))
			// Don't increment step number for notes
		} else {
			result.WriteString(fmt.Sprintf("%d. ", stepNum))

			// Render components in markdown-friendly format
			currentComponent := currentStep.FirstComponent
			for currentComponent != nil {
				mr.renderComponent(&result, currentComponent)
				currentComponent = currentComponent.GetNext()
			}

			result.WriteString("\n\n")
			stepNum++
		}
		currentStep = currentStep.NextStep
	}

	return result.String()
}

// renderComponent renders a single component in markdown format
func (mr MarkdownRenderer) renderComponent(result *strings.Builder, currentComponent cooklang.StepComponent) {
	switch comp := currentComponent.(type) {
	case *cooklang.Ingredient:
		if comp.Quantity > 0 {
			fmt.Fprintf(result, "**%s** (%g %s)", comp.Name, comp.Quantity, comp.Unit)
		} else {
			fmt.Fprintf(result, "**%s**", comp.Name)
		}
		if comp.Annotation != "" {
			fmt.Fprintf(result, " (%s)", comp.Annotation)
		}
		if comp.Optional {
			result.WriteString(" *(optional)*")
		}
	case *cooklang.Cookware:
		if comp.Quantity > 1 {
			fmt.Fprintf(result, "*%s* (x%d)", comp.Name, comp.Quantity)
		} else {
			fmt.Fprintf(result, "*%s*", comp.Name)
		}
		if comp.Annotation != "" {
			fmt.Fprintf(result, " (%s)", comp.Annotation)
		}
	case *cooklang.Timer:
		if comp.Name != "" {
			fmt.Fprintf(result, "⏲️ %s (%s)", comp.Name, comp.Duration)
		} else {
			fmt.Fprintf(result, "⏲️ %s", comp.Duration)
		}
		if comp.Annotation != "" {
			fmt.Fprintf(result, " (%s)", comp.Annotation)
		}
	case *cooklang.Instruction:
		result.WriteString(comp.Text)
	case *cooklang.Section:
		// Sections handled specially in the main render loop
		if comp.Name != "" {
			fmt.Fprintf(result, "\n\n### %s\n\n", comp.Name)
		}
	case *cooklang.Comment:
		// Render comments as italicized text
		fmt.Fprintf(result, "*(%s)*", comp.Text)
	case *cooklang.Note:
		// Render notes as blockquotes (Markdown style)
		fmt.Fprintf(result, "\n\n> %s\n\n", comp.Text)
	}
}

// DefaultMarkdownRenderer is the default instance of MarkdownRenderer
var DefaultMarkdownRenderer = MarkdownRenderer{}
