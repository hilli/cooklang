package cooklang

import (
	"math"
	"strings"
)

// ConversionMode defines how precise unit conversions should be
type ConversionMode int

const (
	// PreciseMode uses exact scientific conversions (29.5735 ml/oz)
	PreciseMode ConversionMode = iota

	// BartenderMode uses practical bartender conversions (30 ml/oz)
	// with rounding to friendly values and smart unit selection
	BartenderMode
)

// UnitSystemUnknown represents an unknown or undetectable unit system
const UnitSystemUnknown UnitSystem = "unknown"

// Bartender-friendly conversion constants
// These are the practical conversions used by bartenders, not scientific ones
const (
	// Volume conversions (bartender-friendly)
	MlPerOz       = 30.0  // Bartender standard (scientific: 29.5735)
	MlPerTbsp     = 15.0  // 1 tablespoon
	MlPerTsp      = 5.0   // 1 teaspoon
	MlPerCup      = 240.0 // Standard US cup
	MlPerDash     = 0.92  // ~1ml, standard bar dash
	MlPerSplash   = 7.5   // Roughly 1/4 oz
	MlPerBarspoon = 5.0   // Same as teaspoon
	MlPerJigger   = 45.0  // Standard jigger (1.5 oz)
	MlPerPony     = 30.0  // Pony shot (1 oz)

	// Scientific conversions for PreciseMode
	MlPerOzPrecise = 29.5735
)

// CocktailUnit represents a unit commonly used in cocktail recipes
type CocktailUnit struct {
	Name       string     // Primary name (e.g., "dash")
	Aliases    []string   // Alternative names
	MlValue    float64    // Value in milliliters
	USValue    float64    // Value in fluid ounces (for bartender mode)
	System     UnitSystem // Which system this unit belongs to
	IsCocktail bool       // True if this is a cocktail-specific unit (dash, splash, etc.)
}

// cocktailUnits defines all cocktail-related units with their conversions
var cocktailUnits = []CocktailUnit{
	// Cocktail-specific units (no system - they're universal)
	{Name: "dash", Aliases: []string{"dashes"}, MlValue: MlPerDash, USValue: MlPerDash / MlPerOz, IsCocktail: true},
	{Name: "splash", Aliases: []string{"splashes"}, MlValue: MlPerSplash, USValue: MlPerSplash / MlPerOz, IsCocktail: true},
	{Name: "barspoon", Aliases: []string{"barspoons", "bar spoon", "bar spoons"}, MlValue: MlPerBarspoon, USValue: MlPerBarspoon / MlPerOz, IsCocktail: true},
	{Name: "jigger", Aliases: []string{"jiggers"}, MlValue: MlPerJigger, USValue: MlPerJigger / MlPerOz, IsCocktail: true},
	{Name: "pony", Aliases: []string{"ponies"}, MlValue: MlPerPony, USValue: MlPerPony / MlPerOz, IsCocktail: true},

	// US Volume units
	{Name: "fl oz", Aliases: []string{"fluid ounce", "fluid ounces", "fl. oz", "fl. oz."}, MlValue: MlPerOz, USValue: 1, System: UnitSystemUS},
	{Name: "oz", Aliases: []string{"ounce", "ounces"}, MlValue: MlPerOz, USValue: 1, System: UnitSystemUS},
	{Name: "tbsp", Aliases: []string{"tablespoon", "tablespoons", "T", "Tbsp"}, MlValue: MlPerTbsp, USValue: 0.5, System: UnitSystemUS},
	{Name: "tsp", Aliases: []string{"teaspoon", "teaspoons", "t"}, MlValue: MlPerTsp, USValue: MlPerTsp / MlPerOz, System: UnitSystemUS},
	{Name: "cup", Aliases: []string{"cups", "c"}, MlValue: MlPerCup, USValue: MlPerCup / MlPerOz, System: UnitSystemUS},
	{Name: "quart", Aliases: []string{"quarts", "qt"}, MlValue: 946.0, USValue: 32, System: UnitSystemUS},
	{Name: "pint", Aliases: []string{"pints", "pt"}, MlValue: 473.0, USValue: 16, System: UnitSystemUS},
	{Name: "gallon", Aliases: []string{"gallons", "gal"}, MlValue: 3785.0, USValue: 128, System: UnitSystemUS},

	// Metric Volume units
	{Name: "ml", Aliases: []string{"milliliter", "milliliters", "millilitre", "millilitres"}, MlValue: 1, USValue: 1 / MlPerOz, System: UnitSystemMetric},
	{Name: "cl", Aliases: []string{"centiliter", "centiliters", "centilitre", "centilitres"}, MlValue: 10, USValue: 10 / MlPerOz, System: UnitSystemMetric},
	{Name: "dl", Aliases: []string{"deciliter", "deciliters", "decilitre", "decilitres"}, MlValue: 100, USValue: 100 / MlPerOz, System: UnitSystemMetric},
	{Name: "l", Aliases: []string{"liter", "liters", "litre", "litres", "L"}, MlValue: 1000, USValue: 1000 / MlPerOz, System: UnitSystemMetric},
}

