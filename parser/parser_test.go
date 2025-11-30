package parser

import (
	"testing"

	"github.com/hilli/cooklang/lexer"
	"github.com/hilli/cooklang/token"
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
					if component.Quantity != "400" || component.Unit != "grams" {
						t.Errorf("Pasta ingredient: got quantity=%q unit=%q, want quantity=400 unit=grams", component.Quantity, component.Unit)
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

func TestTextCompression(t *testing.T) {
	// Test that consecutive text elements are compressed into single elements
	recipeContent := `Add @salt{} and @pepper{} to taste.`

	parser := New()
	recipe, err := parser.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	if len(recipe.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(recipe.Steps))
	}

	step := recipe.Steps[0]

	// Expected components after compression:
	// 1. "Add " (text)
	// 2. salt ingredient
	// 3. " and " (text)
	// 4. pepper ingredient
	// 5. " to taste." (text)

	expectedComponents := 5
	if len(step.Components) != expectedComponents {
		t.Errorf("Expected %d components after compression, got %d", expectedComponents, len(step.Components))
		for i, comp := range step.Components {
			t.Logf("Component %d: Type=%s, Value=%q, Name=%q", i, comp.Type, comp.Value, comp.Name)
		}
	}

	// Check that text components contain expected values
	textComponents := []string{}
	for _, comp := range step.Components {
		if comp.Type == "text" {
			textComponents = append(textComponents, comp.Value)
		}
	}

	expectedTexts := []string{"Add ", " and ", " to taste."}
	if len(textComponents) != len(expectedTexts) {
		t.Errorf("Expected %d text components, got %d", len(expectedTexts), len(textComponents))
	}

	for i, expected := range expectedTexts {
		if i < len(textComponents) && textComponents[i] != expected {
			t.Errorf("Text component %d: expected %q, got %q", i, expected, textComponents[i])
		}
	}
}

func TestTextCompressionMultipleConsecutive(t *testing.T) {
	// Test compression of multiple consecutive text tokens
	recipeContent := `Boil water for cooking pasta.`

	parser := New()
	recipe, err := parser.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	if len(recipe.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(recipe.Steps))
	}

	step := recipe.Steps[0]

	// Should be compressed into a single text component
	if len(step.Components) != 1 {
		t.Errorf("Expected 1 component after compression, got %d", len(step.Components))
		for i, comp := range step.Components {
			t.Logf("Component %d: Type=%s, Value=%q", i, comp.Type, comp.Value)
		}
	}

	if step.Components[0].Type != "text" {
		t.Errorf("Expected text component, got %s", step.Components[0].Type)
	}

	expectedText := "Boil water for cooking pasta."
	if step.Components[0].Value != expectedText {
		t.Errorf("Expected compressed text %q, got %q", expectedText, step.Components[0].Value)
	}
}

func TestDebugCookware(t *testing.T) {
	// Test just the lexer first
	input := "#pan{}"
	l := lexer.New(input)

	tok1 := l.NextToken() // Should be #
	tok2 := l.NextToken() // Should be "pan"
	tok3 := l.NextToken() // Should be {
	tok4 := l.NextToken() // Should be }
	tok5 := l.NextToken() // Should be EOF

	t.Logf("Token 1: Type=%s, Literal='%s'", tok1.Type, tok1.Literal)
	t.Logf("Token 2: Type=%s, Literal='%s'", tok2.Type, tok2.Literal)
	t.Logf("Token 3: Type=%s, Literal='%s'", tok3.Type, tok3.Literal)
	t.Logf("Token 4: Type=%s, Literal='%s'", tok4.Type, tok4.Literal)
	t.Logf("Token 5: Type=%s, Literal='%s'", tok5.Type, tok5.Literal)
}

func TestDebugMultiWordCookware(t *testing.T) {
	// Test multi-word cookware
	input := "#large skillet{1}"
	l := lexer.New(input)

	tokenCount := 0
	for {
		tok := l.NextToken()
		tokenCount++
		t.Logf("Token %d: Type=%s, Literal='%s'", tokenCount, tok.Type, tok.Literal)
		if tok.Type == token.EOF {
			break
		}
		if tokenCount > 10 { // Safety break
			t.Fatal("Too many tokens, possible infinite loop")
		}
	}
}

