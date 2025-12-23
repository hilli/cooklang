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
		if !strings.Contains(output, "---") {
			t.Errorf("Expected YAML frontmatter delimiter '---', got: %s", output)
		}
		if !strings.Contains(output, "title: Test Pasta") {
			t.Errorf("Expected YAML metadata format (without >>), got: %s", output)
		}
		if strings.Contains(output, ">>") {
			t.Errorf("Expected YAML format without >> prefix, got: %s", output)
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

	t.Run("PrintRenderer", func(t *testing.T) {
		output := recipe.RenderWith(Default.Print)

		// Check for complete HTML document structure
		if !strings.Contains(output, "<!DOCTYPE html>") {
			t.Errorf("Expected HTML doctype, got: %s", output)
		}
		if !strings.Contains(output, "<html lang=\"en\">") {
			t.Errorf("Expected html lang attribute, got: %s", output)
		}
		if !strings.Contains(output, "</html>") {
			t.Errorf("Expected closing html tag, got: %s", output)
		}

		// Check for embedded CSS
		if !strings.Contains(output, "<style>") {
			t.Errorf("Expected embedded CSS, got: %s", output)
		}
		if !strings.Contains(output, "@page") {
			t.Errorf("Expected @page CSS rule for printing, got: %s", output)
		}
		if !strings.Contains(output, "@media print") {
			t.Errorf("Expected @media print CSS rules, got: %s", output)
		}

		// Check for title in both <title> and <h1>
		if !strings.Contains(output, "<title>Test Pasta</title>") {
			t.Errorf("Expected HTML title element, got: %s", output)
		}
		if !strings.Contains(output, "<h1 class=\"recipe-title\">Test Pasta</h1>") {
			t.Errorf("Expected h1 title, got: %s", output)
		}

		// Check for recipe metadata
		if !strings.Contains(output, "Servings:</span> 4") {
			t.Errorf("Expected servings in metadata, got: %s", output)
		}
		if !strings.Contains(output, "Cuisine:</span> Italian") {
			t.Errorf("Expected cuisine in metadata, got: %s", output)
		}

		// Check for two-column layout structure
		if !strings.Contains(output, "class=\"recipe-body\"") {
			t.Errorf("Expected recipe-body container for two-column layout, got: %s", output)
		}
		if !strings.Contains(output, "class=\"recipe-ingredients\"") {
			t.Errorf("Expected recipe-ingredients section, got: %s", output)
		}
		if !strings.Contains(output, "class=\"recipe-instructions\"") {
			t.Errorf("Expected recipe-instructions section, got: %s", output)
		}

		// Check for ingredient formatting
		if !strings.Contains(output, "class=\"ingredient-qty\"") {
			t.Errorf("Expected ingredient-qty class, got: %s", output)
		}
		if !strings.Contains(output, "class=\"ingredient-name\"") {
			t.Errorf("Expected ingredient-name class, got: %s", output)
		}
		if !strings.Contains(output, "2 liters") {
			t.Errorf("Expected formatted quantity '2 liters', got: %s", output)
		}

		// Check for instruction formatting with inline styles
		if !strings.Contains(output, "class=\"ing\"") {
			t.Errorf("Expected ing class for inline ingredients, got: %s", output)
		}
		if !strings.Contains(output, "class=\"cw\"") {
			t.Errorf("Expected cw class for cookware, got: %s", output)
		}

		// Check for tags in footer
		if !strings.Contains(output, "class=\"recipe-tags\"") {
			t.Errorf("Expected recipe-tags in footer, got: %s", output)
		}
		if !strings.Contains(output, "pasta, quick, easy") {
			t.Errorf("Expected tags content, got: %s", output)
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

	t.Run("CooklangArrayFormatting", func(t *testing.T) {
		// Test that arrays are properly formatted in YAML frontmatter
		recipeWithArrays := &cooklang.Recipe{
			Title:    "Array Test Recipe",
			Tags:     []string{"tag1", "tag2", "tag3"},
			Images:   []string{"image1.jpg", "image2.jpg"},
			Servings: 4,
		}

		output := recipeWithArrays.RenderWith(Default.Cooklang)

		// Check that tags are formatted as YAML array
		if !strings.Contains(output, "tags:\n") {
			t.Errorf("Expected tags to start with 'tags:\\n', got: %s", output)
		}
		if !strings.Contains(output, "  - tag1") {
			t.Errorf("Expected tag1 formatted as '  - tag1', got: %s", output)
		}
		if !strings.Contains(output, "  - tag2") {
			t.Errorf("Expected tag2 formatted as '  - tag2', got: %s", output)
		}
		if !strings.Contains(output, "  - tag3") {
			t.Errorf("Expected tag3 formatted as '  - tag3', got: %s", output)
		}

		// Check that images are formatted as YAML array
		if !strings.Contains(output, "images:\n") {
			t.Errorf("Expected images to start with 'images:\\n', got: %s", output)
		}
		if !strings.Contains(output, "  - image1.jpg") {
			t.Errorf("Expected image1.jpg formatted as '  - image1.jpg', got: %s", output)
		}
		if !strings.Contains(output, "  - image2.jpg") {
			t.Errorf("Expected image2.jpg formatted as '  - image2.jpg', got: %s", output)
		}

		// Ensure arrays are NOT comma-separated
		if strings.Contains(output, "tags: tag1, tag2, tag3") {
			t.Errorf("Tags should not be comma-separated, got: %s", output)
		}
		if strings.Contains(output, "images: image1.jpg, image2.jpg") {
			t.Errorf("Images should not be comma-separated, got: %s", output)
		}
	})
}
