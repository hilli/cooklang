package main

import (
	"fmt"
	"strings"

	"github.com/hilli/cooklang"
	"github.com/spf13/cobra"
)

var (
	ingredientsJSON        bool
	ingredientsConsolidate bool
	ingredientsTargetUnit  string
)

var ingredientsCmd = &cobra.Command{
	Use:   "ingredients <recipe-file> [recipe-files...]",
	Short: "List ingredients from one or more recipes",
	Long: `Extract and display ingredients from recipe files.

Can list ingredients from a single recipe or consolidate ingredients
from multiple recipes with automatic unit conversion.

Examples:
  cook ingredients recipe.cook
  cook ingredients recipe.cook --consolidate
  cook ingredients recipe1.cook recipe2.cook --consolidate
  cook ingredients *.cook --consolidate --unit=kg
  cook ingredients recipe.cook --unit=metric
  cook ingredients recipe.cook --json`,
	Args:              cobra.MinimumNArgs(1),
	RunE:              runIngredients,
	ValidArgsFunction: completeCookFiles,
}

func init() {
	ingredientsCmd.Flags().BoolVarP(&ingredientsJSON, "json", "j", false, "Output as JSON")
	ingredientsCmd.Flags().BoolVarP(&ingredientsConsolidate, "consolidate", "c", false, "Consolidate ingredients with the same name")
	ingredientsCmd.Flags().StringVarP(&ingredientsTargetUnit, "unit", "u", "", "Convert to target unit or system (e.g., g, kg, ml, metric, imperial, us)")
	rootCmd.AddCommand(ingredientsCmd)

	// Register completion for the --unit flag
	_ = ingredientsCmd.RegisterFlagCompletionFunc("unit", completeUnitFlag)
}

func runIngredients(cmd *cobra.Command, args []string) error {
	recipes, err := readMultipleRecipes(args)
	if err != nil {
		return err
	}

	if len(recipes) == 1 && !ingredientsConsolidate {
		// Single recipe, no consolidation
		return displaySingleRecipeIngredients(recipes[0], args[0])
	}

	// Multiple recipes or consolidation requested
	return displayConsolidatedIngredients(recipes, args)
}

// getUnitSystem checks if the unit string is a system name and returns the system
func getUnitSystem(unit string) (cooklang.UnitSystem, bool) {
	switch strings.ToLower(unit) {
	case "metric":
		return cooklang.UnitSystemMetric, true
	case "imperial":
		return cooklang.UnitSystemImperial, true
	case "us":
		return cooklang.UnitSystemUS, true
	default:
		return cooklang.UnitSystemMetric, false
	}
}

// convertIngredientList converts all ingredients to the target unit or system
func convertIngredientList(ingredients *cooklang.IngredientList, targetUnit string) *cooklang.IngredientList {
	if targetUnit == "" {
		return ingredients
	}

	// Check if it's a unit system
	if system, isSystem := getUnitSystem(targetUnit); isSystem {
		return ingredients.ConvertToSystem(system)
	}

	// Otherwise, try to convert each ingredient to the specific unit
	result := cooklang.NewIngredientList()
	for _, ing := range ingredients.Ingredients {
		if ing.CanConvertTo(targetUnit) {
			converted, err := ing.ConvertTo(targetUnit)
			if err == nil {
				result.Add(converted)
				continue
			}
		}
		// If conversion not possible, keep original
		result.Add(ing)
	}
	return result
}

func displaySingleRecipeIngredients(recipe *cooklang.Recipe, filename string) error {
	ingredients := recipe.GetIngredients()

	// Apply unit conversion if requested
	if ingredientsTargetUnit != "" {
		ingredients = convertIngredientList(ingredients, ingredientsTargetUnit)
	}

	if ingredientsJSON {
		return outputJSON(ingredients)
	}

	fmt.Printf("📄 Ingredients from: %s\n", filename)
	if recipe.Title != "" {
		fmt.Printf("📋 Recipe: %s\n", recipe.Title)
	}
	if recipe.Servings > 0 {
		fmt.Printf("👥 Servings: %.0f\n", recipe.Servings)
	}
	if ingredientsTargetUnit != "" {
		fmt.Printf("✓ Converted to: %s\n", ingredientsTargetUnit)
	}
	fmt.Println()

	displayIngredientList(ingredients)
	return nil
}

func displayConsolidatedIngredients(recipes []*cooklang.Recipe, filenames []string) error {
	// Collect all ingredients
	allIngredients := cooklang.NewIngredientList()
	for _, recipe := range recipes {
		ingredients := recipe.GetIngredients()
		for _, ing := range ingredients.Ingredients {
			allIngredients.Add(ing)
		}
	}

	// Consolidate if requested
	var finalList *cooklang.IngredientList
	var err error

	if ingredientsConsolidate {
		// For consolidation with a unit system, convert first then consolidate
		if system, isSystem := getUnitSystem(ingredientsTargetUnit); isSystem {
			converted := allIngredients.ConvertToSystem(system)
			finalList, err = converted.ConsolidateByName("")
		} else {
			finalList, err = allIngredients.ConsolidateByName(ingredientsTargetUnit)
		}
		if err != nil {
			printWarning("Some ingredients could not be consolidated: %v", err)
			finalList = allIngredients
		}
	} else {
		finalList = allIngredients
	}

	// Apply unit conversion if not already done during consolidation
	if ingredientsTargetUnit != "" && !ingredientsConsolidate {
		finalList = convertIngredientList(finalList, ingredientsTargetUnit)
	}

	if ingredientsJSON {
		return outputJSON(finalList)
	}

	// Display results
	if len(recipes) > 1 {
		fmt.Printf("📚 Ingredients from %d recipes:\n", len(recipes))
		for i, filename := range filenames {
			title := recipes[i].Title
			if title != "" {
				fmt.Printf("  %d. %s (%s)\n", i+1, title, filename)
			} else {
				fmt.Printf("  %d. %s\n", i+1, filename)
			}
		}
	} else {
		fmt.Printf("📄 Consolidated ingredients from: %s\n", filenames[0])
	}

	if ingredientsConsolidate {
		if ingredientsTargetUnit != "" {
			fmt.Printf("✓ Consolidated and converted to: %s\n", ingredientsTargetUnit)
		} else {
			fmt.Println("✓ Consolidated")
		}
	} else if ingredientsTargetUnit != "" {
		fmt.Printf("✓ Converted to: %s\n", ingredientsTargetUnit)
	}
	fmt.Println()

	displayIngredientList(finalList)

	return nil
}

func displayIngredientList(ingredients *cooklang.IngredientList) {
	if len(ingredients.Ingredients) == 0 {
		fmt.Println("No ingredients found.")
		return
	}

	fmt.Printf("Ingredients (%d):\n", len(ingredients.Ingredients))
	for i, ing := range ingredients.Ingredients {
		display := ing.RenderDisplay()
		if ing.Annotation != "" {
			display += fmt.Sprintf(" (%s)", ing.Annotation)
		}
		fmt.Printf("  %2d. %s\n", i+1, display)
	}
}
