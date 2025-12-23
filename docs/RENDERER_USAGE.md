# Renderer System Usage

The Cooklang library now supports a flexible renderer system that allows you to convert recipes to different output formats.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/hilli/cooklang"
    "github.com/hilli/cooklang/renderers"
)

func main() {
    // Parse a recipe
    recipe, err := cooklang.ParseFile("recipe.cook")
    if err != nil {
        panic(err)
    }

    // Use default Cooklang format
    fmt.Println(recipe.Render())

    // Use built-in renderers
    fmt.Println(recipe.RenderWith(renderers.Default.Markdown))
    fmt.Println(recipe.RenderWith(renderers.Default.HTML))
    fmt.Println(recipe.RenderWith(renderers.Default.Cooklang))
    fmt.Println(recipe.RenderWith(renderers.Default.Print))
}
```

## Built-in Renderers

### 1. CooklangRenderer (Default)
Renders recipes in the original Cooklang format with metadata and components.

```go
output := recipe.RenderWith(renderers.Default.Cooklang)
```

### 2. MarkdownRenderer
Renders recipes as Markdown with proper headers and formatting.

```go
output := recipe.RenderWith(renderers.Default.Markdown)
```

### 3. HTMLRenderer
Renders recipes as HTML with CSS classes for styling.

```go
output := recipe.RenderWith(renderers.Default.HTML)
```

### 4. PrintRenderer
Renders recipes as print-optimized HTML with embedded CSS, designed to fit nicely on a single sheet of paper. Features:
- Two-column layout (ingredients left, instructions right)
- Print-specific CSS with `@page` rules
- Clean, professional typography
- No interactive elements or unnecessary UI chrome

```go
output := recipe.RenderWith(renderers.Default.Print)
```

Or via CLI:
```bash
cook render recipe.cook --format=print --output=recipe.html
```

## Custom Renderers

### Using RendererFunc
```go
customRenderer := cooklang.RendererFunc(func(r *cooklang.Recipe) string {
    return fmt.Sprintf("🍽️ %s - Serves %g", r.Title, r.Servings)
})

recipe.SetRenderer(customRenderer)
output := recipe.Render()
```

### Using SetRendererFunc
```go
recipe.SetRendererFunc(func(r *cooklang.Recipe) string {
    return fmt.Sprintf("Recipe: %s", r.Title)
})

output := recipe.Render()
```

### Implementing RecipeRenderer Interface
```go
type JSONRenderer struct{}

func (jr JSONRenderer) RenderRecipe(recipe *cooklang.Recipe) string {
    data := map[string]interface{}{
        "title": recipe.Title,
        "steps": []string{},
    }
    
    // Traverse linked list of steps
    currentStep := recipe.FirstStep
    for currentStep != nil {
        stepText := ""
        currentComponent := currentStep.FirstComponent
        for currentComponent != nil {
            stepText += currentComponent.Render()
            currentComponent = currentComponent.GetNext()
        }
        data["steps"] = append(data["steps"].([]string), stepText)
        currentStep = currentStep.NextStep
    }
    
    jsonBytes, _ := json.Marshal(data)
    return string(jsonBytes)
}

// Usage
jsonRenderer := JSONRenderer{}
output := recipe.RenderWith(jsonRenderer)
```

## Linked List Structure

The recipe structure uses linked lists for efficient traversal:

- `Recipe.FirstStep` points to the first step
- `Step.NextStep` points to the next step (nil for last step)
- `Step.FirstComponent` points to the first component in the step
- Each component has a `GetNext()` method that returns the next component

This allows for memory-efficient iteration through recipe data:

```go
// Iterate through steps
currentStep := recipe.FirstStep
for currentStep != nil {
    // Iterate through components in this step
    currentComponent := currentStep.FirstComponent
    for currentComponent != nil {
        // Process component
        fmt.Println(currentComponent.Render())
        currentComponent = currentComponent.GetNext()
    }
    currentStep = currentStep.NextStep
}
```

## Component Types

All components implement the `StepComponent` interface with:
- `Render() string` - Returns the Cooklang representation
- `SetNext(StepComponent)` - Sets the next component in the chain
- `GetNext() StepComponent` - Gets the next component in the chain

Available component types:
- `Ingredient` - Recipe ingredients with quantities
- `Cookware` - Cooking equipment
- `Timer` - Timing instructions
- `Instruction` - Text instructions
