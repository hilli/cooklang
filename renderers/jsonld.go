// Package renderers provides different renderers for Cooklang recipes.
package renderers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hilli/cooklang"
)

// JSONLDRenderer renders recipes as Schema.org Recipe JSON-LD structured data.
// JSON-LD (JavaScript Object Notation for Linked Data) is the recommended format
// by Google for embedding structured data in web pages to enable rich search results.
//
// The renderer produces output conforming to the Schema.org Recipe specification:
// https://schema.org/Recipe
//
// Example usage:
//
//	recipe, _ := cooklang.ParseFile("margarita.cook")
//	renderer := renderers.JSONLDRenderer{}
//
//	// Get JSON-LD as a map for further manipulation
//	data := renderer.RenderRecipe(recipe, nil)
//
//	// Get JSON-LD as a formatted JSON string
//	jsonStr, _ := renderer.RenderRecipeJSON(recipe, nil)
//
//	// Get a complete <script> tag ready for HTML embedding
//	scriptTag, _ := renderer.RenderRecipeScriptTag(recipe, &renderers.JSONLDOptions{
//	    URL: "https://example.com/recipes/margarita",
//	    AggregateRating: &renderers.AggregateRating{
//	        RatingValue: 4.5,
//	        RatingCount: 42,
//	    },
//	})
type JSONLDRenderer struct{}

// JSONLDOptions allows customization of the JSON-LD output with application-specific data
// that may not be available in the cooklang Recipe itself.
//
// All fields are optional. When nil or zero-valued, the corresponding JSON-LD property
// is either omitted or derived from the cooklang.Recipe fields.
type JSONLDOptions struct {
	// URL is the canonical URL of the recipe page.
	// Maps to Schema.org "url" property.
	URL string

	// DateModified is the last modification date of the recipe.
	// Maps to Schema.org "dateModified" property.
	DateModified *time.Time

	// DatePublished overrides the recipe.Date field.
	// Maps to Schema.org "datePublished" property.
	DatePublished *time.Time

	// Images provides full URLs for recipe images, overriding or supplementing
	// the recipe.Images field which typically contains relative paths.
	// Maps to Schema.org "image" property.
	Images []string

	// AggregateRating provides rating data from your application.
	// Maps to Schema.org "aggregateRating" property.
	AggregateRating *AggregateRating

	// Video provides video information for the recipe.
	// Maps to Schema.org "video" property.
	Video *VideoObject

	// Keywords provides additional keywords beyond the recipe tags.
	// These are merged with recipe.Tags.
	// Maps to Schema.org "keywords" property.
	Keywords []string

	// RecipeCategory overrides or supplements the recipe category.
	// Examples: "Cocktail", "Appetizer", "Main course", "Dessert"
	// Maps to Schema.org "recipeCategory" property.
	RecipeCategory string

	// AuthorURL provides a URL for the author's profile page.
	// When set, the author is rendered as a Person object with a URL.
	AuthorURL string
}

// AggregateRating represents the aggregate rating of a recipe based on multiple user reviews.
// This is used to display star ratings in Google search results.
//
// Example:
//
//	rating := &renderers.AggregateRating{
//	    RatingValue: 4.5,
//	    RatingCount: 42,
//	}
type AggregateRating struct {
	// RatingValue is the average rating value.
	// Required for the rating to be included.
	RatingValue float64

	// RatingCount is the total number of ratings.
	// Required for the rating to be included.
	RatingCount int

	// BestRating is the highest possible rating value.
	// Defaults to 5 if not specified.
	BestRating float64

	// WorstRating is the lowest possible rating value.
	// Defaults to 1 if not specified.
	WorstRating float64
}

