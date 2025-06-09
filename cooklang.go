package cooklang

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bcicen/go-units"
	"github.com/hilli/cooklang/parser"
)

type Recipe struct {
	Title       string    `json:"title,omitempty"`
	Cuisine     string    `json:"cuisine,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Description string    `json:"description,omitempty"`
	Difficulty  string    `json:"difficulty,omitempty"`
	PrepTime    string    `json:"prep_time,omitempty"`
	TotalTime   string    `json:"total_time,omitempty"`
	Metadata    Metadata  `json:"metadata,omitempty"`
	Author      string    `json:"author,omitempty"`
	Images      []string  `json:"images,omitempty"`
	Servings    float32   `json:"servings,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	FirstStep   *Step     `json:"first_step,omitempty"`
	CooklangRenderable
}

type CooklangRenderable struct {
	RenderFunc func() string `json:"-"`
}

type CooklangRecipe interface {
	isRecipe()
}

type Metadata map[string]string

type StepComponent interface {
	isStepComponent()
	Render() string
	SetNext(StepComponent)
	GetNext() StepComponent
}

type Step struct {
	FirstComponent StepComponent `json:"first_component,omitempty"`
	NextStep       *Step         `json:"next_step,omitempty"`
	CooklangRenderable
}

func (Instruction) isStepComponent() {}
func (Timer) isStepComponent()       {}
func (Cookware) isStepComponent()    {}
func (Ingredient) isStepComponent()  {}

func (i Ingredient) Render() string {
	if i.Quantity > 0 {
		return fmt.Sprintf("@%s{%g%%%s}", i.Name, i.Quantity, i.Unit)
	} else if i.Quantity == -1 {
		// -1 indicates "some" quantity
		return fmt.Sprintf("@%s{}", i.Name)
	}
	return fmt.Sprintf("@%s{}", i.Name)
}

func (inst Instruction) Render() string {
	return inst.Text
}

func (t Timer) Render() string {
	if t.Name != "" {
		return fmt.Sprintf("~%s{%s}", t.Name, t.Duration)
	}
	return fmt.Sprintf("~{%s}", t.Duration)
}

func (c Cookware) Render() string {
	if c.Quantity > 1 {
		return fmt.Sprintf("#%s{%d}", c.Name, c.Quantity)
	}
	return fmt.Sprintf("#%s{}", c.Name)
}

type Ingredient struct {
	Name           string        `json:"name,omitempty"`
	Quantity       float32       `json:"quantity,omitempty"`
	Unit           string        `json:"unit,omitempty"`
	TypedUnit      *units.Unit   `json:"typed_unit,omitempty"`
	Subinstruction string        `json:"value,omitempty"`
	NextComponent  StepComponent `json:"next_component,omitempty"`
	CooklangRenderable
}

type Instruction struct {
	Text          string        `json:"text,omitempty"`
	NextComponent StepComponent `json:"next_component,omitempty"`
	CooklangRenderable
}

type Timer struct {
	Duration      string        `json:"duration,omitempty"`
	Name          string        `json:"name,omitempty"`
	Text          string        `json:"text,omitempty"`
	Unit          string        `json:"unit,omitempty"`
	NextComponent StepComponent `json:"next_component,omitempty"`
	CooklangRenderable
}

type Cookware struct {
	Name          string        `json:"name,omitempty"`
	Quantity      int           `json:"quantity,omitempty"`
	NextComponent StepComponent `json:"next_component,omitempty"`
	CooklangRenderable
}

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
	return ToCooklangRecipe(parsedRecipe), nil
}

