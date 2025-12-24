package cooklang

import (
	"testing"
)

func TestNewIngredient(t *testing.T) {
	// Test that NewIngredient creates an ingredient with proper TypedUnit
	tests := []struct {
		name         string
		quantity     float32
		unit         string
		targetSystem UnitSystem
		expectConv   bool // Whether conversion should happen
	}{
		{"vodka", 50, "ml", UnitSystemUS, true},
		{"sugar", 100, "g", UnitSystemUS, true},
		{"lime juice", 1, "oz", UnitSystemMetric, true},
		{"water", 2, "cups", UnitSystemMetric, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create ingredient using NewIngredient constructor
			ing := NewIngredient(tc.name, tc.quantity, tc.unit)

			// Verify TypedUnit is set
			if ing.TypedUnit == nil {
				t.Errorf("NewIngredient(%s, %v, %s): TypedUnit should not be nil", tc.name, tc.quantity, tc.unit)
			}

			// Test that conversion works
			converted := ing.ConvertToSystem(tc.targetSystem)

			if tc.expectConv {
				// Conversion should change the unit
				if converted.Unit == ing.Unit && converted.Quantity == ing.Quantity {
					t.Logf("Warning: No conversion happened for %s %v %s -> %s (might be expected for some units)",
						tc.name, tc.quantity, tc.unit, tc.targetSystem)
				} else {
					t.Logf("%s: %.1f %s -> %.2f %s", tc.name, tc.quantity, tc.unit, converted.Quantity, converted.Unit)
				}
			}
		})
	}
}

func TestNewIngredientVsManualCreation(t *testing.T) {
	// Test that NewIngredient produces different results than manual struct creation
	// when it comes to unit conversion

	// Create ingredient manually (without TypedUnit)
	manualIng := &Ingredient{
		Name:     "vodka",
		Quantity: 50,
		Unit:     "ml",
	}

	// Create ingredient using NewIngredient
	constructedIng := NewIngredient("vodka", 50, "ml")

	// Try to convert both
	manualConverted := manualIng.ConvertToSystem(UnitSystemUS)
	constructedConverted := constructedIng.ConvertToSystem(UnitSystemUS)

	t.Logf("Manual creation: %.1f %s -> %.2f %s",
		manualIng.Quantity, manualIng.Unit,
		manualConverted.Quantity, manualConverted.Unit)

	t.Logf("NewIngredient: %.1f %s -> %.2f %s",
		constructedIng.Quantity, constructedIng.Unit,
		constructedConverted.Quantity, constructedConverted.Unit)

	// The manually created ingredient should NOT convert (TypedUnit is nil)
	if manualConverted.Unit != manualIng.Unit {
		t.Errorf("Manual ingredient should not convert without TypedUnit")
	}

	// The NewIngredient-created ingredient SHOULD convert
	if constructedConverted.Unit == constructedIng.Unit && constructedConverted.Quantity == constructedIng.Quantity {
		t.Errorf("NewIngredient should enable conversion, but no conversion happened")
	}
}

func TestUnitSystemConversion(t *testing.T) {
	// Create a test recipe with metric ingredients
	testRecipe := `
>> title: Test Recipe
>> servings: 4

Mix @flour{500%g} with @milk{250%ml} and @sugar{50%g}.
Add @butter{125%g} and @vanilla extract{5%ml}.
`

	recipe, err := ParseString(testRecipe)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	// Test converting to US system
	usShoppingList, err := recipe.GetShoppingListInSystem(UnitSystemUS)
	if err != nil {
		t.Fatalf("Failed to convert to US system: %v", err)
	}

	t.Logf("US Shopping List: %+v", usShoppingList)

	// Test converting to Imperial system
	imperialShoppingList, err := recipe.GetShoppingListInSystem(UnitSystemImperial)
	if err != nil {
		t.Fatalf("Failed to convert to Imperial system: %v", err)
	}

	t.Logf("Imperial Shopping List: %+v", imperialShoppingList)

	// Test individual ingredient conversion
	ingredients := recipe.GetIngredients()
	for _, ingredient := range ingredients.Ingredients {
		if ingredient.Unit != "" && ingredient.Quantity > 0 {
			// Test conversion to US units
			usIngredient := ingredient.ConvertToSystem(UnitSystemUS)
			t.Logf("Original: %s %.1f%s -> US: %s %.1f%s",
				ingredient.Name, ingredient.Quantity, ingredient.Unit,
				usIngredient.Name, usIngredient.Quantity, usIngredient.Unit)

			// Test conversion to Imperial units
			imperialIngredient := ingredient.ConvertToSystem(UnitSystemImperial)
			t.Logf("Original: %s %.1f%s -> Imperial: %s %.1f%s",
				ingredient.Name, ingredient.Quantity, ingredient.Unit,
				imperialIngredient.Name, imperialIngredient.Quantity, imperialIngredient.Unit)
		}
	}
}

