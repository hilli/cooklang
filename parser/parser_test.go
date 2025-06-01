package parser

import (
	"testing"
)

func TestParseYAMLMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name: "Simple key-value pairs",
			input: `title: Simple Recipe
author: Chef Test
servings: 4`,
			expected: map[string]string{
				"title":    "Simple Recipe",
				"author":   "Chef Test",
				"servings": "4",
			},
		},
		{
			name: "Array format with brackets",
			input: `title: Test Recipe
tags: [ pasta, italian, comfort-food ]
ingredients: [chicken, rice, soy sauce]`,
			expected: map[string]string{
				"title":       "Test Recipe",
				"tags":        "pasta, italian, comfort-food",
				"ingredients": "chicken, rice, soy sauce",
			},
		},
		{
			name: "Array with extra spaces",
			input: `tags: [ spicy , asian,  quick-meal ]
diet: vegetarian`,
			expected: map[string]string{
				"tags": "spicy, asian, quick-meal",
				"diet": "vegetarian",
			},
		},
		{
			name: "Empty array",
			input: `tags: []
title: Empty Tags`,
			expected: map[string]string{
				"tags":  "",
				"title": "Empty Tags",
			},
		},
		{
			name: "Mixed single values and arrays",
			input: `title: Mixed Recipe
author: Chef Antonio
tags: [ italian, pasta ]
difficulty: Medium
prep_time: 10 minutes`,
			expected: map[string]string{
				"title":      "Mixed Recipe",
				"author":     "Chef Antonio",
				"tags":       "italian, pasta",
				"difficulty": "Medium",
				"prep_time":  "10 minutes",
			},
		},
		{
			name: "YAML list format",
			input: `title: List Recipe
tags:
  - spicy
  - asian
  - quick-meal
author: Chef Kim`,
			expected: map[string]string{
				"title":  "List Recipe",
				"tags":   "spicy, asian, quick-meal",
				"author": "Chef Kim",
			},
		},
		{
			name: "Mixed bracket and list format",
			input: `title: Mixed Recipe
ingredients: [tofu, broccoli]
tags:
  - vegetarian
  - healthy
difficulty: Medium`,
			expected: map[string]string{
				"title":       "Mixed Recipe",
				"ingredients": "tofu, broccoli",
				"tags":        "vegetarian, healthy",
				"difficulty":  "Medium",
			},
		},
	}

	parser := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseYAMLMetadata(tt.input)
			if err != nil {
				t.Fatalf("parseYAMLMetadata() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("parseYAMLMetadata() got %d items, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("parseYAMLMetadata() missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("parseYAMLMetadata() key %q = %q, want %q", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestParseRecipeWithArrayTags(t *testing.T) {
	recipeContent := `---
title: Spaghetti Carbonara
author: Chef Antonio
tags: [ pasta, italian, comfort-food ]
servings: 4
---

Add @pasta{400%grams} to boiling @water{}.`

	parser := New()
	recipe, err := parser.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Check metadata
	expectedMetadata := map[string]string{
		"title":    "Spaghetti Carbonara",
		"author":   "Chef Antonio",
		"tags":     "pasta, italian, comfort-food",
		"servings": "4",
	}

	for key, expectedValue := range expectedMetadata {
		if actualValue, exists := recipe.Metadata[key]; !exists {
			t.Errorf("Metadata missing key %q", key)
		} else if actualValue != expectedValue {
			t.Errorf("Metadata key %q = %q, want %q", key, actualValue, expectedValue)
		}
	}

	// Check that we have steps
	if len(recipe.Steps) == 0 {
		t.Error("Expected at least one step in the recipe")
	}

	// Check for ingredients
	foundPasta := false
	foundWater := false
	for _, step := range recipe.Steps {
		for _, component := range step.Components {
			if component.Type == "ingredient" {
				if component.Name == "pasta" {
					foundPasta = true
					if component.Quantity != "400" || component.Units != "grams" {
						t.Errorf("Pasta ingredient: got quantity=%q units=%q, want quantity=400 units=grams", component.Quantity, component.Units)
					}
				}
				if component.Name == "water" {
					foundWater = true
				}
			}
		}
	}

	if !foundPasta {
		t.Error("Expected to find pasta ingredient")
	}
	if !foundWater {
		t.Error("Expected to find water ingredient")
	}
}
