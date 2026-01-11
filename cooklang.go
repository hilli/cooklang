package cooklang

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bcicen/go-units"
	"github.com/hilli/cooklang/parser"
)

// Recipe represents a parsed Cooklang recipe with its metadata and step-by-step instructions.
// The Recipe struct provides access to all recipe information including ingredients, cookware,
// timers, and cooking instructions organized as a linked list of steps.
//
// Recipes can be created by parsing Cooklang files using ParseFile, ParseString, or ParseBytes.
//
// Example:
//
//	recipe, err := cooklang.ParseFile("lasagna.cook")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(recipe.Title)
//	ingredients := recipe.GetIngredients()
type Recipe struct {
	Title       string    `json:"title,omitempty"`       // Recipe title from frontmatter
	Cuisine     string    `json:"cuisine,omitempty"`     // Cuisine type (e.g., "Italian", "Mexican")
	Date        time.Time `json:"date,omitempty"`        // Recipe date in YYYY-MM-DD format
	Description string    `json:"description,omitempty"` // Brief recipe description
	Difficulty  string    `json:"difficulty,omitempty"`  // Difficulty level (e.g., "easy", "medium", "hard")
	PrepTime    string    `json:"prep_time,omitempty"`   // Preparation time (e.g., "15 minutes")
	TotalTime   string    `json:"total_time,omitempty"`  // Total cooking time
	Metadata    Metadata  `json:"metadata,omitempty"`    // Additional custom metadata fields
	Author      string    `json:"author,omitempty"`      // Recipe author name
	Images      []string  `json:"images,omitempty"`      // Image filenames associated with the recipe
	Servings    float32   `json:"servings,omitempty"`    // Number of servings this recipe makes
	Tags        []string  `json:"tags,omitempty"`        // Recipe tags for categorization
	FirstStep   *Step     `json:"first_step,omitempty"`  // First step in the linked list of recipe steps
	CooklangRenderable
}

// CooklangRenderable provides rendering capabilities for recipe components.
// It allows custom rendering functions to be attached to recipes and their components.
type CooklangRenderable struct {
	RenderFunc func() string `json:"-"` // Custom rendering function
}

// Metadata stores arbitrary key-value pairs for recipe metadata not covered by structured fields.
// This allows recipes to include custom fields beyond the standard ones.
//
// Example:
//
//	metadata := Metadata{
//	    "source": "Grandma's cookbook",
//	    "category": "dessert",
//	}
type Metadata map[string]string

// UnitSystem defines supported unit systems for easy conversion
type UnitSystem string

const (
	UnitSystemMetric   UnitSystem = "metric"
	UnitSystemImperial UnitSystem = "imperial"
	UnitSystemUS       UnitSystem = "us"
)

// canonicalUnits maps quantity types to preferred units for each system
var canonicalUnits = map[UnitSystem]map[string]string{
	UnitSystemMetric: {
		"mass":        "g",
		"volume":      "ml",
		"length":      "cm",
		"temperature": "c",
	},
	UnitSystemImperial: {
		"mass":        "oz",
		"volume":      "ml", // Will be converted to appropriate units via mappings
		"length":      "in",
		"temperature": "f",
	},
	UnitSystemUS: {
		"mass":        "oz",
		"volume":      "ml", // Will be converted to appropriate units via mappings
		"length":      "in",
		"temperature": "f",
	},
}

// commonUnitMappings provides alternative units for better recipe display
var commonUnitMappings = map[UnitSystem]map[string]map[string]string{
	UnitSystemMetric: {
		"volume": {
			"large": "l",  // for volumes >= 1000ml
			"small": "ml", // for volumes < 1000ml
		},
		"mass": {
			"large": "kg", // for mass >= 1000g
			"small": "g",  // for mass < 1000g
		},
	},
	UnitSystemUS: {
		"volume": {
			"large":  "qt",   // for large volumes
			"medium": "cup",  // for medium volumes
			"small":  "tbsp", // for small volumes
			"tiny":   "tsp",  // for very small volumes
		},
	},
	UnitSystemImperial: {
		"volume": {
			"large":  "pt",    // for large volumes
			"medium": "fl_oz", // for medium volumes
			"small":  "tbsp",  // for small volumes
			"tiny":   "tsp",   // for very small volumes
		},
	},
}

// cookingUnitConversions provides conversions for common cooking units to ml (for volume)
// This supplements the go-units library which doesn't have all cooking units
var cookingUnitConversions = map[string]float64{
	// Volume conversions to ml
	"cup":    236.588,
	"tbsp":   14.7868,
	"tsp":    4.92892,
	"qt":     946.353,
	"pt":     473.176,
	"fl_oz":  29.5735,
	"gallon": 3785.41,

	// Keep metric units
	"ml": 1.0,
	"l":  1000.0,
}

// cookingMassConversions provides conversions for mass units to grams
var cookingMassConversions = map[string]float64{
	// Mass conversions to grams
	"oz": 28.3495,
	"lb": 453.592,
	"kg": 1000.0,
	"g":  1.0,
}

// convertCookingUnit converts between common cooking units using ml or grams as an intermediate.
// Supports volume units (ml, cup, tbsp, etc.) and mass units (g, kg, oz, lb).
func convertCookingUnit(value float64, fromUnit, toUnit string) (float64, error) {
	// Try volume conversion first
	if fromMl, okFrom := cookingUnitConversions[fromUnit]; okFrom {
		if toMl, okTo := cookingUnitConversions[toUnit]; okTo {
			// Convert from -> ml -> to
			mlValue := value * fromMl
			return mlValue / toMl, nil
		}
	}

	// Try mass conversion
	if fromG, okFrom := cookingMassConversions[fromUnit]; okFrom {
		if toG, okTo := cookingMassConversions[toUnit]; okTo {
			// Convert from -> g -> to
			gValue := value * fromG
			return gValue / toG, nil
		}
	}

	return 0, fmt.Errorf("cannot convert from %s to %s", fromUnit, toUnit)
}

// isCookingUnit checks if a unit is a recognized cooking unit that can be converted.
func isCookingUnit(unit string) bool {
	_, isVolume := cookingUnitConversions[unit]
	_, isMass := cookingMassConversions[unit]
	return isVolume || isMass
}

// getCookingUnitType returns "volume" or "mass" for recognized cooking units, empty string otherwise.
func getCookingUnitType(unit string) string {
	if _, isVolume := cookingUnitConversions[unit]; isVolume {
		return "volume"
	}
	if _, isMass := cookingMassConversions[unit]; isMass {
		return "mass"
	}
	return ""
}

// StepComponent represents a component within a recipe step (ingredient, instruction, timer, or cookware).
// Components are organized as a linked list within each step, allowing iteration through the sequence of actions.
type StepComponent interface {
	isStepComponent()       // Marker method
	Render() string         // Renders the component as Cooklang syntax
	SetNext(StepComponent)  // Sets the next component in the linked list
	GetNext() StepComponent // Gets the next component in the linked list
}

// Step represents a single step in a recipe's instructions.
// Each step contains a linked list of components (ingredients, cookware, timers, text instructions)
// and a link to the next step.
//
// Steps are traversed by following the NextStep pointer to iterate through the recipe's instructions.
type Step struct {
	FirstComponent StepComponent `json:"first_component,omitempty"` // First component in this step
	NextStep       *Step         `json:"next_step,omitempty"`       // Next step in the recipe
	CooklangRenderable
}

func (Instruction) isStepComponent() {}
func (Timer) isStepComponent()       {}
func (Cookware) isStepComponent()    {}
func (Ingredient) isStepComponent()  {}
func (Section) isStepComponent()     {}
func (Comment) isStepComponent()     {}
func (Note) isStepComponent()        {}

// Render returns the Cooklang syntax representation of this ingredient.
// Examples: "@flour{500%g}", "@salt{}", "@milk{2%cups}(cold)", "@yeast{=1%packet}"
func (i Ingredient) Render() string {
	var result string
	fixedPrefix := ""
	if i.Fixed {
		fixedPrefix = "="
	}
	if i.Quantity > 0 {
		result = fmt.Sprintf("@%s{%s%g%%%s}", i.Name, fixedPrefix, i.Quantity, i.Unit)
	} else if i.Quantity == -1 {
		// -1 indicates "some" quantity
		result = fmt.Sprintf("@%s{}", i.Name)
	} else {
		result = fmt.Sprintf("@%s{}", i.Name)
	}
	if i.Annotation != "" {
		result += fmt.Sprintf("(%s)", i.Annotation)
	}
	return result
}

// RenderDisplay returns ingredient in plain text format suitable for display.
// Examples: "2 cups flour", "500 g flour", "salt"
// Uses bartender-friendly fraction formatting (e.g., "1/2 oz" instead of "0.5 oz")
// When quantity is unspecified (e.g., @salt{}), returns just the ingredient name.
func (i Ingredient) RenderDisplay() string {
	var result string
	if i.Quantity > 0 && i.Unit != "" {
		qtyStr := FormatAsFractionDefault(float64(i.Quantity))
		result = fmt.Sprintf("%s %s %s", qtyStr, i.Unit, i.Name)
	} else if i.Quantity > 0 {
		qtyStr := FormatAsFractionDefault(float64(i.Quantity))
		result = fmt.Sprintf("%s %s", qtyStr, i.Name)
	} else {
		// Quantity == -1 (unspecified) or 0: just use the ingredient name
		result = i.Name
	}
	return result
}