// VideoObject represents a video associated with the recipe.
// This enables video rich results in search engines.
type VideoObject struct {
	// Name is the title of the video.
	Name string

	// Description is a description of the video.
	Description string

	// ThumbnailURL is the URL to a thumbnail image for the video.
	ThumbnailURL string

	// ContentURL is the URL to the actual video file.
	ContentURL string

	// EmbedURL is the URL to an embeddable player for the video.
	EmbedURL string

	// UploadDate is the date the video was uploaded.
	UploadDate *time.Time

	// Duration is the video duration in ISO 8601 format (e.g., "PT1M30S").
	Duration string
}

// RenderRecipe returns a JSON-LD object as a map for flexible manipulation.
// This is useful when you need to modify the output before serialization.
//
// The returned map follows the Schema.org Recipe specification with these properties:
//   - @context: Always "https://schema.org"
//   - @type: Always "Recipe"
//   - name: From recipe.Title
//   - description: From recipe.Description
//   - author: From recipe.Author (as Person object)
//   - image: From opts.Images or recipe.Images
//   - recipeYield: From recipe.Servings
//   - prepTime: From recipe.PrepTime (converted to ISO 8601 duration)
//   - totalTime: From recipe.TotalTime (converted to ISO 8601 duration)
//   - recipeCategory: From opts.RecipeCategory or recipe.Metadata["category"]
//   - recipeCuisine: From recipe.Cuisine
//   - keywords: From recipe.Tags merged with opts.Keywords
//   - recipeIngredient: Array of ingredient strings
//   - recipeInstructions: Array of HowToStep objects
//   - tool: Array of cookware/tool names
//   - datePublished: From opts.DatePublished or recipe.Date
//   - dateModified: From opts.DateModified
//   - url: From opts.URL
//   - aggregateRating: From opts.AggregateRating
//   - video: From opts.Video
//
// Parameters:
//   - recipe: The parsed Cooklang recipe to render
//   - opts: Optional configuration (can be nil for defaults)
//
// Returns:
//   - map[string]interface{}: The JSON-LD object as a map
func (jr JSONLDRenderer) RenderRecipe(recipe *cooklang.Recipe, opts *JSONLDOptions) map[string]interface{} {
	if opts == nil {
		opts = &JSONLDOptions{}
	}

	data := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "Recipe",
	}

	// Name (required)
	if recipe.Title != "" {
		data["name"] = recipe.Title
	}

	// Description
	if recipe.Description != "" {
		data["description"] = recipe.Description
	}

	// Author
	if recipe.Author != "" {
		author := map[string]interface{}{
			"@type": "Person",
			"name":  recipe.Author,
		}
		if opts.AuthorURL != "" {
			author["url"] = opts.AuthorURL
		}
		data["author"] = author
	}

	// Images
	images := opts.Images
	if len(images) == 0 && len(recipe.Images) > 0 {
		images = recipe.Images
	}
	if len(images) > 0 {
		if len(images) == 1 {
			data["image"] = images[0]
		} else {
			data["image"] = images
		}
	}

	// Recipe Yield (servings)
	if recipe.Servings > 0 {
		if recipe.Servings == float32(int(recipe.Servings)) {
			data["recipeYield"] = fmt.Sprintf("%d serving", int(recipe.Servings))
			if recipe.Servings != 1 {
				data["recipeYield"] = fmt.Sprintf("%d servings", int(recipe.Servings))
			}
		} else {
			data["recipeYield"] = fmt.Sprintf("%.1f servings", recipe.Servings)
		}
	}

	// Prep Time (ISO 8601 duration)
	if recipe.PrepTime != "" {
		if duration := ParseDurationToISO8601(recipe.PrepTime); duration != "" {
			data["prepTime"] = duration
		}
	}

	// Total Time (ISO 8601 duration)
	if recipe.TotalTime != "" {
		if duration := ParseDurationToISO8601(recipe.TotalTime); duration != "" {
			data["totalTime"] = duration
		}
	}

	// Recipe Category
	category := opts.RecipeCategory
	if category == "" {
		if cat, ok := recipe.Metadata["category"]; ok {
			category = cat
		}
	}
	if category != "" {
		data["recipeCategory"] = category
	}

	// Recipe Cuisine
	if recipe.Cuisine != "" {
		data["recipeCuisine"] = recipe.Cuisine
	}

	// Keywords (merge tags and additional keywords)
	keywords := make([]string, 0, len(recipe.Tags)+len(opts.Keywords))
	keywords = append(keywords, recipe.Tags...)
	keywords = append(keywords, opts.Keywords...)
	if len(keywords) > 0 {
		data["keywords"] = strings.Join(keywords, ", ")
	}

	// Recipe Ingredients
	ingredients := recipe.GetIngredients()
	if len(ingredients.Ingredients) > 0 {
		ingredientStrings := make([]string, 0, len(ingredients.Ingredients))
		for _, ing := range ingredients.Ingredients {
			ingredientStrings = append(ingredientStrings, ing.RenderDisplay())
		}
		data["recipeIngredient"] = ingredientStrings
	}

	// Recipe Instructions (as HowToStep or HowToSection)
	instructions := jr.buildInstructions(recipe)
	if len(instructions) > 0 {
		data["recipeInstructions"] = instructions
	}

	// Tools (cookware)
	cookware := recipe.GetCookware()
	if len(cookware) > 0 {
		// Deduplicate cookware names
		seen := make(map[string]bool)
		tools := make([]string, 0, len(cookware))
		for _, cw := range cookware {
			if !seen[cw.Name] {
				seen[cw.Name] = true
				tools = append(tools, cw.Name)
			}
		}
		data["tool"] = tools
	}

	// Date Published
	var datePublished time.Time
	if opts.DatePublished != nil {
		datePublished = *opts.DatePublished
	} else if !recipe.Date.IsZero() {
		datePublished = recipe.Date
	}
	if !datePublished.IsZero() {
		data["datePublished"] = datePublished.Format("2006-01-02")
	}

	// Date Modified
	if opts.DateModified != nil && !opts.DateModified.IsZero() {
		data["dateModified"] = opts.DateModified.Format("2006-01-02")
	}

	// URL
	if opts.URL != "" {
		data["url"] = opts.URL
	}

	// Aggregate Rating
	if opts.AggregateRating != nil && opts.AggregateRating.RatingCount > 0 {
		rating := map[string]interface{}{
			"@type":       "AggregateRating",
			"ratingValue": opts.AggregateRating.RatingValue,
			"ratingCount": opts.AggregateRating.RatingCount,
		}
		bestRating := opts.AggregateRating.BestRating
		if bestRating == 0 {
			bestRating = 5
		}
		rating["bestRating"] = bestRating

		worstRating := opts.AggregateRating.WorstRating
		if worstRating == 0 {
			worstRating = 1
		}
		rating["worstRating"] = worstRating

		data["aggregateRating"] = rating
	}

	// Video
	if opts.Video != nil && opts.Video.Name != "" {
		video := map[string]interface{}{
			"@type": "VideoObject",
			"name":  opts.Video.Name,
		}
		if opts.Video.Description != "" {
			video["description"] = opts.Video.Description
		}
		if opts.Video.ThumbnailURL != "" {
			video["thumbnailUrl"] = opts.Video.ThumbnailURL
		}
		if opts.Video.ContentURL != "" {
			video["contentUrl"] = opts.Video.ContentURL
		}
		if opts.Video.EmbedURL != "" {
			video["embedUrl"] = opts.Video.EmbedURL
		}
		if opts.Video.UploadDate != nil && !opts.Video.UploadDate.IsZero() {
			video["uploadDate"] = opts.Video.UploadDate.Format("2006-01-02")
		}
		if opts.Video.Duration != "" {
			video["duration"] = opts.Video.Duration
		}
		data["video"] = video
	}

	return data
}

