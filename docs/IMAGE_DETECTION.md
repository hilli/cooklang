# Image Auto-Detection

When parsing `.cook` recipe files, the parser automatically detects and includes image files that match the recipe filename pattern.

## How It Works

When you call `cooklang.ParseFile()`, the parser automatically:

1. Looks for image files matching the recipe's base name
2. Merges detected images with any existing images from frontmatter
3. Avoids duplicates
4. Updates the recipe's `Images` field and metadata

## Image File Patterns

The parser searches for images following these patterns:

### Base Images

For a recipe file named `Recipe.cook`, it searches for:

- `Recipe.jpg`
- `Recipe.jpeg`
- `Recipe.png`

### Numbered Images

For sequential images:

- `Recipe-1.jpg`, `Recipe-1.jpeg`, `Recipe-1.png`
- `Recipe-2.jpg`, `Recipe-2.jpeg`, `Recipe-2.png`
- `Recipe-3.jpg`, etc.

**Important:** The parser stops searching when it encounters a gap in the numbering sequence. For example, if you have `Recipe-1.jpg` and `Recipe-3.jpg` but no `Recipe-2.*`, only `Recipe-1.jpg` will be detected.

## Examples

### Example 1: Basic Auto-Detection

**Directory structure:**

```
recipes/
  Pancakes.cook
  Pancakes.jpg
  Pancakes-1.png
```

**Recipe file (`Pancakes.cook`):**

```cooklang
Add @flour{2%cups} and @milk{1%cup}.
```

**Result:**

```go
recipe, _ := cooklang.ParseFile("recipes/Pancakes.cook")
fmt.Println(recipe.Images)
// Output: [Pancakes.jpg Pancakes-1.png]
```

### Example 2: Merging with Frontmatter

**Directory structure:**

```
recipes/
  Burger.cook
  Burger.jpg
  Burger-1.png
```

**Recipe file (`Burger.cook`):**

```cooklang
---
title: Classic Burger
images: burger-plated.jpg, burger-closeup.jpg
---

Form @ground beef{1%lb} into patties.
```

**Result:**

```go
recipe, _ := cooklang.ParseFile("recipes/Burger.cook")
fmt.Println(recipe.Images)
// Output: [burger-plated.jpg burger-closeup.jpg Burger.jpg Burger-1.png]
```

### Example 3: Avoiding Duplicates

**Directory structure:**

```
recipes/
  Salad.cook
  Salad.jpg
```

**Recipe file (`Salad.cook`):**

```cooklang
---
title: Green Salad
images: Salad.jpg
---

Chop @lettuce{1%head}.
```

**Result:**

```go
recipe, _ := cooklang.ParseFile("recipes/Salad.cook")
fmt.Println(recipe.Images)
// Output: [Salad.jpg]
// Note: No duplicate - Salad.jpg appears only once
```

### Example 4: Multiple Extensions

**Directory structure:**

```
recipes/
  Pasta.cook
  Pasta.jpg
  Pasta.png
  Pasta.jpeg
```

**Result:**

```go
recipe, _ := cooklang.ParseFile("recipes/Pasta.cook")
fmt.Println(recipe.Images)
// Output: [Pasta.jpg Pasta.jpeg Pasta.png]
// Note: All three image formats are detected
```

### Example 5: Gap in Numbering

**Directory structure:**

```
recipes/
  Cake.cook
  Cake.jpg
  Cake-1.jpg
  Cake-3.jpg  ← Note: Cake-2.jpg is missing
```

**Result:**

```go
recipe, _ := cooklang.ParseFile("recipes/Cake.cook")
fmt.Println(recipe.Images)
// Output: [Cake.jpg Cake-1.jpg]
// Note: Cake-3.jpg is NOT included due to gap at -2
```

## Usage

### Parsing with Auto-Detection

```go
import "github.com/hilli/cooklang"

// Auto-detection happens automatically when using ParseFile
recipe, err := cooklang.ParseFile("path/to/recipe.cook")
if err != nil {
    log.Fatal(err)
}

// Access detected images
for _, img := range recipe.Images {
    fmt.Println(img)
}

// Images are also available in metadata
fmt.Println(recipe.Metadata["images"])
```

### Without Auto-Detection

If you're parsing from a string or bytes (not a file), auto-detection won't occur:

```go
// ParseString does NOT auto-detect images (no file context)
recipe, err := cooklang.ParseString(recipeText)

// ParseBytes does NOT auto-detect images (no file context)
recipe, err := cooklang.ParseBytes(recipeBytes)
```

## Implementation Details

### Functions

#### `findRecipeImages(cookFilePath string) []string`

Searches for image files matching the recipe filename pattern.

**Parameters:**

- `cookFilePath`: Full path to the `.cook` file

**Returns:**

- Slice of image filenames (just the filenames, not full paths)

#### `mergeUniqueStrings(slice1, slice2 []string) []string`

Merges two string slices, removing duplicates.

**Parameters:**

- `slice1`: First slice (typically existing frontmatter images)
- `slice2`: Second slice (typically auto-detected images)

**Returns:**

- Merged slice with no duplicates

#### `fileExists(path string) bool`

Checks if a file exists and is not a directory.

**Parameters:**

- `path`: Full path to check

**Returns:**

- `true` if file exists and is not a directory

### Order of Images

When merging images:

1. Existing frontmatter images come first (in their original order)
2. Auto-detected base images come next (e.g., `Recipe.jpg`)
3. Auto-detected numbered images follow (e.g., `Recipe-1.jpg`, `Recipe-2.jpg`)

## Best Practices

1. **Consistent Naming**: Name your image files to match your recipe files
2. **Sequential Numbering**: Use sequential numbers without gaps (1, 2, 3, etc.)
3. **Multiple Angles**: Use numbered images for different angles/stages
4. **File Formats**: Stick to common formats (jpg, jpeg, png)

## Testing

The image auto-detection feature includes comprehensive tests:

```bash
# Run image detection tests
go test -v -run TestFindRecipeImages
go test -v -run TestParseFileWithImageDetection

# Run all tests
go test ./...
```

## See Also

- [Frontmatter CRUD Operations](./FRONTMATTER_CRUD.md) - For programmatically managing recipe metadata including images
- [Cooklang Specification](https://cooklang.org/docs/spec/) - Official Cooklang specification
