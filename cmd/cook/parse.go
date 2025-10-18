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
		fmt.Println("\n🥕 Ingredients:")
		for _, ing := range ingredients.Ingredients {
			if ing.Quantity == -1 {
				fmt.Printf("  • %s (some)\n", ing.Name)
			} else if ing.Unit == "" {
				fmt.Printf("  • %s: %.2g\n", ing.Name, ing.Quantity)
			} else {
				fmt.Printf("  • %s: %.2g %s\n", ing.Name, ing.Quantity, ing.Unit)
			}
		}
	}

	// Display cookware
	cookware := recipe.GetCookware()
	if len(cookware) > 0 {
		fmt.Println("\n🍳 Cookware:")
		for _, item := range cookware {
			if item.Quantity == 0 || item.Quantity == 1 {
				fmt.Printf("  • %s\n", item.Name)
			} else {
				fmt.Printf("  • %s (x%d)\n", item.Name, item.Quantity)
			}
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
			text += comp.Text
		case *cooklang.Ingredient:
			if comp.Quantity == -1 {
				text += fmt.Sprintf("%s (some)", comp.Name)
			} else if comp.Unit == "" {
				text += fmt.Sprintf("%s (%.2g)", comp.Name, comp.Quantity)
			} else {
				text += fmt.Sprintf("%s (%.2g %s)", comp.Name, comp.Quantity, comp.Unit)
			}
		case *cooklang.Cookware:
			text += comp.Name
		case *cooklang.Timer:
			if comp.Name != "" {
				text += fmt.Sprintf("[%s]", comp.Name)
			} else {
				text += fmt.Sprintf("[timer: %s]", comp.Duration)
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
			fmt.Printf("%s 🥕 Ingredient: %s", prefix, comp.Name)
			if comp.Quantity == -1 {
				fmt.Print(" (some)")
			} else if comp.Unit == "" {
				fmt.Printf(" (%.2g)", comp.Quantity)
			} else {
				fmt.Printf(" (%.2g %s)", comp.Quantity, comp.Unit)
			}
			fmt.Println()
		case *cooklang.Cookware:
			fmt.Printf("%s 🍳 Cookware: %s", prefix, comp.Name)
			if comp.Quantity > 1 {
				fmt.Printf(" (qty: %d)", comp.Quantity)
			}
			fmt.Println()
		case *cooklang.Timer:
			fmt.Printf("%s ⏲️  Timer:", prefix)
			if comp.Name != "" {
				fmt.Printf(" %s", comp.Name)
			}
			if comp.Duration != "" {
				fmt.Printf(" (%s)", comp.Duration)
			}
			fmt.Println()
		case *cooklang.Instruction:
			fmt.Printf("%s 📝 Text: %q\n", prefix, comp.Text)
		}
		currentComponent = currentComponent.GetNext()
		i++
	}
}
