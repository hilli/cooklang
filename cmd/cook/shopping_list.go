package main

import (
	"fmt"
	"sort"

	"github.com/hilli/cooklang"
	"github.com/spf13/cobra"
)

var (
	shoppingListJSON     bool
	shoppingListScale    float64
	shoppingListServings int
	shoppingListUnit     string
	shoppingListSimple   bool
)

var shoppingListCmd = &cobra.Command{
	Use:     "shopping-list <recipe-files...>",
	Short:   "Create a shopping list from multiple recipes",
	Aliases: []string{"shop", "list"},
	Long: `Create a consolidated shopping list from one or more recipe files.

Automatically consolidates ingredients with the same name and compatible units.
Perfect for meal planning and batch cooking.

Options:
  --servings N  Scale each recipe to N servings before combining (ideal for meal planning)
  --scale F     Scale the final shopping list by factor F (for batch cooking)
  
Note: --servings and --scale are mutually exclusive.

Examples:
  # Basic shopping list from multiple recipes
  cook shopping-list dinner.cook dessert.cook

  # Meal planning: scale all recipes to 4 servings (household size)
  cook shop monday.cook tuesday.cook wednesday.cook --servings 4

  # Batch cooking: double the entire shopping list
  cook list meal-prep.cook --scale 2.0

  # Convert units while scaling to servings
  cook list recipes/*.cook --servings 4 --unit metric

  # Simple output format
  cook list meal-prep.cook --simple`,
	Args:              cobra.MinimumNArgs(1),
	RunE:              runShoppingList,
	ValidArgsFunction: completeCookFiles,
}

func init() {
	shoppingListCmd.Flags().BoolVarP(&shoppingListJSON, "json", "j", false, "Output as JSON")
	shoppingListCmd.Flags().Float64VarP(&shoppingListScale, "scale", "S", 1.0, "Scale all quantities by this factor")
	shoppingListCmd.Flags().IntVarP(&shoppingListServings, "servings", "s", 0, "Scale each recipe to this many servings before combining")
	shoppingListCmd.Flags().StringVarP(&shoppingListUnit, "unit", "u", "", "Convert to unit system (metric, imperial, us)")
	shoppingListCmd.Flags().BoolVar(&shoppingListSimple, "simple", false, "Simple format (ingredient: quantity)")
	rootCmd.AddCommand(shoppingListCmd)

	// Register flag completions
	_ = shoppingListCmd.RegisterFlagCompletionFunc("servings", completeServingsFlag)
	_ = shoppingListCmd.RegisterFlagCompletionFunc("unit", completeUnitFlag)
}

func runShoppingList(cmd *cobra.Command, args []string) error {
	// Validate mutual exclusivity of --servings and --scale
	if shoppingListServings > 0 && shoppingListScale != 1.0 {
		return fmt.Errorf("cannot specify both --servings and --scale; use --servings to normalize recipes to a household size, or --scale to multiply the final list")
	}

	// Validate unit system if provided
	var unitSystem cooklang.UnitSystem
	hasUnitSystem := false
	if shoppingListUnit != "" {
		switch shoppingListUnit {
		case "metric":
			unitSystem = cooklang.UnitSystemMetric
			hasUnitSystem = true
		case "imperial":
			unitSystem = cooklang.UnitSystemImperial
			hasUnitSystem = true
		case "us":
			unitSystem = cooklang.UnitSystemUS
			hasUnitSystem = true
		default:
			return fmt.Errorf("invalid unit system: %s (use metric, imperial, or us)", shoppingListUnit)
		}
	}

	recipes, err := readMultipleRecipes(args)
	if err != nil {
		return err
	}

	// Create shopping list
	var shoppingList *cooklang.ShoppingList

	if shoppingListServings > 0 {
		// Scale each recipe to target servings before combining
		shoppingList, err = cooklang.CreateShoppingListForServings(float64(shoppingListServings), recipes...)
		if err != nil {
			printWarning("Some ingredients could not be consolidated: %v", err)
		}
		printInfo("Scaled each recipe to %d servings", shoppingListServings)
	} else {
		shoppingList, err = cooklang.CreateShoppingList(recipes...)
		if err != nil {
			printWarning("Some ingredients could not be consolidated: %v", err)
		}
	}

	// Convert to unit system if requested
	if hasUnitSystem {
		shoppingList.Ingredients = shoppingList.Ingredients.ConvertToSystem(unitSystem)
	}

	// Scale if requested (only when not using --servings)
	if shoppingListScale != 1.0 {
		shoppingList = shoppingList.Scale(shoppingListScale)
	}

	// Output
	if shoppingListJSON {
		return outputJSON(shoppingList)
	}

	displayShoppingList(shoppingList, recipes, args)
	return nil
}

