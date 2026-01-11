package cooklang

import (
	"math"
	"testing"
)

func TestDefaultServings(t *testing.T) {
	tests := []struct {
		name             string
		recipe           string
		expectedServings float32
	}{
		{
			name: "explicit servings",
			recipe: `---
servings: 4
---
Add @flour{500%g}.`,
			expectedServings: 4,
		},
		{
			name:             "no servings defaults to 1",
			recipe:           `Add @flour{500%g}.`,
			expectedServings: 1,
		},
		{
			name: "zero servings defaults to 1",
			recipe: `---
servings: 0
---
Add @flour{500%g}.`,
			expectedServings: 1,
		},
		{
			name: "negative servings defaults to 1",
			recipe: `---
servings: -2
---
Add @flour{500%g}.`,
			expectedServings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := ParseString(tt.recipe)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}
			if recipe.Servings != tt.expectedServings {
				t.Errorf("Servings = %v, want %v", recipe.Servings, tt.expectedServings)
			}
		})
	}
}

func TestCreateShoppingListForServings(t *testing.T) {
	// Recipe with 2 servings
	recipe2Servings := `---
servings: 2
---
Add @flour{500%g} and @sugar{100%g}.`

	// Recipe with 4 servings
	recipe4Servings := `---
servings: 4
---
Add @flour{200%g} and @butter{50%g}.`

	// Recipe with no servings (defaults to 1)
	recipe1Serving := `Add @salt{5%g}.`

	r2, err := ParseString(recipe2Servings)
	if err != nil {
		t.Fatalf("Failed to parse recipe2Servings: %v", err)
	}

	r4, err := ParseString(recipe4Servings)
	if err != nil {
		t.Fatalf("Failed to parse recipe4Servings: %v", err)
	}

	r1, err := ParseString(recipe1Serving)
	if err != nil {
		t.Fatalf("Failed to parse recipe1Serving: %v", err)
	}

	t.Run("scale to 4 servings", func(t *testing.T) {
		list, err := CreateShoppingListForServings(4, r2, r4, r1)
		if err != nil {
			t.Fatalf("CreateShoppingListForServings() error = %v", err)
		}

		ingredients := list.Ingredients.Ingredients

		// Find flour: r2 has 500g for 2 servings (scaled 2x = 1000g)
		//             r4 has 200g for 4 servings (scaled 1x = 200g)
		//             Total: 1200g
		flour := findIngredient(ingredients, "flour")
		if flour == nil {
			t.Error("flour not found in shopping list")
		} else if !floatClose(float64(flour.Quantity), 1200, 0.1) {
			t.Errorf("flour quantity = %v, want 1200", flour.Quantity)
		}

		// Sugar: r2 has 100g for 2 servings (scaled 2x = 200g)
		sugar := findIngredient(ingredients, "sugar")
		if sugar == nil {
			t.Error("sugar not found in shopping list")
		} else if !floatClose(float64(sugar.Quantity), 200, 0.1) {
			t.Errorf("sugar quantity = %v, want 200", sugar.Quantity)
		}

		// Butter: r4 has 50g for 4 servings (scaled 1x = 50g)
		butter := findIngredient(ingredients, "butter")
		if butter == nil {
			t.Error("butter not found in shopping list")
		} else if !floatClose(float64(butter.Quantity), 50, 0.1) {
			t.Errorf("butter quantity = %v, want 50", butter.Quantity)
		}

		// Salt: r1 has 5g for 1 serving (scaled 4x = 20g)
		salt := findIngredient(ingredients, "salt")
		if salt == nil {
			t.Error("salt not found in shopping list")
		} else if !floatClose(float64(salt.Quantity), 20, 0.1) {
			t.Errorf("salt quantity = %v, want 20", salt.Quantity)
		}
	})

	t.Run("scale to 1 serving", func(t *testing.T) {
		list, err := CreateShoppingListForServings(1, r2, r4)
		if err != nil {
			t.Fatalf("CreateShoppingListForServings() error = %v", err)
		}

		ingredients := list.Ingredients.Ingredients

		// Flour: r2 has 500g for 2 servings (scaled 0.5x = 250g)
		//        r4 has 200g for 4 servings (scaled 0.25x = 50g)
		//        Total: 300g
		flour := findIngredient(ingredients, "flour")
		if flour == nil {
			t.Error("flour not found in shopping list")
		} else if !floatClose(float64(flour.Quantity), 300, 0.1) {
			t.Errorf("flour quantity = %v, want 300", flour.Quantity)
		}
	})

	t.Run("empty recipes", func(t *testing.T) {
		list, err := CreateShoppingListForServings(4)
		if err != nil {
			t.Fatalf("CreateShoppingListForServings() error = %v", err)
		}
		if list.Count() != 0 {
			t.Errorf("expected empty list, got %d ingredients", list.Count())
		}
	})
}

