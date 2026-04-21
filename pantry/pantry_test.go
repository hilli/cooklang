package pantry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePantrySimple(t *testing.T) {
	input := `[freezer]
cranberries = "500%g"
ice_cream = "2%L"

[pantry]
rice = "5%kg"
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Locations) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(p.Locations))
	}
	// Find freezer
	freezer := findLocation(t, p, "freezer")
	if len(freezer.Items) != 2 {
		t.Fatalf("expected 2 freezer items, got %d", len(freezer.Items))
	}
	cranberries := findItem(t, freezer, "cranberries")
	if cranberries.Quantity != "500%g" {
		t.Errorf("expected quantity '500%%g', got %q", cranberries.Quantity)
	}
}

func TestParsePantryObjects(t *testing.T) {
	input := `[freezer]
spinach = { bought = "05.05.2024", expire = "05.06.2025", quantity = "1%kg" }

[fridge]
milk = { expire = "10.05.2024", quantity = "1%L" }
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	freezer := findLocation(t, p, "freezer")
	spinach := findItem(t, freezer, "spinach")
	if spinach.Quantity != "1%kg" {
		t.Errorf("expected quantity '1%%kg', got %q", spinach.Quantity)
	}
	if spinach.Bought != "05.05.2024" {
		t.Errorf("expected bought '05.05.2024', got %q", spinach.Bought)
	}
	if spinach.Expire != "05.06.2025" {
		t.Errorf("expected expire '05.06.2025', got %q", spinach.Expire)
	}

	fridge := findLocation(t, p, "fridge")
	milk := findItem(t, fridge, "milk")
	if milk.Expire != "10.05.2024" {
		t.Errorf("expected expire '10.05.2024', got %q", milk.Expire)
	}
}

func TestParsePantryMixed(t *testing.T) {
	input := `[freezer]
ice_cream = "2%L"
frozen_peas = { bought = "01.01.2024", quantity = "500%g", low = "200%g" }

[fridge]
cheese = { expire = "15.05.2024" }
yogurt = { bought = "05.05.2024", expire = "12.05.2024", quantity = "500%ml" }

[pantry]
flour = "5%kg"
pasta = { quantity = "1%kg", low = "200%g" }
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Locations) != 3 {
		t.Fatalf("expected 3 locations, got %d", len(p.Locations))
	}

	freezer := findLocation(t, p, "freezer")
	peas := findItem(t, freezer, "frozen_peas")
	if peas.Low != "200%g" {
		t.Errorf("expected low '200%%g', got %q", peas.Low)
	}

	pantryLoc := findLocation(t, p, "pantry")
	pasta := findItem(t, pantryLoc, "pasta")
	if pasta.Quantity != "1%kg" {
		t.Errorf("expected quantity '1%%kg', got %q", pasta.Quantity)
	}
	if pasta.Low != "200%g" {
		t.Errorf("expected low '200%%g', got %q", pasta.Low)
	}
}

func TestParsePantryMultipleLocations(t *testing.T) {
	input := `[freezer]
chicken = "2%kg"

[fridge]
butter = "250%g"

[pantry]
salt = "1%kg"

[cellar]
wine = "6%bottles"
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Locations) != 4 {
		t.Fatalf("expected 4 locations, got %d", len(p.Locations))
	}
	names := make(map[string]bool)
	for _, loc := range p.Locations {
		names[loc.Name] = true
	}
	for _, name := range []string{"freezer", "fridge", "pantry", "cellar"} {
		if !names[name] {
			t.Errorf("missing location %q", name)
		}
	}
}

func TestParsePantryEmpty(t *testing.T) {
	p, err := ParseString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Locations) != 0 {
		t.Errorf("expected 0 locations, got %d", len(p.Locations))
	}
}

func TestParsePantryLowThreshold(t *testing.T) {
	input := `[pantry]
flour = { quantity = "5%kg", low = "500%g" }
sugar = { quantity = "2%kg", low = "200%g" }
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pantryLoc := findLocation(t, p, "pantry")
	flour := findItem(t, pantryLoc, "flour")
	if flour.Low != "500%g" {
		t.Errorf("expected low '500%%g', got %q", flour.Low)
	}
	sugar := findItem(t, pantryLoc, "sugar")
	if sugar.Low != "200%g" {
		t.Errorf("expected low '200%%g', got %q", sugar.Low)
	}
}

func TestParsePantryFromFile(t *testing.T) {
	content := `[fridge]
milk = "1%L"
butter = "250%g"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "pantry.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	p, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fridge := findLocation(t, p, "fridge")
	if len(fridge.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(fridge.Items))
	}
}

func TestParsePantryQuantityFormats(t *testing.T) {
	input := `[pantry]
item1 = "500%g"
item2 = "1%kg"
item3 = "2%L"
item4 = "250%ml"
item5 = "3%lbs"
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pantryLoc := findLocation(t, p, "pantry")
	expected := map[string]string{
		"item1": "500%g",
		"item2": "1%kg",
		"item3": "2%L",
		"item4": "250%ml",
		"item5": "3%lbs",
	}
	for name, qty := range expected {
		item := findItem(t, pantryLoc, name)
		if item.Quantity != qty {
			t.Errorf("item %q: expected quantity %q, got %q", name, qty, item.Quantity)
		}
	}
}

func TestParsePantryDateFormats(t *testing.T) {
	input := `[fridge]
milk = { bought = "01.01.2024", expire = "15.01.2024", quantity = "1%L" }
cheese = { bought = "2024-01-05", expire = "2024-02-05", quantity = "500%g" }
`
	p, err := ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fridge := findLocation(t, p, "fridge")
	milk := findItem(t, fridge, "milk")
	if milk.Bought != "01.01.2024" {
		t.Errorf("expected bought '01.01.2024', got %q", milk.Bought)
	}
	cheese := findItem(t, fridge, "cheese")
	if cheese.Bought != "2024-01-05" {
		t.Errorf("expected bought '2024-01-05', got %q", cheese.Bought)
	}
}

// Helper functions

func findLocation(t *testing.T, p *Pantry, name string) *Location {
	t.Helper()
	for i := range p.Locations {
		if p.Locations[i].Name == name {
			return &p.Locations[i]
		}
	}
	t.Fatalf("location %q not found", name)
	return nil
}

func findItem(t *testing.T, loc *Location, name string) *Item {
	t.Helper()
	for i := range loc.Items {
		if loc.Items[i].Name == name {
			return &loc.Items[i]
		}
	}
	t.Fatalf("item %q not found in location %q", name, loc.Name)
	return nil
}
