package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hilli/cooklang"
	"github.com/hilli/cooklang/lexer"
	"github.com/hilli/cooklang/token"
)

// type Metadata map[string]string

// // Recipe represents a parsed cooklang recipe
// type Recipe struct {
// 	Metadata Metadata `json:"metadata"`
// 	Steps    []Step   `json:"steps"`
// }

// // Step represents a cooking step with its components
// type Step struct {
// 	Components []Component `json:"components"`
// }

// // Component represents a component within a step
// type Component struct {
// 	Type     string `json:"type"` // "text", "ingredient", "cookware", "timer"
// 	Value    string `json:"value,omitempty"`
// 	Name     string `json:"name,omitempty"`
// 	Quantity string `json:"quantity,omitempty"`
// 	Units    string `json:"units,omitempty"`
// }

// CooklangParser handles parsing of cooklang recipes
type CooklangParser struct {
	CooklangSpecVersion int
}

// New creates a new CooklangParser
func New() *CooklangParser {
	return &CooklangParser{
		CooklangSpecVersion: 7,
	}
}

// ParseString parses a cooklang recipe from a string
func (p *CooklangParser) ParseString(input string) (*cooklang.Recipe, error) {
	l := lexer.New(input)
	return p.parseTokens(l)
}

// ParseBytes parses a cooklang recipe from a byte slice
func (p *CooklangParser) ParseBytes(input []byte) (*cooklang.Recipe, error) {
	return p.ParseString(string(input))
}

// ParseReader parses a cooklang recipe from an io.Reader
func (p *CooklangParser) ParseReader(reader io.Reader) (*cooklang.Recipe, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	return p.ParseBytes(content)
}

