package cooklang

// RecipeRenderer interface defines how recipes can be rendered to different output formats.
// Implementations can render recipes as Markdown, HTML, plain text, or any custom format.
//
// Example implementation:
//
//	type JSONRenderer struct{}
//	func (jr JSONRenderer) RenderRecipe(recipe *cooklang.Recipe) string {
//	    data, _ := json.MarshalIndent(recipe, "", "  ")
//	    return string(data)
//	}
type RecipeRenderer interface {
	RenderRecipe(recipe *Recipe) string
}

// RendererFunc is a function type that implements RecipeRenderer.
// This allows using plain functions as renderers without creating a new type.
//
// Example:
//
//	simpleRenderer := cooklang.RendererFunc(func(r *cooklang.Recipe) string {
//	    return fmt.Sprintf("# %s\n\nServings: %.0f", r.Title, r.Servings)
//	})
//	output := recipe.RenderWith(simpleRenderer)
type RendererFunc func(*Recipe) string

// RenderRecipe implements the RecipeRenderer interface for RendererFunc.
func (f RendererFunc) RenderRecipe(recipe *Recipe) string {
	return f(recipe)
}

// SetRenderer allows setting a custom renderer for a recipe.
// Once set, calling Render() will use this custom renderer instead of the default.
//
// Parameters:
//   - renderer: A RecipeRenderer implementation
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("recipe.cook")
//	recipe.SetRenderer(renderers.MarkdownRenderer{})
//	markdown := recipe.Render()
func (r *Recipe) SetRenderer(renderer RecipeRenderer) {
	r.RenderFunc = func() string {
		return renderer.RenderRecipe(r)
	}
}

// SetRendererFunc allows setting a custom renderer function for a recipe
func (r *Recipe) SetRendererFunc(renderFunc func(*Recipe) string) {
	r.RenderFunc = func() string {
		return renderFunc(r)
	}
}

// RenderWith renders the recipe using the provided renderer.
// This allows one-time rendering without setting a permanent renderer on the recipe.
//
// Parameters:
//   - renderer: A RecipeRenderer implementation to use for rendering
//
// Returns:
//   - string: The rendered recipe in the format defined by the renderer
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("recipe.cook")
//	markdown := recipe.RenderWith(renderers.MarkdownRenderer{})
//	html := recipe.RenderWith(renderers.HTMLRenderer{})
func (r *Recipe) RenderWith(renderer RecipeRenderer) string {
	return renderer.RenderRecipe(r)
}
