package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/hilli/cooklang/lexer"
	"github.com/hilli/cooklang/token"
)

// Recipe represents a parsed cooklang recipe
type Recipe struct {
	Metadata map[string]string `json:"metadata"`
	Steps    []Step            `json:"steps"`
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

// CooklangParser handles parsing of cooklang recipes
type CooklangParser struct {
	Version int
}

// New creates a new CooklangParser
func New() *CooklangParser {
	return &CooklangParser{
		Version: 7,
	}
}

// ParseString parses a cooklang recipe from a string
func (p *CooklangParser) ParseString(input string) (*Recipe, error) {
	l := lexer.New(input)
	return p.parseTokens(l)
}

// ParseBytes parses a cooklang recipe from a byte slice
func (p *CooklangParser) ParseBytes(input []byte) (*Recipe, error) {
	return p.ParseString(string(input))
}

// ParseReader parses a cooklang recipe from an io.Reader
func (p *CooklangParser) ParseReader(reader io.Reader) (*Recipe, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	return p.ParseBytes(content)
}

// parseTokens handles the actual parsing logic
func (p *CooklangParser) parseTokens(l *lexer.Lexer) (*Recipe, error) {
	recipe := &Recipe{
		Metadata: make(map[string]string),
		Steps:    []Step{},
	}

	// Parse tokens and build recipe
	currentStep := Step{Components: []Component{}}

	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}

		switch tok.Type {
		case token.YAML_FRONTMATTER:
			// Parse YAML frontmatter into metadata
			metadata, err := p.parseYAMLMetadata(tok.Literal)
			if err != nil {
				return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
			}
			recipe.Metadata = metadata

		case token.INGREDIENT:
			// Parse ingredient
			ingredient, err := p.parseIngredient(l)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ingredient: %w", err)
			}
			currentStep.Components = append(currentStep.Components, ingredient)

		case token.COOKWARE:
			// Parse cookware
			cookware, err := p.parseCookware(l)
			if err != nil {
				return nil, fmt.Errorf("failed to parse cookware: %w", err)
			}
			currentStep.Components = append(currentStep.Components, cookware)

		case token.COOKTIME:
			// Parse timer
			timer, err := p.parseTimer(l)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timer: %w", err)
			}
			currentStep.Components = append(currentStep.Components, timer)

		case token.IDENT:
			// Regular text
			currentStep.Components = append(currentStep.Components, Component{
				Type:  "text",
				Value: tok.Literal,
			})

		default:
			// Other tokens like punctuation, numbers, etc.
			currentStep.Components = append(currentStep.Components, Component{
				Type:  "text",
				Value: tok.Literal,
			})
		}

		// Check if we need to start a new step (simplified logic)
		// In a real implementation, you'd want more sophisticated step detection
	}

	// Add the current step if it has components
	if len(currentStep.Components) > 0 {
		recipe.Steps = append(recipe.Steps, currentStep)
	}

	return recipe, nil
}

// parseYAMLMetadata parses YAML frontmatter into a metadata map
func (p *CooklangParser) parseYAMLMetadata(yamlContent string) (map[string]string, error) {
	metadata := make(map[string]string)

	lines := strings.Split(yamlContent, "\n")
	var currentKey string
	var listItems []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if this is a YAML list item (starts with -)
		if strings.HasPrefix(line, "-") {
			if currentKey != "" {
				// This is a list item for the current key
				item := strings.TrimSpace(line[1:]) // Remove the - and trim
				listItems = append(listItems, item)
				continue
			}
		} else if currentKey != "" && len(listItems) > 0 {
			// We were collecting list items, but this line doesn't start with -
			// so the list is complete
			metadata[currentKey] = strings.Join(listItems, ", ")
			currentKey = ""
			listItems = nil
		}

		// Simple key: value parsing (not full YAML)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Handle YAML array format: [ item1, item2, item3 ]
				if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
					// Remove brackets and parse array
					arrayContent := strings.TrimSpace(value[1 : len(value)-1])
					if arrayContent != "" {
						// Split by comma and clean up each item
						items := strings.Split(arrayContent, ",")
						var cleanItems []string
						for _, item := range items {
							cleanItem := strings.TrimSpace(item)
							cleanItems = append(cleanItems, cleanItem)
						}
						// Join back as comma-separated string for simple storage
						metadata[key] = strings.Join(cleanItems, ", ")
					} else {
						metadata[key] = ""
					}
				} else if value == "" {
					// This might be the start of a YAML list
					currentKey = key
					listItems = nil
				} else {
					metadata[key] = value
				}
			}
		}
	}

	// Handle case where list is at the end of the metadata
	if currentKey != "" && len(listItems) > 0 {
		metadata[currentKey] = strings.Join(listItems, ", ")
	}

	return metadata, nil
}