func TestDebugCookwareParser(t *testing.T) {
	// Test the actual parser
	p := New()
	recipe, err := p.ParseString("#large skillet{1}")
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	t.Logf("Number of steps: %d", len(recipe.Steps))
	for i, step := range recipe.Steps {
		t.Logf("Step %d has %d components", i, len(step.Components))
		for j, comp := range step.Components {
			t.Logf("  Component %d: Type=%s, Name='%s', Value='%s', Quantity='%s'",
				j, comp.Type, comp.Name, comp.Value, comp.Quantity)
		}
	}
}

func TestDebugFullSentence(t *testing.T) {
	// Test the full sentence
	p := New()
	recipe, err := p.ParseString("Heat oil in a #large skillet{1} over medium heat.")
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	t.Logf("Number of steps: %d", len(recipe.Steps))
	for i, step := range recipe.Steps {
		t.Logf("Step %d has %d components", i, len(step.Components))
		for j, comp := range step.Components {
			t.Logf("  Component %d: Type=%s, Name='%s', Value='%s', Quantity='%s'",
				j, comp.Type, comp.Name, comp.Value, comp.Quantity)
		}
	}

	// Verify we have a cookware component
	found := false
	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "cookware" && comp.Name == "large skillet" && comp.Quantity == "1" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Expected cookware component 'large skillet' with quantity '1' not found")
	}
}

func TestCookwareDefaultQuantity(t *testing.T) {
	p := New()

	// Test cookware with empty braces (should default to "1")
	recipe1, err := p.ParseString("#frying pan{}")
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	// Check that default quantity is "1" for cookware
	found := false
	for _, step := range recipe1.Steps {
		for _, comp := range step.Components {
			if comp.Type == "cookware" && comp.Name == "frying pan" {
				found = true
				if comp.Quantity != "1" {
					t.Errorf("Expected cookware default quantity '1', got '%s'", comp.Quantity)
				}
			}
		}
	}
	if !found {
		t.Error("Cookware component not found")
	}

	// Test cookware without braces (should also default to "1")
	recipe2, err := p.ParseString("Simmer in #pan for some time")
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	found = false
	for _, step := range recipe2.Steps {
		for _, comp := range step.Components {
			if comp.Type == "cookware" && comp.Name == "pan" {
				found = true
				if comp.Quantity != "1" {
					t.Errorf("Expected cookware default quantity '1', got '%s'", comp.Quantity)
				}
			}
		}
	}
	if !found {
		t.Error("Cookware component not found")
	}
}

func TestDebugCookwareWithoutBraces(t *testing.T) {
	p := New()
	recipe, err := p.ParseString("Simmer in #pan for some time")
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	t.Logf("Number of steps: %d", len(recipe.Steps))
	for i, step := range recipe.Steps {
		t.Logf("Step %d has %d components", i, len(step.Components))
		for j, comp := range step.Components {
			t.Logf("  Component %d: Type=%s, Name='%s', Value='%s', Quantity='%s'",
				j, comp.Type, comp.Name, comp.Value, comp.Quantity)
		}
	}
}

