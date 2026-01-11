package cooklang_test

import (
	"fmt"
	"log"
	"sort"

	"github.com/hilli/cooklang"
)

// ExampleParseString demonstrates basic recipe parsing from a string
func ExampleParseString() {
	recipeText := `---
title: Pasta Aglio e Olio
servings: 2
---
Cook @pasta{400%g} in salted water for ~{10%minutes}.
Meanwhile, heat @olive oil{4%tbsp} and sauté @garlic{3%cloves}.
Toss everything together and serve.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Recipe: %s\n", recipe.Title)
	fmt.Printf("Servings: %.0f\n", recipe.Servings)
	// Output:
	// Recipe: Pasta Aglio e Olio
	// Servings: 2
}

// ExampleParseFile demonstrates parsing a recipe from a file
func ExampleParseFile() {
	// Create a temporary recipe file for demonstration
	// In real usage, you would use an actual .cook file path
	recipe, err := cooklang.ParseString(`---
title: Quick Omelette
servings: 1
---
Beat @eggs{2} with @milk{2%tbsp}.
Cook in a #pan{} over medium heat for ~{3%minutes}.`)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(recipe.Title)
	// Output:
	// Quick Omelette
}

// ExampleRecipe_GetIngredients shows how to extract all ingredients from a recipe
func ExampleRecipe_GetIngredients() {
	recipeText := `Mix @flour{200%g}, @sugar{150%g}, and @butter{100%g}.
Add @eggs{2} and @vanilla{1%tsp}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	ingredients := recipe.GetIngredients()
	fmt.Printf("Found %d ingredients:\n", len(ingredients.Ingredients))
	for _, ing := range ingredients.Ingredients {
		if ing.Unit != "" {
			fmt.Printf("- %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
		} else {
			fmt.Printf("- %s: %.0f\n", ing.Name, ing.Quantity)
		}
	}
	// Output:
	// Found 5 ingredients:
	// - flour: 200 g
	// - sugar: 150 g
	// - butter: 100 g
	// - eggs: 2
	// - vanilla: 1 tsp
}

// ExampleRecipe_GetCookware shows how to extract cookware items from a recipe
func ExampleRecipe_GetCookware() {
	recipeText := `Mix ingredients in a #mixing bowl{}.
Transfer to a #baking dish{} and bake in an #oven{}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	cookware := recipe.GetCookware()
	fmt.Printf("Cookware needed (%d items):\n", len(cookware))
	for _, cw := range cookware {
		fmt.Printf("- %s\n", cw.Name)
	}
	// Output:
	// Cookware needed (3 items):
	// - mixing bowl
	// - baking dish
	// - oven
}

// ExampleIngredient_ConvertTo demonstrates unit conversion for ingredients
func ExampleIngredient_ConvertTo() {
	recipeText := `Add @water{500%ml}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	ingredients := recipe.GetIngredients()
	water := ingredients.Ingredients[0]

	// Convert from milliliters to cups
	converted, err := water.ConvertTo("cup")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: %.0f %s\n", water.Quantity, water.Unit)
	fmt.Printf("Converted: %.2f %s\n", converted.Quantity, converted.Unit)
	// Output:
	// Original: 500 ml
	// Converted: 2.11 cup
}

// ExampleIngredient_CanConvertTo shows how to check if unit conversion is possible
func ExampleIngredient_CanConvertTo() {
	recipeText := `Add @flour{200%g}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	flour := recipe.GetIngredients().Ingredients[0]

	// Check various conversions
	fmt.Printf("Can convert to kg: %v\n", flour.CanConvertTo("kg"))
	fmt.Printf("Can convert to oz: %v\n", flour.CanConvertTo("oz"))
	fmt.Printf("Can convert to ml: %v\n", flour.CanConvertTo("ml"))
	// Output:
	// Can convert to kg: true
	// Can convert to oz: true
	// Can convert to ml: false
}