// GetCocktailUnit looks up a unit by name (case-insensitive)
func GetCocktailUnit(name string) *CocktailUnit {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for i := range cocktailUnits {
		if strings.ToLower(cocktailUnits[i].Name) == normalizedName {
			return &cocktailUnits[i]
		}
		for _, alias := range cocktailUnits[i].Aliases {
			if strings.ToLower(alias) == normalizedName {
				return &cocktailUnits[i]
			}
		}
	}
	return nil
}

// DetectUnitSystemFromUnit determines the unit system from a single unit name
func DetectUnitSystemFromUnit(unitName string) UnitSystem {
	unit := GetCocktailUnit(unitName)
	if unit != nil && unit.System != "" {
		return unit.System
	}
	return UnitSystemUnknown
}

// DetectIngredientListUnitSystem detects the dominant unit system in an ingredient list
func DetectIngredientListUnitSystem(il *IngredientList) UnitSystem {
	if il == nil {
		return UnitSystemUnknown
	}

	usCount := 0
	metricCount := 0

	for _, ing := range il.Ingredients {
		system := DetectUnitSystemFromUnit(ing.Unit)
		switch system {
		case UnitSystemUS:
			usCount++
		case UnitSystemMetric:
			metricCount++
		}
	}

	// Return the dominant system
	if usCount > metricCount {
		return UnitSystemUS
	} else if metricCount > usCount {
		return UnitSystemMetric
	}

	// Default to US if tied (most cocktail recipes are US-based)
	if usCount > 0 || metricCount > 0 {
		return UnitSystemUS
	}

	return UnitSystemUnknown
}

// SmartUnitResult contains the result of intelligent unit selection
type SmartUnitResult struct {
	Value    float64 // The numeric value
	Unit     string  // The selected unit name
	Original string  // Original formatted value for reference
}

// SelectBestUnit chooses the most appropriate unit for a given volume in ml
// This is used in bartender mode to pick human-friendly units
func SelectBestUnit(mlValue float64, targetSystem UnitSystem) SmartUnitResult {
	// For very small amounts (<= 3 ml), use dashes
	// This threshold captures amounts like 1/12 oz (2.5ml) = ~3 dashes
	if mlValue <= 3 && mlValue > 0 {
		dashes := mlValue / MlPerDash
		roundedDashes := math.Round(dashes)
		if roundedDashes < 1 {
			roundedDashes = 1
		}
		return SmartUnitResult{
			Value: roundedDashes,
			Unit:  "dash",
		}
	}

	if targetSystem == UnitSystemMetric {
		return selectBestMetricUnit(mlValue)
	}

	// For US system (default)
	return selectBestUSUnit(mlValue)
}

// selectBestMetricUnit picks the best metric unit
func selectBestMetricUnit(mlValue float64) SmartUnitResult {
	// Round to nearest 5 for amounts >= 10ml
	if mlValue >= 10 {
		rounded := math.Round(mlValue/5) * 5
		if rounded >= 1000 {
			return SmartUnitResult{Value: rounded / 1000, Unit: "l"}
		}
		if rounded >= 100 {
			return SmartUnitResult{Value: rounded / 10, Unit: "cl"}
		}
		return SmartUnitResult{Value: rounded, Unit: "ml"}
	}

	// For small amounts, round to nearest 2.5ml
	rounded := math.Round(mlValue/2.5) * 2.5
	if rounded < 1 {
		rounded = mlValue // Keep original if too small
	}
	return SmartUnitResult{Value: rounded, Unit: "ml"}
}

