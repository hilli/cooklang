// Package pantry implements parsing for the Cooklang pantry configuration format.
//
// Pantry files use TOML format with sections for storage locations (freezer, fridge, pantry).
// Each item can be a simple quantity string ("500%g") or an object with attributes
// (bought, expire, quantity, low).
package pantry

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Pantry represents a parsed pantry configuration file.
type Pantry struct {
	Locations []Location `json:"locations"`
}

// Location represents a storage location (e.g., freezer, fridge, pantry).
type Location struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

// Item represents a pantry item with optional attributes.
type Item struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity,omitempty"` // Cooklang format e.g. "500%g"
	Bought   string `json:"bought,omitempty"`   // Date when purchased
	Expire   string `json:"expire,omitempty"`   // Expiration date
	Low      string `json:"low,omitempty"`       // Low stock threshold
}

// ParseFile reads and parses a pantry TOML file from disk.
func ParseFile(filename string) (*Pantry, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseString(string(content))
}

// ParseString parses a pantry configuration from a TOML string.
func ParseString(content string) (*Pantry, error) {
	// TOML structure: sections are locations, keys are item names
	// Values are either strings or objects
	var raw map[string]map[string]interface{}
	if _, err := toml.Decode(content, &raw); err != nil {
		return nil, err
	}

	p := &Pantry{}
	for locationName, items := range raw {
		loc := Location{Name: locationName}
		for itemName, value := range items {
			item := Item{Name: itemName}
			switch v := value.(type) {
			case string:
				item.Quantity = v
			case map[string]interface{}:
				if q, ok := v["quantity"].(string); ok {
					item.Quantity = q
				}
				if b, ok := v["bought"].(string); ok {
					item.Bought = b
				}
				if e, ok := v["expire"].(string); ok {
					item.Expire = e
				}
				if l, ok := v["low"].(string); ok {
					item.Low = l
				}
			}
			loc.Items = append(loc.Items, item)
		}
		p.Locations = append(p.Locations, loc)
	}

	return p, nil
}
