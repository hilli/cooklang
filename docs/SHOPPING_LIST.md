# Shopping List Feature

The shopping list functionality allows you to consolidate ingredients from multiple recipes into a single unified shopping list. This is perfect for meal planning, batch cooking, or preparing multiple dishes at once.

## Overview

The cooklang package provides functions to:

1. Combine ingredients from multiple recipes
2. Automatically consolidate duplicate ingredients
3. Convert units when possible
4. Scale quantities for different serving sizes
5. Export to various formats for display or processing

## Core Functions

### CreateShoppingList

Creates a consolidated shopping list from multiple recipes.

```go
func CreateShoppingList(recipes ...*Recipe) (*ShoppingList, error)
```

**Parameters:**

- `recipes ...Recipe`: Variable number of recipe pointers to combine

**Returns:**

- `*ShoppingList`: Consolidated shopping list with all ingredients
- `error`: Error if consolidation fails

**Example:**

```go
pasta, _ := cooklang.ParseString(pastaRecipe)
salad, _ := cooklang.ParseString(saladRecipe)
bread, _ := cooklang.ParseString(breadRecipe)

shoppingList, err := cooklang.CreateShoppingList(pasta, salad, bread)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total ingredients: %d\n", shoppingList.Count())
```

### CreateShoppingListWithUnit

Creates a shopping list with ingredients converted to a target unit when possible.

```go
func CreateShoppingListWithUnit(targetUnit string, recipes ...*Recipe) (*ShoppingList, error)
```

**Parameters:**

- `targetUnit string`: The unit to convert compatible ingredients to (e.g., "g", "ml", "kg")
- `recipes ...*Recipe`: Variable number of recipe pointers to combine

**Returns:**

- `*ShoppingList`: Consolidated shopping list with converted units
- `error`: Error if conversion or consolidation fails

**Example:**

```go
// Convert all mass ingredients to kilograms
shoppingList, err := cooklang.CreateShoppingListWithUnit("kg", recipe1, recipe2)
if err != nil {
    log.Fatal(err)
}
```

### CreateShoppingListForServings

Creates a shopping list by scaling each recipe to the target number of servings before combining. This is ideal for meal planning where recipes have different serving sizes.

```go
func CreateShoppingListForServings(targetServings float64, recipes ...*Recipe) (*ShoppingList, error)
```

**Parameters:**

- `targetServings float64`: The desired number of servings for each recipe
- `recipes ...*Recipe`: Variable number of recipe pointers to combine

**Returns:**

- `*ShoppingList`: Consolidated shopping list with all ingredients scaled
- `error`: Error if consolidation fails

**Example:**

```go
// Recipes with different serving sizes
monday, _ := cooklang.ParseFile("monday.cook")     // servings: 2
tuesday, _ := cooklang.ParseFile("tuesday.cook")  // servings: 8

// Normalize both to 4 servings before combining
shoppingList, err := cooklang.CreateShoppingListForServings(4, monday, tuesday)
// monday scaled 2x, tuesday scaled 0.5x
```

### CreateShoppingListForServingsWithUnit

Combines servings normalization with unit conversion - scales each recipe to the target servings and converts ingredients to the target unit.

```go
func CreateShoppingListForServingsWithUnit(targetServings float64, targetUnit string, recipes ...*Recipe) (*ShoppingList, error)
```

**Parameters:**

- `targetServings float64`: The desired number of servings for each recipe
- `targetUnit string`: The unit to convert compatible ingredients to (e.g., "g", "ml", "kg")
- `recipes ...*Recipe`: Variable number of recipe pointers to combine

**Returns:**

- `*ShoppingList`: Consolidated shopping list with scaled and converted ingredients
- `error`: Error if conversion or consolidation fails

**Example:**

```go
// Scale to 4 servings and convert masses to grams
shoppingList, err := cooklang.CreateShoppingListForServingsWithUnit(4, "g", recipes...)
```

