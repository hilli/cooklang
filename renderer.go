package cooklang

// RecipeRenderer interface defines how recipes can be rendered
type RecipeRenderer interface {
	RenderRecipe(recipe *Recipe) string
}

// RendererFunc is a function type that implements RecipeRenderer
type RendererFunc func(*Recipe) string

func (f RendererFunc) RenderRecipe(recipe *Recipe) string {
	return f(recipe)
}

// SetRenderer allows setting a custom renderer for a recipe
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

// RenderWith renders the recipe using the provided renderer
func (r *Recipe) RenderWith(renderer RecipeRenderer) string {
	return renderer.RenderRecipe(r)
}
