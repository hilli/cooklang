package cooklang

import (
	"fmt"
	"math"
)

// fractionEntry represents a common fraction with its decimal value
type fractionEntry struct {
	numerator   int
	denominator int
	value       float64
}

// commonFractions lists fractions in order of preference (simpler fractions first)
// This ordering ensures we prefer 1/2 over 4/8, etc.
var commonFractions = []fractionEntry{
	{1, 2, 0.5},
	{1, 4, 0.25},
	{3, 4, 0.75},
	{1, 3, 1.0 / 3.0},
	{2, 3, 2.0 / 3.0},
	{1, 8, 0.125},
	{3, 8, 0.375},
	{5, 8, 0.625},
	{7, 8, 0.875},
	{1, 6, 1.0 / 6.0},
	{5, 6, 5.0 / 6.0},
	{1, 12, 1.0 / 12.0},
	{5, 12, 5.0 / 12.0},
	{7, 12, 7.0 / 12.0},
	{11, 12, 11.0 / 12.0},
}

// DefaultFractionTolerance is the default tolerance for matching fractions
const DefaultFractionTolerance = 0.02

// FormatAsFraction converts a float to a human-readable fraction string.
// It handles whole numbers, simple fractions, and mixed numbers.
//
// Examples:
//   - 0.5 → "1/2"
//   - 2.5 → "2 1/2"
//   - 0.0833 → "1/12"
//   - 3.0 → "3"
//   - 1.234 → "1.23" (fallback for non-standard values)
//
// Parameters:
//   - value: The numeric value to format
//   - tolerance: How close a value must be to a fraction to match (e.g., 0.02 = 2%)
//
// Returns:
//   - A human-readable string representation
func FormatAsFraction(value float64, tolerance float64) string {
	if tolerance <= 0 {
		tolerance = DefaultFractionTolerance
	}

	// Handle negative values
	if value < 0 {
		return "-" + FormatAsFraction(-value, tolerance)
	}

	// Handle zero
	if value == 0 {
		return "0"
	}

	// Extract whole number part
	whole := int(value)
	frac := value - float64(whole)

	// If it's very close to a whole number, return just the whole number
	if frac < tolerance {
		return fmt.Sprintf("%d", whole)
	}

	// If the fractional part is very close to 1, round up
	if frac > 1.0-tolerance {
		return fmt.Sprintf("%d", whole+1)
	}

	// Try to match the fractional part to a common fraction
	for _, f := range commonFractions {
		if math.Abs(frac-f.value) < tolerance {
			if whole > 0 {
				return fmt.Sprintf("%d %d/%d", whole, f.numerator, f.denominator)
			}
			return fmt.Sprintf("%d/%d", f.numerator, f.denominator)
		}
	}

	// Fallback: format as decimal
	if whole > 0 {
		// Format fractional part with 1-2 decimal places
		formatted := formatDecimal(frac)
		if formatted == "0" {
			return fmt.Sprintf("%d", whole)
		}
		// Return as decimal (e.g., "2.75" not "2 0.75")
		return formatDecimal(value)
	}

	return formatDecimal(value)
}

// FormatAsFractionDefault uses the default tolerance for fraction matching.
// This is a convenience wrapper around FormatAsFraction with DefaultFractionTolerance.
//
// Parameters:
//   - value: The numeric value to format
//
// Returns:
//   - A human-readable string representation
//
// Example:
//
//	cooklang.FormatAsFractionDefault(0.5)  // "1/2"
//	cooklang.FormatAsFractionDefault(2.25) // "2 1/4"
//	cooklang.FormatAsFractionDefault(3.0)  // "3"
func FormatAsFractionDefault(value float64) string {
	return FormatAsFraction(value, DefaultFractionTolerance)
}

// formatDecimal formats a decimal number nicely, removing unnecessary trailing zeros.
func formatDecimal(value float64) string {
	// For very small values, use more precision
	if value > 0 && value < 0.1 {
		formatted := fmt.Sprintf("%.3f", value)
		return trimTrailingZeros(formatted)
	}

	// For values less than 10, use 2 decimal places
	if value < 10 {
		formatted := fmt.Sprintf("%.2f", value)
		return trimTrailingZeros(formatted)
	}

	// For larger values, use 1 decimal place
	formatted := fmt.Sprintf("%.1f", value)
	return trimTrailingZeros(formatted)
}

