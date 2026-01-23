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

// TestBlockComments tests block comment parsing [- comment -]
func TestBlockComments(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		extendedMode       bool
		expectedSteps      int
		expectBlockComment bool
		blockCommentValue  string
	}{
		{
			name:               "Block comment ignored in canonical mode",
			input:              "Add @milk{4%cup} [- TODO change units to litres -], keep mixing",
			extendedMode:       false,
			expectedSteps:      1,
			expectBlockComment: false,
		},
		{
			name:               "Block comment preserved in extended mode",
			input:              "Add @milk{4%cup} [- TODO change units to litres -], keep mixing",
			extendedMode:       true,
			expectedSteps:      1,
			expectBlockComment: true,
			blockCommentValue:  "TODO change units to litres",
		},
		{
			name:               "Multiple block comments in extended mode",
			input:              "[- comment 1 -] Add @salt{1%tsp} [- comment 2 -] and stir",
			extendedMode:       true,
			expectedSteps:      1,
			expectBlockComment: true,
		},
		{
			name:               "Empty block comment",
			input:              "Add salt [-  -] and pepper",
			extendedMode:       true,
			expectedSteps:      1,
			expectBlockComment: true,
			blockCommentValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			p.ExtendedMode = tt.extendedMode

			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(recipe.Steps) != tt.expectedSteps {
				t.Errorf("Expected %d step(s), got %d", tt.expectedSteps, len(recipe.Steps))
			}

			// Check for block comment component
			foundBlockComment := false
			var blockCommentValue string
			for _, step := range recipe.Steps {
				for _, comp := range step.Components {
					if comp.Type == "blockComment" {
						foundBlockComment = true
						blockCommentValue = comp.Value
					}
				}
			}

			if foundBlockComment != tt.expectBlockComment {
				t.Errorf("Expected block comment present=%v, got %v", tt.expectBlockComment, foundBlockComment)
			}

			if tt.expectBlockComment && tt.blockCommentValue != "" && blockCommentValue != tt.blockCommentValue {
				t.Errorf("Expected block comment value %q, got %q", tt.blockCommentValue, blockCommentValue)
			}
		})
	}
}

// TestSectionParsing tests that sections are correctly parsed
func TestSectionParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSteps int
		sectionName   string
	}{
		{
			name:          "Single section with content",
			input:         "= Dough\nMix @flour{200%g} with water.",
			expectedSteps: 1,
			sectionName:   "Dough",
		},
		{
			name:          "Section with double =",
			input:         "== Filling ==\nCombine @cheese{100%g} and @spinach{50%g}.",
			expectedSteps: 1,
			sectionName:   "Filling",
		},
		{
			name:          "Content before section",
			input:         "Prep the ingredients.\n\n= Main Steps\nCook the @pasta{400%g}.",
			expectedSteps: 2,
			sectionName:   "Main Steps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(recipe.Steps) != tt.expectedSteps {
				t.Errorf("Expected %d steps, got %d", tt.expectedSteps, len(recipe.Steps))
				for i, step := range recipe.Steps {
					t.Logf("Step %d has %d components", i, len(step.Components))
					for j, comp := range step.Components {
						t.Logf("  Component %d: Type=%s, Name=%q, Value=%q", j, comp.Type, comp.Name, comp.Value)
					}
				}
			}

			// Check for section component
			foundSection := false
			var sectionName string
			for _, step := range recipe.Steps {
				for _, comp := range step.Components {
					if comp.Type == "section" {
						foundSection = true
						sectionName = comp.Name
					}
				}
			}

			if !foundSection {
				t.Error("Expected to find a section component")
			}

			if sectionName != tt.sectionName {
				t.Errorf("Expected section name %q, got %q", tt.sectionName, sectionName)
			}
		})
	}
}