// TestFractionParsing tests that fractions and mixed fractions are correctly converted to decimals
func TestFractionParsing(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedQuantity string
		expectedUnit     string
	}{
		{
			name:             "Simple fraction 1/2",
			input:            "@gin{1/2%fl oz}",
			expectedQuantity: "0.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Simple fraction 1/4",
			input:            "@gin{1/4%fl oz}",
			expectedQuantity: "0.25",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Simple fraction 3/4",
			input:            "@gin{3/4%fl oz}",
			expectedQuantity: "0.75",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Simple fraction with spaces 1 / 2",
			input:            "@gin{1 / 2%fl oz}",
			expectedQuantity: "0.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Mixed fraction 1 1/2",
			input:            "@gin{1 1/2%fl oz}",
			expectedQuantity: "1.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Mixed fraction 2 1/2",
			input:            "@gin{2 1/2%fl oz}",
			expectedQuantity: "2.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Mixed fraction 2 3/4",
			input:            "@gin{2 3/4%fl oz}",
			expectedQuantity: "2.75",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Mixed fraction with spaces 1 1 / 2",
			input:            "@gin{1 1 / 2%fl oz}",
			expectedQuantity: "1.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Whole number",
			input:            "@gin{2%fl oz}",
			expectedQuantity: "2",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Decimal",
			input:            "@gin{1.5%fl oz}",
			expectedQuantity: "1.5",
			expectedUnit:     "fl oz",
		},
		// Unicode fraction tests
		{
			name:             "Unicode fraction one half",
			input:            "@gin{½%fl oz}",
			expectedQuantity: "0.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction one quarter",
			input:            "@gin{¼%fl oz}",
			expectedQuantity: "0.25",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction three quarters",
			input:            "@gin{¾%fl oz}",
			expectedQuantity: "0.75",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode mixed fraction 1½",
			input:            "@gin{1½%fl oz}",
			expectedQuantity: "1.5",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode mixed fraction 2¼",
			input:            "@gin{2¼%fl oz}",
			expectedQuantity: "2.25",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction one third",
			input:            "@gin{⅓%fl oz}",
			expectedQuantity: "0.3333333333333333",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction two thirds",
			input:            "@gin{⅔%fl oz}",
			expectedQuantity: "0.6666666666666666",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction one eighth",
			input:            "@gin{⅛%fl oz}",
			expectedQuantity: "0.125",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction three eighths",
			input:            "@gin{⅜%fl oz}",
			expectedQuantity: "0.375",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction five eighths",
			input:            "@gin{⅝%fl oz}",
			expectedQuantity: "0.625",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction seven eighths",
			input:            "@gin{⅞%fl oz}",
			expectedQuantity: "0.875",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction one fifth",
			input:            "@gin{⅕%fl oz}",
			expectedQuantity: "0.2",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction two fifths",
			input:            "@gin{⅖%fl oz}",
			expectedQuantity: "0.4",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction three fifths",
			input:            "@gin{⅗%fl oz}",
			expectedQuantity: "0.6",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction four fifths",
			input:            "@gin{⅘%fl oz}",
			expectedQuantity: "0.8",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction one sixth",
			input:            "@gin{⅙%fl oz}",
			expectedQuantity: "0.16666666666666666",
			expectedUnit:     "fl oz",
		},
		{
			name:             "Unicode fraction five sixths",
			input:            "@gin{⅚%fl oz}",
			expectedQuantity: "0.8333333333333334",
			expectedUnit:     "fl oz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := New()
			recipe, err := parser.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse recipe: %v", err)
			}

			if len(recipe.Steps) != 1 {
				t.Fatalf("Expected 1 step, got %d", len(recipe.Steps))
			}

			// Find the ingredient component
			var ingredientComp *Component
			for _, comp := range recipe.Steps[0].Components {
				if comp.Type == "ingredient" {
					ingredientComp = &comp
					break
				}
			}

			if ingredientComp == nil {
				t.Fatal("No ingredient component found")
			}

			if ingredientComp.Quantity != tt.expectedQuantity {
				t.Errorf("Expected quantity %q, got %q", tt.expectedQuantity, ingredientComp.Quantity)
			}

			if ingredientComp.Unit != tt.expectedUnit {
				t.Errorf("Expected unit %q, got %q", tt.expectedUnit, ingredientComp.Unit)
			}
		})
	}
}

