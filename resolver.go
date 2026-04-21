package cooklang

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RecipeResolver resolves recipe references to parsed Recipe objects.
type RecipeResolver interface {
	// Resolve takes a relative recipe path (e.g., "./sauces/Hollandaise") and returns
	// the parsed Recipe. The path is relative to the recipe root, without the .cook extension.
	Resolve(path string) (*Recipe, error)
}

// FileSystemResolver resolves recipe references by reading .cook files from disk.
type FileSystemResolver struct {
	BasePath string
	cache    map[string]*Recipe
	stack    map[string]bool // cycle detection
}

// NewFileSystemResolver creates a new resolver rooted at the given base directory.
func NewFileSystemResolver(basePath string) *FileSystemResolver {
	return &FileSystemResolver{
		BasePath: basePath,
		cache:    make(map[string]*Recipe),
		stack:    make(map[string]bool),
	}
}

// Resolve reads and parses a .cook file relative to the base path.
// It caches results and detects cycles.
func (r *FileSystemResolver) Resolve(path string) (*Recipe, error) {
	// Normalize path
	cleanPath := filepath.Clean(path)

	// Check cache
	if recipe, ok := r.cache[cleanPath]; ok {
		return recipe, nil
	}

	// Cycle detection
	if r.stack[cleanPath] {
		return nil, fmt.Errorf("cycle detected: recipe %q references itself", cleanPath)
	}
	r.stack[cleanPath] = true
	defer func() { delete(r.stack, cleanPath) }()

	// Build file path
	filePath := filepath.Join(r.BasePath, cleanPath+".cook")
	recipe, err := ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve recipe %q: %w", path, err)
	}

	r.cache[cleanPath] = recipe
	return recipe, nil
}

// ParseYield parses a yield metadata value in Cooklang format (e.g., "500%ml", "2%loaves").
// Returns quantity and unit, or 0 and empty string if parsing fails.
func ParseYield(yieldStr string) (float64, string) {
	if yieldStr == "" {
		return 0, ""
	}

	// Handle "quantity%unit" format
	if idx := strings.Index(yieldStr, "%"); idx >= 0 {
		qtyStr := strings.TrimSpace(yieldStr[:idx])
		unit := strings.TrimSpace(yieldStr[idx+1:])
		qty, err := strconv.ParseFloat(qtyStr, 64)
		if err != nil {
			return 0, ""
		}
		return qty, unit
	}

	// Try plain number
	qty, err := strconv.ParseFloat(strings.TrimSpace(yieldStr), 64)
	if err != nil {
		return 0, ""
	}
	return qty, ""
}

// ScaleByYield calculates a scaling factor for a recipe reference based on yield metadata.
// If the recipe has yield metadata matching the requested unit, it calculates
// targetQuantity / yieldQuantity as the scaling factor.
//
// This is an experimental feature per the Cooklang spec.
func ScaleByYield(recipe *Recipe, targetQuantity float64, targetUnit string) (float64, error) {
	yieldStr, ok := recipe.Metadata["yield"]
	if !ok {
		return 0, fmt.Errorf("recipe has no yield metadata")
	}

	yieldQty, yieldUnit := ParseYield(yieldStr)
	if yieldQty <= 0 {
		return 0, fmt.Errorf("invalid yield quantity in %q", yieldStr)
	}

	if !strings.EqualFold(yieldUnit, targetUnit) {
		return 0, fmt.Errorf("incompatible units: recipe yields %q but %q requested", yieldUnit, targetUnit)
	}

	return targetQuantity / yieldQty, nil
}

// ResolveAndScale resolves a recipe reference and scales it according to the reference's
// quantity and unit. It supports three scaling modes:
//  1. No unit — scales by the given factor
//  2. "servings" unit — scales to target servings
//  3. Other units — uses yield-based scaling (experimental)
func ResolveAndScale(resolver RecipeResolver, ref *RecipeReference) (*Recipe, error) {
	recipe, err := resolver.Resolve(ref.Path)
	if err != nil {
		return nil, err
	}

	if ref.Quantity <= 0 {
		return recipe, nil // No scaling
	}

	switch {
	case ref.Unit == "":
		// Factor-based scaling
		return recipe.Scale(float64(ref.Quantity)), nil

	case strings.EqualFold(ref.Unit, "servings"):
		// Servings-based scaling
		return recipe.ScaleToServings(float64(ref.Quantity)), nil

	default:
		// Units-based scaling (experimental)
		factor, err := ScaleByYield(recipe, float64(ref.Quantity), ref.Unit)
		if err != nil {
			return nil, fmt.Errorf("cannot scale %q: %w", ref.Path, err)
		}
		return recipe.Scale(factor), nil
	}
}

// writeFile is a helper for tests — not exported
func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
