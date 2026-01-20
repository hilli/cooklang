package renderers

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hilli/cooklang"
)

func TestJSONLDRenderer_BasicRecipe(t *testing.T) {
	recipeContent := `---
title: Classic Margarita
description: A refreshing tequila-based cocktail
author: John Bartender
cuisine: Mexican
servings: 1
prep_time: 5 minutes
tags: tequila, citrus, classic
---

Add @tequila blanco{2%fl oz} to a #cocktail shaker{}.

Add @fresh lime juice{1%fl oz} and @triple sec{0.5%fl oz}.

Fill with ice and shake for ~{15%seconds}.

Strain into a salt-rimmed #coupe glass{}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	renderer := JSONLDRenderer{}
	data := renderer.RenderRecipe(recipe, nil)

	// Check required fields
	if data["@context"] != "https://schema.org" {
		t.Errorf("Expected @context to be 'https://schema.org', got %v", data["@context"])
	}
	if data["@type"] != "Recipe" {
		t.Errorf("Expected @type to be 'Recipe', got %v", data["@type"])
	}
	if data["name"] != "Classic Margarita" {
		t.Errorf("Expected name to be 'Classic Margarita', got %v", data["name"])
	}
	if data["description"] != "A refreshing tequila-based cocktail" {
		t.Errorf("Expected description to match, got %v", data["description"])
	}

	// Check author
	author, ok := data["author"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected author to be a map, got %T", data["author"])
	}
	if author["@type"] != "Person" {
		t.Errorf("Expected author @type to be 'Person', got %v", author["@type"])
	}
	if author["name"] != "John Bartender" {
		t.Errorf("Expected author name to be 'John Bartender', got %v", author["name"])
	}

	// Check servings
	if data["recipeYield"] != "1 serving" {
		t.Errorf("Expected recipeYield to be '1 serving', got %v", data["recipeYield"])
	}

	// Check prep time
	if data["prepTime"] != "PT5M" {
		t.Errorf("Expected prepTime to be 'PT5M', got %v", data["prepTime"])
	}

	// Check cuisine
	if data["recipeCuisine"] != "Mexican" {
		t.Errorf("Expected recipeCuisine to be 'Mexican', got %v", data["recipeCuisine"])
	}

	// Check keywords
	if data["keywords"] != "tequila, citrus, classic" {
		t.Errorf("Expected keywords to be 'tequila, citrus, classic', got %v", data["keywords"])
	}

	// Check ingredients
	ingredients, ok := data["recipeIngredient"].([]string)
	if !ok {
		t.Fatalf("Expected recipeIngredient to be []string, got %T", data["recipeIngredient"])
	}
	if len(ingredients) != 3 {
		t.Errorf("Expected 3 ingredients, got %d", len(ingredients))
	}

	// Check instructions
	instructions, ok := data["recipeInstructions"].([]interface{})
	if !ok {
		t.Fatalf("Expected recipeInstructions to be []interface{}, got %T", data["recipeInstructions"])
	}
	if len(instructions) != 4 {
		t.Errorf("Expected 4 instructions, got %d", len(instructions))
	}

	// Check first instruction structure
	firstStep, ok := instructions[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected instruction to be a map, got %T", instructions[0])
	}
	if firstStep["@type"] != "HowToStep" {
		t.Errorf("Expected instruction @type to be 'HowToStep', got %v", firstStep["@type"])
	}
	if firstStep["position"] != 1 {
		t.Errorf("Expected instruction position to be 1, got %v", firstStep["position"])
	}

	// Check tools/cookware
	tools, ok := data["tool"].([]string)
	if !ok {
		t.Fatalf("Expected tool to be []string, got %T", data["tool"])
	}
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d: %v", len(tools), tools)
	}
}

func TestJSONLDRenderer_WithOptions(t *testing.T) {
	recipeContent := `---
title: Test Recipe
author: Test Author
servings: 2
---

Mix @ingredient{1%cup}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	datePublished := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	dateModified := time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC)

	opts := &JSONLDOptions{
		URL:            "https://example.com/recipes/test",
		DatePublished:  &datePublished,
		DateModified:   &dateModified,
		Images:         []string{"https://example.com/images/test.jpg"},
		AuthorURL:      "https://example.com/users/testauthor",
		RecipeCategory: "Cocktail",
		Keywords:       []string{"extra", "keywords"},
		AggregateRating: &AggregateRating{
			RatingValue: 4.5,
			RatingCount: 42,
		},
	}

	renderer := JSONLDRenderer{}
	data := renderer.RenderRecipe(recipe, opts)

	// Check URL
	if data["url"] != "https://example.com/recipes/test" {
		t.Errorf("Expected url to match, got %v", data["url"])
	}

	// Check dates
	if data["datePublished"] != "2024-03-15" {
		t.Errorf("Expected datePublished to be '2024-03-15', got %v", data["datePublished"])
	}
	if data["dateModified"] != "2024-06-20" {
		t.Errorf("Expected dateModified to be '2024-06-20', got %v", data["dateModified"])
	}

	// Check image
	if data["image"] != "https://example.com/images/test.jpg" {
		t.Errorf("Expected image to match, got %v", data["image"])
	}

	// Check author URL
	author, ok := data["author"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected author to be a map, got %T", data["author"])
	}
	if author["url"] != "https://example.com/users/testauthor" {
		t.Errorf("Expected author url to match, got %v", author["url"])
	}

	// Check category
	if data["recipeCategory"] != "Cocktail" {
		t.Errorf("Expected recipeCategory to be 'Cocktail', got %v", data["recipeCategory"])
	}

	// Check aggregate rating
	rating, ok := data["aggregateRating"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected aggregateRating to be a map, got %T", data["aggregateRating"])
	}
	if rating["@type"] != "AggregateRating" {
		t.Errorf("Expected rating @type to be 'AggregateRating', got %v", rating["@type"])
	}
	if rating["ratingValue"] != 4.5 {
		t.Errorf("Expected ratingValue to be 4.5, got %v", rating["ratingValue"])
	}
	if rating["ratingCount"] != 42 {
		t.Errorf("Expected ratingCount to be 42, got %v", rating["ratingCount"])
	}
	if rating["bestRating"] != float64(5) {
		t.Errorf("Expected bestRating to be 5, got %v", rating["bestRating"])
	}
	if rating["worstRating"] != float64(1) {
		t.Errorf("Expected worstRating to be 1, got %v", rating["worstRating"])
	}
}