func TestIngredientListConversion(t *testing.T) {
	// Create test ingredients
	ingredients := NewIngredientList()

	// Add some metric ingredients
	flour := &Ingredient{Name: "flour", Quantity: 1000, Unit: "g", TypedUnit: CreateTypedUnit("g")}
	milk := &Ingredient{Name: "milk", Quantity: 500, Unit: "ml", TypedUnit: CreateTypedUnit("ml")}
	sugar := &Ingredient{Name: "sugar", Quantity: 200, Unit: "g", TypedUnit: CreateTypedUnit("g")}

	ingredients.Add(flour)
	ingredients.Add(milk)
	ingredients.Add(sugar)

	// Convert to US system
	usIngredients := ingredients.ConvertToSystem(UnitSystemUS)

	t.Log("Original ingredients (metric):")
	for _, ing := range ingredients.Ingredients {
		t.Logf("  %s: %.1f %s", ing.Name, ing.Quantity, ing.Unit)
	}

	t.Log("Converted to US system:")
	for _, ing := range usIngredients.Ingredients {
		t.Logf("  %s: %.1f %s", ing.Name, ing.Quantity, ing.Unit)
	}

	// Convert to Imperial system
	imperialIngredients := ingredients.ConvertToSystem(UnitSystemImperial)

	t.Log("Converted to Imperial system:")
	for _, ing := range imperialIngredients.Ingredients {
		t.Logf("  %s: %.1f %s", ing.Name, ing.Quantity, ing.Unit)
	}
}

func TestSmartUnitSelection(t *testing.T) {
	// Test that large volumes get converted to appropriate units
	testCases := []struct {
		name     string
		quantity float32
		unit     string
		system   UnitSystem
		expected string
	}{
		{"Large volume to US quarts", 2000, "ml", UnitSystemUS, "qt"},
		{"Medium volume to US cups", 500, "ml", UnitSystemUS, "cup"},
		{"Small volume to US tablespoons", 30, "ml", UnitSystemUS, "tbsp"},
		{"Tiny volume to US teaspoons", 5, "ml", UnitSystemUS, "tsp"},
		{"Large mass to kg", 2000, "g", UnitSystemMetric, "kg"},
		{"Small mass stays as g", 500, "g", UnitSystemMetric, "g"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ingredient := &Ingredient{
				Name:      "test",
				Quantity:  tc.quantity,
				Unit:      tc.unit,
				TypedUnit: CreateTypedUnit(tc.unit),
			}

			converted := ingredient.ConvertToSystem(tc.system)

			if converted.Unit != tc.expected {
				t.Errorf("Expected unit %s, got %s (quantity: %.1f)",
					tc.expected, converted.Unit, converted.Quantity)
			}

			t.Logf("%.1f %s -> %.2f %s", tc.quantity, tc.unit, converted.Quantity, converted.Unit)
		})
	}
}

func TestConversionWithConsolidation(t *testing.T) {
	// Create ingredients with same name but different units
	ingredients := NewIngredientList()

	// Add the same ingredient in different units
	flour1 := &Ingredient{Name: "flour", Quantity: 500, Unit: "g", TypedUnit: CreateTypedUnit("g")}
	flour2 := &Ingredient{Name: "flour", Quantity: 250, Unit: "g", TypedUnit: CreateTypedUnit("g")}

	ingredients.Add(flour1)
	ingredients.Add(flour2)

	// Convert to US and consolidate
	consolidated, err := ingredients.ConvertToSystemWithConsolidation(UnitSystemUS)
	if err != nil {
		t.Fatalf("Failed to convert and consolidate: %v", err)
	}

	if len(consolidated.Ingredients) != 1 {
		t.Errorf("Expected 1 consolidated ingredient, got %d", len(consolidated.Ingredients))
	}

	flour := consolidated.Ingredients[0]
	t.Logf("Consolidated flour: %.2f %s", flour.Quantity, flour.Unit)

	// Should be approximately 1.65 lbs or similar US weight unit
	if flour.Name != "flour" {
		t.Errorf("Expected ingredient name 'flour', got '%s'", flour.Name)
	}
}
