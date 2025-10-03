package cooklang

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRecipeImages(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cooklang-image-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case 1: Single image with same name
	t.Run("single_matching_image", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "TestRecipe.cook")
		imageFile := filepath.Join(tmpDir, "TestRecipe.jpg")

		// Create the files
		os.WriteFile(cookFile, []byte("Test recipe"), 0644)
		os.WriteFile(imageFile, []byte("fake image"), 0644)

		images := findRecipeImages(cookFile)
		if len(images) != 1 {
			t.Errorf("Expected 1 image, got %d", len(images))
		}
		if images[0] != "TestRecipe.jpg" {
			t.Errorf("Expected 'TestRecipe.jpg', got '%s'", images[0])
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(imageFile)
	})

	// Test case 2: Multiple extensions
	t.Run("multiple_extensions", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "MultiExt.cook")
		jpgFile := filepath.Join(tmpDir, "MultiExt.jpg")
		pngFile := filepath.Join(tmpDir, "MultiExt.png")

		os.WriteFile(cookFile, []byte("Test"), 0644)
		os.WriteFile(jpgFile, []byte("jpg"), 0644)
		os.WriteFile(pngFile, []byte("png"), 0644)

		images := findRecipeImages(cookFile)
		if len(images) != 2 {
			t.Errorf("Expected 2 images, got %d", len(images))
		}

		// Should find both jpg and png
		foundJpg := false
		foundPng := false
		for _, img := range images {
			if img == "MultiExt.jpg" {
				foundJpg = true
			}
			if img == "MultiExt.png" {
				foundPng = true
			}
		}
		if !foundJpg || !foundPng {
			t.Errorf("Expected to find both .jpg and .png, got: %v", images)
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(jpgFile)
		os.Remove(pngFile)
	})

	// Test case 3: Numbered images
	t.Run("numbered_images", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "Numbered.cook")
		image1 := filepath.Join(tmpDir, "Numbered-1.jpg")
		image2 := filepath.Join(tmpDir, "Numbered-2.png")
		image3 := filepath.Join(tmpDir, "Numbered-3.jpeg")

		os.WriteFile(cookFile, []byte("Test"), 0644)
		os.WriteFile(image1, []byte("img1"), 0644)
		os.WriteFile(image2, []byte("img2"), 0644)
		os.WriteFile(image3, []byte("img3"), 0644)

		images := findRecipeImages(cookFile)
		if len(images) != 3 {
			t.Errorf("Expected 3 images, got %d: %v", len(images), images)
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(image1)
		os.Remove(image2)
		os.Remove(image3)
	})

	// Test case 4: Mixed base and numbered images
	t.Run("mixed_base_and_numbered", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "Mixed.cook")
		baseImage := filepath.Join(tmpDir, "Mixed.jpg")
		image1 := filepath.Join(tmpDir, "Mixed-1.png")
		image2 := filepath.Join(tmpDir, "Mixed-2.jpg")

		os.WriteFile(cookFile, []byte("Test"), 0644)
		os.WriteFile(baseImage, []byte("base"), 0644)
		os.WriteFile(image1, []byte("img1"), 0644)
		os.WriteFile(image2, []byte("img2"), 0644)

		images := findRecipeImages(cookFile)
		if len(images) != 3 {
			t.Errorf("Expected 3 images, got %d: %v", len(images), images)
		}

		// Base image should be first
		if images[0] != "Mixed.jpg" {
			t.Errorf("Expected base image 'Mixed.jpg' first, got: %v", images)
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(baseImage)
		os.Remove(image1)
		os.Remove(image2)
	})

	// Test case 5: No images
	t.Run("no_images", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "NoImages.cook")
		os.WriteFile(cookFile, []byte("Test"), 0644)

		images := findRecipeImages(cookFile)
		if len(images) != 0 {
			t.Errorf("Expected 0 images, got %d: %v", len(images), images)
		}

		// Cleanup
		os.Remove(cookFile)
	})

	// Test case 6: Gap in numbering (should stop at gap)
	t.Run("gap_in_numbering", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "Gap.cook")
		image1 := filepath.Join(tmpDir, "Gap-1.jpg")
		image3 := filepath.Join(tmpDir, "Gap-3.jpg") // Gap at 2

		os.WriteFile(cookFile, []byte("Test"), 0644)
		os.WriteFile(image1, []byte("img1"), 0644)
		os.WriteFile(image3, []byte("img3"), 0644)

		images := findRecipeImages(cookFile)
		// Should only find Gap-1.jpg, stop at the gap
		if len(images) != 1 {
			t.Errorf("Expected 1 image (should stop at gap), got %d: %v", len(images), images)
		}
		if images[0] != "Gap-1.jpg" {
			t.Errorf("Expected 'Gap-1.jpg', got '%s'", images[0])
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(image1)
		os.Remove(image3)
	})
}

