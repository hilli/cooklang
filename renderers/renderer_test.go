package renderers

import (
	"strings"
	"testing"

	"github.com/hilli/cooklang"
)

func TestBasicRenderers(t *testing.T) {
	// Create a test recipe
	recipe := &cooklang.Recipe{
		Title:       "Test Pasta",
		Cuisine:     "Italian",
		Description: "A simple pasta recipe",
		Difficulty:  "Easy",
		PrepTime:    "10 minutes",
		TotalTime:   "20 minutes",
		Author:      "Test Chef",
		Servings:    4,
		Tags:        []string{"pasta", "quick", "easy"},
	}

	// Create some test steps and components
	step1 := &cooklang.Step{}

	// Step 1 components
	inst1 := &cooklang.Instruction{Text: "Boil "}
	ing1 := &cooklang.Ingredient{Name: "water", Quantity: 2, Unit: "liters"}
	inst2 := &cooklang.Instruction{Text: " in a "}
	cookware1 := &cooklang.Cookware{Name: "large pot", Quantity: 1}
	inst3 := &cooklang.Instruction{Text: "."}

	// Link step 1 components
	step1.FirstComponent = inst1
	inst1.SetNext(ing1)
	ing1.SetNext(inst2)
	inst2.SetNext(cookware1)
	cookware1.SetNext(inst3)

	// Link steps
	recipe.FirstStep = step1

	t.Run("CooklangRenderer", func(t *testing.T) {
		output := recipe.RenderWith(Default.Cooklang)
		if !strings.Contains(output, ">> title: Test Pasta") {
			t.Errorf("Expected Cooklang metadata format, got: %s", output)
		}
		if !strings.Contains(output, "@water{2%liters}") {
			t.Errorf("Expected Cooklang ingredient format, got: %s", output)
		}
		if !strings.Contains(output, "#large pot{}") {
			t.Errorf("Expected Cooklang cookware format, got: %s", output)
		}
	})

	t.Run("MarkdownRenderer", func(t *testing.T) {
		output := recipe.RenderWith(Default.Markdown)
		if !strings.Contains(output, "# Test Pasta") {
			t.Errorf("Expected Markdown title format, got: %s", output)
		}
		if !strings.Contains(output, "**water**") {
			t.Errorf("Expected Markdown ingredient format, got: %s", output)
		}
		if !strings.Contains(output, "*large pot*") {
			t.Errorf("Expected Markdown cookware format, got: %s", output)
		}
	})

	t.Run("HTMLRenderer", func(t *testing.T) {
		output := recipe.RenderWith(Default.HTML)
		if !strings.Contains(output, "<h1 class=\"recipe-title\">Test Pasta</h1>") {
			t.Errorf("Expected HTML title format, got: %s", output)
		}
		if !strings.Contains(output, "<span class=\"ingredient\">water</span>") {
			t.Errorf("Expected HTML ingredient format, got: %s", output)
		}
		if !strings.Contains(output, "<span class=\"cookware\">large pot</span>") {
			t.Errorf("Expected HTML cookware format, got: %s", output)
		}
	})

	t.Run("CustomRenderer", func(t *testing.T) {
		// Test setting a custom renderer
		customRenderer := cooklang.RendererFunc(func(r *cooklang.Recipe) string {
			return "Custom: " + r.Title
		})

		recipe.SetRenderer(customRenderer)
		output := recipe.Render()
		if output != "Custom: Test Pasta" {
			t.Errorf("Expected custom renderer output 'Custom: Test Pasta', got: %s", output)
		}
	})
}