// buildInstructions converts recipe steps to Schema.org HowToStep/HowToSection objects.
func (jr JSONLDRenderer) buildInstructions(recipe *cooklang.Recipe) []interface{} {
	var instructions []interface{}
	var currentSection *map[string]interface{}
	var sectionSteps []interface{}
	position := 1

	currentStep := recipe.FirstStep
	for currentStep != nil {
		// Build step text from components
		var stepText strings.Builder
		var isSection bool
		var sectionName string

		currentComponent := currentStep.FirstComponent
		for currentComponent != nil {
			switch comp := currentComponent.(type) {
			case *cooklang.Section:
				isSection = true
				sectionName = comp.Name
			case *cooklang.Instruction:
				stepText.WriteString(comp.Text)
			case *cooklang.Ingredient:
				stepText.WriteString(comp.Name)
			case *cooklang.Cookware:
				stepText.WriteString(comp.Name)
			case *cooklang.Timer:
				if comp.Duration != "" {
					if comp.Unit != "" {
						stepText.WriteString(comp.Duration + " " + comp.Unit)
					} else {
						stepText.WriteString(comp.Duration)
					}
				} else if comp.Name != "" {
					stepText.WriteString(comp.Name)
				}
			}
			currentComponent = currentComponent.GetNext()
		}

		if isSection {
			// If we have a current section with steps, finalize it
			if currentSection != nil && len(sectionSteps) > 0 {
				(*currentSection)["itemListElement"] = sectionSteps
				instructions = append(instructions, *currentSection)
			}

			// Start a new section
			newSection := map[string]interface{}{
				"@type": "HowToSection",
				"name":  sectionName,
			}
			currentSection = &newSection
			sectionSteps = nil
		}

		// Add step if it has content
		text := strings.TrimSpace(stepText.String())
		if text != "" {
			step := map[string]interface{}{
				"@type":    "HowToStep",
				"position": position,
				"text":     text,
			}
			position++

			if currentSection != nil {
				sectionSteps = append(sectionSteps, step)
			} else {
				instructions = append(instructions, step)
			}
		}

		currentStep = currentStep.NextStep
	}

	// Finalize any remaining section
	if currentSection != nil && len(sectionSteps) > 0 {
		(*currentSection)["itemListElement"] = sectionSteps
		instructions = append(instructions, *currentSection)
	}

	return instructions
}