// Render returns the plain text instruction.
func (inst Instruction) Render() string {
	return inst.Text
}

// RenderDisplay returns instruction text suitable for display (same as Render for Instruction).
func (inst Instruction) RenderDisplay() string {
	return inst.Text
}

// Render returns the Cooklang syntax representation of this timer.
// Examples: "~{10%minutes}", "~boil{15%min}"
func (t Timer) Render() string {
	var result string
	if t.Name != "" {
		result = fmt.Sprintf("~%s{%s}", t.Name, t.Duration)
	} else {
		result = fmt.Sprintf("~{%s}", t.Duration)
	}
	if t.Annotation != "" {
		result += fmt.Sprintf("(%s)", t.Annotation)
	}
	return result
}

// RenderDisplay returns timer in plain text format suitable for display.
// Returns the duration with unit if available, or just the duration.
// If the timer has a name but no duration, returns the name.
func (t Timer) RenderDisplay() string {
	// If there's a duration, show it (optionally with unit)
	if t.Duration != "" {
		if t.Unit != "" {
			return t.Duration + " " + t.Unit
		}
		return t.Duration
	}
	// Fall back to name if no duration
	if t.Name != "" {
		return t.Name
	}
	return ""
}

// Render returns the Cooklang syntax representation of this cookware.
// Examples: "#pot{}", "#bowl{2}", "#oven{}(preheated)"
func (c Cookware) Render() string {
	var result string
	if c.Quantity > 1 {
		result = fmt.Sprintf("#%s{%d}", c.Name, c.Quantity)
	} else {
		result = fmt.Sprintf("#%s{}", c.Name)
	}
	if c.Annotation != "" {
		result += fmt.Sprintf("(%s)", c.Annotation)
	}
	return result
}

// RenderDisplay returns cookware in plain text format suitable for display.
// Returns just the cookware name.
func (c Cookware) RenderDisplay() string {
	return c.Name
}

// Render returns the Cooklang syntax representation of this section.
// Examples: "== Section Name =="
func (s Section) Render() string {
	if s.Name != "" {
		return fmt.Sprintf("== %s ==", s.Name)
	}
	return "=="
}

// RenderDisplay returns section name suitable for display.
func (s Section) RenderDisplay() string {
	return s.Name
}

// Render returns the Cooklang syntax representation of this comment.
// Examples: "-- comment text" for line comments, "[- comment text -]" for block comments
func (cm Comment) Render() string {
	if cm.IsBlock {
		return fmt.Sprintf("[- %s -]", cm.Text)
	}
	return fmt.Sprintf("-- %s", cm.Text)
}

// RenderDisplay returns comment text suitable for display.
func (cm Comment) RenderDisplay() string {
	return cm.Text
}

// Render returns the Cooklang syntax representation of this note.
// Example: "> This is a note"
func (n Note) Render() string {
	return fmt.Sprintf("> %s", n.Text)
}

// RenderDisplay returns note text suitable for display.
func (n Note) RenderDisplay() string {
	return n.Text
}

