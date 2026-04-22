package shoppinglist

import (
	"testing"
)

func TestParseShoppingListSimple(t *testing.T) {
	input := `./Breakfast/Mexican Style Burrito{2}
./Salads/Boring{2}
olive oil{4%l}
salt
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(list.Items))
	}
	if list.Items[0].Type != "recipe" || list.Items[0].Path != "./Breakfast/Mexican Style Burrito" {
		t.Errorf("item 0: expected recipe './Breakfast/Mexican Style Burrito', got %q %q", list.Items[0].Type, list.Items[0].Path)
	}
	if list.Items[0].Quantity != "2" {
		t.Errorf("item 0: expected quantity '2', got %q", list.Items[0].Quantity)
	}
	if list.Items[2].Type != "ingredient" || list.Items[2].Name != "olive oil" {
		t.Errorf("item 2: expected ingredient 'olive oil', got %q %q", list.Items[2].Type, list.Items[2].Name)
	}
	if list.Items[2].Quantity != "4" || list.Items[2].Unit != "l" {
		t.Errorf("item 2: expected quantity '4' unit 'l', got %q %q", list.Items[2].Quantity, list.Items[2].Unit)
	}
	if list.Items[3].Type != "ingredient" || list.Items[3].Name != "salt" {
		t.Errorf("item 3: expected ingredient 'salt', got %q %q", list.Items[3].Type, list.Items[3].Name)
	}
}

func TestParseShoppingListWithMultiplier(t *testing.T) {
	input := `./bread{2}
./sauce{0.5}
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
	if list.Items[0].Quantity != "2" {
		t.Errorf("expected quantity '2', got %q", list.Items[0].Quantity)
	}
	if list.Items[1].Quantity != "0.5" {
		t.Errorf("expected quantity '0.5', got %q", list.Items[1].Quantity)
	}
}

func TestParseShoppingListNesting(t *testing.T) {
	input := `./Plans/3 Day Plan I
  ./Breakfast/Mexican Style Burrito{2}
  ./Salads/Boring{2}
  ./Slowcooker/Slow-cooker beef stew{0.5}
olive oil{4%l}
salt
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 top-level items, got %d", len(list.Items))
	}
	if list.Items[0].Type != "recipe" {
		t.Fatalf("expected first item to be recipe, got %q", list.Items[0].Type)
	}
	if len(list.Items[0].Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(list.Items[0].Children))
	}
	if list.Items[0].Children[0].Path != "./Breakfast/Mexican Style Burrito" {
		t.Errorf("child 0: expected path './Breakfast/Mexican Style Burrito', got %q", list.Items[0].Children[0].Path)
	}
	if list.Items[0].Children[2].Quantity != "0.5" {
		t.Errorf("child 2: expected quantity '0.5', got %q", list.Items[0].Children[2].Quantity)
	}
}

func TestParseShoppingListDeepNesting(t *testing.T) {
	input := `./Plans/Weekly
  ./Breakfast/Burrito{2}
    ./Components/Guacamole{2}
    ./Components/Beans{2}
  ./Dinner/Stew{1}
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 top-level item, got %d", len(list.Items))
	}
	plan := list.Items[0]
	if len(plan.Children) != 2 {
		t.Fatalf("expected 2 children of plan, got %d", len(plan.Children))
	}
	burrito := plan.Children[0]
	if len(burrito.Children) != 2 {
		t.Fatalf("expected 2 children of burrito, got %d", len(burrito.Children))
	}
	if burrito.Children[0].Path != "./Components/Guacamole" {
		t.Errorf("expected path './Components/Guacamole', got %q", burrito.Children[0].Path)
	}
}

func TestParseShoppingListComments(t *testing.T) {
	input := `./bread{1}
-- remember to check the pantry
salt
[- shopping note -]
pepper
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 items (comments excluded), got %d", len(list.Items))
	}
	if list.Items[0].Path != "./bread" {
		t.Errorf("expected './bread', got %q", list.Items[0].Path)
	}
	if list.Items[1].Name != "salt" {
		t.Errorf("expected 'salt', got %q", list.Items[1].Name)
	}
	if list.Items[2].Name != "pepper" {
		t.Errorf("expected 'pepper', got %q", list.Items[2].Name)
	}
}

func TestParseShoppingListEmpty(t *testing.T) {
	list, err := ParseShoppingList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(list.Items))
	}
}

func TestParseShoppingListFreehandWithQuantity(t *testing.T) {
	input := `olive oil{4%l}
flour{500%g}
sugar{1%kg}
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list.Items))
	}
	tests := []struct {
		name, qty, unit string
	}{
		{"olive oil", "4", "l"},
		{"flour", "500", "g"},
		{"sugar", "1", "kg"},
	}
	for i, tt := range tests {
		if list.Items[i].Name != tt.name {
			t.Errorf("item %d: expected name %q, got %q", i, tt.name, list.Items[i].Name)
		}
		if list.Items[i].Quantity != tt.qty {
			t.Errorf("item %d: expected qty %q, got %q", i, tt.qty, list.Items[i].Quantity)
		}
		if list.Items[i].Unit != tt.unit {
			t.Errorf("item %d: expected unit %q, got %q", i, tt.unit, list.Items[i].Unit)
		}
	}
}