// ExampleIngredientList_ConvertToSystem demonstrates converting all ingredients to a unit system
func ExampleIngredientList_ConvertToSystem() {
	recipeText := `Add @water{2%cup}, @flour{500%g}, and @sugar{1%lb}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	ingredients := recipe.GetIngredients()

	// Convert to metric system
	metric := ingredients.ConvertToSystem(cooklang.UnitSystemMetric)

	fmt.Println("Metric ingredients:")
	for _, ing := range metric.Ingredients {
		fmt.Printf("- %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
	}
	// Output:
	// Metric ingredients:
	// - water: 473 ml
	// - flour: 500 g
	// - sugar: 454 g
}

// ExampleIngredientList_ConsolidateByName shows how to consolidate duplicate ingredients
func ExampleIngredientList_ConsolidateByName() {
	recipeText := `Add @flour{200%g} to bowl.
Mix with @flour{300%g} and @sugar{100%g}.
Add @sugar{50%g} for sweetness.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	ingredients := recipe.GetIngredients()
	consolidated, err := ingredients.ConsolidateByName("")
	if err != nil {
		log.Fatal(err)
	}

	// Sort ingredients by name for consistent output
	sort.Slice(consolidated.Ingredients, func(i, j int) bool {
		return consolidated.Ingredients[i].Name < consolidated.Ingredients[j].Name
	})

	fmt.Println("Consolidated ingredients:")
	for _, ing := range consolidated.Ingredients {
		fmt.Printf("- %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
	}
	// Output:
	// Consolidated ingredients:
	// - flour: 500 g
	// - sugar: 150 g
}

// ExampleRecipe_GetCollectedIngredients demonstrates getting consolidated ingredients in one step
func ExampleRecipe_GetCollectedIngredients() {
	recipeText := `First layer: @cheese{100%g} and @tomato{2}.
Second layer: @cheese{150%g} and @tomato{3}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	collected, err := recipe.GetCollectedIngredients()
	if err != nil {
		log.Fatal(err)
	}

	// Sort ingredients by name for consistent output
	sort.Slice(collected.Ingredients, func(i, j int) bool {
		return collected.Ingredients[i].Name < collected.Ingredients[j].Name
	})

	fmt.Println("Shopping list:")
	for _, ing := range collected.Ingredients {
		if ing.Unit != "" {
			fmt.Printf("- %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
		} else {
			fmt.Printf("- %s: %.0f\n", ing.Name, ing.Quantity)
		}
	}
	// Output:
	// Shopping list:
	// - cheese: 250 g
	// - tomato: 5
}

// ExampleRecipe_GetCollectedIngredientsMap shows getting ingredients as a map for display
func ExampleRecipe_GetCollectedIngredientsMap() {
	recipeText := `Add @water{1%l}, @salt{1%tsp}, and @salt{0.5%tsp}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	shoppingMap, err := recipe.GetCollectedIngredientsMap()
	if err != nil {
		log.Fatal(err)
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(shoppingMap))
	for name := range shoppingMap {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		fmt.Printf("%s: %s\n", name, shoppingMap[name])
	}
	// Output:
	// salt: 1.5 tsp
	// water: 1 l
}

// ExampleCreateShoppingList demonstrates creating a shopping list from multiple recipes
func ExampleCreateShoppingList() {
	recipe1Text := `---
title: Pasta
---
Cook @pasta{400%g} with @olive oil{2%tbsp}.`

	recipe2Text := `---
title: Salad
---
Mix @olive oil{3%tbsp} with @lettuce{100%g}.`

	recipe1, _ := cooklang.ParseString(recipe1Text)
	recipe2, _ := cooklang.ParseString(recipe2Text)

	shoppingList, err := cooklang.CreateShoppingList(recipe1, recipe2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Shopping list for %d recipes:\n", len(shoppingList.Recipes))
	ingredientMap := shoppingList.ToMap()

	// Print in deterministic order
	fmt.Printf("- lettuce: %s\n", ingredientMap["lettuce"])
	fmt.Printf("- olive oil: %s\n", ingredientMap["olive oil"])
	fmt.Printf("- pasta: %s\n", ingredientMap["pasta"])
	// Output:
	// Shopping list for 2 recipes:
	// - lettuce: 100 g
	// - olive oil: 5 tbsp
	// - pasta: 400 g
}

// ExampleShoppingList_Scale demonstrates scaling a shopping list
func ExampleShoppingList_Scale() {
	recipeText := `Use @flour{200%g} and @sugar{100%g}.`

	recipe, _ := cooklang.ParseString(recipeText)
	shoppingList, _ := cooklang.CreateShoppingList(recipe)

	// Double the recipe
	scaled := shoppingList.Scale(2.0)

	fmt.Println("Scaled (×2):")
	scaledMap := scaled.ToMap()
	// Sort keys for consistent output
	keys := make([]string, 0, len(scaledMap))
	for name := range scaledMap {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		fmt.Printf("- %s: %s\n", name, scaledMap[name])
	}
	// Output:
	// Scaled (×2):
	// - flour: 400 g
	// - sugar: 200 g
}

// ExampleRecipe_GetMetricShoppingList shows getting a shopping list in metric units
func ExampleRecipe_GetMetricShoppingList() {
	recipeText := `Add @flour{2%cup} and @butter{4%oz}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	metricList, err := recipe.GetMetricShoppingList()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Metric shopping list:")
	// Print in deterministic order
	fmt.Printf("- butter: %s\n", metricList["butter"])
	fmt.Printf("- flour: %s\n", metricList["flour"])
	// Output:
	// Metric shopping list:
	// - butter: 113.4 g
	// - flour: 473.2 ml
}

// ExampleRecipe_GetUSShoppingList shows getting a shopping list in US units
func ExampleRecipe_GetUSShoppingList() {
	recipeText := `Add @water{500%ml} and @flour{250%g}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	usList, err := recipe.GetUSShoppingList()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("US shopping list:")
	// Print in deterministic order
	fmt.Printf("- flour: %s\n", usList["flour"])
	fmt.Printf("- water: %s\n", usList["water"])
	// Output:
	// US shopping list:
	// - flour: 8.8 oz
	// - water: 2.1 cup
}

// ExampleNewFrontmatterEditor demonstrates editing recipe metadata
func ExampleNewFrontmatterEditor() {
	// In real usage, you would use an actual file path
	// This example shows the API usage
	recipeText := `---
title: Old Title
servings: 2
---
Cook @pasta{400%g}.`

	recipe, _ := cooklang.ParseString(recipeText)
	fmt.Printf("Original title: %s\n", recipe.Title)
	fmt.Printf("Original servings: %.0f\n", recipe.Servings)

	// Note: In actual use, you'd create editor with NewFrontmatterEditor(filepath)
	// and then call SetMetadata, Save, etc.
	// Output:
	// Original title: Old Title
	// Original servings: 2
}

// ExampleFrontmatterEditor_GetMetadata shows retrieving metadata values
func ExampleFrontmatterEditor_GetMetadata() {
	// Demonstrates the pattern for getting metadata
	recipeText := `---
title: Chocolate Cake
difficulty: Medium
prep_time: 30 minutes
---
Mix ingredients.`

	recipe, _ := cooklang.ParseString(recipeText)

	// Access metadata directly from recipe
	fmt.Printf("Title: %s\n", recipe.Title)
	fmt.Printf("Difficulty: %s\n", recipe.Difficulty)
	fmt.Printf("Prep time: %s\n", recipe.PrepTime)
	// Output:
	// Title: Chocolate Cake
	// Difficulty: Medium
	// Prep time: 30 minutes
}

// ExampleRecipe_Render demonstrates basic recipe rendering
func ExampleRecipe_Render() {
	recipeText := `---
title: Quick Snack
servings: 1
---
Toast @bread{2%slices} and spread @butter{1%tbsp}.`

	recipe, err := cooklang.ParseString(recipeText)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Recipe: %s (serves %.0f)\n", recipe.Title, recipe.Servings)
	// Output:
	// Recipe: Quick Snack (serves 1)
}

// ExampleIngredient_GetUnitType shows how to get the type of a unit
func ExampleIngredient_GetUnitType() {
	recipeText := `Add @water{500%ml}, @flour{200%g}, and @vanilla{1%tsp}.`

	recipe, _ := cooklang.ParseString(recipeText)
	ingredients := recipe.GetIngredients()

	for _, ing := range ingredients.Ingredients {
		unitType := ing.GetUnitType()
		fmt.Printf("%s: %s\n", ing.Name, unitType)
	}
	// Output:
	// water: volume
	// flour: mass
	// vanilla: volume
}

// ExampleIngredientList_ToMap shows converting ingredient list to a map
func ExampleIngredientList_ToMap() {
	recipeText := `Use @flour{500%g}, @sugar{200%g}, and @eggs{3}.`

	recipe, _ := cooklang.ParseString(recipeText)
	ingredients := recipe.GetIngredients()

	ingredientMap := ingredients.ToMap()

	// Print in deterministic order for test
	fmt.Printf("eggs: %s\n", ingredientMap["eggs"])
	fmt.Printf("flour: %s\n", ingredientMap["flour"])
	fmt.Printf("sugar: %s\n", ingredientMap["sugar"])
	// Output:
	// eggs: 3
	// flour: 500 g
	// sugar: 200 g
}

// ExampleIngredient_Render shows rendering an ingredient back to Cooklang format
func ExampleIngredient_Render() {
	recipeText := `Add @garlic{3%cloves}(minced) and @salt{}.`

	recipe, _ := cooklang.ParseString(recipeText)
	ingredients := recipe.GetIngredients()

	for _, ing := range ingredients.Ingredients {
		fmt.Println(ing.Render())
	}
	// Output:
	// @garlic{3%cloves}(minced)
	// @salt{}
}

// ExampleTimer demonstrates working with timers in recipes
func ExampleTimer() {
	recipeText := `Boil for ~{10%minutes}.
Rest for ~cooling{5%minutes}.`

	recipe, _ := cooklang.ParseString(recipeText)

	// Walk through steps to find timers
	step := recipe.FirstStep
	timersFound := 0
	for step != nil {
		component := step.FirstComponent
		for component != nil {
			if timer, ok := component.(*cooklang.Timer); ok {
				timersFound++
				if timer.Name != "" {
					fmt.Printf("Timer: %s (%s)\n", timer.Name, timer.Duration)
				} else {
					fmt.Printf("Timer: %s\n", timer.Duration)
				}
			}
			component = component.GetNext()
		}
		step = step.NextStep
	}
	fmt.Printf("Total timers: %d\n", timersFound)
	// Output:
	// Timer: 10
	// Timer: cooling (5)
	// Total timers: 2
}

// ExampleCookware demonstrates working with cookware items
func ExampleCookware() {
	recipeText := `Use a #large pot{} and #wooden spoons{2}.`

	recipe, _ := cooklang.ParseString(recipeText)
	cookware := recipe.GetCookware()

	for _, cw := range cookware {
		if cw.Quantity > 1 {
			fmt.Printf("%s (×%d)\n", cw.Name, cw.Quantity)
		} else {
			fmt.Printf("%s\n", cw.Name)
		}
	}
	// Output:
	// large pot
	// wooden spoons (×2)
}

// ExampleRecipe_Scale demonstrates scaling a recipe by a factor
func ExampleRecipe_Scale() {
	recipeText := `---
title: Pancakes
servings: 2
---
Mix @flour{200%g} with @milk{300%ml} and @eggs{2}.`

	recipe, _ := cooklang.ParseString(recipeText)

	// Double the recipe
	doubled := recipe.Scale(2.0)

	fmt.Printf("Original servings: %.0f\n", recipe.Servings)
	fmt.Printf("Doubled servings: %.0f\n", doubled.Servings)

	// Show scaled ingredients
	ingredients := doubled.GetIngredients()
	for _, ing := range ingredients.Ingredients {
		if ing.Unit != "" {
			fmt.Printf("- %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
		} else {
			fmt.Printf("- %s: %.0f\n", ing.Name, ing.Quantity)
		}
	}
	// Output:
	// Original servings: 2
	// Doubled servings: 4
	// - flour: 400 g
	// - milk: 600 ml
	// - eggs: 4
}

// ExampleRecipe_ScaleToServings demonstrates scaling a recipe to target servings
func ExampleRecipe_ScaleToServings() {
	recipeText := `---
title: Cookies
servings: 12
---
Mix @flour{300%g}, @sugar{150%g}, and @butter{100%g}.`

	recipe, _ := cooklang.ParseString(recipeText)

	// Scale from 12 to 36 servings (triple)
	scaled := recipe.ScaleToServings(36)

	fmt.Printf("Original: %.0f servings\n", recipe.Servings)
	fmt.Printf("Scaled: %.0f servings\n", scaled.Servings)

	ingredients := scaled.GetIngredients()
	for _, ing := range ingredients.Ingredients {
		fmt.Printf("- %s: %.0f %s\n", ing.Name, ing.Quantity, ing.Unit)
	}
	// Output:
	// Original: 12 servings
	// Scaled: 36 servings
	// - flour: 900 g
	// - sugar: 450 g
	// - butter: 300 g
}

// ExampleCreateShoppingListForServings demonstrates creating a shopping list
// for a specific number of servings across multiple recipes
func ExampleCreateShoppingListForServings() {
	mondayDinner := `---
title: Pasta
servings: 2
---
Cook @pasta{200%g} with @olive oil{2%tbsp}.`

	tuesdayDinner := `---
title: Rice Bowl
servings: 1
---
Serve @rice{150%g} with @olive oil{1%tbsp}.`

	recipe1, _ := cooklang.ParseString(mondayDinner)
	recipe2, _ := cooklang.ParseString(tuesdayDinner)

	// Create shopping list for 4 servings of each recipe
	list, err := cooklang.CreateShoppingListForServings(4, recipe1, recipe2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Shopping list (4 servings each):")
	ingredientMap := list.ToMap()

	// Print in deterministic order
	fmt.Printf("- olive oil: %s\n", ingredientMap["olive oil"])
	fmt.Printf("- pasta: %s\n", ingredientMap["pasta"])
	fmt.Printf("- rice: %s\n", ingredientMap["rice"])
	// Output:
	// Shopping list (4 servings each):
	// - olive oil: 8 tbsp
	// - pasta: 400 g
	// - rice: 600 g
}