// RenderRecipeJSON returns JSON-LD as a formatted (indented) JSON string.
// This is suitable for debugging or when you need the raw JSON output.
//
// Parameters:
//   - recipe: The parsed Cooklang recipe to render
//   - opts: Optional configuration (can be nil for defaults)
//
// Returns:
//   - string: The JSON-LD as an indented JSON string
//   - error: Any error during JSON marshaling
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("recipe.cook")
//	renderer := renderers.JSONLDRenderer{}
//	jsonStr, err := renderer.RenderRecipeJSON(recipe, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(jsonStr)
func (jr JSONLDRenderer) RenderRecipeJSON(recipe *cooklang.Recipe, opts *JSONLDOptions) (string, error) {
	data := jr.RenderRecipe(recipe, opts)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-LD: %w", err)
	}
	return string(bytes), nil
}

// RenderRecipeScriptTag returns a complete HTML <script> tag with JSON-LD content.
// This is ready to be embedded directly in an HTML page's <head> section.
//
// The output format is:
//
//	<script type="application/ld+json">
//	{
//	  "@context": "https://schema.org",
//	  "@type": "Recipe",
//	  ...
//	}
//	</script>
//
// Parameters:
//   - recipe: The parsed Cooklang recipe to render
//   - opts: Optional configuration (can be nil for defaults)
//
// Returns:
//   - string: The complete script tag with JSON-LD content
//   - error: Any error during JSON marshaling
//
// Example:
//
//	recipe, _ := cooklang.ParseFile("recipe.cook")
//	renderer := renderers.JSONLDRenderer{}
//	scriptTag, err := renderer.RenderRecipeScriptTag(recipe, &renderers.JSONLDOptions{
//	    URL: "https://example.com/recipes/my-recipe",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Embed scriptTag in your HTML template
func (jr JSONLDRenderer) RenderRecipeScriptTag(recipe *cooklang.Recipe, opts *JSONLDOptions) (string, error) {
	jsonStr, err := jr.RenderRecipeJSON(recipe, opts)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("<script type=\"application/ld+json\">\n%s\n</script>", jsonStr), nil
}

