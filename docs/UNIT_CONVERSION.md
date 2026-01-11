# Unit System Conversion

The Cooklang Go library supports easy conversion between metric, imperial, and US unit systems for recipes. This makes it simple to display recipes in different unit systems based on user preference or regional standards.

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

## Bartender Mode

Bartender mode provides practical, bartender-friendly unit conversions optimized for cocktail recipes. Instead of using scientifically precise conversions, it uses the rounded values that bartenders actually work with.

### Why Bartender Mode?

Standard scientific conversions (e.g., 1 oz = 29.5735 ml) produce awkward values that are impractical for bartending. Bartender mode uses industry-standard rounded conversions:

| Conversion | Scientific | Bartender Mode |
|------------|------------|----------------|
| 1 oz → ml  | 29.5735 ml | 30 ml          |
| 1 tbsp → ml| 14.787 ml  | 15 ml          |
| 1 tsp → ml | 4.929 ml   | 5 ml           |

### Cocktail-Specific Units

Bartender mode recognizes cocktail-specific units that aren't part of standard measurement systems:

| Unit     | Milliliters | Description                    |
|----------|-------------|--------------------------------|
| dash     | ~0.92 ml    | Standard bar dash              |
| splash   | ~7.5 ml     | Roughly 1/4 oz                 |
| barspoon | 5 ml        | Same as teaspoon               |
| jigger   | 45 ml       | Standard jigger (1.5 oz)       |
| pony     | 30 ml       | Pony shot (1 oz)               |

These units are considered "universal" and are not converted between systems — a dash is a dash regardless of whether you're using metric or US units.

### Smart Unit Selection

Bartender mode automatically selects the most appropriate unit based on quantity:

- **Very small amounts (≤3 ml)**: Converted to dashes
  - Example: `1/12 fl oz` → `3 dashes`
- **Small amounts (<10 ml)**: Uses barspoons when appropriate, otherwise oz fractions
- **Standard drink amounts (≤240 ml)**: Uses oz with nice fractions (1/4, 1/2, 3/4, etc.)
- **Larger amounts (>240 ml)**: Uses cups

Values are automatically rounded to bartender-friendly fractions rather than awkward decimals.

### Usage Examples for Bartender Mode

#### Ingredient-Level Conversion in Bartender Mode

```go
ingredients := recipe.GetIngredients()

// Convert using bartender-friendly rounding
usIngredients := ingredients.ConvertToSystemBartender(cooklang.UnitSystemUS)
metricIngredients := ingredients.ConvertToSystemBartender(cooklang.UnitSystemMetric)

// Individual ingredient conversion
gin := ingredients.Ingredients[0]
ginMetric := gin.ConvertToSystemBartender(cooklang.UnitSystemMetric)
```

#### Direct Volume Conversion

```go
// Convert 2 oz to metric with bartender-friendly rounding
result := cooklang.ConvertVolumeBartender(2.0, "oz", cooklang.UnitSystemMetric)
// result.Value = 60, result.Unit = "ml"

// Format for display
formatted := cooklang.FormatBartenderValue(result)
// formatted = "60 ml"
```

#### Smart Unit Selection

```go
// Let the library choose the best unit for 2.5 ml
result := cooklang.SelectBestUnit(2.5, cooklang.UnitSystemUS)
// result.Value = 3, result.Unit = "dash"

// Best unit for 45 ml in US system
result = cooklang.SelectBestUnit(45, cooklang.UnitSystemUS)
// result.Value = 1.5, result.Unit = "oz"
```

### Bartender Mode API Reference

#### IngredientList Methods

- `ConvertToSystemBartender(system UnitSystem) *IngredientList` - Convert all ingredients using bartender-friendly conversions

#### Ingredient Methods

- `ConvertToSystemBartender(system UnitSystem) *Ingredient` - Convert using bartender-friendly conversions
- `FormatQuantityBartender() string` - Format quantity with fractions and proper pluralization

#### Conversion Functions

- `ConvertVolumeBartender(value float64, fromUnit string, toSystem UnitSystem) SmartUnitResult` - Convert volume with bartender rounding
- `SelectBestUnit(mlValue float64, targetSystem UnitSystem) SmartUnitResult` - Choose the most appropriate unit for a volume
- `FormatBartenderValue(result SmartUnitResult) string` - Format a conversion result for display

#### Unit Lookup Functions

- `GetCocktailUnit(name string) *CocktailUnit` - Look up unit information by name (case-insensitive)
- `IsCocktailSpecificUnit(unitName string) bool` - Check if a unit is cocktail-specific (dash, splash, etc.)
- `DetectUnitSystemFromUnit(unitName string) UnitSystem` - Determine unit system from a unit name
- `DetectIngredientListUnitSystem(il *IngredientList) UnitSystem` - Detect dominant unit system in an ingredient list

### Conversion Modes

The library supports two conversion modes defined by the `ConversionMode` type:

```go
const (
    PreciseMode   ConversionMode = iota  // Scientific conversions (29.5735 ml/oz)
    BartenderMode                         // Practical conversions (30 ml/oz)
)
```

Use `PreciseMode` when scientific accuracy matters (e.g., baking chemistry). Use `BartenderMode` for cocktails and general cooking where practical measurements are preferred.