// selectBestUSUnit picks the best US unit for a given ml value
func selectBestUSUnit(mlValue float64) SmartUnitResult {
	ozValue := mlValue / MlPerOz

	// Very small amounts: use dashes (already handled in SelectBestUnit)
	if mlValue <= 3 {
		dashes := mlValue / MlPerDash
		roundedDashes := math.Round(dashes)
		if roundedDashes < 1 {
			roundedDashes = 1
		}
		return SmartUnitResult{Value: roundedDashes, Unit: "dash"}
	}

	// Small amounts (up to ~7ml / 1/4 oz): consider barspoons
	if mlValue < 10 {
		// Try barspoons for amounts that work well with them
		barspoons := mlValue / MlPerBarspoon
		if barspoons <= 2 && IsNiceFraction(barspoons, 0.1) {
			return SmartUnitResult{Value: RoundToNiceFraction(barspoons, 0.1), Unit: "barspoon"}
		}
		// Fall back to oz fraction
		return SmartUnitResult{Value: RoundToNiceFraction(ozValue, 0.05), Unit: "oz"}
	}

	// Standard drink amounts: prefer oz with nice fractions
	if mlValue <= 240 { // Up to 1 cup
		rounded := RoundToNiceFraction(ozValue, 0.05)
		return SmartUnitResult{Value: rounded, Unit: "oz"}
	}

	// Larger amounts: use cups
	cups := mlValue / MlPerCup
	if cups >= 1 {
		rounded := RoundToNiceFraction(cups, 0.05)
		return SmartUnitResult{Value: rounded, Unit: "cup"}
	}

	return SmartUnitResult{Value: RoundToNiceFraction(ozValue, 0.05), Unit: "oz"}
}

// ConvertVolumeBartender converts a volume from one unit to another using bartender-friendly rounding
func ConvertVolumeBartender(value float64, fromUnit string, toSystem UnitSystem) SmartUnitResult {
	// First convert to ml
	fromUnitInfo := GetCocktailUnit(fromUnit)
	if fromUnitInfo == nil {
		// Unknown unit, return as-is
		return SmartUnitResult{Value: value, Unit: fromUnit}
	}

	mlValue := value * fromUnitInfo.MlValue

	// Then select the best unit in the target system
	return SelectBestUnit(mlValue, toSystem)
}

// FormatBartenderValue formats a value with the appropriate unit for display
func FormatBartenderValue(result SmartUnitResult) string {
	// Use fraction formatting for nice display
	valueStr := FormatAsFractionDefault(result.Value)

	// Handle pluralization for certain units
	unit := result.Unit
	if result.Value != 1 && !strings.HasSuffix(unit, "s") {
		// Simple pluralization for common units
		switch unit {
		case "dash":
			unit = "dashes"
		case "splash":
			unit = "splashes"
		case "cup":
			unit = "cups"
		case "barspoon":
			unit = "barspoons"
		}
	}

	return valueStr + " " + unit
}

// IsCocktailSpecificUnit returns true if the unit is cocktail-specific (dash, splash, etc.)
func IsCocktailSpecificUnit(unitName string) bool {
	unit := GetCocktailUnit(unitName)
	return unit != nil && unit.IsCocktail
}

// ShouldSkipConversion returns true if conversion should be skipped
// (e.g., source and target systems are the same)
func ShouldSkipConversion(sourceUnit string, targetSystem UnitSystem) bool {
	sourceSystem := DetectUnitSystemFromUnit(sourceUnit)

	// Skip if source system matches target system
	if sourceSystem == targetSystem && targetSystem != UnitSystemUnknown {
		return true
	}

	// Skip for cocktail-specific units (they're universal)
	if IsCocktailSpecificUnit(sourceUnit) {
		return true
	}

	return false
}
