package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hilli/cooklang/parser"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: cook <recipe-file> [--json]")
		os.Exit(1)
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create parser and parse the recipe
	p := parser.New()
	recipe, err := p.ParseBytes(content)
	if err != nil {
		fmt.Printf("Error parsing recipe: %v\n", err)
		os.Exit(1)
	}

	// Display the parsed recipe
	fmt.Printf("Parsing recipe: %s\n", filename)
	fmt.Printf("Content length: %d characters\n", len(content))
	fmt.Println("=====================================")

	// Display metadata
	if len(recipe.Metadata) > 0 {
		fmt.Println("📋 Recipe Metadata:")
		for key, value := range recipe.Metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}

	// Display steps
	fmt.Println("🍳 Recipe Steps:")
	for i, step := range recipe.Steps {
		fmt.Printf("Step %d:\n", i+1)
		for j, component := range step.Components {
			switch component.Type {
			case "ingredient":
				fmt.Printf("  [%d] 🥕 Ingredient: %s", j+1, component.Name)
				if component.Quantity != "" {
					fmt.Printf(" (%s", component.Quantity)
					if component.Unit != "" {
						fmt.Printf(" %s", component.Unit)
					}
					fmt.Printf(")")
				}
				fmt.Println()
			case "cookware":
				fmt.Printf("  [%d] 🍳 Cookware: %s", j+1, component.Name)
				if component.Quantity != "" {
					fmt.Printf(" (qty: %s)", component.Quantity)
				}
				fmt.Println()
			case "timer":
				fmt.Printf("  [%d] ⏲️  Timer:", j+1)
				if component.Name != "" {
					fmt.Printf(" %s", component.Name)
				}
				if component.Quantity != "" {
					fmt.Printf(" (%s", component.Quantity)
					if component.Unit != "" {
						fmt.Printf(" %s", component.Unit)
					}
					fmt.Printf(")")
				}
				fmt.Println()
			case "text":
				fmt.Printf("  [%d] 📝 Text: %q\n", j+1, component.Value)
			}
		}
		fmt.Println()
	}

	// Optional: Output as JSON for programmatic use
	if len(os.Args) > 2 && os.Args[2] == "--json" {
		fmt.Println("JSON Output:")
		fmt.Println("============")
		jsonOutput, err := json.MarshalIndent(recipe, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	}
}
