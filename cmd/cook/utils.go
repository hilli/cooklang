package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hilli/cooklang"
)

// readRecipeFile reads and parses a recipe file
func readRecipeFile(filename string) (*cooklang.Recipe, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	recipe, err := cooklang.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing recipe: %w", err)
	}

	return recipe, nil
}

// readMultipleRecipes reads and parses multiple recipe files
func readMultipleRecipes(filenames []string) ([]*cooklang.Recipe, error) {
	recipes := make([]*cooklang.Recipe, 0, len(filenames))
	for _, filename := range filenames {
		recipe, err := readRecipeFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", filename, err)
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

// outputJSON outputs data as formatted JSON
func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// printError prints an error message to stderr
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// printSuccess prints a success message
func printSuccess(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// printWarning prints a warning message
func printWarning(format string, args ...interface{}) {
	fmt.Printf("⚠ "+format+"\n", args...)
}

// printInfo prints an info message
func printInfo(format string, args ...interface{}) {
	fmt.Printf("ℹ "+format+"\n", args...)
}
