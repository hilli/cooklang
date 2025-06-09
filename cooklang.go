package cooklang

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
func (Step) Render() string { return "" }

// func (Step)

type Ingredient struct {
	Name           string        `json:"name,omitempty"`
	Quantity       float32       `json:"quantity,omitempty"`
	Unit           string        `json:"unit,omitempty"`
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
	recpie, err := p.ParseBytes(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(recpie), nil
}

func ParseBytes(content []byte) (*Recipe, error) {
	p := parser.New()
	recpie, err := p.ParseBytes(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(recpie), nil
}

func ParseString(content string) (*Recipe, error) {
	p := parser.New()
	recipe, err := p.ParseString(content)
	if err != nil {
		return nil, err
	}
	return ToCooklangRecipe(recipe), nil
}

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

	fmt.Printf("Converting parsed recipe: %s\n", pRecipe.Metadata["title"])

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
				quant64, err := strconv.ParseFloat(component.Quantity, 32)
				quant := float32(quant64)
				if err != nil {
					quant = float32(0)
				}
				stepComp = &Ingredient{
					Name:     component.Name,
					Quantity: quant,
					Unit:     component.Unit,
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
