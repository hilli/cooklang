package cooklang

// Setup types
type Metadata map[string]string

// Recipe represents a parsed cooklang recipe
type Recipe struct {
	Metadata Metadata `json:"metadata"`
	Steps    []Step   `json:"steps"`
}

// Step represents a cooking step with its components
type Step struct {
	Components []Component `json:"components"`
}

// Component represents a component within a step
type Component struct {
	Type     string `json:"type"` // "text", "ingredient", "cookware", "timer"
	Value    string `json:"value,omitempty"`
	Name     string `json:"name,omitempty"`
	Quantity string `json:"quantity,omitempty"`
	Units    string `json:"units,omitempty"`
}
