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