// Ingredient represents a recipe ingredient with quantity, unit, and optional annotations.
// Ingredients support unit conversion and consolidation for shopping lists.
//
// Example Cooklang syntax: @flour{500%g}, @salt{}, @milk{2%cups}
//
// The Quantity field uses -1 to represent "some" (unspecified amount).
// The Fixed field indicates a quantity that should not scale with servings (e.g., @salt{=1%tsp}).
type Ingredient struct {
	Name           string        `json:"name,omitempty"`           // Ingredient name (e.g., "flour", "sugar")
	Quantity       float32       `json:"quantity,omitempty"`       // Amount (-1 means "some", 0 means none specified)
	Unit           string        `json:"unit,omitempty"`           // Unit of measurement (e.g., "g", "cup", "tbsp")
	Fixed          bool          `json:"fixed,omitempty"`          // Fixed quantity doesn't scale with servings
	TypedUnit      *units.Unit   `json:"typed_unit,omitempty"`     // Typed unit for conversion operations
	Subinstruction string        `json:"value,omitempty"`          // Additional preparation instructions
	Annotation     string        `json:"annotation,omitempty"`     // Optional annotation (e.g., "finely chopped")
	NextComponent  StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// NewIngredient creates a new Ingredient with proper unit typing for conversion operations.
// This constructor ensures that the TypedUnit field is properly initialized, which is required
// for unit conversion methods like ConvertTo and ConvertToSystem to work correctly.
//
// Parameters:
//   - name: The ingredient name (e.g., "vodka", "sugar")
//   - quantity: The amount (-1 means "some" unspecified amount)
//   - unit: The unit of measurement (e.g., "ml", "oz", "g", "cups")
//
// Example:
//
//	ing := cooklang.NewIngredient("vodka", 50, "ml")
//	converted := ing.ConvertToSystem(cooklang.UnitSystemUS)
//	fmt.Printf("%v %s\n", converted.Quantity, converted.Unit) // "1.69 oz"
func NewIngredient(name string, quantity float32, unit string) *Ingredient {
	return &Ingredient{
		Name:      name,
		Quantity:  quantity,
		Unit:      unit,
		TypedUnit: CreateTypedUnit(unit),
	}
}

// Instruction represents a text instruction within a recipe step.
// This is plain text that provides cooking directions.
type Instruction struct {
	Text          string        `json:"text,omitempty"`           // Instruction text
	NextComponent StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// Timer represents a duration timer in a recipe step.
// Timers specify how long to perform an action.
//
// Example Cooklang syntax: ~{10%minutes}, ~boil{15%min}
type Timer struct {
	Duration      string        `json:"duration,omitempty"`       // Duration value (e.g., "10")
	Name          string        `json:"name,omitempty"`           // Timer name/description (e.g., "boil", "rest")
	Text          string        `json:"text,omitempty"`           // Full timer text
	Unit          string        `json:"unit,omitempty"`           // Time unit (e.g., "minutes", "hours")
	Annotation    string        `json:"annotation,omitempty"`     // Optional annotation
	NextComponent StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// Cookware represents a cooking utensil or equipment needed for a recipe.
//
// Example Cooklang syntax: #pot{}, #bowl{2}, #oven{}
type Cookware struct {
	Name          string        `json:"name,omitempty"`           // Cookware name (e.g., "pot", "bowl", "oven")
	Quantity      int           `json:"quantity,omitempty"`       // Number of items needed (default 1)
	Annotation    string        `json:"annotation,omitempty"`     // Optional annotation (e.g., "large", "non-stick")
	NextComponent StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// Section represents a section header in a recipe.
// Sections divide complex recipes into logical parts (e.g., "Dough", "Filling").
//
// Example Cooklang syntax: = Dough, == Filling ==
type Section struct {
	Name          string        `json:"name,omitempty"`           // Section name (e.g., "Dough", "Filling")
	NextComponent StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// Comment represents a comment in a recipe.
// Comments are notes that don't affect the cooking instructions.
//
// Example Cooklang syntax:
// - Line comment: -- This is a comment
// - Block comment: [- This is a block comment -]
type Comment struct {
	Text          string        `json:"text,omitempty"`           // Comment text
	IsBlock       bool          `json:"is_block,omitempty"`       // True if this is a block comment [- -]
	NextComponent StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// Note represents a note block in a recipe.
// Notes are supplementary information that appears in recipe details but not during cooking mode.
// They are used for background stories, tips, or personal anecdotes related to the recipe.
//
// Example Cooklang syntax:
// > This dish is even better the next day, after the flavors have melded overnight.
// > This is a multi-line note
// > that continues here.
type Note struct {
	Text          string        `json:"text,omitempty"`           // Note text
	NextComponent StepComponent `json:"next_component,omitempty"` // Next component in the step
	CooklangRenderable
}

// ParseFile reads and parses a Cooklang recipe file, returning a Recipe object.
// It automatically detects and includes associated image files matching the recipe filename.
//
// Image detection looks for files with the same base name:
//   - Recipe.cook → Recipe.jpg, Recipe.png, Recipe.jpeg
//   - Recipe.cook → Recipe-1.jpg, Recipe-2.png, etc. (numbered variants)
//
// Parameters:
//   - filename: Path to the .cook file to parse
//
// Returns:
//   - *Recipe: The parsed recipe with all metadata, steps, and detected images
//   - error: Any error encountered during file reading or parsing
//
// Example:
//
//	recipe, err := cooklang.ParseFile("recipes/lasagna.cook")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Recipe: %s\n", recipe.Title)
//	fmt.Printf("Servings: %.0f\n", recipe.Servings)
func ParseFile(filename string) (*Recipe, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	p := parser.New()
	parsedRecipe, err := p.ParseBytes(content)
	if err != nil {
		return nil, err
	}
	recipe := ToCooklangRecipe(parsedRecipe)

	// Auto-detect and add images from filesystem
	detectedImages := findRecipeImages(filename)
	if len(detectedImages) > 0 {
		// Merge detected images with existing ones, avoiding duplicates
		recipe.Images = mergeUniqueStrings(recipe.Images, detectedImages)
		// Update metadata to reflect the merged images
		if len(recipe.Images) > 0 {
			recipe.Metadata["images"] = strings.Join(recipe.Images, ", ")
		}
	}

	return recipe, nil
}

// findRecipeImages looks for image files matching the recipe filename pattern.
// For a recipe file "Recipe.cook", it searches for:
// - Recipe.jpg, Recipe.jpeg, Recipe.png (base image)
// - Recipe-1.jpg, Recipe-2.jpg, etc. (numbered variants)
// Returns just the filenames (not full paths) of found images.
func findRecipeImages(cookFilePath string) []string {
	dir := filepath.Dir(cookFilePath)
	baseName := strings.TrimSuffix(filepath.Base(cookFilePath), ".cook")

	var images []string
	extensions := []string{".jpg", ".jpeg", ".png"}

	// Check for base image (e.g., Recipe.jpg)
	for _, ext := range extensions {
		imagePath := filepath.Join(dir, baseName+ext)
		if fileExists(imagePath) {
			images = append(images, baseName+ext)
		}
	}

	// Check for numbered images (e.g., Recipe-1.jpg, Recipe-2.jpg)
	// We'll check up to 99 numbered variants
	for i := 1; i <= 99; i++ {
		foundAny := false
		for _, ext := range extensions {
			numberedName := fmt.Sprintf("%s-%d%s", baseName, i, ext)
			imagePath := filepath.Join(dir, numberedName)
			if fileExists(imagePath) {
				images = append(images, numberedName)
				foundAny = true
			}
		}
		// If we didn't find any images for this number, stop searching
		if !foundAny {
			break
		}
	}

	return images
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// mergeUniqueStrings merges two string slices, removing duplicates and empty strings.
func mergeUniqueStrings(slice1, slice2 []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice1)+len(slice2))

	// Add all items from slice1
	for _, item := range slice1 {
		if item != "" && !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// Add items from slice2 that aren't already present
	for _, item := range slice2 {
		if item != "" && !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ParseBytes parses Cooklang recipe content from a byte slice.
// This is useful for parsing recipes from memory, HTTP responses, or other byte sources.
//
// Unlike ParseFile, this function does not perform image detection since no filename is available.
//
// Parameters:
//   - content: The raw Cooklang recipe content as bytes
//
// Returns:
//   - *Recipe: The parsed recipe with all metadata and steps
//   - error: Any error encountered during parsing
//
// Example:
//
//	content := []byte("---\ntitle: Quick Pasta\n---\n\nBoil @water{2%L} and add @pasta{100%g}.")
//	recipe, err := cooklang.ParseBytes(content)
func ParseBytes(content []byte) (*Recipe, error) {
	p := parser.New()
	parsedRecipe, err := p.ParseBytes(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(parsedRecipe), nil
}

// ParseString parses Cooklang recipe content from a string.
// This is a convenience wrapper around ParseBytes for string input.
//
// Parameters:
//   - content: The Cooklang recipe content as a string
//
// Returns:
//   - *Recipe: The parsed recipe with all metadata and steps
//   - error: Any error encountered during parsing
//
// Example:
//
//	content := "---\ntitle: Quick Pasta\n---\n\nBoil @water{2%L}."
//	recipe, err := cooklang.ParseString(content)
func ParseString(content string) (*Recipe, error) {
	p := parser.New()
	recipe, err := p.ParseString(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(recipe), nil
}

// CreateTypedUnit attempts to find a unit in go-units or creates a new one if not found.
// This function is used internally when parsing Cooklang content and can be used externally
// when programmatically creating Ingredient structs that need unit conversion support.
//
// If the unit string is empty, nil is returned.
// If the unit is found in the go-units library, a pointer to that unit is returned.
// Otherwise, a new unit is created with the given string as both name and symbol.
func CreateTypedUnit(unitStr string) *units.Unit {
	if unitStr == "" {
		return nil
	}

	// Try to find the unit first
	if foundUnit, err := units.Find(unitStr); err == nil {
		return &foundUnit
	}

	// If not found, create a new unit (returns by value, so we take address)
	newUnit := units.NewUnit(unitStr, unitStr)
	return &newUnit
}

// ToCooklangRecipe converts a parser.Recipe to a cooklang.Recipe.
// This is the internal function that transforms the parser's output into the high-level Recipe structure
// with all metadata fields populated and step components organized as linked lists.
//
// Most users should use ParseFile, ParseString, or ParseBytes instead of calling this directly.
func ToCooklangRecipe(pRecipe *parser.Recipe) *Recipe {
	recipe := &Recipe{}
	// Copy metadata to recipe fields
	recipe.Metadata = Metadata(pRecipe.Metadata)
	if title, ok := pRecipe.Metadata["title"]; ok {
		recipe.Title = title
	}
	if cuisine, ok := pRecipe.Metadata["cuisine"]; ok {
		recipe.Cuisine = cuisine
	}
	if description, ok := pRecipe.Metadata["description"]; ok {
		recipe.Description = description
	}
	if difficulty, ok := pRecipe.Metadata["difficulty"]; ok {
		recipe.Difficulty = difficulty
	}
	if prepTime, ok := pRecipe.Metadata["prep_time"]; ok {
		recipe.PrepTime = prepTime
	}
	if totalTime, ok := pRecipe.Metadata["total_time"]; ok {
		recipe.TotalTime = totalTime
	}
	if author, ok := pRecipe.Metadata["author"]; ok {
		recipe.Author = author
	}
	if servingsStr, ok := pRecipe.Metadata["servings"]; ok {
		if servings, err := strconv.ParseFloat(servingsStr, 32); err == nil {
			recipe.Servings = float32(servings)
		}
	}
	// Default to 1 serving if not specified or invalid
	if recipe.Servings <= 0 {
		recipe.Servings = 1
	}
	if dateStr, ok := pRecipe.Metadata["date"]; ok {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			recipe.Date = date
		}
	}
	if imgsStr, ok := pRecipe.Metadata["images"]; ok {
		// Assuming images are comma-separated
		recipe.Images = strings.Split(strings.TrimSpace(imgsStr), ",")
		for i := range recipe.Images {
			recipe.Images[i] = strings.TrimSpace(recipe.Images[i])
		}
	}
	if tagsStr, ok := pRecipe.Metadata["tags"]; ok {
		// Assuming tags are comma-separated
		recipe.Tags = strings.Split(strings.TrimSpace(tagsStr), ",")
		for i := range recipe.Tags {
			recipe.Tags[i] = strings.TrimSpace(recipe.Tags[i])
		}
	}

	var prevStep *Step

	for _, step := range pRecipe.Steps {

		newStep := &Step{}

		var prevComponent StepComponent

		for _, component := range step.Components {

			var stepComp StepComponent

			switch component.Type {
			case "ingredient":
				var quant float32
				if component.Quantity == "some" {
					quant = -1 // Use -1 to indicate "some" quantity
				} else {
					quant64, err := strconv.ParseFloat(component.Quantity, 32)
					if err != nil {
						quant = -1 // Default to "some" if parsing fails
					} else {
						quant = float32(quant64)
					}
				}
				stepComp = &Ingredient{
					Name:       component.Name,
					Quantity:   quant,
					Unit:       component.Unit,
					Fixed:      component.Fixed,
					TypedUnit:  CreateTypedUnit(component.Unit),
					Annotation: component.Value,
				}
			case "cookware":
				cookwareQuant, err := strconv.Atoi(component.Quantity)
				if err != nil {
					cookwareQuant = 1 // Default to 1 if parsing fails
				}
				stepComp = &Cookware{
					Name:       component.Name,
					Quantity:   cookwareQuant,
					Annotation: component.Value,
				}
			case "timer":
				stepComp = &Timer{
					Duration:   component.Quantity,
					Unit:       component.Unit,
					Name:       component.Name,
					Annotation: component.Value,
				}
			case "text":
				stepComp = &Instruction{
					Text: component.Value,
				}
			case "section":
				stepComp = &Section{
					Name: component.Name,
				}
			case "comment":
				stepComp = &Comment{
					Text:    component.Value,
					IsBlock: false,
				}
			case "blockComment":
				stepComp = &Comment{
					Text:    component.Value,
					IsBlock: true,
				}
			case "note":
				stepComp = &Note{
					Text: component.Value,
				}
			}

			if stepComp != nil {
				if newStep.FirstComponent == nil {
					newStep.FirstComponent = stepComp
				} else {
					prevComponent.SetNext(stepComp)
				}
				prevComponent = stepComp
			}
		}

		if recipe.FirstStep == nil {
			recipe.FirstStep = newStep
		} else {
			prevStep.NextStep = newStep
		}
		prevStep = newStep
	}

	return recipe
}

// Render returns a human-readable representation of the recipe.
// If a custom renderer has been set via SetRenderer or SetRendererFunc, it will be used.
// Otherwise, a default text format is used showing metadata, ingredients, and steps.
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("lasagna.cook")
//	fmt.Println(recipe.Render())
func (r *Recipe) Render() string {
	if r.RenderFunc != nil {
		return r.RenderFunc()
	}
	// Default rendering logic
	result := fmt.Sprintf("Title: %s\n", r.Title)
	result += fmt.Sprintf("Cuisine: %s\n", r.Cuisine)
	result += fmt.Sprintf("Date: %s\n", r.Date.Format("2006-01-02"))
	result += fmt.Sprintf("Description: %s\n", r.Description)
	result += fmt.Sprintf("Difficulty: %s\n", r.Difficulty)
	result += fmt.Sprintf("Prep Time: %s\n", r.PrepTime)
	result += fmt.Sprintf("Total Time: %s\n", r.TotalTime)
	result += fmt.Sprintf("Author: %s\n", r.Author)
	result += fmt.Sprintf("Servings: %.2f\n", r.Servings)
	if len(r.Images) > 0 {
		result += "Images:\n"
		for _, img := range r.Images {
			result += fmt.Sprintf("- %s\n", img)
		}
	}
	if len(r.Tags) > 0 {
		result += "Tags:\n"
		for _, tag := range r.Tags {
			result += fmt.Sprintf("- %s\n", tag)
		}
	}

	// Iterate through linked list of steps
	stepNum := 1
	currentStep := r.FirstStep
	for currentStep != nil {
		result += fmt.Sprintf("Step %d:\n", stepNum)

		// Iterate through linked list of components in this step
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			result += currentComponent.Render()
			currentComponent = currentComponent.GetNext()
		}

		currentStep = currentStep.NextStep
		stepNum++
	}
	return result
}

// ConvertTo converts the ingredient to a different unit if possible.
// The conversion uses either custom cooking unit conversions (for common units like cups, tbsp, oz)
// or the go-units library for scientific units.
//
// Parameters:
//   - targetUnitStr: The target unit to convert to (e.g., "g", "cup", "ml")
//
// Returns:
//   - *Ingredient: A new ingredient with the converted quantity and unit
//   - error: Error if conversion is not possible (incompatible units, "some" quantity, etc.)
//
// Example:
//
//	ingredient := &Ingredient{Name: "flour", Quantity: 2, Unit: "cup"}
//	converted, err := ingredient.ConvertTo("g")
//	if err == nil {
//	    fmt.Printf("%.0f %s\n", converted.Quantity, converted.Unit) // "473 g"
//	}
func (i *Ingredient) ConvertTo(targetUnitStr string) (*Ingredient, error) {
	if i.TypedUnit == nil {
		return nil, fmt.Errorf("ingredient has no typed unit")
	}

	if i.Quantity == -1 {
		return nil, fmt.Errorf("cannot convert ingredients with 'some' quantity")
	}

	// Try custom cooking unit conversions first
	if isCookingUnit(i.Unit) && isCookingUnit(targetUnitStr) {
		convertedValue, err := convertCookingUnit(float64(i.Quantity), i.Unit, targetUnitStr)
		if err == nil {
			targetUnit := CreateTypedUnit(targetUnitStr)
			converted := &Ingredient{
				Name:           i.Name,
				Quantity:       float32(convertedValue),
				Unit:           targetUnitStr,
				TypedUnit:      targetUnit,
				Subinstruction: i.Subinstruction,
				NextComponent:  i.NextComponent,
			}
			return converted, nil
		}
	}

	// Fall back to go-units for other conversions
	targetUnit, err := units.Find(targetUnitStr)
	if err != nil {
		// If unit not found, create a new one
		targetUnit = units.NewUnit(targetUnitStr, targetUnitStr)
	}

	// Convert using go-units
	convertedValue, err := units.ConvertFloat(float64(i.Quantity), *i.TypedUnit, targetUnit)
	if err != nil {
		return nil, fmt.Errorf("cannot convert from %s to %s: %v", i.Unit, targetUnitStr, err)
	}

	// Create a new ingredient with converted values
	converted := &Ingredient{
		Name:           i.Name,
		Quantity:       float32(convertedValue.Float()),
		Unit:           targetUnitStr,
		TypedUnit:      &targetUnit,
		Subinstruction: i.Subinstruction,
		NextComponent:  i.NextComponent,
	}

	return converted, nil
}

// CanConvertTo checks if the ingredient can be converted to the target unit.
// This allows validating conversions before attempting them.
//
// Parameters:
//   - targetUnitStr: The unit to check conversion compatibility with
//
// Returns:
//   - bool: true if conversion is possible, false otherwise
//
// Example:
//
//	ingredient := &Ingredient{Name: "water", Quantity: 250, Unit: "ml"}
//	if ingredient.CanConvertTo("cup") {
//	    converted, _ := ingredient.ConvertTo("cup")
//	    fmt.Printf("Can convert: %.2f %s\n", converted.Quantity, converted.Unit)
//	}
func (i *Ingredient) CanConvertTo(targetUnitStr string) bool {
	if i.TypedUnit == nil {
		return false
	}

	if i.Quantity == -1 {
		return false // Can't convert "some" quantities
	}

	// Try custom cooking unit conversions first
	if isCookingUnit(i.Unit) && isCookingUnit(targetUnitStr) {
		_, err := convertCookingUnit(float64(i.Quantity), i.Unit, targetUnitStr)
		return err == nil
	}

	// Fall back to go-units
	targetUnit, err := units.Find(targetUnitStr)
	if err != nil {
		// If unit not found, create a new one
		targetUnit = units.NewUnit(targetUnitStr, targetUnitStr)
	}

	_, err = units.ConvertFloat(float64(i.Quantity), *i.TypedUnit, targetUnit)
	return err == nil
}

// GetUnitType returns the ingredient's unit quantity type (e.g., "mass", "volume", "length").
// This helps categorize ingredients and determine valid conversions.
//
// Returns:
//   - string: The quantity type ("mass", "volume", "length", "temperature", "time", "energy", or "")
//
// Example:
//
//	ingredient := &Ingredient{Name: "flour", Quantity: 500, Unit: "g"}
//	fmt.Println(ingredient.GetUnitType()) // "mass"
//
//	ingredient2 := &Ingredient{Name: "milk", Quantity: 2, Unit: "cup"}
//	fmt.Println(ingredient2.GetUnitType()) // "volume"
func (i *Ingredient) GetUnitType() string {
	if i.TypedUnit == nil {
		return ""
	}

	// Check for custom cooking units first
	if cookingType := getCookingUnitType(i.Unit); cookingType != "" {
		return cookingType
	}

	// Check the unit's quantity type from the predefined quantity types
	switch i.TypedUnit.Quantity {
	case "mass":
		return "mass"
	case "volume":
		return "volume"
	case "length":
		return "length"
	case "temperature":
		return "temperature"
	case "time":
		return "time"
	case "energy":
		return "energy"
	default:
		return i.TypedUnit.Quantity
	}
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (i *Ingredient) SetNext(next StepComponent) {
	i.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (i *Ingredient) GetNext() StepComponent {
	return i.NextComponent
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (inst *Instruction) SetNext(next StepComponent) {
	inst.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (inst *Instruction) GetNext() StepComponent {
	return inst.NextComponent
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (t *Timer) SetNext(next StepComponent) {
	t.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (t *Timer) GetNext() StepComponent {
	return t.NextComponent
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (c *Cookware) SetNext(next StepComponent) {
	c.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (c *Cookware) GetNext() StepComponent {
	return c.NextComponent
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (s *Section) SetNext(next StepComponent) {
	s.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (s *Section) GetNext() StepComponent {
	return s.NextComponent
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (cm *Comment) SetNext(next StepComponent) {
	cm.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (cm *Comment) GetNext() StepComponent {
	return cm.NextComponent
}

// SetNext sets the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (n *Note) SetNext(next StepComponent) {
	n.NextComponent = next
}

// GetNext returns the next component in the step's linked list.
// This implements the StepComponent interface for recipe step traversal.
func (n *Note) GetNext() StepComponent {
	return n.NextComponent
}

// IngredientList represents a collection of ingredients with unit consolidation capabilities.
// It provides methods for grouping, converting, and consolidating ingredients for shopping lists
// and recipe scaling operations.
type IngredientList struct {
	Ingredients []*Ingredient // The list of ingredients
}

// NewIngredientList creates a new empty ingredient list.
//
// Returns:
//   - *IngredientList: A new ingredient list ready for use
//
// Example:
//
//	list := cooklang.NewIngredientList()
//	list.Add(&cooklang.Ingredient{Name: "flour", Quantity: 500, Unit: "g"})
func NewIngredientList() *IngredientList {
	return &IngredientList{
		Ingredients: make([]*Ingredient, 0),
	}
}

// Add adds an ingredient to the list.
//
// Parameters:
//   - ingredient: The ingredient to add
//
// Example:
//
//	list := cooklang.NewIngredientList()
//	list.Add(&cooklang.Ingredient{Name: "sugar", Quantity: 100, Unit: "g"})
//	list.Add(&cooklang.Ingredient{Name: "flour", Quantity: 2, Unit: "cup"})
func (il *IngredientList) Add(ingredient *Ingredient) {
	il.Ingredients = append(il.Ingredients, ingredient)
}

// GetIngredientsByName returns all ingredients with the given name.
// This is useful for finding duplicate ingredients before consolidation
// or for extracting specific ingredients from a list.
//
// Parameters:
//   - name: The ingredient name to search for (case-sensitive)
//
// Returns:
//   - []*Ingredient: Slice of matching ingredients (empty if none found)
//
// Example:
//
//	list := recipe.GetIngredients()
//	flourEntries := list.GetIngredientsByName("flour")
//	for _, f := range flourEntries {
//	    fmt.Printf("Found %v %s flour\n", f.Quantity, f.Unit)
//	}
func (il *IngredientList) GetIngredientsByName(name string) []*Ingredient {
	var result []*Ingredient
	for _, ingredient := range il.Ingredients {
		if ingredient.Name == name {
			result = append(result, ingredient)
		}
	}
	return result
}

// ConsolidateByName consolidates ingredients with the same name, converting to a common unit when possible.
// This is useful for creating shopping lists where multiple mentions of the same ingredient
// should be combined into a single entry.
//
// If targetUnit is empty, the method attempts to find a common unit from the ingredients.
// If targetUnit is specified, all compatible ingredients are converted to that unit before consolidation.
//
// Ingredients with "some" quantity (-1) or incompatible units are kept separate.
//
// Parameters:
//   - targetUnit: The unit to convert all ingredients to (empty string to auto-detect)
//
// Returns:
//   - *IngredientList: A new list with consolidated ingredients
//   - error: Any error encountered during consolidation
//
// Example:
//
//	list := cooklang.NewIngredientList()
//	list.Add(&cooklang.Ingredient{Name: "flour", Quantity: 100, Unit: "g"})
//	list.Add(&cooklang.Ingredient{Name: "flour", Quantity: 150, Unit: "g"})
//	consolidated, _ := list.ConsolidateByName("")
//	// consolidated will have one "flour" entry with 250g
func (il *IngredientList) ConsolidateByName(targetUnit string) (*IngredientList, error) {
	consolidated := NewIngredientList()
	ingredientMap := make(map[string][]*Ingredient)

	// Group ingredients by name
	for _, ingredient := range il.Ingredients {
		ingredientMap[ingredient.Name] = append(ingredientMap[ingredient.Name], ingredient)
	}

	// Process each group
	for name, ingredients := range ingredientMap {
		if len(ingredients) == 1 {
			// Single ingredient - convert to target unit if specified
			ing := ingredients[0]
			if targetUnit != "" && ing.Unit != "" && ing.Unit != targetUnit && ing.CanConvertTo(targetUnit) {
				converted, err := ing.ConvertTo(targetUnit)
				if err == nil {
					consolidated.Add(converted)
					continue
				}
			}
			// No conversion needed or possible, add as-is
			consolidated.Add(ing)
			continue
		}

		// Multiple ingredients with same name - try to consolidate
		var totalQuantity float32
		var unitToUse string
		var typedUnit *units.Unit
		var hasConvertibleUnits bool

		// Check if we should use the target unit or find a common unit
		if targetUnit != "" {
			unitToUse = targetUnit
			typedUnit = CreateTypedUnit(targetUnit)
			hasConvertibleUnits = true
		} else {
			// Use the unit from the first ingredient that has a unit
			for _, ing := range ingredients {
				if ing.Unit != "" {
					unitToUse = ing.Unit
					typedUnit = ing.TypedUnit
					hasConvertibleUnits = true
					break
				}
			}
		}

		// Try to convert and sum quantities
		for _, ingredient := range ingredients {
			if ingredient.Quantity == -1 {
				// "Some" quantity - add separately
				consolidated.Add(ingredient)
				continue
			}

			if ingredient.Unit == "" {
				// Unitless ingredient - add to list separately if we have units, or sum if all unitless
				if !hasConvertibleUnits {
					totalQuantity += ingredient.Quantity
				} else {
					// Add unitless ingredient separately
					consolidated.Add(ingredient)
				}
				continue
			}

			if hasConvertibleUnits && ingredient.CanConvertTo(unitToUse) {
				converted, err := ingredient.ConvertTo(unitToUse)
				if err != nil {
					// Can't convert, add separately
					consolidated.Add(ingredient)
					continue
				}
				totalQuantity += converted.Quantity
			} else if ingredient.Unit == unitToUse || unitToUse == "" {
				// Same unit or no target unit specified
				totalQuantity += ingredient.Quantity
				if unitToUse == "" {
					unitToUse = ingredient.Unit
					typedUnit = ingredient.TypedUnit
				}
			} else {
				// Different unit that can't be converted, add separately
				consolidated.Add(ingredient)
			}
		}

		// Add consolidated ingredient if we have something to consolidate
		if totalQuantity > 0 {
			consolidatedIngredient := &Ingredient{
				Name:      name,
				Quantity:  totalQuantity,
				Unit:      unitToUse,
				TypedUnit: typedUnit,
			}
			consolidated.Add(consolidatedIngredient)
		}
	}

	return consolidated, nil
}

// ToMap returns a map of ingredient names to their formatted quantities.
// This is useful for displaying shopping lists in a simple key-value format.
//
// The quantity formatting follows these rules:
//   - Whole numbers are shown without decimals (e.g., "100 g")
//   - Fractional quantities show one decimal place (e.g., "1.5 cup")
//   - "Some" quantities (-1) are displayed as "some" or "some [unit]"
//   - Unitless ingredients show just the quantity or "some"
//
// Returns:
//   - map[string]string: Map of ingredient names to formatted quantity strings
//
// Example:
//
//	list := recipe.GetIngredients()
//	for name, qty := range list.ToMap() {
//	    fmt.Printf("- %s: %s\n", name, qty)
//	}
//	// Output:
//	// - flour: 500 g
//	// - eggs: 3
//	// - salt: some
func (il *IngredientList) ToMap() map[string]string {
	result := make(map[string]string)
	for _, ingredient := range il.Ingredients {
		key := ingredient.Name
		if ingredient.Unit != "" {
			if ingredient.Quantity == -1 {
				result[key] = "some " + ingredient.Unit
			} else {
				// Use %g to avoid scientific notation for reasonable numbers
				quantity := ingredient.Quantity
				if quantity == float32(int(quantity)) {
					// Show as integer if it's a whole number
					result[key] = fmt.Sprintf("%.0f %s", quantity, ingredient.Unit)
				} else {
					result[key] = fmt.Sprintf("%.1f %s", quantity, ingredient.Unit)
				}
			}
		} else {
			if ingredient.Quantity == -1 {
				result[key] = "some"
			} else if ingredient.Quantity > 0 {
				if ingredient.Quantity == float32(int(ingredient.Quantity)) {
					result[key] = fmt.Sprintf("%.0f", ingredient.Quantity)
				} else {
					result[key] = fmt.Sprintf("%.1f", ingredient.Quantity)
				}
			} else {
				result[key] = "some"
			}
		}
	}
	return result
}

// GetIngredients returns all ingredients from a recipe, extracted from all steps.
// This traverses the recipe's linked list structure to collect every ingredient mention.
//
// Returns:
//   - *IngredientList: A list containing all ingredients in order of appearance
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("lasagna.cook")
//	ingredients := recipe.GetIngredients()
//	for _, ing := range ingredients.Ingredients {
//	    fmt.Printf("%s: %.1f %s\n", ing.Name, ing.Quantity, ing.Unit)
//	}
func (r *Recipe) GetIngredients() *IngredientList {
	ingredientList := NewIngredientList()

	// Walk through all steps and components to find ingredients
	currentStep := r.FirstStep
	for currentStep != nil {
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			if ingredient, ok := currentComponent.(*Ingredient); ok {
				ingredientList.Add(ingredient)
			}
			currentComponent = currentComponent.GetNext()
		}
		currentStep = currentStep.NextStep
	}

	return ingredientList
}

// GetCookware returns all cookware items from a recipe, extracted from all steps.
//
// Returns:
//   - []*Cookware: A slice containing all cookware items in order of appearance
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("pasta.cook")
//	cookware := recipe.GetCookware()
//	for _, cw := range cookware {
//	    fmt.Printf("%s (qty: %d)\n", cw.Name, cw.Quantity)
//	}
func (r *Recipe) GetCookware() []*Cookware {
	var cookware []*Cookware

	// Walk through all steps and components to find cookware
	currentStep := r.FirstStep
	for currentStep != nil {
		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			if cw, ok := currentComponent.(*Cookware); ok {
				cookware = append(cookware, cw)
			}
			currentComponent = currentComponent.GetNext()
		}
		currentStep = currentStep.NextStep
	}

	return cookware
}

// ConvertToSystem converts all ingredients in the list to the target unit system.
// Each ingredient is converted individually using Ingredient.ConvertToSystem.
// Ingredients that cannot be converted (no TypedUnit or "some" quantity) are copied as-is.
//
// Parameters:
//   - system: The target unit system (UnitSystemMetric, UnitSystemUS, UnitSystemImperial)
//
// Returns:
//   - *IngredientList: A new list with converted ingredients
//
// Example:
//
//	metricList := recipe.GetIngredients()
//	usList := metricList.ConvertToSystem(cooklang.UnitSystemUS)
func (il *IngredientList) ConvertToSystem(system UnitSystem) *IngredientList {
	result := NewIngredientList()

	for _, ingredient := range il.Ingredients {
		converted := ingredient.ConvertToSystem(system)
		result.Add(converted)
	}

	return result
}

// ConvertToSystem converts an ingredient to the target unit system.
// The conversion selects an appropriate unit based on the ingredient's unit type
// (mass or volume) and converts the quantity accordingly.
//
// If the ingredient has no TypedUnit or has "some" quantity (-1), a copy is returned unchanged.
//
// Parameters:
//   - system: The target unit system (UnitSystemMetric, UnitSystemUS, UnitSystemImperial)
//
// Returns:
//   - *Ingredient: A new ingredient with converted quantity and unit
//
// Example:
//
//	flour := cooklang.NewIngredient("flour", 500, "g")
//	usFlour := flour.ConvertToSystem(cooklang.UnitSystemUS)
//	fmt.Printf("%v %s\n", usFlour.Quantity, usFlour.Unit) // "17.6 oz"
func (i *Ingredient) ConvertToSystem(system UnitSystem) *Ingredient {
	if i.TypedUnit == nil || i.Quantity == -1 {
		// Return a copy of the ingredient if it can't be converted
		return &Ingredient{
			Name:           i.Name,
			Quantity:       i.Quantity,
			Unit:           i.Unit,
			TypedUnit:      i.TypedUnit,
			Subinstruction: i.Subinstruction,
			NextComponent:  i.NextComponent,
		}
	}

	unitType := i.GetUnitType()
	canonicalUnit, ok := canonicalUnits[system][unitType]
	if !ok {
		// No canonical unit for this type in the target system, return as-is
		return &Ingredient{
			Name:           i.Name,
			Quantity:       i.Quantity,
			Unit:           i.Unit,
			TypedUnit:      i.TypedUnit,
			Subinstruction: i.Subinstruction,
			NextComponent:  i.NextComponent,
		}
	}

	// Try to convert to the canonical unit
	if converted, err := i.ConvertTo(canonicalUnit); err == nil {
		// Check if we should use a more appropriate unit based on quantity
		if alternatives, hasAlternatives := commonUnitMappings[system][unitType]; hasAlternatives {
			bestUnit := i.getBestUnit(converted.Quantity, canonicalUnit, alternatives)
			if bestUnit != canonicalUnit {
				if finalConverted, err := converted.ConvertTo(bestUnit); err == nil {
					return finalConverted
				}
			}
		}
		return converted
	}

	// If conversion failed, return the original ingredient
	return &Ingredient{
		Name:           i.Name,
		Quantity:       i.Quantity,
		Unit:           i.Unit,
		TypedUnit:      i.TypedUnit,
		Subinstruction: i.Subinstruction,
		NextComponent:  i.NextComponent,
	}
}

// getBestUnit selects the most appropriate unit based on quantity
func (i *Ingredient) getBestUnit(quantity float32, defaultUnit string, alternatives map[string]string) string {
	unitType := i.GetUnitType()

	switch unitType {
	case "volume":
		switch {
		case quantity >= 946.4 && alternatives["large"] != "": // ~1 quart
			return alternatives["large"]
		case quantity >= 236.6 && alternatives["medium"] != "": // ~1 cup
			return alternatives["medium"]
		case quantity >= 14.8 && alternatives["small"] != "": // ~1 tbsp
			return alternatives["small"]
		case quantity < 14.8 && alternatives["tiny"] != "":
			return alternatives["tiny"]
		}
	case "mass":
		switch {
		case quantity >= 1000 && alternatives["large"] != "": // 1kg or more
			return alternatives["large"]
		case alternatives["small"] != "":
			return alternatives["small"]
		}
	}

	return defaultUnit
}

// ConvertToSystemWithConsolidation converts ingredients to a target system and consolidates by name.
// This combines ConvertToSystem and ConsolidateByName in a single operation.
//
// Parameters:
//   - system: The target unit system (UnitSystemMetric, UnitSystemUS, UnitSystemImperial)
//
// Returns:
//   - *IngredientList: A new list with converted and consolidated ingredients
//   - error: Any error encountered during consolidation
//
// Example:
//
//	ingredients := recipe.GetIngredients()
//	usList, _ := ingredients.ConvertToSystemWithConsolidation(cooklang.UnitSystemUS)
func (il *IngredientList) ConvertToSystemWithConsolidation(system UnitSystem) (*IngredientList, error) {
	converted := il.ConvertToSystem(system)
	return converted.ConsolidateByName("")
}

// ConvertToSystemBartender converts all ingredients using bartender-friendly conversions.
// This uses practical bartender measurements (30ml = 1oz instead of 29.5735ml) and
// smart unit selection (dashes for tiny amounts, friendly fractions).
//
// Parameters:
//   - system: The target unit system (UnitSystemMetric, UnitSystemUS)
//
// Returns:
//   - *IngredientList: A new list with bartender-friendly converted ingredients
//
// Example:
//
//	ingredients := cocktail.GetIngredients()
//	usBar := ingredients.ConvertToSystemBartender(cooklang.UnitSystemUS)
func (il *IngredientList) ConvertToSystemBartender(system UnitSystem) *IngredientList {
	result := NewIngredientList()

	for _, ingredient := range il.Ingredients {
		converted := ingredient.ConvertToSystemBartender(system)
		result.Add(converted)
	}

	return result
}

// ConvertToSystemBartender converts an ingredient using bartender-friendly conversions.
// It uses practical measurements like dashes for very small amounts (<3ml), and skips
// conversion for cocktail-specific units (dash, splash, etc.).
//
// Features:
//   - Practical oz/ml conversion (30ml = 1oz)
//   - Smart unit selection based on quantity
//   - Dashes for tiny amounts (≤3ml)
//   - Preserves cocktail-specific units
//
// Parameters:
//   - system: The target unit system (UnitSystemMetric, UnitSystemUS)
//
// Returns:
//   - *Ingredient: A new ingredient with bartender-friendly quantity and unit
//
// Example:
//
//	vodka := cooklang.NewIngredient("vodka", 45, "ml")
//	usVodka := vodka.ConvertToSystemBartender(cooklang.UnitSystemUS)
//	fmt.Printf("%v %s\n", usVodka.Quantity, usVodka.Unit) // "1.5 oz"
func (i *Ingredient) ConvertToSystemBartender(system UnitSystem) *Ingredient {
	// Skip conversion for ingredients without quantities
	if i.Quantity == -1 {
		return &Ingredient{
			Name:           i.Name,
			Quantity:       i.Quantity,
			Unit:           i.Unit,
			TypedUnit:      i.TypedUnit,
			Subinstruction: i.Subinstruction,
			Annotation:     i.Annotation,
			NextComponent:  i.NextComponent,
		}
	}

	// Get unit info to determine ml value for smart unit selection
	unitInfo := GetCocktailUnit(i.Unit)

	// For cocktail-specific units (dash, splash, etc.), never convert
	if IsCocktailSpecificUnit(i.Unit) {
		return &Ingredient{
			Name:           i.Name,
			Quantity:       i.Quantity,
			Unit:           i.Unit,
			TypedUnit:      i.TypedUnit,
			Subinstruction: i.Subinstruction,
			Annotation:     i.Annotation,
			NextComponent:  i.NextComponent,
		}
	}

	// Calculate ml value for smart unit selection decisions
	var mlValue float64
	if unitInfo != nil && unitInfo.MlValue > 0 {
		mlValue = float64(i.Quantity) * unitInfo.MlValue
	}

	// IMPORTANT: For very small amounts (≤3ml), always convert to dashes
	// regardless of whether we're staying in the same unit system.
	// This handles awkward fractions like "1/12 fl oz" → "3 dashes"
	if mlValue > 0 && mlValue <= 3 {
		result := SelectBestUnit(mlValue, system)
		return &Ingredient{
			Name:           i.Name,
			Quantity:       float32(result.Value),
			Unit:           result.Unit,
			TypedUnit:      nil,
			Subinstruction: i.Subinstruction,
			Annotation:     i.Annotation,
			NextComponent:  i.NextComponent,
		}
	}

	// For larger amounts in the same system, skip conversion to preserve original
	sourceSystem := DetectUnitSystemFromUnit(i.Unit)
	if sourceSystem == system && system != UnitSystemUnknown {
		return &Ingredient{
			Name:           i.Name,
			Quantity:       i.Quantity,
			Unit:           i.Unit,
			TypedUnit:      i.TypedUnit,
			Subinstruction: i.Subinstruction,
			Annotation:     i.Annotation,
			NextComponent:  i.NextComponent,
		}
	}

	// Use bartender conversion for volume units
	if unitInfo != nil && unitInfo.MlValue > 0 {
		result := ConvertVolumeBartender(float64(i.Quantity), i.Unit, system)
		return &Ingredient{
			Name:           i.Name,
			Quantity:       float32(result.Value),
			Unit:           result.Unit,
			TypedUnit:      nil, // Clear typed unit since we're using bartender conversion
			Subinstruction: i.Subinstruction,
			Annotation:     i.Annotation,
			NextComponent:  i.NextComponent,
		}
	}

	// Fall back to standard conversion for non-cocktail units
	return i.ConvertToSystem(system)
}

// FormatQuantityBartender formats an ingredient's quantity using bartender-friendly formatting.
// This uses fractions instead of decimals (e.g., "1/2" instead of "0.5") and
// handles pluralization appropriately.
//
// Returns:
//   - string: Formatted quantity string (e.g., "1 1/2 oz", "some", "")
//
// Example:
//
//	vodka := cooklang.NewIngredient("vodka", 1.5, "oz")
//	fmt.Println(vodka.FormatQuantityBartender()) // "1 1/2 oz"
func (i *Ingredient) FormatQuantityBartender() string {
	if i.Quantity == -1 {
		return "some"
	}
	if i.Quantity == 0 {
		return ""
	}

	result := SmartUnitResult{
		Value: float64(i.Quantity),
		Unit:  i.Unit,
	}
	return FormatBartenderValue(result)
}

// GetShoppingListInSystem returns a shopping list with ingredients converted to the target unit system.
// The ingredients are first converted, then consolidated by name to combine duplicate entries.
//
// Parameters:
//   - system: The target unit system (UnitSystemMetric, UnitSystemUS, UnitSystemImperial)
//
// Returns:
//   - map[string]string: A map of ingredient names to formatted quantities
//   - error: Any error encountered during conversion
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("pasta.cook")
//	usList, _ := recipe.GetShoppingListInSystem(cooklang.UnitSystemUS)
//	for name, qty := range usList {
//	    fmt.Printf("%s: %s\n", name, qty)
//	}
func (r *Recipe) GetShoppingListInSystem(system UnitSystem) (map[string]string, error) {
	ingredients := r.GetIngredients()
	converted := ingredients.ConvertToSystem(system)
	consolidated, err := converted.ConsolidateByName("")
	if err != nil {
		return nil, err
	}
	return consolidated.ToMap(), nil
}

// GetMetricShoppingList returns a shopping list with all ingredients converted to metric units.
// This is a convenience method for GetShoppingListInSystem(UnitSystemMetric).
//
// Returns:
//   - map[string]string: A map of ingredient names to quantities (e.g., "flour": "500 g")
//   - error: Any error encountered during conversion
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("cookies.cook")
//	shoppingList, err := recipe.GetMetricShoppingList()
//	if err == nil {
//	    for ingredient, amount := range shoppingList {
//	        fmt.Printf("%s: %s\n", ingredient, amount)
//	    }
//	}
func (r *Recipe) GetMetricShoppingList() (map[string]string, error) {
	return r.GetShoppingListInSystem(UnitSystemMetric)
}

// GetUSShoppingList returns a shopping list with all ingredients converted to US customary units.
// Common conversions include: cups, tablespoons, teaspoons, ounces, pounds.
//
// Returns:
//   - map[string]string: A map of ingredient names to quantities (e.g., "flour": "2 cup")
//   - error: Any error encountered during conversion
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("cookies.cook")
//	shoppingList, err := recipe.GetUSShoppingList()
func (r *Recipe) GetUSShoppingList() (map[string]string, error) {
	return r.GetShoppingListInSystem(UnitSystemUS)
}

// GetImperialShoppingList returns a shopping list with all ingredients converted to Imperial units.
// Common conversions include: pints, fluid ounces, pounds, ounces.
//
// Returns:
//   - map[string]string: A map of ingredient names to quantities
//   - error: Any error encountered during conversion
func (r *Recipe) GetImperialShoppingList() (map[string]string, error) {
	return r.GetShoppingListInSystem(UnitSystemImperial)
}

// GetCollectedIngredients returns a consolidated list of all ingredients from the recipe.
// This combines GetIngredients() and ConsolidateByName() into a single convenient function.
// Duplicate ingredients with compatible units are combined into single entries.
//
// Returns:
//   - *IngredientList: A consolidated list of ingredients
//   - error: Any error encountered during consolidation
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("cake.cook")
//	ingredients, _ := recipe.GetCollectedIngredients()
//	for _, ing := range ingredients.Ingredients {
//	    fmt.Printf("%s: %v %s\n", ing.Name, ing.Quantity, ing.Unit)
//	}
func (r *Recipe) GetCollectedIngredients() (*IngredientList, error) {
	ingredients := r.GetIngredients()
	return ingredients.ConsolidateByName("")
}

// GetCollectedIngredientsWithUnit returns a consolidated list of all ingredients from the recipe,
// converting them to the specified target unit when possible.
//
// Parameters:
//   - targetUnit: The unit to convert compatible ingredients to (e.g., "g", "ml", "oz")
//
// Returns:
//   - *IngredientList: A consolidated list with converted ingredients
//   - error: Any error encountered during consolidation
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("cake.cook")
//	// Get all ingredients in grams
//	ingredients, _ := recipe.GetCollectedIngredientsWithUnit("g")
func (r *Recipe) GetCollectedIngredientsWithUnit(targetUnit string) (*IngredientList, error) {
	ingredients := r.GetIngredients()
	return ingredients.ConsolidateByName(targetUnit)
}

// GetCollectedIngredientsMap returns a map of ingredient names to their consolidated quantities.
// This is useful for creating simple shopping lists or ingredient summaries.
//
// Returns:
//   - map[string]string: A map of ingredient names to formatted quantity strings
//   - error: Any error encountered during consolidation
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("pasta.cook")
//	ingredientMap, _ := recipe.GetCollectedIngredientsMap()
//	for name, qty := range ingredientMap {
//	    fmt.Printf("- %s: %s\n", name, qty)
//	}
func (r *Recipe) GetCollectedIngredientsMap() (map[string]string, error) {
	collectedIngredients, err := r.GetCollectedIngredients()
	if err != nil {
		return nil, err
	}
	return collectedIngredients.ToMap(), nil
}

// ShoppingList represents a consolidated list of ingredients from multiple recipes.
// It combines ingredients across recipes and provides a unified shopping list with recipe attribution.
type ShoppingList struct {
	Ingredients *IngredientList `json:"ingredients"`       // Consolidated ingredient list
	Recipes     []string        `json:"recipes,omitempty"` // List of recipe titles included
}

// CreateShoppingList creates a consolidated shopping list from multiple recipes.
// All ingredients from all recipes are combined and consolidated by name, automatically
// converting compatible units and summing quantities.
//
// Parameters:
//   - recipes: Variable number of Recipe pointers to include in the shopping list
//
// Returns:
//   - *ShoppingList: A shopping list with consolidated ingredients
//   - error: Any error encountered during consolidation
//
// Example:
//
//	recipe1, _ := cooklang.ParseFile("pasta.cook")
//	recipe2, _ := cooklang.ParseFile("salad.cook")
//	shoppingList, err := cooklang.CreateShoppingList(recipe1, recipe2)
//	if err == nil {
//	    for name, amount := range shoppingList.ToMap() {
//	        fmt.Printf("%s: %s\n", name, amount)
//	    }
//	}
func CreateShoppingList(recipes ...*Recipe) (*ShoppingList, error) {
	if len(recipes) == 0 {
		return &ShoppingList{
			Ingredients: &IngredientList{Ingredients: []*Ingredient{}},
			Recipes:     []string{},
		}, nil
	}

	// Collect all ingredients from all recipes
	allIngredients := []*Ingredient{}
	recipeNames := []string{}

	for _, recipe := range recipes {
		ingredients := recipe.GetIngredients()
		allIngredients = append(allIngredients, ingredients.Ingredients...)
		if recipe.Title != "" {
			recipeNames = append(recipeNames, recipe.Title)
		}
	}

	// Create a combined ingredient list and consolidate
	combinedList := &IngredientList{Ingredients: allIngredients}
	consolidated, err := combinedList.ConsolidateByName("")
	if err != nil {
		return nil, err
	}

	return &ShoppingList{
		Ingredients: consolidated,
		Recipes:     recipeNames,
	}, nil
}

// CreateShoppingListWithUnit creates a shopping list with ingredients converted to the target unit.
// All compatible ingredients are converted to the specified unit before consolidation.
//
// Parameters:
//   - targetUnit: The unit to convert ingredients to (e.g., "g", "ml", "oz", "kg")
//   - recipes: Variable number of Recipe pointers to include
//
// Returns:
//   - *ShoppingList: A shopping list with converted and consolidated ingredients
//   - error: Any error encountered during consolidation
//
// Example:
//
//	recipe1, _ := cooklang.ParseFile("pasta.cook")
//	recipe2, _ := cooklang.ParseFile("salad.cook")
//	// Get shopping list with all weights in grams
//	list, _ := cooklang.CreateShoppingListWithUnit("g", recipe1, recipe2)
func CreateShoppingListWithUnit(targetUnit string, recipes ...*Recipe) (*ShoppingList, error) {
	if len(recipes) == 0 {
		return &ShoppingList{
			Ingredients: &IngredientList{Ingredients: []*Ingredient{}},
			Recipes:     []string{},
		}, nil
	}

	// Collect all ingredients from all recipes
	allIngredients := []*Ingredient{}
	recipeNames := []string{}

	for _, recipe := range recipes {
		ingredients := recipe.GetIngredients()
		allIngredients = append(allIngredients, ingredients.Ingredients...)
		if recipe.Title != "" {
			recipeNames = append(recipeNames, recipe.Title)
		}
	}

	// Create a combined ingredient list and consolidate with unit conversion
	combinedList := &IngredientList{Ingredients: allIngredients}
	consolidated, err := combinedList.ConsolidateByName(targetUnit)
	if err != nil {
		return nil, err
	}

	return &ShoppingList{
		Ingredients: consolidated,
		Recipes:     recipeNames,
	}, nil
}

// CreateShoppingListForServings creates a shopping list by scaling each recipe
// to the target number of servings before combining ingredients.
//
// This is ideal for meal planning where recipes have different serving sizes
// and you want to normalize them all to your household size.
//
// Each recipe is scaled from its original servings to the target servings,
// then all ingredients are combined and consolidated.
//
// Parameters:
//   - targetServings: The desired number of servings for each recipe
//   - recipes: Variable number of recipe pointers to combine
//
// Returns:
//   - *ShoppingList: Consolidated shopping list with all ingredients scaled
//   - error: Error if consolidation fails
//
// Example:
//
//	monday, _ := cooklang.ParseFile("monday.cook")     // servings: 2
//	tuesday, _ := cooklang.ParseFile("tuesday.cook")  // servings: 8
//	wednesday, _ := cooklang.ParseFile("wednesday.cook") // servings: 1 (default)
//
//	// Create shopping list for household of 5
//	list, _ := cooklang.CreateShoppingListForServings(5, monday, tuesday, wednesday)
//	// monday scaled 2.5x, tuesday scaled 0.625x, wednesday scaled 5x
func CreateShoppingListForServings(targetServings float64, recipes ...*Recipe) (*ShoppingList, error) {
	if len(recipes) == 0 {
		return &ShoppingList{
			Ingredients: &IngredientList{Ingredients: []*Ingredient{}},
			Recipes:     []string{},
		}, nil
	}

	// Scale each recipe to target servings, then collect ingredients
	allIngredients := []*Ingredient{}
	recipeNames := []string{}

	for _, recipe := range recipes {
		// Scale this recipe to target servings
		scaledRecipe := recipe.ScaleToServings(targetServings)

		// Collect ingredients from the scaled recipe
		ingredients := scaledRecipe.GetIngredients()
		allIngredients = append(allIngredients, ingredients.Ingredients...)

		if recipe.Title != "" {
			recipeNames = append(recipeNames, recipe.Title)
		}
	}

	// Create a combined ingredient list and consolidate
	combinedList := &IngredientList{Ingredients: allIngredients}
	consolidated, err := combinedList.ConsolidateByName("")
	if err != nil {
		return nil, err
	}

	return &ShoppingList{
		Ingredients: consolidated,
		Recipes:     recipeNames,
	}, nil
}

// CreateShoppingListForServingsWithUnit creates a shopping list by scaling each recipe
// to the target servings and converting ingredients to the target unit.
//
// This combines the functionality of CreateShoppingListForServings and
// CreateShoppingListWithUnit for meal planning with unit standardization.
//
// Parameters:
//   - targetServings: The desired number of servings for each recipe
//   - targetUnit: The unit to convert compatible ingredients to (e.g., "g", "ml", "kg")
//   - recipes: Variable number of recipe pointers to combine
//
// Returns:
//   - *ShoppingList: Consolidated shopping list with scaled and converted ingredients
//   - error: Error if conversion or consolidation fails
//
// Example:
//
//	// Create shopping list for 4 servings with metric units
//	list, _ := cooklang.CreateShoppingListForServingsWithUnit(4, "g", recipes...)
func CreateShoppingListForServingsWithUnit(targetServings float64, targetUnit string, recipes ...*Recipe) (*ShoppingList, error) {
	if len(recipes) == 0 {
		return &ShoppingList{
			Ingredients: &IngredientList{Ingredients: []*Ingredient{}},
			Recipes:     []string{},
		}, nil
	}

	// Scale each recipe to target servings, then collect ingredients
	allIngredients := []*Ingredient{}
	recipeNames := []string{}

	for _, recipe := range recipes {
		// Scale this recipe to target servings
		scaledRecipe := recipe.ScaleToServings(targetServings)

		// Collect ingredients from the scaled recipe
		ingredients := scaledRecipe.GetIngredients()
		allIngredients = append(allIngredients, ingredients.Ingredients...)

		if recipe.Title != "" {
			recipeNames = append(recipeNames, recipe.Title)
		}
	}

	// Create a combined ingredient list and consolidate with unit conversion
	combinedList := &IngredientList{Ingredients: allIngredients}
	consolidated, err := combinedList.ConsolidateByName(targetUnit)
	if err != nil {
		return nil, err
	}

	return &ShoppingList{
		Ingredients: consolidated,
		Recipes:     recipeNames,
	}, nil
}

// ToMap returns the shopping list as a map of ingredient names to formatted quantities.
// This is a convenience method that delegates to IngredientList.ToMap().
//
// Returns:
//   - map[string]string: A map of ingredient names to quantity strings
//
// Example:
//
//	list, _ := cooklang.CreateShoppingList(recipe1, recipe2)
//	for name, qty := range list.ToMap() {
//	    fmt.Printf("- %s: %s\n", name, qty)
//	}
func (sl *ShoppingList) ToMap() map[string]string {
	if sl.Ingredients == nil {
		return map[string]string{}
	}
	return sl.Ingredients.ToMap()
}

// Scale scales all ingredients in the shopping list by the given multiplier.
// This is useful when adjusting recipe servings or batch cooking.
// Ingredients with "some" quantity (-1) are not scaled.
//
// Parameters:
//   - multiplier: The scaling factor (e.g., 2.0 for double, 0.5 for half)
//
// Returns:
//   - *ShoppingList: A new shopping list with scaled quantities
//
// Example:
//
//	shoppingList, _ := cooklang.CreateShoppingList(recipe)
//	doubled := shoppingList.Scale(2.0)  // Double all quantities
//	halved := shoppingList.Scale(0.5)   // Half all quantities
func (sl *ShoppingList) Scale(multiplier float64) *ShoppingList {
	if sl.Ingredients == nil || len(sl.Ingredients.Ingredients) == 0 {
		return sl
	}

	scaledIngredients := make([]*Ingredient, len(sl.Ingredients.Ingredients))
	for i, ingredient := range sl.Ingredients.Ingredients {
		scaledIngredient := &Ingredient{
			Name:     ingredient.Name,
			Quantity: ingredient.Quantity,
			Unit:     ingredient.Unit,
		}
		if ingredient.Quantity > 0 {
			scaledIngredient.Quantity = ingredient.Quantity * float32(multiplier)
		}
		scaledIngredients[i] = scaledIngredient
	}

	return &ShoppingList{
		Ingredients: &IngredientList{Ingredients: scaledIngredients},
		Recipes:     sl.Recipes,
	}
}

// Count returns the number of unique ingredients in the shopping list.
//
// Returns:
//   - int: The number of ingredients (0 if the list is empty or nil)
//
// Example:
//
//	list, _ := cooklang.CreateShoppingList(recipe1, recipe2)
//	fmt.Printf("You need %d ingredients\n", list.Count())
func (sl *ShoppingList) Count() int {
	if sl.Ingredients == nil {
		return 0
	}
	return len(sl.Ingredients.Ingredients)
}

// Scale creates a new recipe with all ingredient quantities scaled by the given factor.
// This is useful for adjusting recipe servings or batch cooking.
// Timers, cookware, and instructions are copied unchanged.
// Ingredients with "some" quantity (-1) are not scaled.
//
// The servings metadata is also updated if present.
//
// Parameters:
//   - factor: The scaling factor (e.g., 2.0 for double, 0.5 for half)
//
// Returns:
//   - *Recipe: A new recipe with scaled quantities
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("cookies.cook")
//	doubled := recipe.Scale(2.0)  // Double all quantities
//	halved := recipe.Scale(0.5)   // Half all quantities
func (r *Recipe) Scale(factor float64) *Recipe {
	// Create a copy of the recipe
	scaledRecipe := &Recipe{
		Title:       r.Title,
		Cuisine:     r.Cuisine,
		Date:        r.Date,
		Description: r.Description,
		Difficulty:  r.Difficulty,
		PrepTime:    r.PrepTime,
		TotalTime:   r.TotalTime,
		Author:      r.Author,
		Images:      r.Images,
		Tags:        r.Tags,
		Metadata:    make(map[string]string),
	}

	// Copy metadata
	for k, v := range r.Metadata {
		scaledRecipe.Metadata[k] = v
	}

	// Update servings if present
	if r.Servings > 0 {
		scaledRecipe.Servings = r.Servings * float32(factor)
		scaledRecipe.Metadata["servings"] = strconv.FormatFloat(float64(scaledRecipe.Servings), 'f', -1, 32)
	}

	// Scale the steps and ingredients
	var lastStep *Step
	for step := r.FirstStep; step != nil; step = step.NextStep {
		newStep := &Step{}

		// Copy and scale components
		var lastComponent StepComponent
		for component := step.FirstComponent; component != nil; component = component.GetNext() {
			var newComponent StepComponent

			switch comp := component.(type) {
			case *Ingredient:
				// Scale the ingredient (unless it's fixed or "some")
				newQty := comp.Quantity
				if newQty > 0 && !comp.Fixed { // Don't scale "some" (-1), zero, or fixed quantities
					newQty = comp.Quantity * float32(factor)
				}
				newComponent = &Ingredient{
					Name:           comp.Name,
					Quantity:       newQty,
					Unit:           comp.Unit,
					Fixed:          comp.Fixed,
					TypedUnit:      comp.TypedUnit,
					Subinstruction: comp.Subinstruction,
					Annotation:     comp.Annotation,
				}

			case *Timer:
				// Copy timer unchanged
				newComponent = &Timer{
					Name:       comp.Name,
					Duration:   comp.Duration,
					Unit:       comp.Unit,
					Text:       comp.Text,
					Annotation: comp.Annotation,
				}

			case *Cookware:
				// Copy cookware unchanged
				newComponent = &Cookware{
					Name:       comp.Name,
					Quantity:   comp.Quantity,
					Annotation: comp.Annotation,
				}

			case *Instruction:
				// Copy instruction unchanged
				newComponent = &Instruction{
					Text: comp.Text,
				}
			}

			// Link components
			if lastComponent == nil {
				newStep.FirstComponent = newComponent
			} else {
				lastComponent.SetNext(newComponent)
			}
			lastComponent = newComponent
		}

		// Link steps
		if scaledRecipe.FirstStep == nil {
			scaledRecipe.FirstStep = newStep
		} else {
			lastStep.NextStep = newStep
		}
		lastStep = newStep
	}

	return scaledRecipe
}

// ScaleToServings creates a new recipe scaled to the target number of servings.
// If the recipe doesn't have servings specified, it assumes 1 serving.
//
// Parameters:
//   - targetServings: The desired number of servings
//
// Returns:
//   - *Recipe: A new recipe scaled to the target servings
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("cookies.cook") // 12 servings
//	scaled := recipe.ScaleToServings(24)            // Double the recipe
func (r *Recipe) ScaleToServings(targetServings float64) *Recipe {
	originalServings := float64(r.Servings)
	if originalServings <= 0 {
		originalServings = 1 // Assume 1 serving if not specified
	}
	factor := targetServings / originalServings
	return r.Scale(factor)
}