func TestParseFileWithImageDetection(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cooklang-parse-image-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case 1: Auto-detect images with no frontmatter
	t.Run("auto_detect_no_frontmatter", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "AutoDetect.cook")
		imageFile := filepath.Join(tmpDir, "AutoDetect.jpg")

		recipeContent := `Add @water{2%cups} to a #pot{}.

Heat for ~{10%minutes}.`

		os.WriteFile(cookFile, []byte(recipeContent), 0644)
		os.WriteFile(imageFile, []byte("fake image"), 0644)

		recipe, err := ParseFile(cookFile)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		if len(recipe.Images) != 1 {
			t.Errorf("Expected 1 auto-detected image, got %d: %v", len(recipe.Images), recipe.Images)
		}
		if recipe.Images[0] != "AutoDetect.jpg" {
			t.Errorf("Expected 'AutoDetect.jpg', got '%s'", recipe.Images[0])
		}

		// Check metadata was updated
		if imgMeta, ok := recipe.Metadata["images"]; !ok || imgMeta != "AutoDetect.jpg" {
			t.Errorf("Expected metadata images='AutoDetect.jpg', got '%s'", imgMeta)
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(imageFile)
	})

	// Test case 2: Merge with existing frontmatter images
	t.Run("merge_with_existing", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "Merge.cook")
		image1File := filepath.Join(tmpDir, "Merge.jpg")
		image2File := filepath.Join(tmpDir, "Merge-1.png")

		recipeContent := `---
title: Test Recipe
images: existing.jpg
---

Add @water{2%cups}.`

		os.WriteFile(cookFile, []byte(recipeContent), 0644)
		os.WriteFile(image1File, []byte("img1"), 0644)
		os.WriteFile(image2File, []byte("img2"), 0644)

		recipe, err := ParseFile(cookFile)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		// Should have existing.jpg + detected images
		if len(recipe.Images) != 3 {
			t.Errorf("Expected 3 images (1 existing + 2 detected), got %d: %v", len(recipe.Images), recipe.Images)
		}

		// Check all images are present
		imageMap := make(map[string]bool)
		for _, img := range recipe.Images {
			imageMap[img] = true
		}

		if !imageMap["existing.jpg"] {
			t.Errorf("Missing existing.jpg from frontmatter")
		}
		if !imageMap["Merge.jpg"] {
			t.Errorf("Missing detected Merge.jpg")
		}
		if !imageMap["Merge-1.png"] {
			t.Errorf("Missing detected Merge-1.png")
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(image1File)
		os.Remove(image2File)
	})

	// Test case 3: Avoid duplicates
	t.Run("avoid_duplicates", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "Duplicate.cook")
		imageFile := filepath.Join(tmpDir, "Duplicate.jpg")

		recipeContent := `---
title: Test Recipe
images: Duplicate.jpg
---

Add @water{2%cups}.`

		os.WriteFile(cookFile, []byte(recipeContent), 0644)
		os.WriteFile(imageFile, []byte("img"), 0644)

		recipe, err := ParseFile(cookFile)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		// Should only have one Duplicate.jpg (not duplicated)
		if len(recipe.Images) != 1 {
			t.Errorf("Expected 1 image (no duplicates), got %d: %v", len(recipe.Images), recipe.Images)
		}
		if recipe.Images[0] != "Duplicate.jpg" {
			t.Errorf("Expected 'Duplicate.jpg', got '%s'", recipe.Images[0])
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(imageFile)
	})

	// Test case 4: Multiple detected images
	t.Run("multiple_detected_images", func(t *testing.T) {
		cookFile := filepath.Join(tmpDir, "Multiple.cook")
		baseImage := filepath.Join(tmpDir, "Multiple.jpg")
		image1 := filepath.Join(tmpDir, "Multiple-1.png")
		image2 := filepath.Join(tmpDir, "Multiple-2.jpeg")

		recipeContent := `---
title: Test Recipe
---

Add @water{2%cups}.`

		os.WriteFile(cookFile, []byte(recipeContent), 0644)
		os.WriteFile(baseImage, []byte("base"), 0644)
		os.WriteFile(image1, []byte("img1"), 0644)
		os.WriteFile(image2, []byte("img2"), 0644)

		recipe, err := ParseFile(cookFile)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		if len(recipe.Images) != 3 {
			t.Errorf("Expected 3 images, got %d: %v", len(recipe.Images), recipe.Images)
		}

		// Base image should be first
		if recipe.Images[0] != "Multiple.jpg" {
			t.Errorf("Expected base image 'Multiple.jpg' first, got: %v", recipe.Images)
		}

		// Check metadata
		imgMeta := recipe.Metadata["images"]
		if !strings.Contains(imgMeta, "Multiple.jpg") {
			t.Errorf("Metadata should contain Multiple.jpg")
		}
		if !strings.Contains(imgMeta, "Multiple-1.png") {
			t.Errorf("Metadata should contain Multiple-1.png")
		}
		if !strings.Contains(imgMeta, "Multiple-2.jpeg") {
			t.Errorf("Metadata should contain Multiple-2.jpeg")
		}

		// Cleanup
		os.Remove(cookFile)
		os.Remove(baseImage)
		os.Remove(image1)
		os.Remove(image2)
	})
}

func TestMergeUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "no_duplicates",
			slice1:   []string{"a", "b"},
			slice2:   []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "with_duplicates",
			slice1:   []string{"a", "b", "c"},
			slice2:   []string{"b", "c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "empty_slices",
			slice1:   []string{},
			slice2:   []string{},
			expected: []string{},
		},
		{
			name:     "one_empty",
			slice1:   []string{"a", "b"},
			slice2:   []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "with_empty_strings",
			slice1:   []string{"a", "", "b"},
			slice2:   []string{"", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeUniqueStrings(tt.slice1, tt.slice2)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("At index %d: expected '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "fileexists-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test file that exists
	existingFile := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	if !fileExists(existingFile) {
		t.Errorf("fileExists returned false for existing file")
	}

	// Test file that doesn't exist
	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.txt")
	if fileExists(nonExistentFile) {
		t.Errorf("fileExists returned true for non-existent file")
	}

	// Test directory (should return false)
	if fileExists(tmpDir) {
		t.Errorf("fileExists returned true for directory")
	}
}
