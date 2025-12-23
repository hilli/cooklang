package cooklang

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FrontmatterEditor provides CRUD operations for recipe frontmatter metadata.
// It allows reading, updating, and managing recipe metadata without manually parsing YAML.
//
// The editor works with the structured Recipe fields (title, cuisine, servings, etc.)
// as well as custom metadata fields, providing a unified interface for metadata management.
//
// Example:
//
//	editor, err := cooklang.NewFrontmatterEditor("recipe.cook")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	editor.SetMetadata("title", "Improved Pasta")
//	editor.SetMetadata("servings", "4")
//	editor.Save()
type FrontmatterEditor struct {
	filePath string
	content  string
	recipe   *Recipe
}

// NewFrontmatterEditor creates a new FrontmatterEditor for the given recipe file.
// It reads and parses the file, making the metadata available for manipulation.
//
// Parameters:
//   - filePath: Path to the .cook file to edit
//
// Returns:
//   - *FrontmatterEditor: An editor instance ready for metadata operations
//   - error: Any error encountered during file reading or parsing
//
// Example:
//
//	editor, err := cooklang.NewFrontmatterEditor("lasagna.cook")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	title, _ := editor.GetMetadata("title")
//	fmt.Println(title)
func NewFrontmatterEditor(filePath string) (*FrontmatterEditor, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	recipe, err := ParseBytes(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipe: %w", err)
	}

	return &FrontmatterEditor{
		filePath: filePath,
		content:  string(content),
		recipe:   recipe,
	}, nil
}

// GetMetadata retrieves a metadata value by key.
// It checks structured fields first (title, cuisine, etc.) then falls back to the generic metadata map.
//
// Parameters:
//   - key: The metadata key to retrieve
//
// Returns:
//   - string: The metadata value
//   - bool: true if the key exists, false otherwise
//
// Example:
//
//	editor, _ := cooklang.NewFrontmatterEditor("recipe.cook")
//	if title, ok := editor.GetMetadata("title"); ok {
//	    fmt.Printf("Recipe title: %s\n", title)
//	}
func (fe *FrontmatterEditor) GetMetadata(key string) (string, bool) {
	// Check structured fields first
	switch key {
	case "title":
		if fe.recipe.Title != "" {
			return fe.recipe.Title, true
		}
	case "cuisine":
		if fe.recipe.Cuisine != "" {
			return fe.recipe.Cuisine, true
		}
	case "description":
		if fe.recipe.Description != "" {
			return fe.recipe.Description, true
		}
	case "difficulty":
		if fe.recipe.Difficulty != "" {
			return fe.recipe.Difficulty, true
		}
	case "prep_time":
		if fe.recipe.PrepTime != "" {
			return fe.recipe.PrepTime, true
		}
	case "total_time":
		if fe.recipe.TotalTime != "" {
			return fe.recipe.TotalTime, true
		}
	case "author":
		if fe.recipe.Author != "" {
			return fe.recipe.Author, true
		}
	case "servings":
		if fe.recipe.Servings > 0 {
			return fmt.Sprintf("%g", fe.recipe.Servings), true
		}
	case "date":
		if !fe.recipe.Date.IsZero() {
			return fe.recipe.Date.Format("2006-01-02"), true
		}
	case "tags":
		if len(fe.recipe.Tags) > 0 {
			return strings.Join(fe.recipe.Tags, ", "), true
		}
	case "images", "image":
		if len(fe.recipe.Images) > 0 {
			return strings.Join(fe.recipe.Images, ", "), true
		}
	}

	// Check generic metadata map
	if val, ok := fe.recipe.Metadata[key]; ok {
		return val, true
	}

	return "", false
}

