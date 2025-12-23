package cooklang

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontmatterEditor_GetMetadata(t *testing.T) {
	// Create a temporary test file
	content := `---
title: Test Recipe
cuisine: Italian
tags:
  - pasta
  - vegetarian
images:
  - test.jpg
servings: 4
date: 2025-10-03
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	tests := []struct {
		key      string
		expected string
		exists   bool
	}{
		{"title", "Test Recipe", true},
		{"cuisine", "Italian", true},
		{"tags", "pasta, vegetarian", true},
		{"images", "test.jpg", true},
		{"servings", "4", true},
		{"date", "2025-10-03", true},
		{"nonexistent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, exists := editor.GetMetadata(tt.key)
			if exists != tt.exists {
				t.Errorf("GetMetadata(%q) exists = %v, want %v", tt.key, exists, tt.exists)
			}
			if value != tt.expected {
				t.Errorf("GetMetadata(%q) = %q, want %q", tt.key, value, tt.expected)
			}
		})
	}
}

func TestFrontmatterEditor_GetAllMetadata(t *testing.T) {
	content := `---
title: Test Recipe
cuisine: Italian
custom_field: custom_value
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	allMeta := editor.GetAllMetadata()

	expectedKeys := []string{"title", "cuisine", "custom_field"}
	for _, key := range expectedKeys {
		if _, exists := allMeta[key]; !exists {
			t.Errorf("GetAllMetadata() missing key %q", key)
		}
	}

	if allMeta["title"] != "Test Recipe" {
		t.Errorf("GetAllMetadata()[title] = %q, want %q", allMeta["title"], "Test Recipe")
	}
	if allMeta["custom_field"] != "custom_value" {
		t.Errorf("GetAllMetadata()[custom_field] = %q, want %q", allMeta["custom_field"], "custom_value")
	}
}

func TestFrontmatterEditor_SetMetadata(t *testing.T) {
	content := `---
title: Original Title
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Test updating existing field
	if err := editor.SetMetadata("title", "New Title"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if value, _ := editor.GetMetadata("title"); value != "New Title" {
		t.Errorf("After SetMetadata, title = %q, want %q", value, "New Title")
	}

	// Test adding new field
	if err := editor.SetMetadata("author", "John Doe"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if value, exists := editor.GetMetadata("author"); !exists || value != "John Doe" {
		t.Errorf("After SetMetadata, author = %q (exists=%v), want %q", value, exists, "John Doe")
	}

	// Test setting array field
	if err := editor.SetMetadata("tags", "italian, quick, easy"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if value, _ := editor.GetMetadata("tags"); value != "italian, quick, easy" {
		t.Errorf("After SetMetadata, tags = %q, want %q", value, "italian, quick, easy")
	}

	// Test setting servings
	if err := editor.SetMetadata("servings", "6"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if value, _ := editor.GetMetadata("servings"); value != "6" {
		t.Errorf("After SetMetadata, servings = %q, want %q", value, "6")
	}
}

func TestFrontmatterEditor_DeleteMetadata(t *testing.T) {
	content := `---
title: Test Recipe
cuisine: Italian
author: John Doe
custom_field: custom_value
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Delete structured field
	if err := editor.DeleteMetadata("cuisine"); err != nil {
		t.Fatalf("DeleteMetadata failed: %v", err)
	}
	if _, exists := editor.GetMetadata("cuisine"); exists {
		t.Error("After DeleteMetadata, cuisine still exists")
	}

	// Delete custom field
	if err := editor.DeleteMetadata("custom_field"); err != nil {
		t.Fatalf("DeleteMetadata failed: %v", err)
	}
	if _, exists := editor.GetMetadata("custom_field"); exists {
		t.Error("After DeleteMetadata, custom_field still exists")
	}

	// Title should still exist
	if _, exists := editor.GetMetadata("title"); !exists {
		t.Error("After DeleteMetadata, title should still exist")
	}
}

