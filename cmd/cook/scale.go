package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilli/cooklang"
	"github.com/hilli/cooklang/renderers"
	"github.com/spf13/cobra"
)

var (
	scaleServings int
	scaleFactor   float64
	scaleUnit     string
	scaleOutput   string
	scaleFormat   string
	scaleJSON     bool
)

var scaleCmd = &cobra.Command{
	Use:   "scale <recipe.cook> [--servings N | --factor F]",
	Short: "Scale a recipe's ingredients",
	Long: `Scale a recipe's ingredients by servings or a custom factor.

The scale command adjusts all ingredient quantities in a recipe based on:
  - Target number of servings (--servings)
  - Custom scaling factor (--factor)

Examples:
  # Scale to 4 servings
  cook scale recipe.cook --servings 4

  # Scale by 1.5x
  cook scale recipe.cook --factor 1.5

  # Scale and convert units
  cook scale recipe.cook --servings 6 --unit metric

  # Scale and save to file
  cook scale recipe.cook --servings 2 --output scaled.cook

  # Scale and output as JSON
  cook scale recipe.cook --factor 0.5 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runScale,
}

func init() {
	rootCmd.AddCommand(scaleCmd)

	scaleCmd.Flags().IntVarP(&scaleServings, "servings", "s", 0, "Target number of servings")
	scaleCmd.Flags().Float64VarP(&scaleFactor, "factor", "f", 0, "Scaling factor (e.g., 0.5 for half, 2 for double)")
	scaleCmd.Flags().StringVarP(&scaleUnit, "unit", "u", "", "Convert to unit system (metric/imperial)")
	scaleCmd.Flags().StringVarP(&scaleOutput, "output", "o", "", "Output file (default: stdout)")
	scaleCmd.Flags().StringVar(&scaleFormat, "format", "cooklang", "Output format: cooklang, markdown, html, json")
	scaleCmd.Flags().BoolVar(&scaleJSON, "json", false, "Output as JSON")
}

func runScale(cmd *cobra.Command, args []string) error {
	filename := args[0]
	recipe, err := readRecipeFile(filename)
	if err != nil {
		return err
	}

	// Validate scaling parameters
	if scaleServings == 0 && scaleFactor == 0 {
		return fmt.Errorf("must specify either --servings or --factor")
	}
	if scaleServings > 0 && scaleFactor > 0 {
		return fmt.Errorf("cannot specify both --servings and --factor")
	}

	// Scale the recipe using library methods
	var scaledRecipe *cooklang.Recipe
	var scale float64

	if scaleServings > 0 {
		// Use library's ScaleToServings method
		originalServings := recipe.Servings
		if originalServings == 0 {
			printWarning("Recipe doesn't specify servings, assuming 1 serving")
			originalServings = 1
		}
		scale = float64(scaleServings) / float64(originalServings)
		printInfo("Scaling from %.0f to %d servings (factor: %.2fx)", originalServings, scaleServings, scale)
		scaledRecipe = recipe.ScaleToServings(float64(scaleServings))
	} else {
		scale = scaleFactor
		printInfo("Scaling by factor: %.2fx", scale)
		scaledRecipe = recipe.Scale(scaleFactor)
	}

	// Apply unit conversion if requested
	if scaleUnit != "" {
		if err := convertRecipeUnits(scaledRecipe, scaleUnit); err != nil {
			return fmt.Errorf("unit conversion failed: %w", err)
		}
		printInfo("Converted to %s units", scaleUnit)
	}

	// Generate output
	var output string
	if scaleJSON {
		output, err = formatScaledJSON(scaledRecipe, scale)
		if err != nil {
			return err
		}
	} else {
		output, err = formatScaledRecipe(scaledRecipe, scaleFormat)
		if err != nil {
			return err
		}
	}

	// Write output
	if scaleOutput != "" {
		// Create directory if needed
		dir := filepath.Dir(scaleOutput)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		if err := os.WriteFile(scaleOutput, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printSuccess("Scaled recipe saved to: %s", scaleOutput)
	} else {
		fmt.Println(output)
	}

	return nil
}

// convertRecipeUnits converts all ingredients in a recipe to the specified unit system
func convertRecipeUnits(recipe *cooklang.Recipe, unitSystem string) error {
	ingredients := recipe.GetIngredients()
	if ingredients == nil || len(ingredients.Ingredients) == 0 {
		return nil
	}

	// Determine target unit system
	var system cooklang.UnitSystem
	switch strings.ToLower(unitSystem) {
	case "metric":
		system = cooklang.UnitSystemMetric
	case "imperial":
		system = cooklang.UnitSystemImperial
	case "us":
		system = cooklang.UnitSystemUS
	default:
		return fmt.Errorf("unknown unit system: %s (use metric, imperial, or us)", unitSystem)
	}

	// Convert ingredients in-place by walking the recipe steps
	for step := recipe.FirstStep; step != nil; step = step.NextStep {
		for component := step.FirstComponent; component != nil; component = component.GetNext() {
			if ing, ok := component.(*cooklang.Ingredient); ok {
				converted := ing.ConvertToSystem(system)
				ing.Quantity = converted.Quantity
				ing.Unit = converted.Unit
				ing.TypedUnit = converted.TypedUnit
			}
		}
	}

	return nil
}

// formatScaledRecipe renders the scaled recipe in the specified format
func formatScaledRecipe(recipe *cooklang.Recipe, format string) (string, error) {
	switch strings.ToLower(format) {
	case "cooklang", "cook":
		renderer := renderers.NewCooklangRenderer()
		return renderer.RenderRecipe(recipe), nil
	case "markdown", "md":
		renderer := renderers.NewMarkdownRenderer()
		return renderer.RenderRecipe(recipe), nil
	case "html":
		renderer := renderers.NewHTMLRenderer()
		return renderer.RenderRecipe(recipe), nil
	case "json":
		return formatScaledJSON(recipe, 1.0)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatScaledJSON formats the scaled recipe as JSON
func formatScaledJSON(recipe *cooklang.Recipe, factor float64) (string, error) {
	result := map[string]interface{}{
		"metadata":     recipe.Metadata,
		"scale_factor": factor,
	}

	// Get ingredients using RenderDisplay for cleaner output
	ingredients := recipe.GetIngredients()
	if ingredients != nil && len(ingredients.Ingredients) > 0 {
		ingList := []map[string]interface{}{}
		for _, ing := range ingredients.Ingredients {
			ingMap := map[string]interface{}{
				"name":    ing.Name,
				"display": ing.RenderDisplay(),
			}
			if ing.Quantity != 0 {
				ingMap["quantity"] = ing.Quantity
			}
			if ing.Unit != "" {
				ingMap["unit"] = ing.Unit
			}
			if ing.Annotation != "" {
				ingMap["annotation"] = ing.Annotation
			}
			ingList = append(ingList, ingMap)
		}
		result["ingredients"] = ingList
	}

	// Get steps
	steps := []string{}
	for step := recipe.FirstStep; step != nil; step = step.NextStep {
		stepText := getStepText(step)
		steps = append(steps, stepText)
	}
	result["steps"] = steps

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}
