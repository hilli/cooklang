# Cooklang Go Library Examples

This document provides an overview of all the runnable code examples available in the cooklang Go library. These examples are executable and tested, ensuring they always work with the current version of the library.

## Running Examples

You can run any example using `go test`:

```bash
go test -v -run ExampleParseString
```

Or view them in the Go documentation:

```bash
go doc -all cooklang | less
```

## Available Examples

### Basic Parsing

#### ExampleParseString
Demonstrates basic recipe parsing from a string with YAML frontmatter.

**Key concepts:** Recipe parsing, metadata extraction

#### ExampleParseFile  
Shows how to parse a recipe from a file (demonstrated with string parsing for portability).

**Key concepts:** File-based recipe loading

### Working with Ingredients

#### ExampleRecipe_GetIngredients
Extracts all ingredients from a recipe and displays them with quantities and units.

**Key concepts:** Ingredient extraction, iteration

#### ExampleIngredient_ConvertTo
Demonstrates converting an ingredient from one unit to another (ml to cups).

**Key concepts:** Unit conversion

#### ExampleIngredient_CanConvertTo
Shows how to check if an ingredient can be converted to a specific unit before attempting conversion.

**Key concepts:** Conversion validation

#### ExampleIngredient_GetUnitType
Demonstrates determining the type of unit (mass, volume, etc.) for ingredients.

**Key concepts:** Unit classification

#### ExampleIngredient_Render
Shows how to render an ingredient back to Cooklang format, including annotations.

**Key concepts:** Cooklang syntax generation

### Ingredient Lists and Consolidation

#### ExampleIngredientList_ConvertToSystem
Converts all ingredients in a recipe to a specific unit system (metric, US, or imperial).

**Key concepts:** System-wide unit conversion

#### ExampleIngredientList_ConsolidateByName
Consolidates duplicate ingredients by name, combining quantities of compatible units.

**Key concepts:** Ingredient deduplication, quantity summation

#### ExampleIngredientList_ToMap
Converts an ingredient list to a map for easy display or processing.

**Key concepts:** Data structure conversion

#### ExampleRecipe_GetCollectedIngredients
One-step function to get all ingredients from a recipe consolidated by name.

**Key concepts:** Convenience methods, consolidated shopping lists

#### ExampleRecipe_GetCollectedIngredientsMap
Gets consolidated ingredients as a map ready for display in shopping lists.

**Key concepts:** Shopping list preparation

### Shopping Lists

#### ExampleCreateShoppingList
Creates a consolidated shopping list from multiple recipes, combining common ingredients.

**Key concepts:** Multi-recipe planning, ingredient aggregation

#### ExampleShoppingList_Scale
Demonstrates scaling a shopping list by a multiplier (e.g., doubling a recipe).

**Key concepts:** Recipe scaling, portion adjustment

#### ExampleRecipe_GetMetricShoppingList
Generates a shopping list with all ingredients converted to metric units.

**Key concepts:** Metric conversion, international compatibility

#### ExampleRecipe_GetUSShoppingList
Generates a shopping list with all ingredients converted to US customary units.

**Key concepts:** US unit conversion

### Cookware and Timers

#### ExampleRecipe_GetCookware
Extracts all cookware items needed for a recipe.

**Key concepts:** Equipment listing, cookware extraction

#### ExampleCookware
Demonstrates working with cookware objects, including quantities.

**Key concepts:** Cookware handling, equipment counts

#### ExampleTimer
Shows how to work with timers in recipes, including named and unnamed timers.

**Key concepts:** Timer extraction, recipe steps

### Metadata and Frontmatter

#### ExampleFrontmatterEditor_GetMetadata
Demonstrates accessing recipe metadata like title, difficulty, and prep time.

**Key concepts:** Metadata access, recipe properties

#### ExampleRecipe_Render
Basic recipe rendering showing how to display recipe information.

**Key concepts:** Recipe display, formatting

## Example Coverage

The examples cover the following areas:

- ✅ Recipe parsing (string, file)
- ✅ Ingredient extraction and manipulation
- ✅ Unit conversions (individual and system-wide)
- ✅ Ingredient consolidation
- ✅ Shopping list generation (single and multi-recipe)
- ✅ Recipe scaling
- ✅ Cookware extraction
- ✅ Timer handling
- ✅ Metadata access
- ✅ Cooklang format rendering

## Writing New Examples

When adding new examples, follow these guidelines:

1. **Naming:** Use `Example` prefix followed by the function/method name
2. **Self-contained:** Examples should be complete and runnable
3. **Output comments:** Include `// Output:` comments showing expected output
4. **Deterministic:** Ensure examples produce consistent output (avoid map iteration)
5. **Documented:** Add a brief comment explaining what the example demonstrates
6. **Simple:** Keep examples focused on one concept
7. **Realistic:** Use practical, real-world scenarios

Example template:

```go
// ExampleFunctionName demonstrates how to use FunctionName for a specific purpose.
func ExampleFunctionName() {
    recipeText := `Recipe content here`
    
    recipe, err := cooklang.ParseString(recipeText)
    if err != nil {
        log.Fatal(err)
    }
    
    // Demonstrate the functionality
    fmt.Println(recipe.Title)
    // Output:
    // Expected output here
}
```

## Testing Examples

All examples are automatically tested when you run:

```bash
go test ./...
```

Failed examples will show a diff between expected and actual output.

## Further Reading

- See the [main README](../README.md) for library overview
- Check [UNIT_CONVERSION.md](./UNIT_CONVERSION.md) for detailed unit conversion information
- Review [SHOPPING_LIST.md](./SHOPPING_LIST.md) for shopping list features
- Read [COLLECTED_INGREDIENTS.md](./COLLECTED_INGREDIENTS.md) for ingredient consolidation details
