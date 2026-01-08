package main

import (
	"fmt"

	"github.com/hilli/cooklang"
	"github.com/spf13/cobra"
)

var (
	parseJSON     bool
	parseDetailed bool
)

var parseCmd = &cobra.Command{
	Use:   "parse <recipe-file>",
	Short: "Parse and display a Cooklang recipe",
	Long: `Parse a Cooklang recipe file and display its contents in a human-readable format.

The parse command validates the recipe syntax and displays:
  • Recipe metadata (title, servings, etc.)
  • Ingredients with quantities and units
  • Cookware requirements
  • Timers
  • Step-by-step instructions

Examples:
  cook parse recipe.cook
  cook parse recipe.cook --json
  cook parse recipe.cook --detailed`,
	Args: cobra.ExactArgs(1),
	RunE: runParse,
}

func init() {
	parseCmd.Flags().BoolVarP(&parseJSON, "json", "j", false, "Output as JSON")
	parseCmd.Flags().BoolVarP(&parseDetailed, "detailed", "d", false, "Show detailed component breakdown")
	rootCmd.AddCommand(parseCmd)
}

func runParse(cmd *cobra.Command, args []string) error {
	filename := args[0]

	recipe, err := readRecipeFile(filename)
	if err != nil {
		return err
	}

	if parseJSON {
		return outputJSON(recipe)
	}

	displayRecipe(recipe, filename, parseDetailed)
	return nil
}

func displayRecipe(recipe *cooklang.Recipe, filename string, detailed bool) {
	fmt.Printf("📄 Recipe: %s\n", filename)
	fmt.Println(string(make([]byte, 60)))

	// Display metadata
	if recipe.Title != "" {
		fmt.Printf("📋 Title: %s\n", recipe.Title)
	}
	if recipe.Servings > 0 {
		fmt.Printf("👥 Servings: %.0f\n", recipe.Servings)
	}
	if recipe.PrepTime != "" {
		fmt.Printf("⏱️  Prep Time: %s\n", recipe.PrepTime)
	}
	if recipe.TotalTime != "" {
		fmt.Printf("⏱️  Total Time: %s\n", recipe.TotalTime)
	}
	if recipe.Description != "" {
		fmt.Printf("📝 Description: %s\n", recipe.Description)
	}
	if len(recipe.Tags) > 0 {
		fmt.Printf("🏷️  Tags: %v\n", recipe.Tags)
	}

	// Display additional metadata
	if len(recipe.Metadata) > 0 {
		fmt.Println("\n📊 Additional Metadata:")
		for key, value := range recipe.Metadata {
			// Skip already displayed fields
			if key != "title" && key != "servings" && key != "prep_time" && key != "total_time" && key != "description" {
				fmt.Printf("  • %s: %s\n", key, value)
			}
		}
	}

	// Display ingredients summary
	ingredients := recipe.GetIngredients()
	if len(ingredients.Ingredients) > 0 {
		fmt.Println("\nIngredients:")
		for _, ing := range ingredients.Ingredients {
			display := ing.RenderDisplay()
			if ing.Annotation != "" {
				display += fmt.Sprintf(" (%s)", ing.Annotation)
			}
			fmt.Printf("  - %s\n", display)
		}
	}

	// Display cookware
	cookware := recipe.GetCookware()
	if len(cookware) > 0 {
		fmt.Println("\nCookware:")
		for _, item := range cookware {
			display := item.RenderDisplay()
			if item.Quantity > 1 {
				display = fmt.Sprintf("%s (x%d)", display, item.Quantity)
			}
			if item.Annotation != "" {
				display += fmt.Sprintf(" (%s)", item.Annotation)
			}
			fmt.Printf("  - %s\n", display)
		}
	}

	// Display steps
	fmt.Println("\n📖 Instructions:")
	step := recipe.FirstStep
	stepNum := 1
	for step != nil {
		if detailed {
			fmt.Printf("\nStep %d (detailed):\n", stepNum)
			displayDetailedStep(step)
		} else {
			text := getStepText(step)
			if text != "" {
				fmt.Printf("%d. %s\n", stepNum, text)
			}
		}
		step = step.NextStep
		stepNum++
	}

	fmt.Println()
}

// getStepText builds the text representation of a step
func getStepText(step *cooklang.Step) string {
	var text string
	currentComponent := step.FirstComponent
	for currentComponent != nil {
		switch comp := currentComponent.(type) {
		case *cooklang.Instruction:
			text += comp.RenderDisplay()
		case *cooklang.Ingredient:
			display := comp.RenderDisplay()
			if comp.Annotation != "" {
				display += ", " + comp.Annotation
			}
			text += display
		case *cooklang.Cookware:
			display := comp.RenderDisplay()
			if comp.Annotation != "" {
				display += " (" + comp.Annotation + ")"
			}
			text += display
		case *cooklang.Timer:
			display := comp.RenderDisplay()
			if display != "" {
				text += "[" + display + "]"
			}
		}
		currentComponent = currentComponent.GetNext()
	}
	return text
}

func displayDetailedStep(step *cooklang.Step) {
	currentComponent := step.FirstComponent
	i := 1
	for currentComponent != nil {
		prefix := fmt.Sprintf("    [%d]", i)
		switch comp := currentComponent.(type) {
		case *cooklang.Ingredient:
			display := comp.RenderDisplay()
			if comp.Annotation != "" {
				display += fmt.Sprintf(" (%s)", comp.Annotation)
			}
			fmt.Printf("%s Ingredient: %s\n", prefix, display)
		case *cooklang.Cookware:
			display := comp.RenderDisplay()
			if comp.Quantity > 1 {
				display = fmt.Sprintf("%s (qty: %d)", display, comp.Quantity)
			}
			if comp.Annotation != "" {
				display += fmt.Sprintf(" (%s)", comp.Annotation)
			}
			fmt.Printf("%s Cookware: %s\n", prefix, display)
		case *cooklang.Timer:
			display := comp.RenderDisplay()
			if comp.Annotation != "" {
				display += fmt.Sprintf(" (%s)", comp.Annotation)
			}
			fmt.Printf("%s Timer: %s\n", prefix, display)
		case *cooklang.Instruction:
			fmt.Printf("%s Text: %q\n", prefix, comp.RenderDisplay())
		}
		currentComponent = currentComponent.GetNext()
		i++
	}
}