// GetAllMetadata returns all metadata as a map
func (fe *FrontmatterEditor) GetAllMetadata() map[string]string {
	result := make(map[string]string)

	// Add structured fields
	if fe.recipe.Title != "" {
		result["title"] = fe.recipe.Title
	}
	if fe.recipe.Cuisine != "" {
		result["cuisine"] = fe.recipe.Cuisine
	}
	if fe.recipe.Description != "" {
		result["description"] = fe.recipe.Description
	}
	if fe.recipe.Difficulty != "" {
		result["difficulty"] = fe.recipe.Difficulty
	}
	if fe.recipe.PrepTime != "" {
		result["prep_time"] = fe.recipe.PrepTime
	}
	if fe.recipe.TotalTime != "" {
		result["total_time"] = fe.recipe.TotalTime
	}
	if fe.recipe.Author != "" {
		result["author"] = fe.recipe.Author
	}
	if fe.recipe.Servings > 0 {
		result["servings"] = fmt.Sprintf("%g", fe.recipe.Servings)
	}
	if !fe.recipe.Date.IsZero() {
		result["date"] = fe.recipe.Date.Format("2006-01-02")
	}
	if len(fe.recipe.Tags) > 0 {
		result["tags"] = strings.Join(fe.recipe.Tags, ", ")
	}
	if len(fe.recipe.Images) > 0 {
		result["images"] = strings.Join(fe.recipe.Images, ", ")
	}

	// Add generic metadata
	for k, v := range fe.recipe.Metadata {
		result[k] = v
	}

	return result
}

// SetMetadata sets or updates a metadata value.
// For array fields (tags, images), the value can be comma-separated.
// For structured fields (servings, date), the value is validated and parsed.
//
// Parameters:
//   - key: The metadata key to set
//   - value: The value to set (format depends on the field type)
//
// Returns:
//   - error: Validation error for structured fields (e.g., invalid date format)
//
// Example:
//
//	editor, _ := cooklang.NewFrontmatterEditor("recipe.cook")
//	editor.SetMetadata("title", "Amazing Lasagna")
//	editor.SetMetadata("servings", "6")
//	editor.SetMetadata("tags", "italian, pasta, main course")
//	editor.SetMetadata("date", "2024-01-15")
//	editor.Save()
func (fe *FrontmatterEditor) SetMetadata(key, value string) error {
	// Update the recipe object
	switch key {
	case "title":
		fe.recipe.Title = value
		fe.recipe.Metadata[key] = value
	case "cuisine":
		fe.recipe.Cuisine = value
		fe.recipe.Metadata[key] = value
	case "description":
		fe.recipe.Description = value
		fe.recipe.Metadata[key] = value
	case "difficulty":
		fe.recipe.Difficulty = value
		fe.recipe.Metadata[key] = value
	case "prep_time":
		fe.recipe.PrepTime = value
		fe.recipe.Metadata[key] = value
	case "total_time":
		fe.recipe.TotalTime = value
		fe.recipe.Metadata[key] = value
	case "author":
		fe.recipe.Author = value
		fe.recipe.Metadata[key] = value
	case "servings":
		servings, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("invalid servings value: %w", err)
		}
		fe.recipe.Servings = float32(servings)
		fe.recipe.Metadata[key] = value
	case "date":
		date, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
		fe.recipe.Date = date
		fe.recipe.Metadata[key] = value
	case "tags":
		fe.recipe.Tags = splitAndTrim(value)
		fe.recipe.Metadata[key] = value
	case "images", "image":
		fe.recipe.Images = splitAndTrim(value)
		fe.recipe.Metadata["images"] = value
		if key == "image" {
			fe.recipe.Metadata["image"] = value
		}
	default:
		// Store in generic metadata
		fe.recipe.Metadata[key] = value
	}

	return nil
}

