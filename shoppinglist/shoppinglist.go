// Package shoppinglist implements parsing for the Cooklang shopping list file format.
//
// The format uses two hidden files per directory:
//   - .shopping-list: list definition with recipe references and freehand ingredients
//   - .shopping-checked: append-only log of checked/unchecked ingredients
package shoppinglist

import (
	"bufio"
	"os"
	"strings"
)

// ShoppingList represents a parsed .shopping-list file.
type ShoppingList struct {
	Items []ListItem `json:"items"`
}

// ListItem represents an item in a shopping list — either a recipe reference or a freehand ingredient.
type ListItem struct {
	Type     string     `json:"type"`               // "recipe" or "ingredient"
	Path     string     `json:"path,omitempty"`      // Recipe path (for recipe items)
	Name     string     `json:"name,omitempty"`      // Ingredient name (for ingredient items)
	Quantity string     `json:"quantity,omitempty"`   // Optional quantity string
	Unit     string     `json:"unit,omitempty"`       // Optional unit
	Children []ListItem `json:"children,omitempty"`   // Nested items (only for recipe refs)
}

// CheckLog represents a parsed .shopping-checked file.
type CheckLog struct {
	Entries []CheckEntry `json:"entries"`
}

// CheckEntry represents a single check/uncheck entry.
type CheckEntry struct {
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

// ParseShoppingList parses a .shopping-list file from a string.
func ParseShoppingList(content string) (*ShoppingList, error) {
	list := &ShoppingList{}
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Stack of parent items at each indentation level
	type stackEntry struct {
		item  *ListItem
		depth int
	}
	var stack []stackEntry
	inBlockComment := false

	for scanner.Scan() {
		line := scanner.Text()

		// Handle block comments
		if inBlockComment {
			if idx := strings.Index(line, "-]"); idx >= 0 {
				inBlockComment = false
				line = line[idx+2:]
				if strings.TrimSpace(line) == "" {
					continue
				}
			} else {
				continue
			}
		}

		// Check for block comment start
		if idx := strings.Index(line, "[-"); idx >= 0 {
			endIdx := strings.Index(line[idx:], "-]")
			if endIdx >= 0 {
				// Block comment starts and ends on same line — remove it
				line = line[:idx] + line[idx+endIdx+2:]
			} else {
				line = line[:idx]
				inBlockComment = true
			}
		}

		// Remove inline comments
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Calculate indentation depth (2 spaces per level)
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else {
				break
			}
		}
		depth := indent / 2

		item := parseLine(trimmed)

		// Pop stack entries that are at same or deeper level
		for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
			stack = stack[:len(stack)-1]
		}

		if len(stack) > 0 {
			// Add as child of parent
			parent := stack[len(stack)-1].item
			parent.Children = append(parent.Children, item)
			childPtr := &parent.Children[len(parent.Children)-1]
			if item.Type == "recipe" {
				stack = append(stack, stackEntry{item: childPtr, depth: depth})
			}
		} else {
			// Top-level item
			list.Items = append(list.Items, item)
			itemPtr := &list.Items[len(list.Items)-1]
			if item.Type == "recipe" {
				stack = append(stack, stackEntry{item: itemPtr, depth: depth})
			}
		}
	}

	return list, scanner.Err()
}

func parseLine(line string) ListItem {
	// Recipe reference: starts with ./
	if strings.HasPrefix(line, "./") || strings.HasPrefix(line, "../") {
		return parseRecipeLine(line)
	}
	// Freehand ingredient
	return parseIngredientLine(line)
}

func parseRecipeLine(line string) ListItem {
	item := ListItem{Type: "recipe"}

	// Check for quantity in braces
	if braceIdx := strings.Index(line, "{"); braceIdx >= 0 {
		item.Path = strings.TrimSpace(line[:braceIdx])
		endBrace := strings.Index(line[braceIdx:], "}")
		if endBrace >= 0 {
			qtyStr := line[braceIdx+1 : braceIdx+endBrace]
			if pctIdx := strings.Index(qtyStr, "%"); pctIdx >= 0 {
				item.Quantity = qtyStr[:pctIdx]
				item.Unit = qtyStr[pctIdx+1:]
			} else if qtyStr != "" {
				item.Quantity = qtyStr
			}
		}
	} else {
		item.Path = strings.TrimSpace(line)
	}

	return item
}

func parseIngredientLine(line string) ListItem {
	item := ListItem{Type: "ingredient"}

	if braceIdx := strings.Index(line, "{"); braceIdx >= 0 {
		item.Name = strings.TrimSpace(line[:braceIdx])
		endBrace := strings.Index(line[braceIdx:], "}")
		if endBrace >= 0 {
			qtyStr := line[braceIdx+1 : braceIdx+endBrace]
			if pctIdx := strings.Index(qtyStr, "%"); pctIdx >= 0 {
				item.Quantity = qtyStr[:pctIdx]
				item.Unit = qtyStr[pctIdx+1:]
			} else if qtyStr != "" {
				item.Quantity = qtyStr
			}
		}
	} else {
		item.Name = strings.TrimSpace(line)
	}

	return item
}

// ParseCheckLog parses a .shopping-checked file from a string.
func ParseCheckLog(content string) (*CheckLog, error) {
	log := &CheckLog{}
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "+ ") {
			log.Entries = append(log.Entries, CheckEntry{
				Name:    strings.TrimSpace(line[2:]),
				Checked: true,
			})
		} else if strings.HasPrefix(line, "- ") {
			log.Entries = append(log.Entries, CheckEntry{
				Name:    strings.TrimSpace(line[2:]),
				Checked: false,
			})
		}
	}

	return log, scanner.Err()
}

// IsChecked returns whether the given ingredient is currently checked.
// Matching is case-insensitive. The last entry for a given name wins.
func (cl *CheckLog) IsChecked(name string) bool {
	nameLower := strings.ToLower(name)
	checked := false
	for _, entry := range cl.Entries {
		if strings.ToLower(entry.Name) == nameLower {
			checked = entry.Checked
		}
	}
	return checked
}

// Compact replays the check log, removes stale entries not in the shopping list,
// and returns a new CheckLog with only `+ name` lines for currently checked ingredients.
func (cl *CheckLog) Compact(list *ShoppingList) *CheckLog {
	// Collect all valid ingredient names from the list
	validNames := make(map[string]bool)
	var collectNames func(items []ListItem)
	collectNames = func(items []ListItem) {
		for _, item := range items {
			if item.Type == "ingredient" {
				validNames[strings.ToLower(item.Name)] = true
			}
			collectNames(item.Children)
		}
	}
	collectNames(list.Items)

	// Replay log to find final state
	state := make(map[string]bool)
	for _, entry := range cl.Entries {
		lower := strings.ToLower(entry.Name)
		if validNames[lower] {
			state[lower] = entry.Checked
		}
	}

	// Build compacted log with only checked items
	compacted := &CheckLog{}
	for name, checked := range state {
		if checked {
			compacted.Entries = append(compacted.Entries, CheckEntry{
				Name:    name,
				Checked: true,
			})
		}
	}

	return compacted
}

// ParseShoppingListFile reads and parses a .shopping-list file from disk.
func ParseShoppingListFile(filename string) (*ShoppingList, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseShoppingList(string(content))
}

// ParseCheckLogFile reads and parses a .shopping-checked file from disk.
func ParseCheckLogFile(filename string) (*CheckLog, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseCheckLog(string(content))
}
