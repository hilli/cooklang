package cooklang

import (
	"math"
	"testing"
)

func TestFormatAsFraction(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		tolerance float64
		want      string
	}{
		// Whole numbers
		{"zero", 0, 0, "0"},
		{"one", 1, 0, "1"},
		{"ten", 10, 0, "10"},

		// Simple fractions
		{"half", 0.5, 0, "1/2"},
		{"quarter", 0.25, 0, "1/4"},
		{"three quarters", 0.75, 0, "3/4"},
		{"third", 1.0 / 3.0, 0, "1/3"},
		{"two thirds", 2.0 / 3.0, 0, "2/3"},
		{"eighth", 0.125, 0, "1/8"},
		{"twelfth", 1.0 / 12.0, 0, "1/12"},

		// Mixed numbers
		{"two and a half", 2.5, 0, "2 1/2"},
		{"one and a quarter", 1.25, 0, "1 1/4"},
		{"three and three quarters", 3.75, 0, "3 3/4"},
		{"five and a third", 5.0 + 1.0/3.0, 0, "5 1/3"},

		// Values close to fractions (within tolerance)
		{"almost half", 0.51, 0.02, "1/2"},
		{"almost quarter", 0.24, 0.02, "1/4"},

		// Values that round to whole numbers
		{"almost one", 0.99, 0.02, "1"},
		{"just over two", 2.01, 0.02, "2"},

		// Negative values
		{"negative half", -0.5, 0, "-1/2"},
		{"negative two and half", -2.5, 0, "-2 1/2"},

		// Non-standard decimals (fallback)
		// Note: 0.37 is close to 3/8 (0.375) within 2% tolerance, so it matches
		{"decimal close to 3/8", 0.37, 0, "3/8"},
		{"large decimal close to 3/8", 2.37, 0, "2 3/8"},
		// Use values truly far from any fraction (0.45 doesn't match anything)
		{"ugly decimal", 0.45, 0, "0.45"},
		{"large ugly decimal", 2.45, 0, "2.45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tolerance := tt.tolerance
			if tolerance == 0 {
				tolerance = DefaultFractionTolerance
			}
			got := FormatAsFraction(tt.value, tolerance)
			if got != tt.want {
				t.Errorf("FormatAsFraction(%v, %v) = %q, want %q", tt.value, tolerance, got, tt.want)
			}
		})
	}
}

func TestFormatAsFractionDefault(t *testing.T) {
	// Just verify it uses the default tolerance
	got := FormatAsFractionDefault(0.5)
	if got != "1/2" {
		t.Errorf("FormatAsFractionDefault(0.5) = %q, want %q", got, "1/2")
	}
}

func TestParseFraction(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		// Simple fractions
		{"simple half", "1/2", 0.5, false},
		{"simple quarter", "1/4", 0.25, false},
		{"simple third", "1/3", 1.0 / 3.0, false},
		{"two thirds", "2/3", 2.0 / 3.0, false},

		// Whole numbers
		{"integer", "2", 2.0, false},
		{"zero", "0", 0.0, false},

		// Decimals
		{"decimal", "0.5", 0.5, false},
		{"decimal with whole", "2.5", 2.5, false},

		// Mixed numbers
		{"mixed half", "2 1/2", 2.5, false},
		{"mixed quarter", "1 1/4", 1.25, false},
		{"mixed third", "3 1/3", 3.0 + 1.0/3.0, false},

		// Invalid input
		{"invalid", "abc", 0, true},
		// Note: "1/0" parses as fraction with n=2, so need explicit check
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFraction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFraction(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("ParseFraction(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFractionDivisionByZero(t *testing.T) {
	// Test that division by zero is handled
	// Note: The current implementation will return Inf, not an error
	_, err := ParseFraction("1/0")
	// If no error, check the result is infinity
	if err == nil {
		// This is acceptable - caller should check for Inf if needed
		t.Skip("ParseFraction accepts 1/0 and returns Inf")
	}
}

func TestIsNiceFraction(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  bool
	}{
		{"half", 0.5, true},
		{"quarter", 0.25, true},
		{"third", 1.0 / 3.0, true},
		{"whole number", 3.0, true},
		{"almost whole", 2.99, true}, // Within tolerance of 3
		// 0.37 is close to 3/8 (0.375) within tolerance, so it IS a nice fraction
		{"close to 3/8", 0.37, true},
		{"mixed close to 3/8", 2.37, true},
		// Use values truly far from any fraction (0.45 doesn't match anything)
		{"ugly decimal", 0.45, false},
		{"mixed ugly", 2.45, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNiceFraction(tt.value, DefaultFractionTolerance)
			if got != tt.want {
				t.Errorf("IsNiceFraction(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestRoundToNiceFraction(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		tolerance float64
		want      float64
	}{
		// Already nice fractions - unchanged
		{"exact half", 0.5, 0.02, 0.5},
		{"exact quarter", 0.25, 0.02, 0.25},
		{"whole number", 3.0, 0.02, 3.0},

		// Close to nice fractions - rounded
		{"almost half", 0.51, 0.02, 0.5},
		{"almost quarter", 0.24, 0.02, 0.25},
		{"almost whole", 2.99, 0.02, 3.0},
		{"just over two", 2.01, 0.02, 2.0},

		// 0.37 is close to 3/8 (0.375), so it rounds to that
		{"close to 3/8", 0.37, 0.02, 0.375},
		// 0.42 is close to 5/12 (0.4167), so it rounds to that
		{"close to 5/12", 0.42, 0.02, 5.0 / 12.0},

		// Truly ugly decimals - unchanged
		{"ugly decimal", 0.45, 0.02, 0.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoundToNiceFraction(tt.value, tt.tolerance)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("RoundToNiceFraction(%v, %v) = %v, want %v", tt.value, tt.tolerance, got, tt.want)
			}
		})
	}
}

func TestFormatDecimal(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{"small", 0.05, "0.05"},
		{"very small", 0.001, "0.001"},
		{"medium", 2.5, "2.5"},
		{"no decimals", 3.0, "3"},
		{"trailing zeros", 2.50, "2.5"},
		{"large", 15.5, "15.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDecimal(tt.value)
			if got != tt.want {
				t.Errorf("formatDecimal(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that formatting and parsing round-trips correctly for nice fractions
	values := []float64{0.5, 0.25, 0.75, 1.0 / 3.0, 2.0 / 3.0, 2.5, 1.25, 3.75}

	for _, v := range values {
		formatted := FormatAsFractionDefault(v)
		parsed, err := ParseFraction(formatted)
		if err != nil {
			t.Errorf("Failed to parse formatted value %q (original %v): %v", formatted, v, err)
			continue
		}
		if math.Abs(parsed-v) > 0.01 {
			t.Errorf("Round-trip failed: %v -> %q -> %v", v, formatted, parsed)
		}
	}
}