func TestCreateShoppingListForServingsWithUnit(t *testing.T) {
	recipe := `---
servings: 2
---
Add @flour{500%g} and @butter{100%g}.`

	r, err := ParseString(recipe)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	t.Run("scale to 4 servings with kg conversion", func(t *testing.T) {
		list, err := CreateShoppingListForServingsWithUnit(4, "kg", r)
		if err != nil {
			t.Fatalf("CreateShoppingListForServingsWithUnit() error = %v", err)
		}

		ingredients := list.Ingredients.Ingredients

		// Flour: 500g for 2 servings (scaled 2x = 1000g = 1kg)
		flour := findIngredient(ingredients, "flour")
		if flour == nil {
			t.Error("flour not found in shopping list")
		} else {
			if flour.Unit != "kg" {
				t.Errorf("flour unit = %v, want kg", flour.Unit)
			}
			if !floatClose(float64(flour.Quantity), 1, 0.01) {
				t.Errorf("flour quantity = %v, want 1", flour.Quantity)
			}
		}

		// Butter: 100g for 2 servings (scaled 2x = 200g = 0.2kg)
		butter := findIngredient(ingredients, "butter")
		if butter == nil {
			t.Error("butter not found in shopping list")
		} else {
			if butter.Unit != "kg" {
				t.Errorf("butter unit = %v, want kg", butter.Unit)
			}
			if !floatClose(float64(butter.Quantity), 0.2, 0.01) {
				t.Errorf("butter quantity = %v, want 0.2", butter.Quantity)
			}
		}
	})
}

func TestScaleToServings(t *testing.T) {
	recipe := `---
servings: 2
---
Add @flour{500%g} and @salt{=10%g}.`

	r, err := ParseString(recipe)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	t.Run("scale from 2 to 4 servings", func(t *testing.T) {
		scaled := r.ScaleToServings(4)

		if scaled.Servings != 4 {
			t.Errorf("scaled servings = %v, want 4", scaled.Servings)
		}

		ingredients := scaled.GetIngredients().Ingredients

		// Flour should be scaled (500g * 2 = 1000g)
		flour := findIngredient(ingredients, "flour")
		if flour == nil {
			t.Error("flour not found")
		} else if !floatClose(float64(flour.Quantity), 1000, 0.1) {
			t.Errorf("flour quantity = %v, want 1000", flour.Quantity)
		}

		// Salt is fixed (=10g), should not scale
		salt := findIngredient(ingredients, "salt")
		if salt == nil {
			t.Error("salt not found")
		} else if !floatClose(float64(salt.Quantity), 10, 0.1) {
			t.Errorf("salt quantity = %v, want 10 (fixed)", salt.Quantity)
		}
	})

	t.Run("scale from 2 to 1 serving", func(t *testing.T) {
		scaled := r.ScaleToServings(1)

		if scaled.Servings != 1 {
			t.Errorf("scaled servings = %v, want 1", scaled.Servings)
		}

		ingredients := scaled.GetIngredients().Ingredients

		// Flour should be scaled (500g * 0.5 = 250g)
		flour := findIngredient(ingredients, "flour")
		if flour == nil {
			t.Error("flour not found")
		} else if !floatClose(float64(flour.Quantity), 250, 0.1) {
			t.Errorf("flour quantity = %v, want 250", flour.Quantity)
		}
	})
}

// Helper functions

func findIngredient(ingredients []*Ingredient, name string) *Ingredient {
	for _, ing := range ingredients {
		if ing.Name == name {
			return ing
		}
	}
	return nil
}

func floatClose(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
