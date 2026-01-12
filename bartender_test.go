package cooklang

import (
	"math"
	"testing"
)

func TestGetCocktailUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantNil  bool
	}{
		// Cocktail units
		{"dash", "dash", "dash", false},
		{"dashes", "dashes", "dash", false},
		{"splash", "splash", "splash", false},
		{"barspoon", "barspoon", "barspoon", false},
		{"bar spoon", "bar spoon", "barspoon", false},
		{"jigger", "jigger", "jigger", false},

		// US units
		{"oz", "oz", "oz", false},
		{"ounce", "ounce", "oz", false},
		{"fl oz", "fl oz", "fl oz", false},
		{"fluid ounce", "fluid ounce", "fl oz", false},
		{"tbsp", "tbsp", "tbsp", false},
		{"tablespoon", "tablespoon", "tbsp", false},
		{"tsp", "tsp", "tsp", false},
		{"cup", "cup", "cup", false},

		// Metric units
		{"ml", "ml", "ml", false},
		{"milliliter", "milliliter", "ml", false},
		{"cl", "cl", "cl", false},
		{"l", "l", "l", false},
		{"liter", "liter", "l", false},

		// Case insensitive
		{"OZ uppercase", "OZ", "oz", false},
		{"Ml mixed case", "Ml", "ml", false},

		// Unknown
		{"unknown", "pieces", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCocktailUnit(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetCocktailUnit(%q) = %v, want nil", tt.input, got.Name)
				}
				return
			}
			if got == nil {
				t.Errorf("GetCocktailUnit(%q) = nil, want %q", tt.input, tt.wantName)
				return
			}
			if got.Name != tt.wantName {
				t.Errorf("GetCocktailUnit(%q).Name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}

func TestDetectUnitSystemFromUnit(t *testing.T) {
	tests := []struct {
		name string
		unit string
		want UnitSystem
	}{
		// US units
		{"oz is US", "oz", UnitSystemUS},
		{"fl oz is US", "fl oz", UnitSystemUS},
		{"cup is US", "cup", UnitSystemUS},
		{"tbsp is US", "tbsp", UnitSystemUS},
		{"tsp is US", "tsp", UnitSystemUS},

		// Metric units
		{"ml is metric", "ml", UnitSystemMetric},
		{"cl is metric", "cl", UnitSystemMetric},
		{"l is metric", "l", UnitSystemMetric},

		// Cocktail units have no system (universal)
		{"dash has no system", "dash", UnitSystemUnknown},
		{"splash has no system", "splash", UnitSystemUnknown},
		{"barspoon has no system", "barspoon", UnitSystemUnknown},

		// Unknown
		{"unknown unit", "pieces", UnitSystemUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectUnitSystemFromUnit(tt.unit)
			if got != tt.want {
				t.Errorf("DetectUnitSystemFromUnit(%q) = %q, want %q", tt.unit, got, tt.want)
			}
		})
	}
}

func TestSelectBestUnit(t *testing.T) {
	tests := []struct {
		name         string
		mlValue      float64
		targetSystem UnitSystem
		wantUnit     string
		wantValue    float64
		tolerance    float64 // For float comparison
	}{
		// Very small amounts -> dashes (threshold is <= 3ml)
		{"1ml to US = 1 dash", 1.0, UnitSystemUS, "dash", 1, 0.1},
		{"2ml to US = 2 dashes", 2.0, UnitSystemUS, "dash", 2, 0.1},
		{"2.5ml to US = 3 dashes", 2.5, UnitSystemUS, "dash", 3, 0.1}, // 2.5/0.92 ≈ 2.7 → rounds to 3
		{"3ml to US = 3 dashes", 3.0, UnitSystemUS, "dash", 3, 0.1},

		// Small amounts with barspoons
		{"5ml to US = 1 barspoon", 5.0, UnitSystemUS, "barspoon", 1, 0.1},
		{"7.5ml to US = 1.5 barspoons", 7.5, UnitSystemUS, "barspoon", 1.5, 0.1},

		// Standard drink amounts
		{"15ml to US = 0.5 oz", 15.0, UnitSystemUS, "oz", 0.5, 0.05},
		{"30ml to US = 1 oz", 30.0, UnitSystemUS, "oz", 1.0, 0.05},
		{"45ml to US = 1.5 oz", 45.0, UnitSystemUS, "oz", 1.5, 0.05},
		{"60ml to US = 2 oz", 60.0, UnitSystemUS, "oz", 2.0, 0.05},
		{"240ml to US = 8 oz", 240.0, UnitSystemUS, "oz", 8.0, 0.1}, // 8 oz is fine, not cups

		// Metric conversions
		{"30ml to metric = 30ml", 30.0, UnitSystemMetric, "ml", 30, 1},
		{"45ml to metric = 45ml", 45.0, UnitSystemMetric, "ml", 45, 1},
		{"100ml to metric = 10cl", 100.0, UnitSystemMetric, "cl", 10, 1},
		{"1000ml to metric = 1l", 1000.0, UnitSystemMetric, "l", 1, 0.1},

		// Metric rounding: nearest 2.5ml for <30ml, nearest 5ml for >=30ml
		{"22ml to metric = 22.5ml", 22.0, UnitSystemMetric, "ml", 22.5, 0.1},
		{"22.5ml to metric = 22.5ml", 22.5, UnitSystemMetric, "ml", 22.5, 0.1},
		{"32ml to metric = 30ml", 32.0, UnitSystemMetric, "ml", 30, 1},
		{"58ml to metric = 60ml", 58.0, UnitSystemMetric, "ml", 60, 1},

		// Large US amounts -> cups
		{"480ml to US = 2 cups", 480.0, UnitSystemUS, "cup", 2, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectBestUnit(tt.mlValue, tt.targetSystem)
			if got.Unit != tt.wantUnit {
				t.Errorf("SelectBestUnit(%v, %v).Unit = %q, want %q", tt.mlValue, tt.targetSystem, got.Unit, tt.wantUnit)
			}
			if math.Abs(got.Value-tt.wantValue) > tt.tolerance {
				t.Errorf("SelectBestUnit(%v, %v).Value = %v, want %v (tolerance %v)", tt.mlValue, tt.targetSystem, got.Value, tt.wantValue, tt.tolerance)
			}
		})
	}
}

func TestConvertVolumeBartender(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fromUnit  string
		toSystem  UnitSystem
		wantUnit  string
		wantValue float64
		tolerance float64
	}{
		// US to metric
		{"1 oz to metric = 30ml", 1.0, "oz", UnitSystemMetric, "ml", 30, 1},
		{"2 oz to metric = 60ml", 2.0, "oz", UnitSystemMetric, "ml", 60, 1},
		{"0.5 oz to metric = 15ml", 0.5, "oz", UnitSystemMetric, "ml", 15, 1},

		// Metric to US
		{"30ml to US = 1 oz", 30.0, "ml", UnitSystemUS, "oz", 1, 0.1},
		{"60ml to US = 2 oz", 60.0, "ml", UnitSystemUS, "oz", 2, 0.1},
		{"15ml to US = 0.5 oz", 15.0, "ml", UnitSystemUS, "oz", 0.5, 0.1},

		// Very small amounts become dashes
		{"1ml to US = 1 dash", 1.0, "ml", UnitSystemUS, "dash", 1, 0.5},

		// Unknown unit returns as-is
		{"unknown unit", 5.0, "pieces", UnitSystemUS, "pieces", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertVolumeBartender(tt.value, tt.fromUnit, tt.toSystem)
			if got.Unit != tt.wantUnit {
				t.Errorf("ConvertVolumeBartender(%v, %q, %v).Unit = %q, want %q", tt.value, tt.fromUnit, tt.toSystem, got.Unit, tt.wantUnit)
			}
			if math.Abs(got.Value-tt.wantValue) > tt.tolerance {
				t.Errorf("ConvertVolumeBartender(%v, %q, %v).Value = %v, want %v", tt.value, tt.fromUnit, tt.toSystem, got.Value, tt.wantValue)
			}
		})
	}
}

