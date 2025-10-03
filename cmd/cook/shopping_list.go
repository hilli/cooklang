package main

import (
	"fmt"

	"github.com/hilli/cooklang"
	"github.com/spf13/cobra"
)

var (
	shoppingListJSON   bool
	shoppingListScale  float64
	shoppingListUnit   string
	shoppingListSimple bool
)

var shoppingListCmd = &cobra.Command{
	Use:     "shopping-list <recipe-files...>",
	Short:   "Create a shopping list from multiple recipes",
	Aliases: []string{"shop", "list"},
	Long: `Create a consolidated shopping list from one or more recipe files.

Automatically consolidates ingredients with the same name and compatible units.
Perfect for meal planning and batch cooking.

Examples:
  cook shopping-list dinner.cook dessert.cook
  cook shop monday.cook tuesday.cook wednesday.cook
  cook list *.cook --scale=2.0
  cook list recipes/*.cook --unit=kg
  cook list meal-prep.cook --simple`,
	Args: cobra.MinimumNArgs(1),
	RunE: runShoppingList,
}

func init() {
	shoppingListCmd.Flags().BoolVarP(&shoppingListJSON, "json", "j", false, "Output as JSON")
	shoppingListCmd.Flags().Float64VarP(&shoppingListScale, "scale", "s", 1.0, "Scale all quantities by this factor")
	shoppingListCmd.Flags().StringVarP(&shoppingListUnit, "unit", "u", "", "Convert to target unit (e.g., g, kg, ml)")
	shoppingListCmd.Flags().BoolVar(&shoppingListSimple, "simple", false, "Simple format (ingredient: quantity)")
	rootCmd.AddCommand(shoppingListCmd)
}

func runShoppingList(cmd *cobra.Command, args []string) error {
	recipes, err := readMultipleRecipes(args)
	if err != nil {
		return err
	}

	// Create shopping list
	var shoppingList *cooklang.ShoppingList
	if shoppingListUnit != "" {
		shoppingList, err = cooklang.CreateShoppingListWithUnit(shoppingListUnit, recipes...)
	} else {
		shoppingList, err = cooklang.CreateShoppingList(recipes...)
	}
	if err != nil {
		printWarning("Some ingredients could not be consolidated: %v", err)
	}

	// Scale if requested
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
	fmt.Println("🛒 Shopping List")
	fmt.Println(string(make([]byte, 60)))

	// Show recipe sources
	if len(recipes) > 1 {
		fmt.Printf("📚 From %d recipes:\n", len(recipes))
		for i, recipe := range recipes {
			if recipe.Title != "" {
				fmt.Printf("  %d. %s\n", i+1, recipe.Title)
			} else {
				fmt.Printf("  %d. %s\n", i+1, filenames[i])
			}
		}
	} else {
		if recipes[0].Title != "" {
			fmt.Printf("📄 Recipe: %s\n", recipes[0].Title)
		} else {
			fmt.Printf("📄 From: %s\n", filenames[0])
		}
	}

	// Show scaling info
	if shoppingListScale != 1.0 {
		fmt.Printf("⚖️  Scaled by: %.1fx\n", shoppingListScale)
	}

	// Show unit conversion info
	if shoppingListUnit != "" {
		fmt.Printf("🔄 Converted to: %s (where possible)\n", shoppingListUnit)
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
	fmt.Printf("📊 Total: %d unique ingredients\n", list.Count())
}

func displaySimpleShoppingList(list *cooklang.ShoppingList) {
	shoppingMap := list.ToMap()
	for name, quantity := range shoppingMap {
		fmt.Printf("%s: %s\n", name, quantity)
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

func displayCategory(title string, ingredients []*cooklang.Ingredient) {
	if len(ingredients) == 0 {
		return
	}

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
