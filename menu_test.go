package cooklang

import (
	"testing"
	"time"
)

func TestParseMenuSimple(t *testing.T) {
	input := `== Monday ==

@./breakfast/omelette{2%servings}
@./lunch/sandwich{1%servings}

== Tuesday ==

@./breakfast/pancakes{3%servings}
@./dinner/pasta{4%servings}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(menu.Days))
	}
	if menu.Days[0].Name != "Monday" {
		t.Errorf("expected day name 'Monday', got %q", menu.Days[0].Name)
	}
	if len(menu.Days[0].Recipes) != 2 {
		t.Errorf("expected 2 recipes on Monday, got %d", len(menu.Days[0].Recipes))
	}
	if menu.Days[0].Recipes[0].Path != "./breakfast/omelette" {
		t.Errorf("expected path './breakfast/omelette', got %q", menu.Days[0].Recipes[0].Path)
	}
	if menu.Days[0].Recipes[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %v", menu.Days[0].Recipes[0].Quantity)
	}
	if menu.Days[0].Recipes[0].Unit != "servings" {
		t.Errorf("expected unit 'servings', got %q", menu.Days[0].Recipes[0].Unit)
	}
	if menu.Days[1].Name != "Tuesday" {
		t.Errorf("expected day name 'Tuesday', got %q", menu.Days[1].Name)
	}
	if len(menu.Days[1].Recipes) != 2 {
		t.Errorf("expected 2 recipes on Tuesday, got %d", len(menu.Days[1].Recipes))
	}
}

func TestParseMenuWithDates(t *testing.T) {
	input := `== Day 1 (2026-03-07) ==

@./pasta{4%servings}

== Day 2 (2026-03-08) ==

@./salad{2%servings}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(menu.Days))
	}
	if menu.Days[0].Date == nil {
		t.Fatal("expected date for Day 1, got nil")
	}
	expectedDate := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)
	if !menu.Days[0].Date.Equal(expectedDate) {
		t.Errorf("expected date 2026-03-07, got %v", menu.Days[0].Date)
	}
	if menu.Days[0].Name != "Day 1 (2026-03-07)" {
		t.Errorf("expected name 'Day 1 (2026-03-07)', got %q", menu.Days[0].Name)
	}
	expectedDate2 := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	if menu.Days[1].Date == nil {
		t.Fatal("expected date for Day 2, got nil")
	}
	if !menu.Days[1].Date.Equal(expectedDate2) {
		t.Errorf("expected date 2026-03-08, got %v", menu.Days[1].Date)
	}
}

func TestParseMenuWithServings(t *testing.T) {
	input := `== Dinner ==

@./pasta{4%servings}
@./bread{2%loaves}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(menu.Days))
	}
	recipes := menu.Days[0].Recipes
	if len(recipes) != 2 {
		t.Fatalf("expected 2 recipes, got %d", len(recipes))
	}
	if recipes[0].Unit != "servings" {
		t.Errorf("expected unit 'servings', got %q", recipes[0].Unit)
	}
	if recipes[0].Quantity != 4 {
		t.Errorf("expected quantity 4, got %v", recipes[0].Quantity)
	}
	if recipes[1].Unit != "loaves" {
		t.Errorf("expected unit 'loaves', got %q", recipes[1].Unit)
	}
	if recipes[1].Quantity != 2 {
		t.Errorf("expected quantity 2, got %v", recipes[1].Quantity)
	}
}

func TestParseMenuWithScalingFactor(t *testing.T) {
	input := `== Lunch ==

@./bread{2}
@./soup{3}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	recipes := menu.Days[0].Recipes
	if len(recipes) != 2 {
		t.Fatalf("expected 2 recipes, got %d", len(recipes))
	}
	if recipes[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %v", recipes[0].Quantity)
	}
	if recipes[0].Unit != "" {
		t.Errorf("expected empty unit, got %q", recipes[0].Unit)
	}
	if recipes[1].Quantity != 3 {
		t.Errorf("expected quantity 3, got %v", recipes[1].Quantity)
	}
}

func TestParseMenuEmpty(t *testing.T) {
	input := ""
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 0 {
		t.Errorf("expected 0 days, got %d", len(menu.Days))
	}
}

func TestParseMenuNoSections(t *testing.T) {
	input := `@./pasta{4%servings}
@./salad{2%servings}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Recipes without a section should be placed in a default day
	if len(menu.Days) != 1 {
		t.Fatalf("expected 1 day (default), got %d", len(menu.Days))
	}
	if menu.Days[0].Name != "" {
		t.Errorf("expected empty name for default day, got %q", menu.Days[0].Name)
	}
	if len(menu.Days[0].Recipes) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(menu.Days[0].Recipes))
	}
}

func TestParseMenuMultipleDays(t *testing.T) {
	input := `== Monday ==
@./breakfast/eggs{2%servings}

== Tuesday ==
@./lunch/soup{4%servings}

== Wednesday ==
@./dinner/steak{2%servings}

== Thursday ==
@./breakfast/toast{1%servings}

== Friday ==
@./lunch/salad{3%servings}

== Saturday ==
@./dinner/pizza{8%servings}

== Sunday ==
@./brunch/waffles{4%servings}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 7 {
		t.Fatalf("expected 7 days, got %d", len(menu.Days))
	}
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for i, d := range days {
		if menu.Days[i].Name != d {
			t.Errorf("day %d: expected %q, got %q", i, d, menu.Days[i].Name)
		}
	}
}

func TestParseMenuRecipeNoQuantity(t *testing.T) {
	input := `== Lunch ==

@./salad{}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(menu.Days))
	}
	recipes := menu.Days[0].Recipes
	if len(recipes) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(recipes))
	}
	if recipes[0].Quantity != 0 {
		t.Errorf("expected quantity 0, got %v", recipes[0].Quantity)
	}
	if recipes[0].Unit != "" {
		t.Errorf("expected empty unit, got %q", recipes[0].Unit)
	}
	if recipes[0].Path != "./salad" {
		t.Errorf("expected path './salad', got %q", recipes[0].Path)
	}
}

func TestParseMenuWithComments(t *testing.T) {
	input := `-- This is a weekly menu plan
== Monday ==
-- Light breakfast
@./breakfast/oatmeal{2%servings}

== Tuesday ==
@./dinner/stew{4%servings}
-- Very filling
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(menu.Days))
	}
	if len(menu.Days[0].Recipes) != 1 {
		t.Errorf("expected 1 recipe on Monday, got %d", len(menu.Days[0].Recipes))
	}
	if len(menu.Days[1].Recipes) != 1 {
		t.Errorf("expected 1 recipe on Tuesday, got %d", len(menu.Days[1].Recipes))
	}
}

func TestParseMenuSingleDay(t *testing.T) {
	input := `== Special Dinner ==

@./appetizer/bruschetta{6%servings}
@./main/lasagna{8%servings}
@./dessert/tiramisu{8%servings}
`
	menu, err := ParseMenuString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menu.Days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(menu.Days))
	}
	if menu.Days[0].Name != "Special Dinner" {
		t.Errorf("expected name 'Special Dinner', got %q", menu.Days[0].Name)
	}
	if len(menu.Days[0].Recipes) != 3 {
		t.Errorf("expected 3 recipes, got %d", len(menu.Days[0].Recipes))
	}
}