func TestFormatBartenderValue(t *testing.T) {
	tests := []struct {
		name   string
		result SmartUnitResult
		want   string
	}{
		{"1 oz", SmartUnitResult{Value: 1, Unit: "oz"}, "1 oz"},
		{"2 oz", SmartUnitResult{Value: 2, Unit: "oz"}, "2 oz"},
		{"1/2 oz", SmartUnitResult{Value: 0.5, Unit: "oz"}, "1/2 oz"},
		{"1 1/2 oz", SmartUnitResult{Value: 1.5, Unit: "oz"}, "1 1/2 oz"},
		{"1 dash", SmartUnitResult{Value: 1, Unit: "dash"}, "1 dash"},
		{"2 dashes", SmartUnitResult{Value: 2, Unit: "dash"}, "2 dashes"},
		{"30 ml", SmartUnitResult{Value: 30, Unit: "ml"}, "30 ml"},
		{"1 cup", SmartUnitResult{Value: 1, Unit: "cup"}, "1 cup"},
		{"2 cups", SmartUnitResult{Value: 2, Unit: "cup"}, "2 cups"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBartenderValue(tt.result)
			if got != tt.want {
				t.Errorf("FormatBartenderValue(%+v) = %q, want %q", tt.result, got, tt.want)
			}
		})
	}
}