func displayShoppingList(list *cooklang.ShoppingList, recipes []*cooklang.Recipe, filenames []string) {
	fmt.Println("Shopping List")
	fmt.Println(string(make([]byte, 60)))

	// Show recipe sources with their original servings
	if len(recipes) > 1 {
		fmt.Printf("From %d recipes:\n", len(recipes))
		for i, recipe := range recipes {
			title := recipe.Title
			if title == "" {
				title = filenames[i]
			}
			fmt.Printf("  %d. %s (serves %.0f)\n", i+1, title, recipe.Servings)
		}
	} else {
		title := recipes[0].Title
		if title == "" {
			title = filenames[0]
		}
		fmt.Printf("Recipe: %s (serves %.0f)\n", title, recipes[0].Servings)
	}

	// Show servings scaling info
	if shoppingListServings > 0 {
		fmt.Printf("Scaled to: %d servings per recipe\n", shoppingListServings)
	}

	// Show scaling info
	if shoppingListScale != 1.0 {
		fmt.Printf("Scaled by: %.1fx\n", shoppingListScale)
	}

	// Show unit conversion info
	if shoppingListUnit != "" {
		fmt.Printf("Converted to: %s (where possible)\n", shoppingListUnit)
	}

	fmt.Println()

	// Display ingredients
	if shoppingListSimple {
		displaySimpleShoppingList(list)
	} else {
		displayDetailedShoppingList(list)
	}

	// Summary
	fmt.Println()
	fmt.Printf("Total: %d unique ingredients\n", list.Count())
}

func displaySimpleShoppingList(list *cooklang.ShoppingList) {
	shoppingMap := list.ToMap()

	// Sort keys alphabetically
	keys := make([]string, 0, len(shoppingMap))
	for name := range shoppingMap {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		fmt.Printf("%s: %s\n", name, shoppingMap[name])
	}
}

func displayDetailedShoppingList(list *cooklang.ShoppingList) {
	// Group by category (rough heuristic based on common ingredient names)
	produce := []*cooklang.Ingredient{}
	dairy := []*cooklang.Ingredient{}
	meat := []*cooklang.Ingredient{}
	pantry := []*cooklang.Ingredient{}
	spices := []*cooklang.Ingredient{}
	other := []*cooklang.Ingredient{}

	for _, ing := range list.Ingredients.Ingredients {
		name := ing.Name

		// Categorize (very simple heuristic)
		switch {
		case contains(name, []string{"lettuce", "tomato", "onion", "garlic", "carrot", "potato", "cucumber", "pepper", "spinach", "kale"}):
			produce = append(produce, ing)
		case contains(name, []string{"milk", "cream", "cheese", "butter", "yogurt", "parmesan"}):
			dairy = append(dairy, ing)
		case contains(name, []string{"chicken", "beef", "pork", "fish", "lamb", "turkey", "bacon", "sausage"}):
			meat = append(meat, ing)
		case contains(name, []string{"salt", "pepper", "cinnamon", "cumin", "paprika", "oregano", "basil", "thyme", "vanilla"}):
			spices = append(spices, ing)
		case contains(name, []string{"flour", "sugar", "rice", "pasta", "bread", "oil", "vinegar", "sauce"}):
			pantry = append(pantry, ing)
		default:
			other = append(other, ing)
		}
	}

	displayCategory("🥬 Produce", produce)
	displayCategory("🧀 Dairy & Eggs", dairy)
	displayCategory("🥩 Meat & Seafood", meat)
	displayCategory("🏺 Pantry", pantry)
	displayCategory("🌶️  Spices & Seasonings", spices)
	displayCategory("📦 Other", other)
}

func contains(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// sortIngredients sorts a slice of ingredients alphabetically by name
func sortIngredients(ingredients []*cooklang.Ingredient) {
	sort.Slice(ingredients, func(i, j int) bool {
		return ingredients[i].Name < ingredients[j].Name
	})
}

func displayCategory(title string, ingredients []*cooklang.Ingredient) {
	if len(ingredients) == 0 {
		return
	}

	// Sort alphabetically
	sortIngredients(ingredients)

	fmt.Printf("\n%s:\n", title)
	for _, ing := range ingredients {
		if ing.Quantity == -1 {
			fmt.Printf("  ☐ %s (some)\n", ing.Name)
		} else if ing.Unit == "" {
			fmt.Printf("  ☐ %s: %.2g\n", ing.Name, ing.Quantity)
		} else {
			if ing.Quantity == float32(int(ing.Quantity)) {
				fmt.Printf("  ☐ %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
			} else {
				fmt.Printf("  ☐ %s: %.2g %s\n", ing.Name, ing.Quantity, ing.Unit)
			}
		}
	}
}
