package parser

import (
	"fmt"
	"testing"

	"github.com/hilli/cooklang/lexer"
	"github.com/hilli/cooklang/token"
)

func TestCookwareTokens(t *testing.T) {
	input := "#large skillet{1}"
	l := lexer.New(input)

	fmt.Printf("Tokenizing: %s\n", input)
	for {
		tok := l.NextToken()
		fmt.Printf("Token: Type=%s, Literal='%s'\n", tok.Type, tok.Literal)
		if tok.Type == token.EOF {
			break
		}
	}
}

func TestCookwareParsing(t *testing.T) {
	p := New()
	recipe, err := p.ParseString("Heat oil in a #large skillet{1} over medium heat.")
	if err != nil {
		t.Fatalf("Error parsing recipe: %v", err)
	}

	fmt.Printf("Recipe: %+v\n", recipe)
	for i, step := range recipe.Steps {
		fmt.Printf("Step %d:\n", i+1)
		for j, comp := range step.Components {
			fmt.Printf("  Component %d: Type=%s, Name='%s', Value='%s', Quantity='%s'\n",
				j, comp.Type, comp.Name, comp.Value, comp.Quantity)
		}
	}

	// Check if we have the expected cookware component
	found := false
	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "cookware" {
				found = true
				if comp.Name != "large skillet" {
					t.Errorf("Expected cookware name 'large skillet', got '%s'", comp.Name)
				}
				if comp.Quantity != "1" {
					t.Errorf("Expected cookware quantity '1', got '%s'", comp.Quantity)
				}
			}
		}
	}

	if !found {
		t.Error("No cookware component found in parsed recipe")
	}
}