// TestStepSeparationNewlineVariants tests that step separation works correctly
// with Unix LF (\n), Windows CRLF (\r\n), and old Mac CR (\r) line endings
func TestStepSeparationNewlineVariants(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSteps int
		stepContents  []string
	}{
		{
			name:          "Unix LF blank line",
			input:         "Step one.\n\nStep two.",
			expectedSteps: 2,
			stepContents:  []string{"Step one.", "Step two."},
		},
		{
			name:          "Windows CRLF blank line",
			input:         "Step one.\r\n\r\nStep two.",
			expectedSteps: 2,
			stepContents:  []string{"Step one.", "Step two."},
		},
		{
			name:          "Old Mac CR blank line",
			input:         "Step one.\r\rStep two.",
			expectedSteps: 2,
			stepContents:  []string{"Step one.", "Step two."},
		},
		{
			name:          "Single Unix LF (no step break)",
			input:         "Step one.\nStep two.",
			expectedSteps: 1,
			stepContents:  []string{"Step one. Step two."},
		},
		{
			name:          "Single Windows CRLF (no step break)",
			input:         "Step one.\r\nStep two.",
			expectedSteps: 1,
			stepContents:  []string{"Step one. Step two."},
		},
		{
			name:          "Single Old Mac CR (no step break)",
			input:         "Step one.\rStep two.",
			expectedSteps: 1,
			stepContents:  []string{"Step one. Step two."},
		},
		{
			name:          "Mixed: CRLF then LF blank lines",
			input:         "Step one.\r\n\r\nStep two.\n\nStep three.",
			expectedSteps: 3,
			stepContents:  []string{"Step one.", "Step two.", "Step three."},
		},
		{
			name:          "Multiple blank lines (Unix) collapse to one step break",
			input:         "Step one.\n\n\n\nStep two.",
			expectedSteps: 2,
			stepContents:  []string{"Step one.", "Step two."},
		},
		{
			name:          "Multiple blank lines (Windows) collapse to one step break",
			input:         "Step one.\r\n\r\n\r\n\r\nStep two.",
			expectedSteps: 2,
			stepContents:  []string{"Step one.", "Step two."},
		},
		{
			name:          "With ingredients - Unix LF",
			input:         "Add @salt{1%tsp}.\n\nStir well.",
			expectedSteps: 2,
		},
		{
			name:          "With ingredients - Windows CRLF",
			input:         "Add @salt{1%tsp}.\r\n\r\nStir well.",
			expectedSteps: 2,
		},
		{
			name:          "With ingredients - Old Mac CR",
			input:         "Add @salt{1%tsp}.\r\rStir well.",
			expectedSteps: 2,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if len(recipe.Steps) != tt.expectedSteps {
				t.Errorf("Expected %d steps, got %d", tt.expectedSteps, len(recipe.Steps))
				for i, step := range recipe.Steps {
					var content string
					for _, comp := range step.Components {
						content += comp.Value + comp.Name
					}
					t.Logf("Step %d: %q", i+1, content)
				}
			}
			// Check step contents if provided
			for i, expected := range tt.stepContents {
				if i >= len(recipe.Steps) {
					break
				}
				var actual string
				for _, comp := range recipe.Steps[i].Components {
					actual += comp.Value
				}
				if actual != expected {
					t.Errorf("Step %d: expected %q, got %q", i+1, expected, actual)
				}
			}
		})
	}
}

// TestYAMLFrontmatterWithCRLF tests that YAML frontmatter works with Windows line endings
func TestYAMLFrontmatterWithCRLF(t *testing.T) {
	// Windows-style CRLF line endings in YAML frontmatter
	input := "---\r\ntitle: Test Recipe\r\nauthor: Chef\r\n---\r\n\r\nCook the @pasta{500%g}."

	p := New()
	recipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse recipe with CRLF frontmatter: %v", err)
	}

	if recipe.Metadata["title"] != "Test Recipe" {
		t.Errorf("Expected title 'Test Recipe', got %q", recipe.Metadata["title"])
	}

	if len(recipe.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(recipe.Steps))
	}
}

// TestCommentsWithCRLF tests that comments work correctly with different line endings
func TestCommentsWithCRLF(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSteps int
	}{
		{
			name:          "Comment with Unix LF",
			input:         "-- This is a comment\nCook something.",
			expectedSteps: 1,
		},
		{
			name:          "Comment with Windows CRLF",
			input:         "-- This is a comment\r\nCook something.",
			expectedSteps: 1,
		},
		{
			name:          "Comment with Old Mac CR",
			input:         "-- This is a comment\rCook something.",
			expectedSteps: 1,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if len(recipe.Steps) != tt.expectedSteps {
				t.Errorf("Expected %d step(s), got %d", tt.expectedSteps, len(recipe.Steps))
			}
		})
	}
}
