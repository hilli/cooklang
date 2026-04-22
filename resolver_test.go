package cooklang

import (
	"path/filepath"
	"testing"
)

func TestParseYield(t *testing.T) {
	tests := []struct {
		input   string
		wantQty float64
		wantUnit string
	}{
		{"500%ml", 500, "ml"},
		{"2%loaves", 2, "loaves"},
		{"1.5%kg", 1.5, "kg"},
		{"100%g", 100, "g"},
		{"3", 3, ""},
		{"", 0, ""},
		{"invalid", 0, ""},
	}
	for _, tt := range tests {
		qty, unit := ParseYield(tt.input)
		if qty != tt.wantQty || unit != tt.wantUnit {
			t.Errorf("ParseYield(%q) = (%v, %q), want (%v, %q)", tt.input, qty, unit, tt.wantQty, tt.wantUnit)
		}
	}
}

func TestScaleByYield(t *testing.T) {
	recipe := &Recipe{
		Metadata: map[string]string{"yield": "500%ml"},
	}

	factor, err := ScaleByYield(recipe, 150, "ml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if factor != 0.3 {
		t.Errorf("expected factor 0.3, got %v", factor)
	}
}

func TestScaleByYieldDouble(t *testing.T) {
	recipe := &Recipe{
		Metadata: map[string]string{"yield": "500%ml"},
	}

	factor, err := ScaleByYield(recipe, 1000, "ml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if factor != 2.0 {
		t.Errorf("expected factor 2.0, got %v", factor)
	}
}

func TestScaleByYieldIncompatibleUnits(t *testing.T) {
	recipe := &Recipe{
		Metadata: map[string]string{"yield": "500%ml"},
	}

	_, err := ScaleByYield(recipe, 150, "g")
	if err == nil {
		t.Fatal("expected error for incompatible units")
	}
}

func TestScaleByYieldNoMetadata(t *testing.T) {
	recipe := &Recipe{
		Metadata: map[string]string{},
	}

	_, err := ScaleByYield(recipe, 150, "ml")
	if err == nil {
		t.Fatal("expected error for missing yield metadata")
	}
}

func TestScaleByYieldInvalidQuantity(t *testing.T) {
	recipe := &Recipe{
		Metadata: map[string]string{"yield": "invalid"},
	}

	_, err := ScaleByYield(recipe, 150, "ml")
	if err == nil {
		t.Fatal("expected error for invalid yield")
	}
}

func TestFileSystemResolverBasic(t *testing.T) {
	dir := t.TempDir()

	// Create a recipe file
	cookContent := `---
servings: 4
---

Mix @flour{500%g} with @water{300%ml}.
`
	if err := writeFile(filepath.Join(dir, "bread.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	recipe, err := resolver.Resolve("bread")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipe.Servings != 4 {
		t.Errorf("expected servings 4, got %v", recipe.Servings)
	}
}

func TestFileSystemResolverNestedPath(t *testing.T) {
	dir := t.TempDir()

	cookContent := `---
servings: 2
---

Whisk @eggs{3} with @milk{200%ml}.
`
	if err := writeFile(filepath.Join(dir, "sauces", "hollandaise.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	recipe, err := resolver.Resolve("sauces/hollandaise")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipe.Servings != 2 {
		t.Errorf("expected servings 2, got %v", recipe.Servings)
	}
}

func TestFileSystemResolverCaching(t *testing.T) {
	dir := t.TempDir()

	cookContent := `Mix @flour{500%g}.
`
	if err := writeFile(filepath.Join(dir, "bread.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)

	// Resolve twice — second should come from cache
	r1, err := resolver.Resolve("bread")
	if err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	r2, err := resolver.Resolve("bread")
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	if r1 != r2 {
		t.Error("expected same pointer from cache")
	}
}

func TestFileSystemResolverMissingFile(t *testing.T) {
	dir := t.TempDir()
	resolver := NewFileSystemResolver(dir)

	_, err := resolver.Resolve("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestResolveAndScaleFactor(t *testing.T) {
	dir := t.TempDir()

	cookContent := `---
servings: 2
---

Mix @flour{500%g} with @water{300%ml}.
`
	if err := writeFile(filepath.Join(dir, "bread.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	ref := &RecipeReference{
		Path:     "bread",
		Quantity: 2,
		Unit:     "",
	}

	scaled, err := ResolveAndScale(resolver, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scaled.Servings != 4 {
		t.Errorf("expected servings 4, got %v", scaled.Servings)
	}
}

func TestResolveAndScaleServings(t *testing.T) {
	dir := t.TempDir()

	cookContent := `---
servings: 2
---

Mix @flour{500%g} with @water{300%ml}.
`
	if err := writeFile(filepath.Join(dir, "pasta.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	ref := &RecipeReference{
		Path:     "pasta",
		Quantity: 4,
		Unit:     "servings",
	}

	scaled, err := ResolveAndScale(resolver, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scaled.Servings != 4 {
		t.Errorf("expected servings 4, got %v", scaled.Servings)
	}
}

func TestResolveAndScaleYield(t *testing.T) {
	dir := t.TempDir()

	cookContent := `---
yield: 500%ml
---

Whisk @butter{250%g} with @lemon juice{50%ml}.
`
	if err := writeFile(filepath.Join(dir, "sauces", "hollandaise.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	ref := &RecipeReference{
		Path:     "sauces/hollandaise",
		Quantity: 150,
		Unit:     "ml",
	}

	scaled, err := ResolveAndScale(resolver, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should scale by 150/500 = 0.3
	ingredients := scaled.GetIngredients()
	foundButter := false
	for _, ing := range ingredients.Ingredients {
		if ing.Name == "butter" {
			foundButter = true
			// 250 * 0.3 = 75
			if ing.Quantity != 75 {
				t.Errorf("expected butter quantity 75, got %v", ing.Quantity)
			}
		}
	}
	if !foundButter {
		t.Error("butter ingredient not found in scaled recipe")
	}
}

func TestResolveAndScaleYieldIncompatibleUnits(t *testing.T) {
	dir := t.TempDir()

	cookContent := `---
yield: 500%ml
---

Mix @sauce{500%ml}.
`
	if err := writeFile(filepath.Join(dir, "sauce.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	ref := &RecipeReference{
		Path:     "sauce",
		Quantity: 150,
		Unit:     "g", // wrong unit
	}

	_, err := ResolveAndScale(resolver, ref)
	if err == nil {
		t.Fatal("expected error for incompatible units")
	}
}

func TestResolveAndScaleNoQuantity(t *testing.T) {
	dir := t.TempDir()

	cookContent := `Mix @flour{500%g}.
`
	if err := writeFile(filepath.Join(dir, "bread.cook"), cookContent); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	resolver := NewFileSystemResolver(dir)
	ref := &RecipeReference{
		Path:     "bread",
		Quantity: 0,
	}

	// Should return unscaled recipe
	recipe, err := ResolveAndScale(resolver, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ingredients := recipe.GetIngredients()
	for _, ing := range ingredients.Ingredients {
		if ing.Name == "flour" && ing.Quantity != 500 {
			t.Errorf("expected flour quantity 500 (unscaled), got %v", ing.Quantity)
		}
	}
}