func TestIsCocktailSpecificUnit(t *testing.T) {
	tests := []struct {
		name string
		unit string
		want bool
	}{
		{"dash is cocktail", "dash", true},
		{"splash is cocktail", "splash", true},
		{"barspoon is cocktail", "barspoon", true},
		{"jigger is cocktail", "jigger", true},
		{"oz is not cocktail", "oz", false},
		{"ml is not cocktail", "ml", false},
		{"cup is not cocktail", "cup", false},
		{"unknown is not cocktail", "pieces", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCocktailSpecificUnit(tt.unit)
			if got != tt.want {
				t.Errorf("IsCocktailSpecificUnit(%q) = %v, want %v", tt.unit, got, tt.want)
			}
		})
	}
}

func TestShouldSkipConversion(t *testing.T) {
	tests := []struct {
		name       string
		sourceUnit string
		targetSys  UnitSystem
		want       bool
	}{
		// Same system - skip
		{"oz to US = skip", "oz", UnitSystemUS, true},
		{"ml to metric = skip", "ml", UnitSystemMetric, true},
		{"cup to US = skip", "cup", UnitSystemUS, true},

		// Different system - don't skip
		{"oz to metric = convert", "oz", UnitSystemMetric, false},
		{"ml to US = convert", "ml", UnitSystemUS, false},

		// Cocktail units - skip (they're universal)
		{"dash to US = skip", "dash", UnitSystemUS, true},
		{"dash to metric = skip", "dash", UnitSystemMetric, true},
		{"splash to US = skip", "splash", UnitSystemUS, true},

		// Unknown source unit - don't skip (let caller decide)
		{"unknown to US = convert", "pieces", UnitSystemUS, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSkipConversion(tt.sourceUnit, tt.targetSys)
			if got != tt.want {
				t.Errorf("ShouldSkipConversion(%q, %v) = %v, want %v", tt.sourceUnit, tt.targetSys, got, tt.want)
			}
		})
	}
}

func TestDetectIngredientListUnitSystem(t *testing.T) {
	tests := []struct {
		name  string
		units []string
		want  UnitSystem
	}{
		{"all US units", []string{"oz", "cup", "tbsp"}, UnitSystemUS},
		{"all metric units", []string{"ml", "cl", "l"}, UnitSystemMetric},
		{"mixed mostly US", []string{"oz", "oz", "ml"}, UnitSystemUS},
		{"mixed mostly metric", []string{"ml", "ml", "oz"}, UnitSystemMetric},
		{"tied defaults to US", []string{"oz", "ml"}, UnitSystemUS},
		{"cocktail units only = unknown", []string{"dash", "splash"}, UnitSystemUnknown},
		{"empty = unknown", []string{}, UnitSystemUnknown},
		{"nil = unknown", nil, UnitSystemUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var il *IngredientList
			if tt.units != nil {
				il = NewIngredientList()
				for i, unit := range tt.units {
					il.Add(&Ingredient{
						Name: "ingredient" + string(rune('A'+i)),
						Unit: unit,
					})
				}
			}
			got := DetectIngredientListUnitSystem(il)
			if got != tt.want {
				t.Errorf("DetectIngredientListUnitSystem with units %v = %q, want %q", tt.units, got, tt.want)
			}
		})
	}
}

// TestRealWorldConversions tests conversions that match real cocktail scenarios
func TestRealWorldConversions(t *testing.T) {
	tests := []struct {
		name       string
		value      float64
		fromUnit   string
		toSystem   UnitSystem
		wantFormat string // Expected formatted output
	}{
		// The problematic case from the issue: 1/12 fl oz should become "3 dashes"
		// 1/12 oz = 2.5 ml, 2.5 / 0.92 ≈ 2.7, rounds to 3 dashes
		{"1/12 oz to US = 3 dashes", 1.0 / 12.0, "oz", UnitSystemUS, "3 dashes"},

		// Common cocktail measurements
		{"1.5 oz to metric = 45 ml", 1.5, "oz", UnitSystemMetric, "45 ml"},
		{"2 oz to metric = 60 ml", 2.0, "oz", UnitSystemMetric, "60 ml"},
		// 0.75 oz = 22.5 ml, stays at 22.5ml (nearest 2.5 for <30ml)
		{"0.75 oz to metric = 22 1/2 ml", 0.75, "oz", UnitSystemMetric, "22 1/2 ml"},

		// Metric to US
		{"30 ml to US = 1 oz", 30.0, "ml", UnitSystemUS, "1 oz"},
		{"45 ml to US = 1 1/2 oz", 45.0, "ml", UnitSystemUS, "1 1/2 oz"},
		{"60 ml to US = 2 oz", 60.0, "ml", UnitSystemUS, "2 oz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertVolumeBartender(tt.value, tt.fromUnit, tt.toSystem)
			formatted := FormatBartenderValue(result)
			if formatted != tt.wantFormat {
				t.Errorf("ConvertVolumeBartender(%v, %q, %v) formatted = %q, want %q",
					tt.value, tt.fromUnit, tt.toSystem, formatted, tt.wantFormat)
			}
		})
	}
}