func TestFrontmatterEditor_SaveAndReload(t *testing.T) {
	content := `---
title: Test Recipe
cuisine: Italian
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Update metadata
	if err := editor.SetMetadata("title", "Updated Title"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if err := editor.SetMetadata("author", "Jane Doe"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if err := editor.DeleteMetadata("cuisine"); err != nil {
		t.Fatalf("DeleteMetadata failed: %v", err)
	}

	// Save the file
	if err := editor.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload the file
	newEditor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to reload editor: %v", err)
	}

	// Verify changes persisted
	if value, _ := newEditor.GetMetadata("title"); value != "Updated Title" {
		t.Errorf("After reload, title = %q, want %q", value, "Updated Title")
	}
	if value, _ := newEditor.GetMetadata("author"); value != "Jane Doe" {
		t.Errorf("After reload, author = %q, want %q", value, "Jane Doe")
	}
	if _, exists := newEditor.GetMetadata("cuisine"); exists {
		t.Error("After reload, cuisine should not exist")
	}

	// Verify recipe body is preserved
	savedContent, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(savedContent), "Mix @flour{500%g} with @water{300%ml}") {
		t.Error("Recipe body was not preserved")
	}
}

func TestFrontmatterEditor_GetUpdatedContent(t *testing.T) {
	content := `---
title: Test Recipe
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Update metadata
	if err := editor.SetMetadata("title", "New Title"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}
	if err := editor.SetMetadata("author", "John Doe"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Get updated content without saving
	updatedContent := editor.GetUpdatedContent()

	// Verify frontmatter is updated
	if !strings.Contains(updatedContent, "title: New Title") {
		t.Error("Updated content should contain new title")
	}
	if !strings.Contains(updatedContent, "author: John Doe") {
		t.Error("Updated content should contain new author")
	}

	// Verify recipe body is preserved
	if !strings.Contains(updatedContent, "Mix @flour{500%g} with @water{300%ml}") {
		t.Error("Recipe body was not preserved in updated content")
	}

	// Verify original file is unchanged
	originalContent, _ := os.ReadFile(tmpFile)
	if strings.Contains(string(originalContent), "New Title") {
		t.Error("Original file should not be modified by GetUpdatedContent")
	}
}

func TestFrontmatterEditor_AppendToArray(t *testing.T) {
	content := `---
title: Test Recipe
tags:
  - italian
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Append to tags
	if err := editor.AppendToArray("tags", "quick"); err != nil {
		t.Fatalf("AppendToArray failed: %v", err)
	}
	if err := editor.AppendToArray("tags", "easy"); err != nil {
		t.Fatalf("AppendToArray failed: %v", err)
	}

	tags, _ := editor.GetMetadata("tags")
	if !strings.Contains(tags, "italian") || !strings.Contains(tags, "quick") || !strings.Contains(tags, "easy") {
		t.Errorf("After AppendToArray, tags = %q, should contain all items", tags)
	}

	// Append to images
	if err := editor.AppendToArray("images", "photo1.jpg"); err != nil {
		t.Fatalf("AppendToArray failed: %v", err)
	}
	if err := editor.AppendToArray("images", "photo2.jpg"); err != nil {
		t.Fatalf("AppendToArray failed: %v", err)
	}

	images, _ := editor.GetMetadata("images")
	if !strings.Contains(images, "photo1.jpg") || !strings.Contains(images, "photo2.jpg") {
		t.Errorf("After AppendToArray, images = %q, should contain all items", images)
	}

	// Try to append to non-array field (should fail)
	if err := editor.AppendToArray("title", "something"); err == nil {
		t.Error("AppendToArray on non-array field should fail")
	}
}

func TestFrontmatterEditor_RemoveFromArray(t *testing.T) {
	content := `---
title: Test Recipe
tags:
  - italian
  - quick
  - easy
images:
  - photo1.jpg
  - photo2.jpg
  - photo3.jpg
---

Mix @flour{500%g} with @water{300%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Remove from tags
	if err := editor.RemoveFromArray("tags", "quick"); err != nil {
		t.Fatalf("RemoveFromArray failed: %v", err)
	}

	tags, _ := editor.GetMetadata("tags")
	if strings.Contains(tags, "quick") {
		t.Errorf("After RemoveFromArray, tags should not contain 'quick', got %q", tags)
	}
	if !strings.Contains(tags, "italian") || !strings.Contains(tags, "easy") {
		t.Errorf("After RemoveFromArray, tags = %q, should still contain other items", tags)
	}

	// Remove from images
	if err := editor.RemoveFromArray("images", "photo2.jpg"); err != nil {
		t.Fatalf("RemoveFromArray failed: %v", err)
	}

	images, _ := editor.GetMetadata("images")
	if strings.Contains(images, "photo2.jpg") {
		t.Errorf("After RemoveFromArray, images should not contain 'photo2.jpg', got %q", images)
	}
	if !strings.Contains(images, "photo1.jpg") || !strings.Contains(images, "photo3.jpg") {
		t.Errorf("After RemoveFromArray, images = %q, should still contain other items", images)
	}
}

func TestFrontmatterEditor_NoFrontmatter(t *testing.T) {
	// Test with a file that has no frontmatter
	content := `Mix @flour{500%g} with @water{300%ml}.`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Should be able to add metadata
	if err := editor.SetMetadata("title", "New Recipe"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Save and verify
	if err := editor.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload and verify
	newEditor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to reload editor: %v", err)
	}

	if value, _ := newEditor.GetMetadata("title"); value != "New Recipe" {
		t.Errorf("After adding frontmatter, title = %q, want %q", value, "New Recipe")
	}

	// Verify recipe body is preserved
	savedContent, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(savedContent), "Mix @flour{500%g} with @water{300%ml}") {
		t.Error("Recipe body was not preserved")
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_LiteralStrip(t *testing.T) {
	// Test parsing literal block scalar with strip (|-)
	content := `---
title: Spicy Margarita
description: |-
  The cocktail is a
  sweet and intensely flavoured
  Beast.
servings: 2
---

Mix @tequila{60%ml} with @lime juice{30%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Verify description is parsed correctly
	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	expected := "The cocktail is a\nsweet and intensely flavoured\nBeast."
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}

	// Verify other fields are still parsed correctly
	if title, _ := editor.GetMetadata("title"); title != "Spicy Margarita" {
		t.Errorf("title = %q, want %q", title, "Spicy Margarita")
	}
	if servings, _ := editor.GetMetadata("servings"); servings != "2" {
		t.Errorf("servings = %q, want %q", servings, "2")
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_Literal(t *testing.T) {
	// Test parsing literal block scalar with clip (|) - keeps single trailing newline
	content := `---
title: Test Recipe
description: |
  Line 1
  Line 2
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	// Clip chomping keeps a single trailing newline
	expected := "Line 1\nLine 2\n"
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_LiteralKeep(t *testing.T) {
	// Test parsing literal block scalar with keep (|+)
	content := `---
title: Test Recipe
description: |+
  Line 1
  Line 2

---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	// Keep chomping preserves all trailing newlines
	expected := "Line 1\nLine 2\n\n"
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_FoldedStrip(t *testing.T) {
	// Test parsing folded block scalar with strip (>-)
	content := `---
title: Test Recipe
description: >-
  This is a long
  description that
  spans multiple lines.
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	// Folded style converts newlines to spaces, strip removes trailing
	expected := "This is a long description that spans multiple lines."
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_Folded(t *testing.T) {
	// Test parsing folded block scalar with clip (>)
	content := `---
title: Test Recipe
description: >
  This is a long
  description.
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	// Folded style converts newlines to spaces, clip keeps single trailing newline
	expected := "This is a long description.\n"
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_AtEnd(t *testing.T) {
	// Test parsing block scalar at end of frontmatter (no following field)
	content := `---
title: Test Recipe
description: |-
  This is the last
  field in frontmatter.
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	expected := "This is the last\nfield in frontmatter."
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}
}

func TestFrontmatterEditor_MultilineBlockScalar_WithBlankLines(t *testing.T) {
	// Test parsing block scalar with blank lines inside
	content := `---
title: Test Recipe
description: |-
  First paragraph.

  Second paragraph.
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	description, exists := editor.GetMetadata("description")
	if !exists {
		t.Fatal("description should exist")
	}

	expected := "First paragraph.\n\nSecond paragraph."
	if description != expected {
		t.Errorf("description = %q, want %q", description, expected)
	}
}

func TestFrontmatterEditor_WriteMultilineDescription(t *testing.T) {
	// Test that multi-line descriptions are written as block scalars
	content := `---
title: Test Recipe
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Set a multi-line description
	multilineDesc := "Line 1\nLine 2\nLine 3"
	if err := editor.SetMetadata("description", multilineDesc); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Get the updated content
	updatedContent := editor.GetUpdatedContent()

	// Verify it uses block scalar syntax
	if !strings.Contains(updatedContent, "description: |-") {
		t.Errorf("Multi-line description should use |- block scalar syntax, got:\n%s", updatedContent)
	}

	// Verify the indented lines are present
	if !strings.Contains(updatedContent, "  Line 1") {
		t.Error("Block scalar content should be indented with 2 spaces")
	}
}

func TestFrontmatterEditor_WriteSingleLineDescription(t *testing.T) {
	// Test that single-line descriptions stay inline
	content := `---
title: Test Recipe
---

Mix ingredients.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	editor, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	// Set a single-line description
	if err := editor.SetMetadata("description", "A simple recipe"); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Get the updated content
	updatedContent := editor.GetUpdatedContent()

	// Verify it stays inline
	if !strings.Contains(updatedContent, "description: A simple recipe") {
		t.Errorf("Single-line description should stay inline, got:\n%s", updatedContent)
	}

	// Verify it does NOT use block scalar syntax
	if strings.Contains(updatedContent, "description: |-") {
		t.Error("Single-line description should not use block scalar syntax")
	}
}

func TestFrontmatterEditor_MultilineRoundTrip(t *testing.T) {
	// Test that multi-line descriptions survive a round-trip (parse -> save -> parse)
	content := `---
title: Spicy Margarita
description: |-
  The cocktail is a
  sweet and intensely flavoured
  Beast.
---

Mix @tequila{60%ml} with @lime juice{30%ml}.
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cook")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// First parse
	editor1, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}

	originalDesc, _ := editor1.GetMetadata("description")

	// Save
	if err := editor1.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Parse again
	editor2, err := NewFrontmatterEditor(tmpFile)
	if err != nil {
		t.Fatalf("Failed to reload editor: %v", err)
	}

	reloadedDesc, _ := editor2.GetMetadata("description")

	// Verify content is preserved
	if originalDesc != reloadedDesc {
		t.Errorf("Round-trip failed:\nOriginal: %q\nReloaded: %q", originalDesc, reloadedDesc)
	}

	// Verify recipe body is preserved
	savedContent, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(savedContent), "Mix @tequila{60%ml} with @lime juice{30%ml}") {
		t.Error("Recipe body was not preserved")
	}
}