// trimTrailingZeros removes unnecessary trailing zeros and decimal point from a formatted number.
func trimTrailingZeros(s string) string {
	// Find the decimal point
	dotIndex := -1
	for i, c := range s {
		if c == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		return s // No decimal point
	}

	// Remove trailing zeros
	end := len(s)
	for end > dotIndex+1 && s[end-1] == '0' {
		end--
	}

	// Remove decimal point if no fractional part remains
	if end == dotIndex+1 {
		end = dotIndex
	}

	return s[:end]
}

// ParseFraction parses a fraction string into a float64.
// Handles multiple formats: simple fractions ("1/2"), mixed numbers ("2 1/2"),
// decimals ("0.5"), and integers ("2").
//
// Parameters:
//   - s: The string to parse
//
// Returns:
//   - float64: The parsed numeric value
//   - error: An error if the string cannot be parsed
//
// Example:
//
//	cooklang.ParseFraction("1/2")     // 0.5, nil
//	cooklang.ParseFraction("2 1/2")   // 2.5, nil
//	cooklang.ParseFraction("0.75")    // 0.75, nil
//	cooklang.ParseFraction("invalid") // 0, error
func ParseFraction(s string) (float64, error) {
	// Try parsing as mixed number first (e.g., "2 1/2")
	var whole, num, den int
	if n, err := fmt.Sscanf(s, "%d %d/%d", &whole, &num, &den); err == nil && n == 3 && den != 0 {
		return float64(whole) + float64(num)/float64(den), nil
	}

	// Try parsing as simple fraction (e.g., "1/2")
	if n, err := fmt.Sscanf(s, "%d/%d", &num, &den); err == nil && n == 2 && den != 0 {
		return float64(num) / float64(den), nil
	}

	// Try parsing as simple float/integer last
	var value float64
	if _, err := fmt.Sscanf(s, "%f", &value); err == nil {
		return value, nil
	}

	return 0, fmt.Errorf("cannot parse fraction: %s", s)
}

// IsNiceFraction checks if a value is close to a common fraction.
// This is useful for determining if a value will format nicely as a fraction.
//
// Parameters:
//   - value: The numeric value to check
//   - tolerance: How close to a fraction the value must be (e.g., 0.02 = 2%)
//
// Returns:
//   - bool: true if the value matches a common fraction within tolerance
//
// Example:
//
//	cooklang.IsNiceFraction(0.5, 0.02)   // true (1/2)
//	cooklang.IsNiceFraction(0.333, 0.02) // true (1/3)
//	cooklang.IsNiceFraction(0.37, 0.02)  // false
func IsNiceFraction(value float64, tolerance float64) bool {
	if tolerance <= 0 {
		tolerance = DefaultFractionTolerance
	}

	// Check if it's a whole number
	frac := value - float64(int(value))
	if frac < tolerance || frac > 1.0-tolerance {
		return true
	}

	// Check common fractions
	for _, f := range commonFractions {
		if math.Abs(frac-f.value) < tolerance {
			return true
		}
	}

	return false
}

// RoundToNiceFraction rounds a value to the nearest common fraction.
// This is useful for bartender mode where clean measurements are preferred.
// If no common fraction is within tolerance, the original value is returned.
//
// Parameters:
//   - value: The numeric value to round
//   - tolerance: Maximum difference to accept for rounding (e.g., 0.02 = 2%)
//
// Returns:
//   - float64: The rounded value, or original if no good match
//
// Example:
//
//	cooklang.RoundToNiceFraction(0.48, 0.05)  // 0.5 (rounds to 1/2)
//	cooklang.RoundToNiceFraction(0.26, 0.02)  // 0.25 (rounds to 1/4)
//	cooklang.RoundToNiceFraction(0.37, 0.02)  // 0.37 (no good match)
func RoundToNiceFraction(value float64, tolerance float64) float64 {
	if tolerance <= 0 {
		tolerance = DefaultFractionTolerance
	}

	whole := float64(int(value))
	frac := value - whole

	// Check if already close to whole number
	if frac < tolerance {
		return whole
	}
	if frac > 1.0-tolerance {
		return whole + 1
	}

	// Find the nearest common fraction
	bestMatch := frac
	bestDiff := tolerance + 1 // Start with a value that won't match

	for _, f := range commonFractions {
		diff := math.Abs(frac - f.value)
		if diff < bestDiff {
			bestDiff = diff
			bestMatch = f.value
		}
	}

	// Only use the match if it's within tolerance
	if bestDiff <= tolerance {
		return whole + bestMatch
	}

	return value // Return unchanged if no good match
}