// TestMultipleSections tests that multiple sections are correctly parsed
func TestMultipleSections(t *testing.T) {
	input := `= Dough
Mix @flour{200%g} and @water{100%ml} until smooth.

== Filling ==
Combine @cheese{100%g} and @spinach{50%g}.

=== Assembly ===
Place filling on dough and roll.`

	p := New()
	recipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have 3 steps (one per section)
	if len(recipe.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(recipe.Steps))
		for i, step := range recipe.Steps {
			t.Logf("Step %d:", i)
			for j, comp := range step.Components {
				t.Logf("  Component %d: Type=%s, Name=%q, Value=%q", j, comp.Type, comp.Name, comp.Value)
			}
		}
	}

	// Find all section names
	sectionNames := []string{}
	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "section" {
				sectionNames = append(sectionNames, comp.Name)
			}
		}
	}

	expectedSections := []string{"Dough", "Filling", "Assembly"}
	if len(sectionNames) != len(expectedSections) {
		t.Errorf("Expected %d sections, got %d", len(expectedSections), len(sectionNames))
	}

	for i, expected := range expectedSections {
		if i < len(sectionNames) && sectionNames[i] != expected {
			t.Errorf("Section %d: expected %q, got %q", i, expected, sectionNames[i])
		}
	}
}

// TestSectionWithIngredients tests that sections work correctly with ingredients
func TestSectionWithIngredients(t *testing.T) {
	input := `= Preparation
Add @salt{1%tsp} and @pepper{} to taste.`

	p := New()
	recipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find section and ingredients
	foundSection := false
	foundSalt := false
	foundPepper := false

	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "section" && comp.Name == "Preparation" {
				foundSection = true
			}
			if comp.Type == "ingredient" && comp.Name == "salt" {
				foundSalt = true
				if comp.Quantity != "1" || comp.Unit != "tsp" {
					t.Errorf("Salt: expected quantity=1 unit=tsp, got quantity=%s unit=%s",
						comp.Quantity, comp.Unit)
				}
			}
			if comp.Type == "ingredient" && comp.Name == "pepper" {
				foundPepper = true
			}
		}
	}

	if !foundSection {
		t.Error("Expected to find section 'Preparation'")
	}
	if !foundSalt {
		t.Error("Expected to find ingredient 'salt'")
	}
	if !foundPepper {
		t.Error("Expected to find ingredient 'pepper'")
	}
}

// TestSectionWithCRLF tests that sections work with Windows line endings
func TestSectionWithCRLF(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		sectionName string
	}{
		{
			name:        "Section with Unix LF",
			input:       "= Dough\nMix flour.",
			sectionName: "Dough",
		},
		{
			name:        "Section with Windows CRLF",
			input:       "= Dough\r\nMix flour.",
			sectionName: "Dough",
		},
		{
			name:        "Section with Old Mac CR",
			input:       "= Dough\rMix flour.",
			sectionName: "Dough",
		},
		{
			name:        "Section after CRLF newline",
			input:       "Step one.\r\n\r\n= Section\r\nStep two.",
			sectionName: "Section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// Find section
			foundSection := false
			var sectionName string
			for _, step := range recipe.Steps {
				for _, comp := range step.Components {
					if comp.Type == "section" {
						foundSection = true
						sectionName = comp.Name
					}
				}
			}

			if !foundSection {
				t.Error("Expected to find a section")
			}

			if sectionName != tt.sectionName {
				t.Errorf("Expected section name %q, got %q", tt.sectionName, sectionName)
			}
		})
	}
}

// TestBlockCommentWithIngredients tests that block comments work alongside ingredients
func TestBlockCommentWithIngredients(t *testing.T) {
	p := New()
	p.ExtendedMode = true

	input := "Slowly add @milk{4%cup} [- TODO change units to litres -], keep mixing"
	recipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have milk ingredient
	foundMilk := false
	foundBlockComment := false
	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "ingredient" && comp.Name == "milk" {
				foundMilk = true
				if comp.Quantity != "4" || comp.Unit != "cup" {
					t.Errorf("Milk ingredient: expected quantity=4, unit=cup, got quantity=%s, unit=%s",
						comp.Quantity, comp.Unit)
				}
			}
			if comp.Type == "blockComment" {
				foundBlockComment = true
			}
		}
	}

	if !foundMilk {
		t.Error("Expected to find milk ingredient")
	}
	if !foundBlockComment {
		t.Error("Expected to find block comment")
	}
}

