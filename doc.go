// Package cooklang provides a parser and tools for working with Cooklang recipe files.
//
// Cooklang is a markup language for cooking recipes that makes it easy to manage recipes
// as plain text files while providing rich semantic information about ingredients,
// cookware, timers, and instructions.
//
// # Basic Usage
//
// Parse a recipe file:
//
//	recipe, err := cooklang.ParseFile("lasagna.cook")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Recipe: %s (serves %.0f)\n", recipe.Title, recipe.Servings)
//
// # Working with Ingredients
//
// Extract and consolidate ingredients for shopping lists:
//
//	ingredients := recipe.GetIngredients()
//	consolidated, err := ingredients.ConsolidateByName("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, ing := range consolidated.Ingredients {
//	    fmt.Printf("- %s: %.1f %s\n", ing.Name, ing.Quantity, ing.Unit)
//	}
//
// # Unit Conversion
//
// Convert ingredients between measurement systems:
//
//	// Convert to metric
//	shoppingList, err := recipe.GetMetricShoppingList()
//
//	// Convert to US customary
//	shoppingList, err := recipe.GetUSShoppingList()
//
//	// Convert individual ingredients
//	ingredient := &cooklang.Ingredient{Name: "flour", Quantity: 2, Unit: "cup"}
//	converted, err := ingredient.ConvertTo("g")
//
// # Shopping Lists
//
// Create shopping lists from multiple recipes:
//
//	recipe1, _ := cooklang.ParseFile("pasta.cook")
//	recipe2, _ := cooklang.ParseFile("salad.cook")
//	recipe3, _ := cooklang.ParseFile("dessert.cook")
//
//	shoppingList, err := cooklang.CreateShoppingList(recipe1, recipe2, recipe3)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Scale for meal prep
//	doubled := shoppingList.Scale(2.0)
//
//	// Print the list
//	for ingredient, amount := range doubled.ToMap() {
//	    fmt.Printf("☐ %s: %s\n", ingredient, amount)
//	}
//
// # Metadata Management
//
// Edit recipe frontmatter metadata:
//
//	editor, err := cooklang.NewFrontmatterEditor("recipe.cook")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	editor.SetMetadata("title", "Improved Lasagna")
//	editor.SetMetadata("servings", "8")
//	editor.SetMetadata("tags", "italian, pasta, main course")
//	editor.SetMetadata("difficulty", "medium")
//
//	if err := editor.Save(); err != nil {
//	    log.Fatal(err)
//	}
//
// # Rendering Recipes
//
// Render recipes in different formats:
//
//	import "github.com/hilli/cooklang/renderers"
//
//	recipe, _ := cooklang.ParseFile("recipe.cook")
//
//	// Render as Markdown
//	markdown := recipe.RenderWith(renderers.MarkdownRenderer{})
//	fmt.Println(markdown)
//
//	// Render as HTML
//	html := recipe.RenderWith(renderers.HTMLRenderer{})
//
//	// Set a custom renderer
//	recipe.SetRenderer(renderers.CooklangRenderer{})
//	cooklangText := recipe.Render()
//
// # Recipe Structure
//
// Recipes are organized as linked lists of steps, where each step contains a linked list
// of components (ingredients, instructions, timers, cookware). This structure allows for
// efficient traversal and manipulation:
//
//	// Walk through all steps
//	currentStep := recipe.FirstStep
//	stepNum := 1
//	for currentStep != nil {
//	    fmt.Printf("Step %d:\n", stepNum)
//
//	    // Walk through components in this step
//	    component := currentStep.FirstComponent
//	    for component != nil {
//	        switch c := component.(type) {
//	        case *cooklang.Ingredient:
//	            fmt.Printf("  Add %s (%.1f %s)\n", c.Name, c.Quantity, c.Unit)
//	        case *cooklang.Instruction:
//	            fmt.Printf("  %s\n", c.Text)
//	        case *cooklang.Timer:
//	            fmt.Printf("  Wait for %s %s\n", c.Duration, c.Unit)
//	        case *cooklang.Cookware:
//	            fmt.Printf("  Using: %s\n", c.Name)
//	        }
//	        component = component.GetNext()
//	    }
//
//	    currentStep = currentStep.NextStep
//	    stepNum++
//	}
//
// # Cooklang Syntax
//
// The parser supports standard Cooklang syntax:
//
//   - Ingredients: @flour{500%g}, @salt{}, @milk{2%cups}(room temperature)
//   - Cookware: #pot{}, #bowl{2}, #oven{}(preheated)
//   - Timers: ~{10%minutes}, ~boil{15%min}
//   - Comments: -- This is a comment
//   - Metadata: YAML frontmatter between --- delimiters
//
// See https://cooklang.org for full specification details.
package cooklang
