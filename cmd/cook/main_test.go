package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain builds the CLI binary before running tests
func TestMain(m *testing.M) {
	// Build the CLI binary for testing
	cmd := exec.Command("go", "build", "-o", "cook_test", ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		panic("failed to build CLI for testing: " + err.Error())
	}

	code := m.Run()

	// Cleanup - ignore error as file may not exist
	_ = os.Remove("cook_test")

	os.Exit(code)
}

// runCLI executes the CLI with given arguments and returns stdout, stderr, and error
func runCLI(args ...string) (string, string, error) {
	cmd := exec.Command("./cook_test", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// getExampleRecipePath returns the path to an example recipe
func getExampleRecipePath(name string) string {
	return filepath.Join("..", "..", "example_recipes", name)
}

func TestCLI_Version(t *testing.T) {
	stdout, _, err := runCLI("--version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(stdout, "cook version") {
		t.Errorf("expected version output, got: %s", stdout)
	}
}

func TestCLI_Help(t *testing.T) {
	stdout, _, err := runCLI("--help")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	// Check for main features mentioned in help
	expectedStrings := []string{
		"Cooklang",
		"parse",
		"render",
		"scale",
		"ingredients",
	}
	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("help output missing %q", expected)
		}
	}
}

func TestCLI_Parse(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("parse", recipePath)
	if err != nil {
		t.Fatalf("parse command failed: %v\nstderr: %s", err, stderr)
	}

	// Check that key elements are present in output
	expectedStrings := []string{
		"Negroni",
		"gin",
		"vermouth",
		"Campari",
	}
	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("parse output missing %q", expected)
		}
	}
}

func TestCLI_Parse_JSON(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("parse", recipePath, "--json")
	if err != nil {
		t.Fatalf("parse --json command failed: %v\nstderr: %s", err, stderr)
	}

	// Check that output is valid JSON-like structure
	if !strings.Contains(stdout, "{") || !strings.Contains(stdout, "}") {
		t.Errorf("expected JSON output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "\"title\"") {
		t.Errorf("JSON output missing title field")
	}
}

func TestCLI_Parse_Detailed(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("parse", recipePath, "--detailed")
	if err != nil {
		t.Fatalf("parse --detailed command failed: %v\nstderr: %s", err, stderr)
	}

	// Detailed output should include component breakdowns
	if !strings.Contains(stdout, "Ingredient:") {
		t.Errorf("detailed output missing ingredient breakdown")
	}
}

func TestCLI_Ingredients(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("ingredients", recipePath)
	if err != nil {
		t.Fatalf("ingredients command failed: %v\nstderr: %s", err, stderr)
	}

	// Check that ingredients are listed
	expectedIngredients := []string{
		"gin",
		"vermouth",
		"Campari",
		"orange zest",
	}
	for _, ing := range expectedIngredients {
		if !strings.Contains(stdout, ing) {
			t.Errorf("ingredients output missing %q", ing)
		}
	}
}

func TestCLI_Ingredients_JSON(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("ingredients", recipePath, "--json")
	if err != nil {
		t.Fatalf("ingredients --json command failed: %v\nstderr: %s", err, stderr)
	}

	// Check JSON structure
	if !strings.Contains(stdout, "[") || !strings.Contains(stdout, "]") {
		t.Errorf("expected JSON array output, got: %s", stdout)
	}
}

func TestCLI_Render_Markdown(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("render", recipePath, "--format", "markdown")
	if err != nil {
		t.Fatalf("render markdown command failed: %v\nstderr: %s", err, stderr)
	}

	// Check for markdown formatting
	if !strings.Contains(stdout, "#") {
		t.Errorf("markdown output missing headers")
	}
	if !strings.Contains(stdout, "Negroni") {
		t.Errorf("markdown output missing recipe title")
	}
}

func TestCLI_Render_HTML(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("render", recipePath, "--format", "html")
	if err != nil {
		t.Fatalf("render html command failed: %v\nstderr: %s", err, stderr)
	}

	// Check for HTML elements
	if !strings.Contains(stdout, "<") || !strings.Contains(stdout, ">") {
		t.Errorf("expected HTML output, got: %s", stdout)
	}
}

func TestCLI_Render_Cooklang(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("render", recipePath, "--format", "cooklang")
	if err != nil {
		t.Fatalf("render cooklang command failed: %v\nstderr: %s", err, stderr)
	}

	// Check for cooklang syntax elements
	if !strings.Contains(stdout, "@") {
		t.Errorf("cooklang output missing ingredient markers (@)")
	}
}

func TestCLI_Scale(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	stdout, stderr, err := runCLI("scale", recipePath, "--servings", "2")
	if err != nil {
		t.Fatalf("scale command failed: %v\nstderr: %s", err, stderr)
	}

	// Scaled recipe should have doubled quantities (50ml -> 100ml)
	if !strings.Contains(stdout, "100") {
		t.Errorf("scaled output doesn't show doubled quantities")
	}
}

func TestCLI_ShoppingList(t *testing.T) {
	negroniPath := getExampleRecipePath("Negroni.cook")
	alaskaPath := getExampleRecipePath("Alaska.cook")

	stdout, stderr, err := runCLI("shopping-list", negroniPath, alaskaPath)
	if err != nil {
		t.Fatalf("shopping-list command failed: %v\nstderr: %s", err, stderr)
	}

	// Shopping list should contain ingredients from both recipes
	if !strings.Contains(stdout, "gin") {
		t.Errorf("shopping list missing gin")
	}
}

func TestCLI_ShoppingListServings(t *testing.T) {
	negroniPath := getExampleRecipePath("Negroni.cook")

	// Test --servings flag
	stdout, stderr, err := runCLI("shopping-list", negroniPath, "--servings", "4")
	if err != nil {
		t.Fatalf("shopping-list --servings failed: %v\nstderr: %s", err, stderr)
	}

	// Output should mention scaling to servings
	if !strings.Contains(stdout, "4 servings") {
		t.Errorf("expected '4 servings' in output, got: %s", stdout)
	}
}

func TestCLI_ShoppingListServingsScaleMutualExclusion(t *testing.T) {
	negroniPath := getExampleRecipePath("Negroni.cook")

	// Test that --servings and --scale together produce an error
	_, _, err := runCLI("shopping-list", negroniPath, "--servings", "4", "--scale", "2.0")
	if err == nil {
		t.Error("expected error when using both --servings and --scale, got none")
	}
}

func TestCLI_CanonicalMode(t *testing.T) {
	recipePath := getExampleRecipePath("Negroni.cook")

	// Test that --canonical flag is accepted
	_, stderr, err := runCLI("parse", recipePath, "--canonical")
	if err != nil {
		t.Fatalf("parse with --canonical failed: %v\nstderr: %s", err, stderr)
	}
}

func TestCLI_InvalidFile(t *testing.T) {
	_, _, err := runCLI("parse", "nonexistent.cook")
	if err == nil {
		t.Error("expected error for nonexistent file, got none")
	}
}

func TestCLI_NoArgs(t *testing.T) {
	// Running without subcommand should show help
	stdout, _, _ := runCLI()
	if !strings.Contains(stdout, "cook") {
		t.Errorf("expected help output when no args provided")
	}
}