// parseIngredient parses an ingredient token and its quantity/units
func (p *CooklangParser) parseIngredient(l *lexer.Lexer) (Component, error) {
	component := Component{Type: "ingredient"}

	// Collect ingredient name (may be multiple words)
	var nameParts []string

	for {
		tok := l.NextToken()
		if tok.Type == token.IDENT {
			nameParts = append(nameParts, tok.Literal)
		} else if tok.Type == token.LBRACE {
			// Found opening brace, parse quantity/units
			quantity, units, err := p.parseQuantityAndUnits(l)
			if err != nil {
				return component, err
			}
			component.Quantity = quantity
			component.Units = units
			break
		} else {
			// No braces found, ingredient has no quantity
			// Put the token back by not consuming it
			break
		}
	}

	if len(nameParts) > 0 {
		component.Name = strings.Join(nameParts, " ")
	}

	return component, nil
}

// parseCookware parses a cookware token
func (p *CooklangParser) parseCookware(l *lexer.Lexer) (Component, error) {
	component := Component{Type: "cookware"}

	// Collect cookware name (may be multiple words)
	var nameParts []string

	for {
		tok := l.NextToken()
		if tok.Type == token.IDENT {
			nameParts = append(nameParts, tok.Literal)
		} else if tok.Type == token.LBRACE {
			// Found opening brace, parse quantity
			quantity, _, err := p.parseQuantityAndUnits(l)
			if err != nil {
				return component, err
			}
			component.Quantity = quantity
			break
		} else {
			// No braces found, cookware has no quantity
			break
		}
	}

	if len(nameParts) > 0 {
		component.Name = strings.Join(nameParts, " ")
	}

	return component, nil
}

// parseTimer parses a timer token
func (p *CooklangParser) parseTimer(l *lexer.Lexer) (Component, error) {
	component := Component{Type: "timer"}

	// Check if next token is an identifier (timer name) or brace (anonymous timer)
	tok := l.NextToken()
	if tok.Type == token.IDENT {
		component.Name = tok.Literal
		tok = l.NextToken()
	}

	// Check for quantity/units in braces
	if tok.Type == token.LBRACE {
		quantity, units, err := p.parseQuantityAndUnits(l)
		if err != nil {
			return component, err
		}
		component.Quantity = quantity
		component.Units = units
	}

	return component, nil
}

// parseQuantityAndUnits parses quantity and units from within braces
func (p *CooklangParser) parseQuantityAndUnits(l *lexer.Lexer) (string, string, error) {
	var quantityParts []string
	var units string
	var foundPercent bool

	for {
		tok := l.NextToken()
		if tok.Type == token.RBRACE {
			break
		}

		if tok.Type == token.EOF {
			return "", "", fmt.Errorf("unexpected EOF while parsing quantity/units")
		}

		if tok.Type == token.PERCENT {
			foundPercent = true
			continue
		}

		if foundPercent {
			// Everything after % is units
			if tok.Type == token.IDENT {
				if units == "" {
					units = tok.Literal
				} else {
					units += " " + tok.Literal
				}
			}
		} else {
			// Before % is quantity
			if tok.Type == token.INT || tok.Type == token.IDENT || tok.Type == token.DASH {
				quantityParts = append(quantityParts, tok.Literal)
			}
		}
	}

	// Join quantity parts
	quantity := strings.Join(quantityParts, "")
	if quantity == "" {
		quantity = "some"
	}

	return quantity, units, nil
}