// TestNoteParsing tests that notes are correctly parsed
func TestNoteParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSteps int
		noteText      string
	}{
		{
			name:          "Single note",
			input:         "> This is a helpful tip.",
			expectedSteps: 1,
			noteText:      "This is a helpful tip.",
		},
		{
			name:          "Note followed by content",
			input:         "> A tip for the cook.\n\nMix @flour{200%g} with water.",
			expectedSteps: 2,
			noteText:      "A tip for the cook.",
		},
		{
			name:          "Content followed by note",
			input:         "Mix @flour{200%g} with water.\n\n> This dish tastes better the next day.",
			expectedSteps: 2,
			noteText:      "This dish tastes better the next day.",
		},
		{
			name:          "Multi-line note",
			input:         "> First line of the note\n> Second line continues here.",
			expectedSteps: 1,
			noteText:      "First line of the note Second line continues here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(recipe.Steps) != tt.expectedSteps {
				t.Errorf("Expected %d steps, got %d", tt.expectedSteps, len(recipe.Steps))
				for i, step := range recipe.Steps {
					t.Logf("Step %d has %d components", i, len(step.Components))
					for j, comp := range step.Components {
						t.Logf("  Component %d: Type=%s, Value=%q", j, comp.Type, comp.Value)
					}
				}
			}

			// Find note component
			foundNote := false
			var noteValue string
			for _, step := range recipe.Steps {
				for _, comp := range step.Components {
					if comp.Type == "note" {
						foundNote = true
						noteValue = comp.Value
						break
					}
				}
			}

			if !foundNote {
				t.Error("Expected to find a note component")
			}

			if noteValue != tt.noteText {
				t.Errorf("Expected note text %q, got %q", tt.noteText, noteValue)
			}
		})
	}
}

// TestMultipleNotes tests that multiple notes are correctly parsed
func TestMultipleNotes(t *testing.T) {
	input := `> First helpful tip.

Mix @flour{200%g} and @water{100%ml}.

> Second tip for serving.

> Third tip about storage.`

	p := New()
	recipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Count notes
	noteCount := 0
	noteTexts := []string{}
	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "note" {
				noteCount++
				noteTexts = append(noteTexts, comp.Value)
			}
		}
	}

	expectedNotes := []string{
		"First helpful tip.",
		"Second tip for serving.",
		"Third tip about storage.",
	}

	if noteCount != len(expectedNotes) {
		t.Errorf("Expected %d notes, got %d", len(expectedNotes), noteCount)
		t.Logf("Found notes: %v", noteTexts)
	}

	for i, expected := range expectedNotes {
		if i < len(noteTexts) && noteTexts[i] != expected {
			t.Errorf("Note %d: expected %q, got %q", i, expected, noteTexts[i])
		}
	}
}

// TestNoteWithSections tests that notes work correctly with sections
func TestNoteWithSections(t *testing.T) {
	input := `= Preparation
> Prep tip: mise en place makes everything easier.

Add @salt{1%tsp} to the bowl.

= Cooking
> Cooking tip: don't rush this step.

Cook for ~{10%minutes}.`

	p := New()
	recipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find sections and notes
	foundPrepSection := false
	foundCookingSection := false
	foundPrepNote := false
	foundCookingNote := false

	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "section" && comp.Name == "Preparation" {
				foundPrepSection = true
			}
			if comp.Type == "section" && comp.Name == "Cooking" {
				foundCookingSection = true
			}
			if comp.Type == "note" && comp.Value == "Prep tip: mise en place makes everything easier." {
				foundPrepNote = true
			}
			if comp.Type == "note" && comp.Value == "Cooking tip: don't rush this step." {
				foundCookingNote = true
			}
		}
	}

	if !foundPrepSection {
		t.Error("Expected to find Preparation section")
	}
	if !foundCookingSection {
		t.Error("Expected to find Cooking section")
	}
	if !foundPrepNote {
		t.Error("Expected to find prep note")
	}
	if !foundCookingNote {
		t.Error("Expected to find cooking note")
	}
}

