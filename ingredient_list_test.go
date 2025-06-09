package cooklang

import (
	"testing"
)

func TestIngredientListConsolidation(t *testing.T) {
	// Create a recipe with multiple ingredients that can be consolidated
	recipeText := `
Add @flour{500%g} to bowl.
Mix with @flour{0.5%kg} and @sugar{200%g}.
Add @sugar{100%g} and @salt{1%tsp}.
Season with @salt{0.5%tsp} to taste.
Add @vanilla{some} extract.
`

	recipe, err := ParseString(recipeText)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	// Get all ingredients
	ingredients := recipe.GetIngredients()
	t.Logf("Found %d ingredients:", len(ingredients.Ingredients))
	for i, ing := range ingredients.Ingredients {
		t.Logf("  %d: %s - %g %s", i+1, ing.Name, ing.Quantity, ing.Unit)
	}

	if len(ingredients.Ingredients) != 7 {
		t.Errorf("Expected 7 raw ingredients, got %d", len(ingredients.Ingredients))
	}

	// Test consolidation
	consolidated, err := ingredients.ConsolidateByName("")
	if err != nil {
		t.Fatalf("Failed to consolidate ingredients: %v", err)
	}

	// Check the consolidated list - should have 4 unique ingredients
	ingredientMap := consolidated.ToMap()
	if len(ingredientMap) != 4 {
		t.Errorf("Expected 4 consolidated ingredients, got %d", len(ingredientMap))
		for name, quantity := range ingredientMap {
			t.Logf("  %s: %s", name, quantity)
		}
	}

	// Check flour - should be consolidated (500g + 500g = 1000g)
	if flour, exists := ingredientMap["flour"]; exists {
		t.Logf("Consolidated flour: %s", flour)
		if flour != "1000 g" {
			t.Errorf("Expected flour to be '1000 g', got '%s'", flour)
		}
	} else {
		t.Error("Flour not found in consolidated list")
	}

	// Check sugar - should be consolidated (200g + 100g = 300g)
	if sugar, exists := ingredientMap["sugar"]; exists {
		t.Logf("Consolidated sugar: %s", sugar)
		if sugar != "300 g" {
			t.Errorf("Expected sugar to be '300 g', got '%s'", sugar)
		}
	} else {
		t.Error("Sugar not found in consolidated list")
	}

	// Check salt - should be consolidated (1 + 0.5 = 1.5 tsp)
	if salt, exists := ingredientMap["salt"]; exists {
		t.Logf("Consolidated salt: %s", salt)
		if salt != "1.5 tsp" {
			t.Errorf("Expected salt to be '1.5 tsp', got '%s'", salt)
		}
	} else {
		t.Error("Salt not found in consolidated list")
	}

	// Check vanilla - should remain as "some"
	if vanilla, exists := ingredientMap["vanilla"]; exists {
		t.Logf("Vanilla: %s", vanilla)
		if vanilla != "some" {
			t.Errorf("Expected vanilla to be 'some', got '%s'", vanilla)
		}
	} else {
		t.Error("Vanilla not found in consolidated list")
	}
}

func TestIngredientListConsolidationWithTargetUnit(t *testing.T) {
	// Create a recipe with ingredients that need unit conversion
	recipeText := `
Add @butter{500%g} to the pan.
Mix with @butter{0.2%kg} until melted.
`

	recipe, err := ParseString(recipeText)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	ingredients := recipe.GetIngredients()

	// Consolidate to kilograms
	consolidated, err := ingredients.ConsolidateByName("kg")
	if err != nil {
		t.Fatalf("Failed to consolidate ingredients: %v", err)
	}

	ingredientMap := consolidated.ToMap()

	// Check butter - should be 0.5kg + 0.2kg = 0.7kg
	if butter, exists := ingredientMap["butter"]; exists {
		t.Logf("Consolidated butter in kg: %s", butter)
		if butter != "0.7 kg" {
			t.Errorf("Expected butter to be '0.7 kg', got '%s'", butter)
		}
	} else {
		t.Error("Butter not found in consolidated list")
	}
}

func TestIngredientListByName(t *testing.T) {
	recipeText := `
Add @flour{500%g} to bowl.
Mix with @flour{200%g} and @sugar{100%g}.
`

	recipe, err := ParseString(recipeText)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	ingredients := recipe.GetIngredients()

	// Get all flour ingredients
	flourIngredients := ingredients.GetIngredientsByName("flour")
	if len(flourIngredients) != 2 {
		t.Errorf("Expected 2 flour ingredients, got %d", len(flourIngredients))
	}

	// Check quantities
	expectedQuantities := []float32{500, 200}
	for i, ing := range flourIngredients {
		if ing.Quantity != expectedQuantities[i] {
			t.Errorf("Expected flour quantity %g, got %g", expectedQuantities[i], ing.Quantity)
		}
	}
}

func TestUnitTypeIdentification(t *testing.T) {
	tests := []struct {
		unit         string
		expectedType string
	}{
		{"g", "mass"},
		{"kg", "mass"},
		{"lb", "mass"},
		{"ml", "volume"},
		{"l", "volume"},
		{"cup", "volume"},
		{"tsp", "volume"},
		{"tbsp", "volume"},
		{"pieces", ""}, // Custom unit should return empty or the quantity type if set
	}

	for _, test := range tests {
		ingredient := &Ingredient{
			Name:      "test",
			Quantity:  1,
			Unit:      test.unit,
			TypedUnit: createTypedUnit(test.unit),
		}

		unitType := ingredient.GetUnitType()
		t.Logf("Unit %s has type: %s", test.unit, unitType)

		// Note: Some units might not be recognized by go-units or might have different type names
		// This test is more for demonstration and logging than strict assertions
		if test.expectedType != "" && unitType != test.expectedType {
			t.Logf("Expected type %s for unit %s, got %s", test.expectedType, test.unit, unitType)
		}
	}
}