## ShoppingList Type

```go
type ShoppingList struct {
    Ingredients *IngredientList `json:"ingredients"`
    Recipes     []string        `json:"recipes,omitempty"`
}
```

### Methods

#### ToMap()

Returns the shopping list as a map of ingredient names to their consolidated quantities.

```go
func (sl *ShoppingList) ToMap() map[string]string
```

**Example:**

```go
shoppingMap := shoppingList.ToMap()
for ingredient, quantity := range shoppingMap {
    fmt.Printf("• %s: %s\n", ingredient, quantity)
}
```

#### Scale()

Scales all ingredients in the shopping list by the given multiplier.

```go
func (sl *ShoppingList) Scale(multiplier float64) *ShoppingList
```

**Parameters:**

- `multiplier float64`: The scaling factor (e.g., 2.0 for double, 0.5 for half)

**Returns:**

- `*ShoppingList`: New shopping list with scaled quantities

**Example:**

```go
// Scale for a dinner party (double the quantities)
scaledList := shoppingList.Scale(2.0)
```

#### Count()

Returns the number of unique ingredients in the shopping list.

```go
func (sl *ShoppingList) Count() int
```

**Example:**

```go
fmt.Printf("You need %d different ingredients\n", shoppingList.Count())
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
    // Define multiple recipes
    pastaRecipe := `Add @pasta{500%g} and @olive oil{3%tbsp}.
Season with @salt{1%tsp} and @pepper{0.5%tsp}.`

    saladRecipe := `Mix @lettuce{200%g} with @olive oil{2%tbsp}.
Season with @salt{0.5%tsp} and @pepper{0.25%tsp}.`

    // Parse recipes
    pasta, _ := cooklang.ParseString(pastaRecipe)
    salad, _ := cooklang.ParseString(saladRecipe)

    // Create consolidated shopping list
    shoppingList, err := cooklang.CreateShoppingList(pasta, salad)
    if err != nil {
        log.Fatal(err)
    }

    // Display as a map
    fmt.Println("Shopping List:")
    for ingredient, quantity := range shoppingList.ToMap() {
        fmt.Printf("  • %s: %s\n", ingredient, quantity)
    }

    // Scale for 4 people instead of 2
    scaled := shoppingList.Scale(2.0)
    fmt.Println("\nFor 4 people:")
    for ingredient, quantity := range scaled.ToMap() {
        fmt.Printf("  • %s: %s\n", ingredient, quantity)
    }
}
```

**Output:**

```text
Shopping List:
  • pasta: 500 g
  • olive oil: 5 tbsp
  • salt: 1.5 tsp
  • pepper: 0.8 tsp
  • lettuce: 200 g

For 4 people:
  • pasta: 1000 g
  • olive oil: 10 tbsp
  • salt: 3 tsp
  • pepper: 1.5 tsp
  • lettuce: 400 g
```

## Features

### Automatic Consolidation

The shopping list automatically consolidates ingredients with the same name:

```go
// Recipe 1: @flour{500%g}
// Recipe 2: @flour{300%g}
// Result: flour: 800 g
```

### Unit Conversion

When ingredients have compatible units, they are automatically converted and summed:

```go
// Recipe 1: @butter{500%g}
// Recipe 2: @butter{0.5%kg}
// Result: butter: 1000 g (or 1 kg if converted)
```

### Handling "Some" Quantities

Ingredients with unspecified quantities are preserved as-is:

```go
// Recipe: @vanilla{some}
// Result: vanilla: some
```

### Scaling

The `Scale()` method multiplies all quantified ingredients while preserving "some" quantities:

```go
shoppingList.Scale(2.0)  // Double all quantities
shoppingList.Scale(0.5)  // Halve all quantities
shoppingList.Scale(1.5)  // 1.5x quantities
```

## Use Cases

### Meal Planning with Servings Normalization

