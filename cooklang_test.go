package cooklang

import (
	"testing"

	"github.com/hilli/cooklang/parser"
)

var (
	input = `---
title: Test Recipe
date: 2023-10-01
cuisine: Italian
difficulty: Easy
prep_time: 10 minutes
total_time: 30 minutes
author: John Doe
description: A simple test recipe.
images:
  - image1.jpg
  - image2.jpeg
servings: 2
tags: [ test, cooking]
---
Mince @garlic{2%cloves} and sauté in @olive oil{2%tbsp}.

Cook the @pasta{500%g} in a #pot until al dente - Approximately ~{8%minutes}.
`
)

func Test_Reparser(t *testing.T) {
	t.Log(input)
	p := parser.New()
	pRecipe, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}
	t.Logf("Parsed Recipe: %#v", pRecipe)
	recipe := ToCooklangRecipe(pRecipe)
	if recipe == nil {
		t.Fatal("Failed to convert parsed recipe to CooklangRecipe")
	}
	// fmt.Printf("Recipe: %+v\n", recipe)

	// Test basic rendering (using default renderer)
	// output := recipe.Render()
	// if output == "" {
	// 	t.Error("Recipe rendering should not be empty")
	// }

}

func TestGetCookware(t *testing.T) {
	recipe := `>> title: Test Recipe
>> servings: 4

Cook the @pasta{500%g} in a #pot{} until al dente.

Transfer to a #serving bowl{} and add @cheese{100%g}.

Bake in an #oven{} at 180°C for ~{30%minutes}.
`
	parsed, err := ParseString(recipe)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	cookware := parsed.GetCookware()

	if len(cookware) != 3 {
		t.Errorf("Expected 3 cookware items, got %d", len(cookware))
	}

	// Check that we got the expected cookware
	expectedNames := []string{"pot", "serving bowl", "oven"}
	for i, expected := range expectedNames {
		if cookware[i].Name != expected {
			t.Errorf("Expected cookware[%d].Name to be %q, got %q", i, expected, cookware[i].Name)
		}
	}
}

func TestGetCookwareWithQuantities(t *testing.T) {
	recipe := `>> title: Test Recipe

Mix ingredients in #bowl{2}.
`
	parsed, err := ParseString(recipe)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	cookware := parsed.GetCookware()

	if len(cookware) != 1 {
		t.Fatalf("Expected 1 cookware item, got %d", len(cookware))
	}

	if cookware[0].Name != "bowl" {
		t.Errorf("Expected cookware name to be 'bowl', got %q", cookware[0].Name)
	}

	if cookware[0].Quantity != 2 {
		t.Errorf("Expected cookware quantity to be 2, got %d", cookware[0].Quantity)
	}
}

func TestGetCookwareEmpty(t *testing.T) {
	recipe := `>> title: Test Recipe

Mix @flour{200%g} with @water{100%ml}.
`
	parsed, err := ParseString(recipe)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	cookware := parsed.GetCookware()

	if len(cookware) != 0 {
		t.Errorf("Expected 0 cookware items, got %d", len(cookware))
	}
}