// parseTokens handles the actual parsing logic
func (p *CooklangParser) parseTokens(l *lexer.Lexer) (*cooklang.Recipe, error) {
	recipe := &cooklang.Recipe{
		Metadata: make(map[string]string),
		Steps:    []cooklang.Step{},
	}

	// Parse tokens and build recipe
	currentStep := cooklang.Step{Components: []cooklang.Component{}}

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

		case token.NEWLINE:
			// End current step and start a new one
			if len(currentStep.Components) > 0 {
				recipe.Steps = append(recipe.Steps, currentStep)
				currentStep = cooklang.Step{Components: []cooklang.Component{}}
			}

		case token.COMMENT:
			// Comments are ignored completely
			continue

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
			currentStep.Components = append(currentStep.Components, cooklang.Component{
				Type:  "text",
				Value: tok.Literal,
			})

		default:
			// Other tokens like punctuation, numbers, etc.
			currentStep.Components = append(currentStep.Components, cooklang.Component{
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

	// Compress consecutive text elements in all steps
	p.compressTextElements(recipe)

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
func (p *CooklangParser) parseIngredient(l *lexer.Lexer) (cooklang.Component, error) {
	component := cooklang.Component{Type: "ingredient"}

	// Collect IDENT and INT tokens and look for braces
	var nameTokens []token.Token

	// Collect all consecutive IDENT and INT tokens
	for {
		tok := l.NextToken()

		if tok.Type == token.IDENT || tok.Type == token.INT || tok.Type == token.DASH {
			nameTokens = append(nameTokens, tok)
		} else if tok.Type == token.LBRACE {
			// Found braces - all the tokens we collected are part of the name
			var nameParts []string
			for _, t := range nameTokens {
				nameParts = append(nameParts, t.Literal)
			}
			quantity, units, err := p.parseQuantityAndUnits(l)
			if err != nil {
				return component, err
			}
			component.Quantity = quantity
			component.Units = units
			component.Name = strings.Join(nameParts, " ")
			return component, nil
		} else {
			// Hit something that's not IDENT or LBRACE
			// Put this token back and stop
			l.PutBackToken(tok)
			break
		}
	}

	// No braces found - for ingredients without braces, only use the first token
	if len(nameTokens) > 0 {
		component.Name = nameTokens[0].Literal // Only use the first token for ingredients without braces
		component.Quantity = "some"            // Default quantity for ingredients

		// Put back any additional tokens that were collected (in reverse order)
		for i := len(nameTokens) - 1; i > 0; i-- {
			l.PutBackToken(nameTokens[i])
		}
	}

	return component, nil
}

// parseCookware parses a cookware token
func (p *CooklangParser) parseCookware(l *lexer.Lexer) (cooklang.Component, error) {
	component := cooklang.Component{Type: "cookware", Quantity: "1"}

	// Collect IDENT and INT tokens and look for braces
	var nameParts []string

	// Collect all consecutive IDENT and INT tokens
	for {
		tok := l.NextToken()

		if tok.Type == token.IDENT || tok.Type == token.INT || tok.Type == token.DASH {
			nameParts = append(nameParts, tok.Literal)
		} else if tok.Type == token.LBRACE {
			// Found braces - all the IDENTs we collected are part of the name
			quantity, _, err := p.parseQuantityAndUnits(l)
			if err != nil {
				return component, err
			}
			if quantity == "some" {
				quantity = "1"
			}
			component.Quantity = quantity
			component.Name = strings.Join(nameParts, " ")
			return component, nil
		} else {
			// Hit something that's not IDENT or LBRACE
			// Put this token back and stop
			l.PutBackToken(tok)
			break
		}
	}

	// No braces found - use only the first IDENT for single-word cookware
	if len(nameParts) > 0 {
		component.Name = nameParts[0] // Only first word without braces

		// Put back any extra IDENT tokens we consumed (in reverse order)
		for i := len(nameParts) - 1; i >= 1; i-- {
			// Reconstruct the IDENT tokens we need to put back
			l.PutBackToken(token.Token{Type: token.IDENT, Literal: nameParts[i]})
		}
	}

	return component, nil
}

// parseTimer parses a timer token
func (p *CooklangParser) parseTimer(l *lexer.Lexer) (cooklang.Component, error) {
	component := cooklang.Component{Type: "timer"}

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
			if tok.Type == token.INT || tok.Type == token.IDENT || tok.Type == token.DASH || tok.Type == token.DIVIDE || tok.Type == token.PERIOD {
				quantityParts = append(quantityParts, tok.Literal)
			}
		}
	}

	// Join quantity parts
	quantity := strings.Join(quantityParts, "")
	if quantity == "" {
		quantity = "some"
	} else {
		// Convert fractions to decimals
		quantity = p.evaluateFraction(quantity)
	}

	return quantity, units, nil
}

// evaluateFraction converts fraction strings like "1/2" to decimal representation "0.5"
// but preserves original format for fractions with leading zeros like "01/2"
func (p *CooklangParser) evaluateFraction(quantity string) string {
	// Check if this looks like a fraction (contains "/")
	if !strings.Contains(quantity, "/") {
		return quantity
	}

	// Split by "/" to get numerator and denominator
	parts := strings.Split(quantity, "/")
	if len(parts) != 2 {
		return quantity // Not a simple fraction, return as-is
	}

	// Check if either part has leading zeros - if so, preserve original format
	numeratorStr := strings.TrimSpace(parts[0])
	denominatorStr := strings.TrimSpace(parts[1])

	if (len(numeratorStr) > 1 && numeratorStr[0] == '0') ||
		(len(denominatorStr) > 1 && denominatorStr[0] == '0') {
		return quantity // Preserve fractions with leading zeros
	}

	// Parse numerator and denominator
	numerator, err1 := strconv.ParseFloat(numeratorStr, 64)
	denominator, err2 := strconv.ParseFloat(denominatorStr, 64)

	// If either part can't be parsed as a number, return original
	if err1 != nil || err2 != nil || denominator == 0 {
		return quantity
	}

	// Calculate the decimal result
	result := numerator / denominator

	// Format as string, removing unnecessary trailing zeros
	return strconv.FormatFloat(result, 'f', -1, 64)
}

// compressTextElements merges consecutive text components into single components
func (p *CooklangParser) compressTextElements(recipe *cooklang.Recipe) {
	for stepIndex := range recipe.Steps {
		step := &recipe.Steps[stepIndex]
		if len(step.Components) <= 1 {
			continue // No compression needed for steps with 0 or 1 components
		}

		var compressed []cooklang.Component
		var textBuffer []string

		for _, component := range step.Components {
			if component.Type == "text" {
				// Accumulate text components
				textBuffer = append(textBuffer, component.Value)
			} else {
				// Non-text component: flush any accumulated text first
				if len(textBuffer) > 0 {
					compressedText := strings.Join(textBuffer, " ")
					compressed = append(compressed, cooklang.Component{
						Type:  "text",
						Value: compressedText,
					})
					textBuffer = nil
				}
				// Add the non-text component
				compressed = append(compressed, component)
			}
		}

		// Flush any remaining text at the end
		if len(textBuffer) > 0 {
			compressedText := strings.Join(textBuffer, " ")
			compressed = append(compressed, cooklang.Component{
				Type:  "text",
				Value: compressedText,
			})
		}

		// Replace the step's components with the compressed version
		step.Components = compressed
	}
}