// DeleteMetadata removes a metadata key
func (fe *FrontmatterEditor) DeleteMetadata(key string) error {
	// Clear structured fields
	switch key {
	case "title":
		fe.recipe.Title = ""
		delete(fe.recipe.Metadata, key)
	case "cuisine":
		fe.recipe.Cuisine = ""
		delete(fe.recipe.Metadata, key)
	case "description":
		fe.recipe.Description = ""
		delete(fe.recipe.Metadata, key)
	case "difficulty":
		fe.recipe.Difficulty = ""
		delete(fe.recipe.Metadata, key)
	case "prep_time":
		fe.recipe.PrepTime = ""
		delete(fe.recipe.Metadata, key)
	case "total_time":
		fe.recipe.TotalTime = ""
		delete(fe.recipe.Metadata, key)
	case "author":
		fe.recipe.Author = ""
		delete(fe.recipe.Metadata, key)
	case "servings":
		fe.recipe.Servings = 0
		delete(fe.recipe.Metadata, key)
	case "date":
		fe.recipe.Date = time.Time{}
		delete(fe.recipe.Metadata, key)
	case "tags":
		fe.recipe.Tags = nil
		delete(fe.recipe.Metadata, key)
	case "images", "image":
		fe.recipe.Images = nil
		delete(fe.recipe.Metadata, key)
		delete(fe.recipe.Metadata, "images")
		delete(fe.recipe.Metadata, "image")
	default:
		// Remove from generic metadata
		delete(fe.recipe.Metadata, key)
	}

	return nil
}

// Save writes the updated recipe back to the original file.
// The recipe body (instructions) is preserved; only the frontmatter is updated.
//
// Returns:
//   - error: Any error encountered during file writing
//
// Example:
//
//	editor, _ := cooklang.NewFrontmatterEditor("recipe.cook")
//	editor.SetMetadata("title", "Updated Title")
//	if err := editor.Save(); err != nil {
//	    log.Fatal(err)
//	}
func (fe *FrontmatterEditor) Save() error {
	return fe.SaveAs(fe.filePath)
}

// SaveAs writes the updated recipe to a specified file
func (fe *FrontmatterEditor) SaveAs(filePath string) error {
	// Get the recipe content after frontmatter
	recipeBody := fe.extractRecipeBody()

	// Render the frontmatter
	frontmatter := fe.renderFrontmatter()

	// Combine frontmatter and body
	newContent := frontmatter + "\n" + recipeBody

	// Write to file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update internal state if saving to the same file
	if filePath == fe.filePath {
		fe.content = newContent
	}

	return nil
}

// GetContent returns the current file content
func (fe *FrontmatterEditor) GetContent() string {
	return fe.content
}

// GetUpdatedContent returns the updated content without saving to disk
func (fe *FrontmatterEditor) GetUpdatedContent() string {
	recipeBody := fe.extractRecipeBody()
	frontmatter := fe.renderFrontmatter()
	return frontmatter + "\n" + recipeBody
}

