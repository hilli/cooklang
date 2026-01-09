package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hilli/cooklang/lexer"
	"github.com/hilli/cooklang/token"
)

type Metadata map[string]string

// Recipe represents a parsed cooklang recipe
type Recipe struct {
	Metadata Metadata `json:"metadata"`
	Steps    []Step   `json:"steps"`
}

// Step represents a cooking step with its components
type Step struct {
	Components []Component `json:"components" yaml:"steps"`
}

// Component represents a component within a step
type Component struct {
	Type     string `json:"type" yaml:"type"` // "text", "ingredient", "cookware", "timer"
	Value    string `json:"value,omitempty" yaml:"value,omitempty"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	Quantity string `json:"quantity,omitempty" yaml:"quantity,omitempty"`
	Unit     string `json:"unit,omitempty" yaml:"units,omitempty"`
	Fixed    bool   `json:"fixed,omitempty" yaml:"fixed,omitempty"` // Fixed quantity doesn't scale with servings
}

// CooklangParser handles parsing of cooklang recipes
type CooklangParser struct {
	CooklangSpecVersion int
	ExtendedMode        bool // Enable extended spec features
}

// New creates a new CooklangParser
func New() *CooklangParser {
	return &CooklangParser{
		CooklangSpecVersion: 7,
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

		case token.NEWLINE:
			// Handle newlines: single newline = space, double newline = new step
			nextTok := l.NextToken()
			if nextTok.Type == token.NEWLINE {
				// Double newline (blank line) - create new step
				if len(currentStep.Components) > 0 {
					recipe.Steps = append(recipe.Steps, currentStep)
					currentStep = Step{Components: []Component{}}
				}
			} else if nextTok.Type == token.EOF {
				// End of file after newline - don't add space, just break
				break
			} else {
				// Single newline - convert to space
				if len(currentStep.Components) > 0 {
					currentStep.Components = append(currentStep.Components, Component{
						Type:  "text",
						Value: " ",
					})
				}
				// Process the next token immediately here
				switch nextTok.Type {
				case token.INGREDIENT:
					ingredient, err := p.parseIngredient(l)
					if err != nil {
						return nil, fmt.Errorf("failed to parse ingredient: %w", err)
					}
					currentStep.Components = append(currentStep.Components, ingredient)
				case token.COOKWARE:
					cookware, err := p.parseCookware(l)
					if err != nil {
						return nil, fmt.Errorf("failed to parse cookware: %w", err)
					}
					currentStep.Components = append(currentStep.Components, cookware)
				case token.COOKTIME:
					timer, err := p.parseTimer(l)
					if err != nil {
						return nil, fmt.Errorf("failed to parse timer: %w", err)
					}
					currentStep.Components = append(currentStep.Components, timer)
				case token.WHITESPACE:
					currentStep.Components = append(currentStep.Components, Component{
						Type:  "text",
						Value: nextTok.Literal,
					})
				case token.IDENT:
					currentStep.Components = append(currentStep.Components, Component{
						Type:  "text",
						Value: nextTok.Literal,
					})
				case token.COMMENT:
					// Only create comment components in extended mode
					if p.ExtendedMode {
						currentStep.Components = append(currentStep.Components, Component{
							Type:  "comment",
							Value: nextTok.Literal,
						})
					}
					// In canonical mode, ignore comments
				case token.BLOCK_COMMENT:
					// Only create block comment components in extended mode
					if p.ExtendedMode {
						currentStep.Components = append(currentStep.Components, Component{
							Type:  "blockComment",
							Value: nextTok.Literal,
						})
					}
					// In canonical mode, ignore block comments
				case token.SECTION_HEADER:
					// Section headers start a new step
					if len(currentStep.Components) > 0 {
						recipe.Steps = append(recipe.Steps, currentStep)
						currentStep = Step{Components: []Component{}}
					}
					currentStep.Components = append(currentStep.Components, Component{
						Type: "section",
						Name: nextTok.Literal,
					})
				case token.NOTE:
					// Notes are standalone blocks that appear in recipe details but not during cooking
					// Notes always start a new step to keep them separate from cooking instructions
					if len(currentStep.Components) > 0 {
						recipe.Steps = append(recipe.Steps, currentStep)
						currentStep = Step{Components: []Component{}}
					}
					currentStep.Components = append(currentStep.Components, Component{
						Type:  "note",
						Value: nextTok.Literal,
					})
					// Add the note step immediately and start fresh for next content
					recipe.Steps = append(recipe.Steps, currentStep)
					currentStep = Step{Components: []Component{}}
				default:
					currentStep.Components = append(currentStep.Components, Component{
						Type:  "text",
						Value: nextTok.Literal,
					})
				}
			}

		case token.COMMENT:
			// Only create comment components in extended mode
			if p.ExtendedMode {
				currentStep.Components = append(currentStep.Components, Component{
					Type:  "comment",
					Value: tok.Literal,
				})
			}
			// In canonical mode, ignore comments

		case token.BLOCK_COMMENT:
			// Only create block comment components in extended mode
			if p.ExtendedMode {
				currentStep.Components = append(currentStep.Components, Component{
					Type:  "blockComment",
					Value: tok.Literal,
				})
			}
			// In canonical mode, ignore block comments

		case token.SECTION_HEADER:
			// Section headers start a new step (if current has content) and add section component
			// In canonical mode, sections are treated as step separators
			// In extended mode, sections create section components
			if len(currentStep.Components) > 0 {
				recipe.Steps = append(recipe.Steps, currentStep)
				currentStep = Step{Components: []Component{}}
			}
			// Add section as a component (in both modes for now, renderers can decide what to do)
			currentStep.Components = append(currentStep.Components, Component{
				Type: "section",
				Name: tok.Literal, // Section name
			})

		case token.NOTE:
			// Notes are standalone blocks that appear in recipe details but not during cooking
			// Notes always start a new step to keep them separate from cooking instructions
			if len(currentStep.Components) > 0 {
				recipe.Steps = append(recipe.Steps, currentStep)
				currentStep = Step{Components: []Component{}}
			}
			currentStep.Components = append(currentStep.Components, Component{
				Type:  "note",
				Value: tok.Literal,
			})
			// Add the note step immediately and start fresh for next content
			recipe.Steps = append(recipe.Steps, currentStep)
			currentStep = Step{Components: []Component{}}

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

		case token.WHITESPACE:
			// Handle whitespace as text component
			currentStep.Components = append(currentStep.Components, Component{
				Type:  "text",
				Value: tok.Literal,
			})

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

	// Compress consecutive text elements in all steps
	p.compressTextElements(recipe)

	return recipe, nil
}

// blockScalarType represents the type and chomping mode of a YAML block scalar
type blockScalarType struct {
	style    string // "literal" (|) or "folded" (>)
	chomping string // "strip" (-), "keep" (+), or "clip" (default)
}

// parseBlockScalarIndicator parses a block scalar indicator like |, |-, |+, >, >-, >+
// Returns the block scalar type and whether it's a valid block scalar indicator
func parseBlockScalarIndicator(value string) (blockScalarType, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return blockScalarType{}, false
	}

	var result blockScalarType

	// Check first character for style
	switch value[0] {
	case '|':
		result.style = "literal"
	case '>':
		result.style = "folded"
	default:
		return blockScalarType{}, false
	}

	// Check for chomping indicator
	if len(value) == 1 {
		result.chomping = "clip" // default
	} else {
		switch value[1] {
		case '-':
			result.chomping = "strip"
		case '+':
			result.chomping = "keep"
		default:
			// Could be an indentation indicator (digit) - treat as clip
			result.chomping = "clip"
		}
	}

	return result, true
}

// countLeadingSpaces returns the number of leading spaces in a line
func countLeadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			count++
		case '\t':
			count += 2 // Treat tab as 2 spaces
		default:
			return count
		}
	}
	return count
}

// isBlankOrWhitespace checks if a line is empty or contains only whitespace
func isBlankOrWhitespace(line string) bool {
	return strings.TrimSpace(line) == ""
}

// applyChomping applies the chomping rules to the block scalar content
func applyChomping(content string, chomping string) string {
	switch chomping {
	case "strip":
		// Remove all trailing newlines
		return strings.TrimRight(content, "\n")
	case "keep":
		// Keep all trailing newlines as-is
		return content
	default: // "clip"
		// Keep a single trailing newline
		content = strings.TrimRight(content, "\n")
		if content != "" {
			content += "\n"
		}
		return content
	}
}

// foldLines applies folded block scalar rules:
// - Single newlines become spaces
// - Multiple consecutive newlines preserve paragraph breaks
// - Leading whitespace on a line preserves the newline before it
func foldLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var result strings.Builder
	prevWasBlank := false

	for i, line := range lines {
		if line == "" {
			// Blank line - preserve it
			if i > 0 {
				result.WriteString("\n")
			}
			prevWasBlank = true
			continue
		}

		// Check if line starts with whitespace (literal block within folded)
		startsWithSpace := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')

		if i > 0 && !prevWasBlank {
			if startsWithSpace {
				// Preserve newline before indented content
				result.WriteString("\n")
			} else {
				// Fold: replace newline with space
				result.WriteString(" ")
			}
		} else if prevWasBlank && i > 0 {
			// After blank line, start new paragraph
			result.WriteString("\n")
		}

		result.WriteString(line)
		prevWasBlank = false
	}

	return result.String()
}

// parseYAMLMetadata parses YAML frontmatter into a metadata map
// Supports block scalars (|, |-, |+, >, >-, >+) for multi-line values
func (p *CooklangParser) parseYAMLMetadata(yamlContent string) (map[string]string, error) {
	metadata := make(map[string]string)

	lines := strings.Split(yamlContent, "\n")
	var currentKey string
	var listItems []string
	var blockScalar *blockScalarType
	var blockLines []string
	var blockBaseIndent int

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// If we're inside a block scalar, check if this line continues it
		if blockScalar != nil {
			// Check if line is indented (part of block) or not
			if isBlankOrWhitespace(line) {
				// Blank lines are part of the block
				blockLines = append(blockLines, "")
				continue
			}

			indent := countLeadingSpaces(line)

			// First content line establishes the base indentation
			if blockBaseIndent == 0 && indent > 0 {
				blockBaseIndent = indent
			}

			// If line is indented at least as much as base, it's part of the block
			if indent >= blockBaseIndent && blockBaseIndent > 0 {
				// Strip the base indentation
				blockLines = append(blockLines, line[blockBaseIndent:])
				continue
			}

			// Line is not indented enough - block scalar is complete
			var content string
			if blockScalar.style == "literal" {
				content = strings.Join(blockLines, "\n")
			} else {
				content = foldLines(blockLines)
			}
			content = applyChomping(content, blockScalar.chomping)
			metadata[currentKey] = content

			// Reset block scalar state
			blockScalar = nil
			blockLines = nil
			blockBaseIndent = 0
			currentKey = ""

			// Fall through to process this line normally
		}

		// Skip empty lines when not in block scalar
		if trimmedLine == "" {
			continue
		}

		// Check if this is a YAML list item (starts with -)
		if strings.HasPrefix(trimmedLine, "-") && !strings.HasPrefix(trimmedLine, "---") {
			if currentKey != "" && blockScalar == nil {
				// This is a list item for the current key
				item := strings.TrimSpace(trimmedLine[1:]) // Remove the - and trim
				listItems = append(listItems, item)
				continue
			}
		} else if currentKey != "" && len(listItems) > 0 && blockScalar == nil {
			// We were collecting list items, but this line doesn't start with -
			// so the list is complete
			metadata[currentKey] = strings.Join(listItems, ", ")
			currentKey = ""
			listItems = nil
		}

		// Simple key: value parsing
		if strings.Contains(trimmedLine, ":") {
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Check for block scalar indicator
				if bsType, isBlockScalar := parseBlockScalarIndicator(value); isBlockScalar {
					currentKey = key
					blockScalar = &bsType
					blockLines = []string{}
					blockBaseIndent = 0
					continue
				}

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

	// Handle case where block scalar is at the end of the metadata
	if blockScalar != nil && currentKey != "" {
		var content string
		if blockScalar.style == "literal" {
			content = strings.Join(blockLines, "\n")
		} else {
			content = foldLines(blockLines)
		}
		content = applyChomping(content, blockScalar.chomping)
		metadata[currentKey] = content
	}

	// Handle case where list is at the end of the metadata
	if currentKey != "" && len(listItems) > 0 && blockScalar == nil {
		metadata[currentKey] = strings.Join(listItems, ", ")
	}

	return metadata, nil
}

// parseIngredient parses an ingredient token and its quantity/unit
func (p *CooklangParser) parseIngredient(l *lexer.Lexer) (Component, error) {
	component := Component{Type: "ingredient"}

	// Collect IDENT and INT tokens and look for braces
	var nameTokens []token.Token

	// Collect all consecutive IDENT, INT, DASH, and WHITESPACE tokens
	for {
		tok := l.NextToken()

		if tok.Type == token.IDENT || tok.Type == token.INT || tok.Type == token.DASH || tok.Type == token.WHITESPACE {
			nameTokens = append(nameTokens, tok)
		} else if tok.Type == token.LBRACE {
			// Found braces - all the tokens we collected are part of the name
			var nameParts []string
			for _, t := range nameTokens {
				nameParts = append(nameParts, t.Literal)
			}
			quantity, unit, isFixed, err := p.parseQuantityAndUnit(l)
			if err != nil {
				return component, err
			}
			component.Quantity = quantity
			// Use the parsed unit in both canonical and extended modes
			component.Unit = unit
			component.Name = strings.Join(nameParts, "")
			component.Fixed = isFixed

			// Check for instruction in parentheses
			tok := l.NextToken()
			if tok.Type == token.LPAREN {
				// Parse instruction until closing parenthesis
				var instructionParts []string
				for {
					tok = l.NextToken()
					if tok.Type == token.RPAREN || tok.Type == token.EOF {
						break
					}
					instructionParts = append(instructionParts, tok.Literal)
				}
				component.Value = strings.Join(instructionParts, "")
			} else {
				// Put back the token we peeked at
				l.PutBackToken(tok)
			}

			return component, nil
		} else {
			// Hit something that's not IDENT or LBRACE
			// Put this token back and stop
			l.PutBackToken(tok)
			break
		}
	}

	// No braces found - for ingredients without braces, collect consecutive alphanumeric tokens
	if len(nameTokens) > 0 {
		// For ingredients without braces, join consecutive IDENT/INT tokens
		var nameParts []string
		var tokensUsed int

		for i, tok := range nameTokens {
			if tok.Type == token.IDENT || tok.Type == token.INT {
				nameParts = append(nameParts, tok.Literal)
				tokensUsed = i + 1
			} else {
				// Stop at first non-alphanumeric token
				break
			}
		}

		component.Name = strings.Join(nameParts, "")
		component.Quantity = "some" // Default quantity for ingredients

		// Put back any tokens we didn't use (in reverse order)
		for i := len(nameTokens) - 1; i >= tokensUsed; i-- {
			l.PutBackToken(nameTokens[i])
		}

		// Check for instruction in parentheses
		tok := l.NextToken()
		if tok.Type == token.LPAREN {
			// Parse instruction until closing parenthesis
			var instructionParts []string
			for {
				tok = l.NextToken()
				if tok.Type == token.RPAREN || tok.Type == token.EOF {
					break
				}
				instructionParts = append(instructionParts, tok.Literal)
			}
			component.Value = strings.Join(instructionParts, "")
		} else {
			// Put back the token we peeked at
			l.PutBackToken(tok)
		}
	}

	return component, nil
}

// parseCookware parses a cookware token
func (p *CooklangParser) parseCookware(l *lexer.Lexer) (Component, error) {
	component := Component{Type: "cookware", Quantity: "1"} // Always default to "1"

	// Collect IDENT, INT, DASH, WHITESPACE, and other valid name tokens and look for braces
	var nameTokens []token.Token

	// Collect all consecutive valid name tokens (everything except reserved characters)
	for {
		tok := l.NextToken()

		// Accept most tokens as part of the name, stop only at braces, parens, or newlines
		if tok.Type == token.IDENT || tok.Type == token.INT || tok.Type == token.DASH ||
			tok.Type == token.WHITESPACE || tok.Type == token.PERIOD || tok.Type == token.COMMA ||
			tok.Type == token.ILLEGAL {
			nameTokens = append(nameTokens, tok)
		} else if tok.Type == token.LBRACE {
			// Found braces - all the tokens we collected are part of the name
			var nameParts []string
			for _, t := range nameTokens {
				nameParts = append(nameParts, t.Literal)
			}
			quantity, _, _, err := p.parseQuantityAndUnit(l) // isFixed ignored - cookware doesn't scale
			if err != nil {
				return component, err
			}
			if quantity == "some" {
				quantity = "1"
			}
			component.Quantity = quantity // Always set quantity for cookware
			component.Name = strings.Join(nameParts, "")

			// Check for instruction in parentheses
			tok := l.NextToken()
			if tok.Type == token.LPAREN {
				// Parse instruction until closing parenthesis
				var instructionParts []string
				for {
					tok = l.NextToken()
					if tok.Type == token.RPAREN || tok.Type == token.EOF {
						break
					}
					instructionParts = append(instructionParts, tok.Literal)
				}
				component.Value = strings.Join(instructionParts, "")
			} else {
				// Put back the token we peeked at
				l.PutBackToken(tok)
			}

			return component, nil
		} else {
			// Hit something that's not IDENT or LBRACE
			// Put this token back and stop
			l.PutBackToken(tok)
			break
		}
	}

	// No braces found - use only the first IDENT for single-word cookware
	if len(nameTokens) > 0 {
		component.Name = nameTokens[0].Literal // Only first word without braces

		// Put back any extra tokens we consumed (in reverse order)
		for i := len(nameTokens) - 1; i >= 1; i-- {
			// Put back the tokens we need to return
			l.PutBackToken(nameTokens[i])
		}

		// Check for instruction in parentheses
		tok := l.NextToken()
		if tok.Type == token.LPAREN {
			// Parse instruction until closing parenthesis
			var instructionParts []string
			for {
				tok = l.NextToken()
				if tok.Type == token.RPAREN || tok.Type == token.EOF {
					break
				}
				instructionParts = append(instructionParts, tok.Literal)
			}
			component.Value = strings.Join(instructionParts, "")
		} else {
			// Put back the token we peeked at
			l.PutBackToken(tok)
		}
	}

	return component, nil
}

// parseTimer parses a timer token
func (p *CooklangParser) parseTimer(l *lexer.Lexer) (Component, error) {
	component := Component{Type: "timer"}

	// Check if next token is an identifier (timer name) or brace (anonymous timer)
	tok := l.NextToken()
	switch tok.Type {
	case token.IDENT:
		if p.ExtendedMode {
			// Extended mode: allow multi-word timer names
			var nameTokens []string
			nameTokens = append(nameTokens, tok.Literal)

			for {
				nextTok := l.NextToken()
				switch nextTok.Type {
				case token.LBRACE:
					// Found braces - parse quantity/unit
					component.Name = strings.Join(nameTokens, "")
					quantity, unit, _, err := p.parseQuantityAndUnit(l) // isFixed ignored - timers don't scale
					if err != nil {
						return component, err
					}
					// Use Quantity and Unit fields in both modes
					component.Quantity = quantity
					component.Unit = unit
					return component, nil
				case token.WHITESPACE:
					// Add whitespace to name parts
					nameTokens = append(nameTokens, nextTok.Literal)
				case token.IDENT:
					// Additional word in timer name
					nameTokens = append(nameTokens, nextTok.Literal)
				default:
					// Hit something else - put it back and stop
					l.PutBackToken(nextTok)
					component.Name = strings.TrimSpace(strings.Join(nameTokens, ""))
				}
			}
		} else {
			// Canonical mode: single word timer names only
			component.Name = tok.Literal

			// Check if next token is braces
			nextTok := l.NextToken()
			if nextTok.Type == token.LBRACE {
				// Parse quantity/unit
				quantity, unit, _, err := p.parseQuantityAndUnit(l) // isFixed ignored - timers don't scale
				if err != nil {
					return component, err
				}
				// In canonical mode, use Quantity and Unit fields
				component.Quantity = quantity
				component.Unit = unit
			} else {
				// Put the token back
				l.PutBackToken(nextTok)
			}
		}
	case token.LBRACE:
		// Anonymous timer - parse quantity/unit directly
		quantity, unit, _, err := p.parseQuantityAndUnit(l) // isFixed ignored - timers don't scale
		if err != nil {
			return component, err
		}
		// Use Quantity and Unit fields in both modes
		component.Quantity = quantity
		component.Unit = unit
	default:
		// Put the token back if it's neither IDENT nor LBRACE
		l.PutBackToken(tok)
	}

	// Check for annotation in parentheses (extended mode feature)
	tok = l.NextToken()
	if tok.Type == token.LPAREN {
		// Parse annotation until closing parenthesis
		var annotationParts []string
		for {
			tok = l.NextToken()
			if tok.Type == token.RPAREN || tok.Type == token.EOF {
				break
			}
			annotationParts = append(annotationParts, tok.Literal)
		}
		component.Value = strings.Join(annotationParts, "")
	} else {
		// Put back the token we peeked at
		l.PutBackToken(tok)
	}

	return component, nil
}

// parseQuantityAndUnit parses quantity and units from within braces
// Returns quantity, unit, isFixed (true if quantity has = prefix), and error
func (p *CooklangParser) parseQuantityAndUnit(l *lexer.Lexer) (string, string, bool, error) {
	var quantityParts []string
	var unit string
	var foundPercent bool
	var isFixed bool

	// Check for leading = which indicates fixed quantity (doesn't scale with servings)
	tok := l.NextToken()
	if tok.Type == token.SECTION { // SECTION token is "="
		isFixed = true
		tok = l.NextToken() // consume the =, get next token
	}

	// Process tokens starting with the current one
	for tok.Type != token.RBRACE {
		if tok.Type == token.EOF {
			return "", "", false, fmt.Errorf("unexpected EOF while parsing quantity/unit")
		}

		if tok.Type == token.PERCENT {
			foundPercent = true
			tok = l.NextToken()
			continue
		}

		if foundPercent {
			// Everything after % is unit
			if tok.Type == token.IDENT || tok.Type == token.WHITESPACE {
				if unit == "" {
					unit = tok.Literal
				} else {
					unit += tok.Literal
				}
			}
		} else {
			// Before % is quantity
			if tok.Type == token.INT || tok.Type == token.IDENT || tok.Type == token.DASH || tok.Type == token.DIVIDE || tok.Type == token.PERIOD || tok.Type == token.WHITESPACE {
				quantityParts = append(quantityParts, tok.Literal)
			}
		}
		tok = l.NextToken()
	}

	// Join quantity parts preserving original spacing
	quantity := strings.Join(quantityParts, "")

	// Trim whitespace from quantity, but preserve internal spaces
	quantity = strings.TrimSpace(quantity)

	if quantity == "" {
		quantity = "some"
	} else {
		// Convert fractions to decimals
		quantity = p.evaluateFraction(quantity)
	}

	// Trim whitespace from units
	unit = strings.TrimSpace(unit)

	// Don't set default units - spec expects empty string when no units provided
	return quantity, unit, isFixed, nil
}

// evaluateFraction converts fraction strings like "1/2" to decimal representation "0.5"
// and mixed fractions like "1 1/2" to "1.5"
// and Unicode fractions like "½" to "0.5" and "1½" to "1.5"
// but preserves original format for fractions with leading zeros like "01/2"
func (p *CooklangParser) evaluateFraction(quantity string) string {
	// Trim the string to handle any leading/trailing spaces
	quantity = strings.TrimSpace(quantity)

	// First, try to convert Unicode fractions to decimal
	converted := p.convertUnicodeFractions(quantity)
	if converted != quantity {
		// Unicode fractions were converted
		return converted
	}

	// Check if this looks like a fraction (contains "/")
	if !strings.Contains(quantity, "/") {
		return quantity
	}

	// First, try to parse as a simple fraction (potentially with spaces around the /)
	parts := strings.Split(quantity, "/")
	if len(parts) == 2 {
		numeratorStr := strings.TrimSpace(parts[0])
		denominatorStr := strings.TrimSpace(parts[1])

		// Check if either part has leading zeros - if so, preserve original format
		if (len(numeratorStr) > 1 && numeratorStr[0] == '0') ||
			(len(denominatorStr) > 1 && denominatorStr[0] == '0') {
			return quantity // Preserve fractions with leading zeros
		}

		// Try to parse the numerator - if it contains spaces, it might be a mixed fraction
		numeratorParts := strings.Fields(numeratorStr) // Split by whitespace
		if len(numeratorParts) == 2 {
			// This looks like a mixed fraction: "1 1/2" becomes numeratorParts=["1", "1"]
			wholeNumber, err1 := strconv.ParseFloat(numeratorParts[0], 64)
			numerator, err2 := strconv.ParseFloat(numeratorParts[1], 64)
			denominator, err3 := strconv.ParseFloat(denominatorStr, 64)

			if err1 == nil && err2 == nil && err3 == nil && denominator != 0 {
				// It's a mixed fraction
				result := wholeNumber + (numerator / denominator)
				return strconv.FormatFloat(result, 'f', -1, 64)
			}
		}

		// Try as simple fraction
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

	// Not a simple fraction pattern, return as-is
	return quantity
}

// convertUnicodeFractions converts Unicode fraction characters to decimal
// Supports both simple fractions (½) and mixed fractions (1½)
func (p *CooklangParser) convertUnicodeFractions(quantity string) string {
	// Map of Unicode fraction characters to their decimal values
	unicodeFractions := map[rune]float64{
		'½': 0.5,     // VULGAR FRACTION ONE HALF
		'¼': 0.25,    // VULGAR FRACTION ONE QUARTER
		'¾': 0.75,    // VULGAR FRACTION THREE QUARTERS
		'⅓': 1.0 / 3, // VULGAR FRACTION ONE THIRD
		'⅔': 2.0 / 3, // VULGAR FRACTION TWO THIRDS
		'⅕': 0.2,     // VULGAR FRACTION ONE FIFTH
		'⅖': 0.4,     // VULGAR FRACTION TWO FIFTHS
		'⅗': 0.6,     // VULGAR FRACTION THREE FIFTHS
		'⅘': 0.8,     // VULGAR FRACTION FOUR FIFTHS
		'⅙': 1.0 / 6, // VULGAR FRACTION ONE SIXTH
		'⅚': 5.0 / 6, // VULGAR FRACTION FIVE SIXTHS
		'⅐': 1.0 / 7, // VULGAR FRACTION ONE SEVENTH
		'⅛': 0.125,   // VULGAR FRACTION ONE EIGHTH
		'⅜': 0.375,   // VULGAR FRACTION THREE EIGHTHS
		'⅝': 0.625,   // VULGAR FRACTION FIVE EIGHTHS
		'⅞': 0.875,   // VULGAR FRACTION SEVEN EIGHTHS
		'⅑': 1.0 / 9, // VULGAR FRACTION ONE NINTH
		'⅒': 0.1,     // VULGAR FRACTION ONE TENTH
	}

	// Check if the string contains any Unicode fractions
	hasFraction := false
	for _, r := range quantity {
		if _, ok := unicodeFractions[r]; ok {
			hasFraction = true
			break
		}
	}

	if !hasFraction {
		return quantity // No Unicode fractions found
	}

	// Parse the quantity to handle mixed fractions like "1½"
	var wholeNumber float64
	var fractionValue float64
	var foundWhole bool

	// Scan through the string
	runes := []rune(quantity)
	numberStart := -1

	for i, r := range runes {
		if r >= '0' && r <= '9' {
			// Start of a number
			if numberStart == -1 {
				numberStart = i
			}
		} else if val, ok := unicodeFractions[r]; ok {
			// Found a Unicode fraction
			if numberStart != -1 {
				// Parse the whole number part before the fraction
				wholeStr := string(runes[numberStart:i])
				if num, err := strconv.ParseFloat(wholeStr, 64); err == nil {
					wholeNumber = num
					foundWhole = true
				}
				numberStart = -1
			}
			fractionValue = val
		} else if r != ' ' && r != '\t' {
			// Non-numeric, non-fraction, non-whitespace character
			// This means it's not a pure numeric fraction string
			return quantity
		}
	}

	// If we found a fraction character, convert to decimal
	if fractionValue > 0 {
		result := wholeNumber + fractionValue
		// If there was no whole number part, just return the fraction
		if !foundWhole && wholeNumber == 0 {
			return strconv.FormatFloat(fractionValue, 'f', -1, 64)
		}
		return strconv.FormatFloat(result, 'f', -1, 64)
	}

	return quantity
}

// compressTextElements merges consecutive text components into single components
func (p *CooklangParser) compressTextElements(recipe *Recipe) {
	for stepIndex := range recipe.Steps {
		step := &recipe.Steps[stepIndex]
		if len(step.Components) <= 1 {
			continue // No compression needed for steps with 0 or 1 components
		}

		var compressed []Component
		var textBuffer []string

		for _, component := range step.Components {
			if component.Type == "text" {
				// Accumulate text components without adding spaces
				textBuffer = append(textBuffer, component.Value)
			} else {
				// Non-text component: flush any accumulated text first
				if len(textBuffer) > 0 {
					compressedText := strings.Join(textBuffer, "")
					compressed = append(compressed, Component{
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
			compressedText := strings.Join(textBuffer, "")
			compressed = append(compressed, Component{
				Type:  "text",
				Value: compressedText,
			})
		}

		// Replace the step's components with the compressed version
		step.Components = compressed
	}
}
