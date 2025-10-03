# Collected Ingredients Functions

This document describes the new convenience functions for collecting and consolidating ingredients from a recipe.

## Overview

The cooklang package now provides three new functions that make it easier to work with ingredients from recipes:

1. `GetCollectedIngredients()` - Returns a consolidated list of all ingredients
2. `GetCollectedIngredientsWithUnit(targetUnit string)` - Returns consolidated ingredients with unit conversion
3. `GetCollectedIngredientsMap()` - Returns a map suitable for shopping lists

## Functions

### GetCollectedIngredients()

Returns a consolidated list of all ingredients from the recipe. This function combines the functionality of `GetIngredients()` and `ConsolidateByName("")` into a single convenient call.

```go
func (r *Recipe) GetCollectedIngredients() (*IngredientList, error)
```

**Example:**

```go
collectedIngredients, err := recipe.GetCollectedIngredients()
if err != nil {
    log.Fatal(err)
}

for _, ingredient := range collectedIngredients.Ingredients {
    fmt.Printf("%s: %g %s\n", ingredient.Name, ingredient.Quantity, ingredient.Unit)
}
```

### GetCollectedIngredientsWithUnit(targetUnit string)

Returns a consolidated list of all ingredients from the recipe, converting them to the specified target unit when possible.

```go
func (r *Recipe) GetCollectedIngredientsWithUnit(targetUnit string) (*IngredientList, error)
```

**Example:**

```go
// Convert all volume ingredients to milliliters
collectedIngredients, err := recipe.GetCollectedIngredientsWithUnit("ml")
if err != nil {
    log.Fatal(err)
}
```

### GetCollectedIngredientsMap()

Returns a map of ingredient names to their consolidated quantities. This is particularly useful for creating shopping lists or ingredient summaries.

```go
func (r *Recipe) GetCollectedIngredientsMap() (map[string]string, error)
```

**Example:**

```go
shoppingList, err := recipe.GetCollectedIngredientsMap()
if err != nil {
    log.Fatal(err)
}

fmt.Println("Shopping List:")
for ingredient, quantity := range shoppingList {
    fmt.Printf("• %s: %s\n", ingredient, quantity)
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/hilli/cooklang"
)

func main() {
    recipeText := `
Add @flour{500%g} to bowl.
Mix with @flour{0.5%kg} and @sugar{200%g}.
Add @sugar{100%g} and @salt{1%tsp}.
Season with @salt{0.5%tsp} to taste.
`

    recipe, err := cooklang.ParseString(recipeText)
    if err != nil {
        log.Fatal(err)
    }

    // Get consolidated ingredients
    ingredients, err := recipe.GetCollectedIngredients()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d consolidated ingredients:\n", len(ingredients.Ingredients))
    
    // Get as a shopping list map
    shoppingList, err := recipe.GetCollectedIngredientsMap()
    if err != nil {
        log.Fatal(err)
    }

    for ingredient, quantity := range shoppingList {
        fmt.Printf("• %s: %s\n", ingredient, quantity)
    }
}
```

## Benefits

These new functions provide several benefits:

1. **Convenience**: Single function call instead of chaining multiple operations
2. **Error Handling**: Proper error propagation for consolidation failures
3. **Flexibility**: Option to specify target units for conversion
4. **Usability**: Direct map output suitable for shopping lists and UI display

## Unit Consolidation

The functions automatically consolidate ingredients with the same name:

- Ingredients with compatible units are converted and summed
- Ingredients with "some" quantities are preserved as-is
- Unit conversions work for mass (g, kg, oz, lb) and volume (ml, l, cup, tbsp, tsp)
- Incompatible units are kept separate in the result list

## Error Handling

All functions return proper errors for:

- Unit conversion failures
- Incompatible ingredient combinations
- Invalid recipe structures

Always check the returned error before using the results.
