package parser

import (
	"testing"

	"github.com/hilli/cooklang/lexer"
	"github.com/hilli/cooklang/token"
)

func TestCookwareTokens(t *testing.T) {
	input := "#large skillet{1}"
	l := lexer.New(input)

	t.Logf("Tokenizing: %s\n", input)
	for {
		tok := l.NextToken()
		t.Logf("Token: Type=%s, Literal='%s'\n", tok.Type, tok.Literal)
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

	t.Logf("Recipe: %+v\n", recipe)
	for i, step := range recipe.Steps {
		t.Logf("Step %d:\n", i+1)
		for j, comp := range step.Components {
			t.Logf("  Component %d: Type=%s, Name='%s', Value='%s', Quantity='%s'\n",
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

func TestCookwareWithAmpersand(t *testing.T) {
	p := New()
	recipe, err := p.ParseString("Strain into a #Nick & Nora glass{}(Chilled).")
	if err != nil {
		t.Fatalf("Error parsing recipe: %v", err)
	}

	t.Logf("Recipe: %+v\n", recipe)
	for i, step := range recipe.Steps {
		t.Logf("Step %d:\n", i+1)
		for j, comp := range step.Components {
			t.Logf("  Component %d: Type=%s, Name='%s', Value='%s', Quantity='%s'\n",
				j, comp.Type, comp.Name, comp.Value, comp.Quantity)
		}
	}

	// Check if we have the expected cookware component with & in the name
	found := false
	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			if comp.Type == "cookware" {
				found = true
				if comp.Name != "Nick & Nora glass" {
					t.Errorf("Expected cookware name 'Nick & Nora glass', got '%s'", comp.Name)
				}
				if comp.Value != "Chilled" {
					t.Errorf("Expected cookware instruction 'Chilled', got '%s'", comp.Value)
				}
			}
		}
	}

	if !found {
		t.Error("No cookware component found in parsed recipe")
	}
}

func TestCookwareWithPunctuation(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
	}{
		{
			name:         "Ampersand in name",
			input:        "Use a #Nick & Nora glass{}.",
			expectedName: "Nick & Nora glass",
		},
		{
			name:         "Comma in name",
			input:        "Use a #Dutch oven, large{}.",
			expectedName: "Dutch oven, large",
		},
		{
			name:         "Period in name",
			input:        "Use a #8-in. skillet{}.",
			expectedName: "8-in. skillet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			recipe, err := p.ParseString(tt.input)
			if err != nil {
				t.Fatalf("Error parsing recipe: %v", err)
			}

			found := false
			for _, step := range recipe.Steps {
				for _, comp := range step.Components {
					if comp.Type == "cookware" {
						found = true
						if comp.Name != tt.expectedName {
							t.Errorf("Expected cookware name '%s', got '%s'", tt.expectedName, comp.Name)
						}
					}
				}
			}

			if !found {
				t.Errorf("No cookware component found in parsed recipe")
			}
		})
	}
}