// ParseDurationToISO8601 converts human-readable duration strings to ISO 8601 format.
// This is used internally to convert prep_time and total_time fields.
//
// Supported input formats:
//   - "5 minutes", "5 min", "5m"
//   - "1 hour", "1 hr", "1h"
//   - "1 hour 30 minutes", "1h 30m"
//   - "90 minutes" (converted to "PT1H30M")
//   - "2 hours" → "PT2H"
//   - "30 seconds", "30 sec", "30s"
//
// Returns empty string if the input cannot be parsed.
//
// Examples:
//
//	ParseDurationToISO8601("15 minutes")     // "PT15M"
//	ParseDurationToISO8601("1 hour")         // "PT1H"
//	ParseDurationToISO8601("1h 30m")         // "PT1H30M"
//	ParseDurationToISO8601("90 minutes")     // "PT1H30M"
//	ParseDurationToISO8601("2 hours 15 min") // "PT2H15M"
func ParseDurationToISO8601(duration string) string {
	if duration == "" {
		return ""
	}

	// Normalize the input
	duration = strings.ToLower(strings.TrimSpace(duration))

	var hours, minutes, seconds int

	// Pattern: "1 hour 30 minutes" or "1h 30m" or similar
	hourPatterns := []string{
		`(\d+)\s*(?:hours?|hrs?|h)\b`,
	}
	minutePatterns := []string{
		`(\d+)\s*(?:minutes?|mins?|m)\b`,
	}
	secondPatterns := []string{
		`(\d+)\s*(?:seconds?|secs?|s)\b`,
	}

	// Extract hours
	for _, pattern := range hourPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(duration); len(matches) > 1 {
			if h, err := strconv.Atoi(matches[1]); err == nil {
				hours = h
			}
			break
		}
	}

	// Extract minutes
	for _, pattern := range minutePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(duration); len(matches) > 1 {
			if m, err := strconv.Atoi(matches[1]); err == nil {
				minutes = m
			}
			break
		}
	}

	// Extract seconds
	for _, pattern := range secondPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(duration); len(matches) > 1 {
			if s, err := strconv.Atoi(matches[1]); err == nil {
				seconds = s
			}
			break
		}
	}

	// If nothing was parsed, try to parse as a plain number (assume minutes)
	if hours == 0 && minutes == 0 && seconds == 0 {
		re := regexp.MustCompile(`^(\d+)$`)
		if matches := re.FindStringSubmatch(duration); len(matches) > 1 {
			if m, err := strconv.Atoi(matches[1]); err == nil {
				minutes = m
			}
		}
	}

	// Convert excess minutes to hours
	if minutes >= 60 {
		hours += minutes / 60
		minutes = minutes % 60
	}

	// Build ISO 8601 duration string
	if hours == 0 && minutes == 0 && seconds == 0 {
		return ""
	}

	result := "PT"
	if hours > 0 {
		result += fmt.Sprintf("%dH", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dM", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%dS", seconds)
	}

	return result
}

// NewJSONLDRenderer creates a new JSON-LD renderer.
// This is a convenience function that returns a configured JSONLDRenderer instance.
func NewJSONLDRenderer() JSONLDRenderer {
	return JSONLDRenderer{}
}

// DefaultJSONLDRenderer is the default instance of JSONLDRenderer.
var DefaultJSONLDRenderer = JSONLDRenderer{}