func ParseBytes(content []byte) (*Recipe, error) {
	p := parser.New()
	parsedRecipe, err := p.ParseBytes(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(parsedRecipe), nil
}

func ParseString(content string) (*Recipe, error) {
	p := parser.New()
	recipe, err := p.ParseString(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(recipe), nil
}

// createTypedUnit attempts to find a unit in go-units or creates a new one if not found
func createTypedUnit(unitStr string) *units.Unit {
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

// ToCooklangRecipe converts a parser.Recipe to a cooklang.Recipe
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

	for stepIndex, step := range pRecipe.Steps {
		fmt.Println("Converting step:", stepIndex+1, "with components:", len(step.Components))

		newStep := &Step{}

		var prevComponent StepComponent

		for _, component := range step.Components {
			fmt.Printf("  Component: %#v\n", component)

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
					Name:      component.Name,
					Quantity:  quant,
					Unit:      component.Unit,
					TypedUnit: createTypedUnit(component.Unit),
				}
			case "cookware":
				cookwareQuant, err := strconv.Atoi(component.Quantity)
				if err != nil {
					cookwareQuant = 1 // Default to 1 if parsing fails
				}
				stepComp = &Cookware{
					Name:     component.Name,
					Quantity: cookwareQuant,
				}
			case "timer":
				stepComp = &Timer{
					Duration: component.Quantity,
					Name:     component.Name,
				}
			case "text":
				fmt.Printf("Adding text component: %s\n", component.Value)
				stepComp = &Instruction{
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

// ConvertTo converts the ingredient to a different unit if possible
func (i *Ingredient) ConvertTo(targetUnitStr string) (*Ingredient, error) {
	if i.TypedUnit == nil {
		return nil, fmt.Errorf("ingredient has no typed unit")
	}

	if i.Quantity == -1 {
		return nil, fmt.Errorf("cannot convert ingredients with 'some' quantity")
	}

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

// CanConvertTo checks if the ingredient can be converted to the target unit
func (i *Ingredient) CanConvertTo(targetUnitStr string) bool {
	if i.TypedUnit == nil {
		return false
	}

	if i.Quantity == -1 {
		return false // Can't convert "some" quantities
	}

	targetUnit, err := units.Find(targetUnitStr)
	if err != nil {
		// If unit not found, create a new one
		targetUnit = units.NewUnit(targetUnitStr, targetUnitStr)
	}

	_, err = units.ConvertFloat(float64(i.Quantity), *i.TypedUnit, targetUnit)
	return err == nil
}

// GetUnitType returns the unit's quantity type (e.g., "mass", "volume", "length") if available
func (i *Ingredient) GetUnitType() string {
	if i.TypedUnit == nil {
		return ""
	}

	// Check the unit's quantity type from the predefined quantity types
	switch {
	case i.TypedUnit.Quantity == "mass":
		return "mass"
	case i.TypedUnit.Quantity == "volume":
		return "volume"
	case i.TypedUnit.Quantity == "length":
		return "length"
	case i.TypedUnit.Quantity == "temperature":
		return "temperature"
	case i.TypedUnit.Quantity == "time":
		return "time"
	case i.TypedUnit.Quantity == "energy":
		return "energy"
	default:
		return i.TypedUnit.Quantity
	}
}

// SetNext and GetNext methods for Ingredient
func (i *Ingredient) SetNext(next StepComponent) {
	i.NextComponent = next
}

func (i *Ingredient) GetNext() StepComponent {
	return i.NextComponent
}

// SetNext and GetNext methods for Instruction
func (inst *Instruction) SetNext(next StepComponent) {
	inst.NextComponent = next
}

func (inst *Instruction) GetNext() StepComponent {
	return inst.NextComponent
}

// SetNext and GetNext methods for Timer
func (t *Timer) SetNext(next StepComponent) {
	t.NextComponent = next
}

func (t *Timer) GetNext() StepComponent {
	return t.NextComponent
}

// SetNext and GetNext methods for Cookware
func (c *Cookware) SetNext(next StepComponent) {
	c.NextComponent = next
}

func (c *Cookware) GetNext() StepComponent {
	return c.NextComponent
}

// IngredientList represents a collection of ingredients with unit consolidation capabilities
type IngredientList struct {
	Ingredients []*Ingredient
}

// NewIngredientList creates a new ingredient list
func NewIngredientList() *IngredientList {
	return &IngredientList{
		Ingredients: make([]*Ingredient, 0),
	}
}

// Add adds an ingredient to the list
func (il *IngredientList) Add(ingredient *Ingredient) {
	il.Ingredients = append(il.Ingredients, ingredient)
}

// GetIngredientsByName returns all ingredients with the given name
func (il *IngredientList) GetIngredientsByName(name string) []*Ingredient {
	var result []*Ingredient
	for _, ingredient := range il.Ingredients {
		if ingredient.Name == name {
			result = append(result, ingredient)
		}
	}
	return result
}

// ConsolidateByName consolidates ingredients with the same name, converting to a common unit when possible
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
			// Single ingredient, just add it
			consolidated.Add(ingredients[0])
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
			typedUnit = createTypedUnit(targetUnit)
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

// ToMap returns a map of ingredient names to their quantities (useful for shopping lists)
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

// GetIngredients returns all ingredients from a recipe
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