// extractRecipeBody extracts the recipe content after the frontmatter
func (fe *FrontmatterEditor) extractRecipeBody() string {
	// Match YAML frontmatter delimited by ---
	re := regexp.MustCompile(`(?s)^---\n.*?\n---\n`)
	body := re.ReplaceAllString(fe.content, "")

	// If no frontmatter found, return original content
	if body == fe.content {
		// Check if content starts with frontmatter
		if strings.HasPrefix(fe.content, "---\n") {
			// Find the end of frontmatter
			parts := strings.SplitN(fe.content, "\n---\n", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
		return fe.content
	}

	return strings.TrimLeft(body, "\n")
}

// renderFrontmatter renders the current recipe metadata as YAML frontmatter
func (fe *FrontmatterEditor) renderFrontmatter() string {
	var lines []string
	lines = append(lines, "---")

	// Add structured fields in a logical order
	if fe.recipe.Title != "" {
		lines = append(lines, renderYAMLValue("title", fe.recipe.Title)...)
	}
	if fe.recipe.Cuisine != "" {
		lines = append(lines, renderYAMLValue("cuisine", fe.recipe.Cuisine)...)
	}
	if fe.recipe.Description != "" {
		lines = append(lines, renderYAMLValue("description", fe.recipe.Description)...)
	}
	if fe.recipe.Difficulty != "" {
		lines = append(lines, renderYAMLValue("difficulty", fe.recipe.Difficulty)...)
	}
	if fe.recipe.Author != "" {
		lines = append(lines, renderYAMLValue("author", fe.recipe.Author)...)
	}
	if !fe.recipe.Date.IsZero() {
		lines = append(lines, fmt.Sprintf("date: %s", fe.recipe.Date.Format("2006-01-02")))
	}
	if fe.recipe.Servings > 0 {
		lines = append(lines, fmt.Sprintf("servings: %g", fe.recipe.Servings))
	}
	if fe.recipe.PrepTime != "" {
		lines = append(lines, renderYAMLValue("prep_time", fe.recipe.PrepTime)...)
	}
	if fe.recipe.TotalTime != "" {
		lines = append(lines, renderYAMLValue("total_time", fe.recipe.TotalTime)...)
	}
	if len(fe.recipe.Tags) > 0 {
		lines = append(lines, "tags:")
		for _, tag := range fe.recipe.Tags {
			lines = append(lines, fmt.Sprintf("  - %s", tag))
		}
	}
	if len(fe.recipe.Images) > 0 {
		lines = append(lines, "images:")
		for _, img := range fe.recipe.Images {
			lines = append(lines, fmt.Sprintf("  - %s", img))
		}
	}

	// Add generic metadata fields (alphabetically)
	keys := make([]string, 0, len(fe.recipe.Metadata))
	for k := range fe.recipe.Metadata {
		// Skip fields we've already added
		if !isStandardField(k) {
			keys = append(keys, k)
		}
	}

	// Simple sort
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		value := fe.recipe.Metadata[key]
		lines = append(lines, renderYAMLValue(key, value)...)
	}

	lines = append(lines, "---")
	return strings.Join(lines, "\n")
}

// renderYAMLValue renders a key-value pair, using block scalar syntax for multi-line values
func renderYAMLValue(key, value string) []string {
	// Check if value contains newlines
	if strings.Contains(value, "\n") {
		// Use literal block scalar with strip chomping (|-)
		lines := []string{fmt.Sprintf("%s: |-", key)}
		for _, line := range strings.Split(value, "\n") {
			lines = append(lines, "  "+line)
		}
		return lines
	}
	// Single-line value - use inline format
	return []string{fmt.Sprintf("%s: %s", key, value)}
}

// isStandardField checks if a key is a standard structured field
func isStandardField(key string) bool {
	standardFields := []string{
		"title", "cuisine", "description", "difficulty", "author",
		"date", "servings", "prep_time", "total_time", "tags", "images", "image",
	}
	for _, field := range standardFields {
		if field == key {
			return true
		}
	}
	return false
}

// splitAndTrim splits a comma-separated string and trims whitespace
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// AppendToArray appends a value to an array field (tags or images)
func (fe *FrontmatterEditor) AppendToArray(key, value string) error {
	switch key {
	case "tags":
		fe.recipe.Tags = append(fe.recipe.Tags, value)
	case "images", "image":
		fe.recipe.Images = append(fe.recipe.Images, value)
	default:
		return fmt.Errorf("field %s is not an array field", key)
	}
	return nil
}

// RemoveFromArray removes a value from an array field (tags or images)
func (fe *FrontmatterEditor) RemoveFromArray(key, value string) error {
	switch key {
	case "tags":
		fe.recipe.Tags = removeFromSlice(fe.recipe.Tags, value)
	case "images", "image":
		fe.recipe.Images = removeFromSlice(fe.recipe.Images, value)
	default:
		return fmt.Errorf("field %s is not an array field", key)
	}
	return nil
}

// removeFromSlice removes all occurrences of a value from a slice
func removeFromSlice(slice []string, value string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}
