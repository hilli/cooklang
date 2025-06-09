# Unit System Conversion

The Cooklang Go library now supports easy conversion between metric, imperial, and US unit systems for recipes. This makes it simple to display recipes in different unit systems based on user preference or regional standards.

## Features

### Supported Unit Systems

- **Metric**: grams (g), kilograms (kg), milliliters (ml), liters (l)
- **US**: ounces (oz), pounds (lb), teaspoons (tsp), tablespoons (tbsp), cups (cup), quarts (qt)
- **Imperial**: ounces (oz), pounds (lb), teaspoons (tsp), tablespoons (tbsp), fluid ounces (fl_oz), pints (pt)

### Smart Unit Selection

The system automatically selects the most appropriate unit based on quantity:

- **Volume**:
  - Very small (< 15ml): teaspoons
  - Small (15-60ml): tablespoons
  - Medium (60-1000ml): cups (US) or fluid ounces (Imperial)
  - Large (> 1000ml): quarts (US) or pints (Imperial)

- **Mass**:
  - Large quantities (> 1000g): kilograms (metric)
  - All other mass: grams (metric) or ounces (US/Imperial)

## Usage Examples

### Recipe-Level Conversion

```go
recipe, _ := cooklang.ParseString(recipeText)

// Get shopping lists in different unit systems
metricList, _ := recipe.GetMetricShoppingList()
usList, _ := recipe.GetUSShoppingList() 
imperialList, _ := recipe.GetImperialShoppingList()

// Or convert to any specific system
anySystemList, _ := recipe.GetShoppingListInSystem(cooklang.UnitSystemUS)
```

### Ingredient-Level Conversion

```go
ingredients := recipe.GetIngredients()

// Convert entire ingredient list to a system
usIngredients := ingredients.ConvertToSystem(cooklang.UnitSystemUS)

// Convert individual ingredients
flour := ingredients.Ingredients[0]
usFlour := flour.ConvertToSystem(cooklang.UnitSystemUS)

// Convert to specific unit
flourInPounds, err := flour.ConvertTo("lb")
```

### Conversion with Consolidation

```go
// Convert to US units and consolidate duplicate ingredients
consolidated, err := ingredients.ConvertToSystemWithConsolidation(cooklang.UnitSystemUS)
```

## Supported Conversions

The library supports conversions between:

- **Volume units**: ml, l, tsp, tbsp, cup, fl_oz, pt, qt, gallon
- **Mass units**: g, kg, oz, lb
- **Go-units library**: For other scientific units when available

## API Reference

### Recipe Methods
- `GetMetricShoppingList() (map[string]string, error)`
- `GetUSShoppingList() (map[string]string, error)`
- `GetImperialShoppingList() (map[string]string, error)`
- `GetShoppingListInSystem(system UnitSystem) (map[string]string, error)`

### IngredientList Methods
- `ConvertToSystem(system UnitSystem) *IngredientList`
- `ConvertToSystemWithConsolidation(system UnitSystem) (*IngredientList, error)`

### Ingredient Methods
- `ConvertToSystem(system UnitSystem) *Ingredient`
- `ConvertTo(targetUnit string) (*Ingredient, error)`
- `CanConvertTo(targetUnit string) bool`
- `GetUnitType() string`

### Unit Systems
- `UnitSystemMetric`
- `UnitSystemUS`
- `UnitSystemImperial`

## Error Handling

The conversion functions gracefully handle:
- Ingredients without units
- Ingredients with "some" quantities
- Incompatible unit conversions
- Unknown units

When conversion fails, ingredients are returned unchanged rather than causing errors.