When planning meals for a household, recipes often have different serving sizes. Use `CreateShoppingListForServings` to normalize each recipe to your household size before combining:

```go
// Recipes with different serving sizes
monday, _ := cooklang.ParseFile("monday.cook")     // servings: 2
tuesday, _ := cooklang.ParseFile("tuesday.cook")  // servings: 8  
wednesday, _ := cooklang.ParseFile("wednesday.cook") // servings: 4

// Normalize all to household of 5 people
weeklyList, _ := cooklang.CreateShoppingListForServings(5, monday, tuesday, wednesday)
// monday scaled 2.5x, tuesday scaled 0.625x, wednesday scaled 1.25x
```

This is much more practical than using `Scale()` which would require calculating individual factors for each recipe.

#### With Unit Conversion

Combine servings normalization with unit standardization:

```go
// Scale to 4 servings and convert to metric
list, _ := cooklang.CreateShoppingListForServingsWithUnit(4, "g", recipes...)
```

#### CLI Usage

```bash
# Scale all recipes to 4 servings (household size)
cook shopping-list monday.cook tuesday.cook --servings 4

# With unit conversion
cook shop recipes/*.cook --servings 4 --unit kg
```

### Basic Meal Planning

Create a weekly shopping list by combining all your planned recipes:

```go
monday := cooklang.ParseString(mondayRecipe)
tuesday := cooklang.ParseString(tuesdayRecipe)
wednesday := cooklang.ParseString(wednesdayRecipe)

weeklyList, _ := cooklang.CreateShoppingList(monday, tuesday, wednesday)
```

### Batch Cooking

Scale a recipe for meal prep:

```go
recipe := cooklang.ParseString(recipeText)
list, _ := cooklang.CreateShoppingList(recipe)
batchList := list.Scale(5.0) // Make 5x the recipe
```

### Party Planning

Combine appetizers, mains, and desserts for an event:

```go
appetizer := cooklang.ParseString(appetizerRecipe)
main := cooklang.ParseString(mainRecipe)
dessert := cooklang.ParseString(dessertRecipe)

partyList, _ := cooklang.CreateShoppingList(appetizer, main, dessert)
guestList := partyList.Scale(float64(numberOfGuests) / 4.0)
```

### Unit Standardization

Convert all ingredients to a preferred unit system:

```go
// Convert all mass ingredients to kilograms
metricList, _ := cooklang.CreateShoppingListWithUnit("kg", recipes...)

// Convert all volume ingredients to milliliters
volumeList, _ := cooklang.CreateShoppingListWithUnit("ml", recipes...)
```

## Error Handling

All functions return appropriate errors for:

- Unit conversion failures
- Incompatible ingredient combinations
- Invalid recipe structures

Always check the returned error:

```go
shoppingList, err := cooklang.CreateShoppingList(recipes...)
if err != nil {
    // Handle error - perhaps some ingredients couldn't be consolidated
    log.Printf("Warning: %v", err)
}
```

## JSON Export

The `ShoppingList` struct supports JSON marshaling for API responses or storage:

```go
import "encoding/json"

jsonData, err := json.Marshal(shoppingList)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(jsonData))
```

## Tips

1. **Group Similar Recipes**: Consolidate recipes that share many ingredients for more efficient shopping.

2. **Use Scaling Strategically**: Scale individual recipes before combining, or scale the final shopping list - choose based on your needs.

3. **Check Counts**: Use `Count()` to quickly see how many unique ingredients you need.

4. **Verify Conversions**: When using `CreateShoppingListWithUnit`, be aware that not all ingredients may convert successfully.

5. **Map Format for Display**: Use `ToMap()` for easy iteration and display in user interfaces.

## Demo

Run the included demo to see all features in action:

```bash
go run cmd/shopping-list-demo/main.go
```

This will demonstrate:

- Creating a shopping list from multiple recipes
- Automatic ingredient consolidation
- Scaling for different serving sizes
- Unit conversion examples
- Different output formats
