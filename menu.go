package cooklang

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/hilli/cooklang/parser"
)

// Menu represents a parsed .menu file containing a meal plan organized by days.
type Menu struct {
	Days []MenuDay `json:"days"`
}

// MenuDay represents a single day or section in a menu, containing recipe references.
type MenuDay struct {
	Name    string       `json:"name"`
	Date    *time.Time   `json:"date,omitempty"`
	Recipes []MenuRecipe `json:"recipes"`
}

// MenuRecipe represents a recipe reference within a menu day.
type MenuRecipe struct {
	Path     string  `json:"path"`
	Quantity float32 `json:"quantity,omitempty"`
	Unit     string  `json:"unit,omitempty"`
}

var dateRegexp = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)

// ParseMenuString parses a menu from a string. A .menu file is a valid Cooklang file
// using sections for days and recipe references for dishes.
func ParseMenuString(s string) (*Menu, error) {
	p := parser.New()
	p.ExtendedMode = true
	recipe, err := p.ParseString(s)
	if err != nil {
		return nil, err
	}

	menu := &Menu{}
	var currentDay *MenuDay

	for _, step := range recipe.Steps {
		for _, comp := range step.Components {
			switch comp.Type {
			case "section":
				// Start a new day
				day := MenuDay{
					Name:    comp.Name,
					Recipes: []MenuRecipe{},
				}
				// Try to extract date from section name
				if match := dateRegexp.FindString(comp.Name); match != "" {
					if t, err := time.Parse("2006-01-02", match); err == nil {
						day.Date = &t
					}
				}
				menu.Days = append(menu.Days, day)
				currentDay = &menu.Days[len(menu.Days)-1]

			case "recipeReference":
				ref := MenuRecipe{
					Path: comp.Name,
					Unit: comp.Unit,
				}
				// Parse quantity
				if comp.Quantity != "" && comp.Quantity != "some" {
					if q, err := strconv.ParseFloat(comp.Quantity, 32); err == nil {
						ref.Quantity = float32(q)
					}
				}

				if currentDay == nil {
					// No section yet — create a default unnamed day
					menu.Days = append(menu.Days, MenuDay{
						Name:    "",
						Recipes: []MenuRecipe{},
					})
					currentDay = &menu.Days[len(menu.Days)-1]
				}
				currentDay.Recipes = append(currentDay.Recipes, ref)
			}
		}
	}

	return menu, nil
}

// ParseMenuFile reads and parses a .menu file from disk.
func ParseMenuFile(filename string) (*Menu, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseMenuString(string(content))
}