func TestParseShoppingListFreehandNoQuantity(t *testing.T) {
	input := `salt
pepper
garlic
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list.Items))
	}
	for i, name := range []string{"salt", "pepper", "garlic"} {
		if list.Items[i].Name != name {
			t.Errorf("item %d: expected %q, got %q", i, name, list.Items[i].Name)
		}
		if list.Items[i].Quantity != "" {
			t.Errorf("item %d: expected empty quantity, got %q", i, list.Items[i].Quantity)
		}
	}
}

func TestParseCheckLogSimple(t *testing.T) {
	input := `+ salt
+ avocados
+ olive oil
`
	log, err := ParseCheckLog(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(log.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(log.Entries))
	}
	if !log.Entries[0].Checked || log.Entries[0].Name != "salt" {
		t.Errorf("entry 0: expected checked 'salt', got %v %q", log.Entries[0].Checked, log.Entries[0].Name)
	}
}

func TestParseCheckLogLastWins(t *testing.T) {
	input := `+ salt
+ avocados
+ olive oil
- avocados
+ avocados
`
	log, err := ParseCheckLog(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(log.Entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(log.Entries))
	}
	// avocados was unchecked then re-checked — should be checked
	if !log.IsChecked("avocados") {
		t.Error("expected avocados to be checked (last wins)")
	}
}

func TestParseCheckLogCaseInsensitive(t *testing.T) {
	input := `+ Salt
- salt
+ SALT
`
	log, err := ParseCheckLog(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Last entry is "+ SALT" so should be checked
	if !log.IsChecked("salt") {
		t.Error("expected 'salt' to be checked (case-insensitive, last wins)")
	}
	if !log.IsChecked("SALT") {
		t.Error("expected 'SALT' to be checked (case-insensitive)")
	}
	if !log.IsChecked("Salt") {
		t.Error("expected 'Salt' to be checked (case-insensitive)")
	}
}

func TestIsChecked(t *testing.T) {
	input := `+ salt
+ butter
- pepper
`
	log, err := ParseCheckLog(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !log.IsChecked("salt") {
		t.Error("expected salt to be checked")
	}
	if !log.IsChecked("butter") {
		t.Error("expected butter to be checked")
	}
	if log.IsChecked("pepper") {
		t.Error("expected pepper to be unchecked")
	}
	if log.IsChecked("sugar") {
		t.Error("expected sugar (not in log) to be unchecked")
	}
}

func TestCompact(t *testing.T) {
	listInput := `./bread{1}
salt
pepper
butter
`
	list, err := ParseShoppingList(listInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkInput := `+ salt
+ butter
+ stale_item
- pepper
+ pepper
- pepper
`
	log, err := ParseCheckLog(checkInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	compacted := log.Compact(list)

	// stale_item should be gone (not in shopping list)
	// pepper was checked then unchecked — final state is unchecked
	// salt and butter are checked
	checkedCount := 0
	for _, entry := range compacted.Entries {
		if !entry.Checked {
			t.Errorf("compacted entry %q should be checked", entry.Name)
		}
		checkedCount++
	}
	// salt and butter should remain
	if checkedCount != 2 {
		t.Errorf("expected 2 checked entries after compaction, got %d", checkedCount)
	}

	// Verify salt and butter are present
	names := make(map[string]bool)
	for _, entry := range compacted.Entries {
		names[entry.Name] = true
	}
	if !names["salt"] {
		t.Error("expected 'salt' in compacted log")
	}
	if !names["butter"] {
		t.Error("expected 'butter' in compacted log")
	}
	if names["stale_item"] {
		t.Error("expected 'stale_item' to be removed from compacted log")
	}
}

func TestParseShoppingListMixed(t *testing.T) {
	input := `./Plans/3 Day Plan I
  ./Breakfast/Mexican Style Burrito{2}
    ./Components/Guacamole{2}
    ./Components/Beans{2}
  ./Salads/Boring{2}
  ./Slowcooker/Slow-cooker beef stew{0.5}
olive oil{4%l}
salt
-- remember to check the pantry
`
	list, err := ParseShoppingList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 top-level items, got %d", len(list.Items))
	}

	// First item: plan with children
	plan := list.Items[0]
	if plan.Type != "recipe" || plan.Path != "./Plans/3 Day Plan I" {
		t.Errorf("expected recipe './Plans/3 Day Plan I', got %q %q", plan.Type, plan.Path)
	}
	if len(plan.Children) != 3 {
		t.Fatalf("expected 3 children of plan, got %d", len(plan.Children))
	}

	// Burrito with sub-components
	burrito := plan.Children[0]
	if len(burrito.Children) != 2 {
		t.Fatalf("expected 2 children of burrito, got %d", len(burrito.Children))
	}

	// Freehand ingredients
	if list.Items[1].Type != "ingredient" || list.Items[1].Name != "olive oil" {
		t.Errorf("expected ingredient 'olive oil', got %q %q", list.Items[1].Type, list.Items[1].Name)
	}
	if list.Items[2].Type != "ingredient" || list.Items[2].Name != "salt" {
		t.Errorf("expected ingredient 'salt', got %q %q", list.Items[2].Type, list.Items[2].Name)
	}
}