// TestOptionalIngredient tests parsing of optional ingredients with @? syntax
func TestOptionalIngredient(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Component
	}{
		{
			name:  "simple optional ingredient with quantity",
			input: "Add @?thyme{2%sprigs} if desired.",
			expected: []Component{
				{Type: "text", Value: "Add "},
				{Type: "ingredient", Name: "thyme", Quantity: "2", Unit: "sprigs", Optional: true},
				{Type: "text", Value: " if desired."},
			},
		},
		{
			name:  "optional ingredient without braces",
			input: "Garnish with @?parsley.",
			expected: []Component{
				{Type: "text", Value: "Garnish with "},
				{Type: "ingredient", Name: "parsley", Quantity: "some", Optional: true},
				{Type: "text", Value: "."},
			},
		},
		{
			name:  "optional ingredient with empty braces",
			input: "Add @?herbs{} for flavor.",
			expected: []Component{
				{Type: "text", Value: "Add "},
				{Type: "ingredient", Name: "herbs", Quantity: "some", Optional: true},
				{Type: "text", Value: " for flavor."},
			},
		},
		{
			name:  "optional and fixed combined",
			input: "Add @?salt{=1%pinch} to taste.",
			expected: []Component{
				{Type: "text", Value: "Add "},
				{Type: "ingredient", Name: "salt", Quantity: "1", Unit: "pinch", Optional: true, Fixed: true},
				{Type: "text", Value: " to taste."},
			},
		},
		{
			name:  "mixed regular and optional ingredients",
			input: "Mix @flour{500%g} with @?herbs{}.",
			expected: []Component{
				{Type: "text", Value: "Mix "},
				{Type: "ingredient", Name: "flour", Quantity: "500", Unit: "g"},
				{Type: "text", Value: " with "},
				{Type: "ingredient", Name: "herbs", Quantity: "some", Optional: true},
				{Type: "text", Value: "."},
			},
		},
		{
			name:  "optional ingredient with multi-word name",
			input: "Add @?fresh thyme{1%sprig} for aroma.",
			expected: []Component{
				{Type: "text", Value: "Add "},
				{Type: "ingredient", Name: "fresh thyme", Quantity: "1", Unit: "sprig", Optional: true},
				{Type: "text", Value: " for aroma."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(recipe.Steps) == 0 {
				t.Fatal("Expected at least one step")
			}

			step := recipe.Steps[0]
			if len(step.Components) != len(tt.expected) {
				t.Fatalf("Expected %d components, got %d: %+v", len(tt.expected), len(step.Components), step.Components)
			}

			for i, expected := range tt.expected {
				actual := step.Components[i]
				if actual.Type != expected.Type {
					t.Errorf("Component %d: expected type %q, got %q", i, expected.Type, actual.Type)
				}
				if actual.Name != expected.Name {
					t.Errorf("Component %d: expected name %q, got %q", i, expected.Name, actual.Name)
				}
				if actual.Value != expected.Value {
					t.Errorf("Component %d: expected value %q, got %q", i, expected.Value, actual.Value)
				}
				if actual.Quantity != expected.Quantity {
					t.Errorf("Component %d: expected quantity %q, got %q", i, expected.Quantity, actual.Quantity)
				}
				if actual.Unit != expected.Unit {
					t.Errorf("Component %d: expected unit %q, got %q", i, expected.Unit, actual.Unit)
				}
				if actual.Optional != expected.Optional {
					t.Errorf("Component %d: expected optional %v, got %v", i, expected.Optional, actual.Optional)
				}
				if actual.Fixed != expected.Fixed {
					t.Errorf("Component %d: expected fixed %v, got %v", i, expected.Fixed, actual.Fixed)
				}
			}
		})
	}
}