func TestJSONLDRenderer_MultipleImages(t *testing.T) {
	recipeContent := `---
title: Test Recipe
---

Mix @ingredient{}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	opts := &JSONLDOptions{
		Images: []string{
			"https://example.com/images/test1.jpg",
			"https://example.com/images/test2.jpg",
			"https://example.com/images/test3.jpg",
		},
	}

	renderer := JSONLDRenderer{}
	data := renderer.RenderRecipe(recipe, opts)

	images, ok := data["image"].([]string)
	if !ok {
		t.Fatalf("Expected image to be []string for multiple images, got %T", data["image"])
	}
	if len(images) != 3 {
		t.Errorf("Expected 3 images, got %d", len(images))
	}
}

func TestJSONLDRenderer_RenderRecipeJSON(t *testing.T) {
	recipeContent := `---
title: Test Recipe
description: A test recipe
---

Mix @ingredient{1%cup}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	renderer := JSONLDRenderer{}
	jsonStr, err := renderer.RenderRecipeJSON(recipe, nil)
	if err != nil {
		t.Fatalf("Failed to render JSON: %v", err)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if data["@context"] != "https://schema.org" {
		t.Errorf("Expected @context to be 'https://schema.org', got %v", data["@context"])
	}
}

func TestJSONLDRenderer_RenderRecipeScriptTag(t *testing.T) {
	recipeContent := `---
title: Test Recipe
---

Mix @ingredient{}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	renderer := JSONLDRenderer{}
	scriptTag, err := renderer.RenderRecipeScriptTag(recipe, nil)
	if err != nil {
		t.Fatalf("Failed to render script tag: %v", err)
	}

	if !strings.HasPrefix(scriptTag, `<script type="application/ld+json">`) {
		t.Errorf("Expected script tag to start with correct opening tag, got: %s", scriptTag[:50])
	}
	if !strings.HasSuffix(scriptTag, "</script>") {
		t.Errorf("Expected script tag to end with </script>")
	}
	if !strings.Contains(scriptTag, `"@context": "https://schema.org"`) {
		t.Errorf("Expected script tag to contain @context")
	}
}

func TestParseDurationToISO8601(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"15 minutes", "PT15M"},
		{"15 min", "PT15M"},
		{"15m", "PT15M"},
		{"1 hour", "PT1H"},
		{"1 hr", "PT1H"},
		{"1h", "PT1H"},
		{"2 hours", "PT2H"},
		{"1 hour 30 minutes", "PT1H30M"},
		{"1h 30m", "PT1H30M"},
		{"90 minutes", "PT1H30M"},
		{"120 minutes", "PT2H"},
		{"30 seconds", "PT30S"},
		{"30 sec", "PT30S"},
		{"30s", "PT30S"},
		{"1 hour 30 minutes 45 seconds", "PT1H30M45S"},
		{"5", "PT5M"}, // Plain number assumed to be minutes
		{"", ""},
		{"invalid", ""},
		{"2 HOURS", "PT2H"},     // Case insensitive
		{"  15 min  ", "PT15M"}, // Whitespace handling
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ParseDurationToISO8601(test.input)
			if result != test.expected {
				t.Errorf("ParseDurationToISO8601(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestJSONLDRenderer_WithSections(t *testing.T) {
	recipeContent := `---
title: Layered Cocktail
---

== Base Layer ==

Add @dark rum{1%fl oz} to #glass{}.

== Top Layer ==

Float @light rum{0.5%fl oz} on top.

Garnish with @lime wheel{1}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	renderer := JSONLDRenderer{}
	data := renderer.RenderRecipe(recipe, nil)

	instructions, ok := data["recipeInstructions"].([]interface{})
	if !ok {
		t.Fatalf("Expected recipeInstructions to be []interface{}, got %T", data["recipeInstructions"])
	}

	// Should have HowToSection objects
	if len(instructions) < 2 {
		t.Fatalf("Expected at least 2 instruction items, got %d", len(instructions))
	}

	// Check first section
	firstSection, ok := instructions[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected first instruction to be a map, got %T", instructions[0])
	}
	if firstSection["@type"] != "HowToSection" {
		t.Errorf("Expected first instruction @type to be 'HowToSection', got %v", firstSection["@type"])
	}
	if firstSection["name"] != "Base Layer" {
		t.Errorf("Expected first section name to be 'Base Layer', got %v", firstSection["name"])
	}
}

func TestJSONLDRenderer_WithVideo(t *testing.T) {
	recipeContent := `---
title: Test Recipe
---

Mix @ingredient{}.
`

	recipe, err := cooklang.ParseString(recipeContent)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	uploadDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	opts := &JSONLDOptions{
		Video: &VideoObject{
			Name:         "How to make Test Recipe",
			Description:  "A video tutorial",
			ThumbnailURL: "https://example.com/thumb.jpg",
			ContentURL:   "https://example.com/video.mp4",
			EmbedURL:     "https://example.com/embed/video",
			UploadDate:   &uploadDate,
			Duration:     "PT5M30S",
		},
	}

	renderer := JSONLDRenderer{}
	data := renderer.RenderRecipe(recipe, opts)

	video, ok := data["video"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected video to be a map, got %T", data["video"])
	}
	if video["@type"] != "VideoObject" {
		t.Errorf("Expected video @type to be 'VideoObject', got %v", video["@type"])
	}
	if video["name"] != "How to make Test Recipe" {
		t.Errorf("Expected video name to match, got %v", video["name"])
	}
	if video["duration"] != "PT5M30S" {
		t.Errorf("Expected video duration to be 'PT5M30S', got %v", video["duration"])
	}
}

func TestJSONLDRenderer_MultipleServings(t *testing.T) {
	tests := []struct {
		servings float32
		expected string
	}{
		{1, "1 serving"},
		{2, "2 servings"},
		{4, "4 servings"},
		{1.5, "1.5 servings"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			recipe := &cooklang.Recipe{
				Title:    "Test",
				Servings: test.servings,
			}

			renderer := JSONLDRenderer{}
			data := renderer.RenderRecipe(recipe, nil)

			if data["recipeYield"] != test.expected {
				t.Errorf("Expected recipeYield to be %q, got %v", test.expected, data["recipeYield"])
			}
		})
	}
}

func TestNewJSONLDRenderer(t *testing.T) {
	renderer := NewJSONLDRenderer()
	if renderer != (JSONLDRenderer{}) {
		t.Error("NewJSONLDRenderer should return an empty JSONLDRenderer struct")
	}
}
