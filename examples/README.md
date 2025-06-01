# Cooklang Parser Examples

This directory contains example cooklang recipes that demonstrate various features of the parser, particularly around YAML frontmatter and tag handling.

## Examples

### 1. `array_tags.cook`
Demonstrates recipes with **array-format tags** in YAML frontmatter:
```yaml
tags: [ pasta, italian, comfort-food ]
```

**Features shown:**
- YAML frontmatter with metadata
- Array-format tags with multiple items
- Complex recipe with ingredients, cookware, and timers
- Named timers (`~cooking{10-12%minutes}`)

### 2. `single_tag.cook`
Demonstrates recipes with **single-value tags**:
```yaml
tags: breakfast
```

**Features shown:**
- Simple string tag (not an array)
- Basic recipe structure
- Anonymous timers (`~{2-3%minutes}`)

### 3. `mixed_arrays.cook`
Demonstrates recipes with **multiple types of arrays** and **YAML list format**:
```yaml
tags:
  - spicy
  - asian
  - quick-meal
  - vegetarian
ingredients: [tofu, broccoli, soy sauce, garlic, ginger]
dietary_restrictions: [ gluten-free, dairy-free ]
cuisine: asian-fusion
```

**Features shown:**
- **YAML list format** with `-` items (for tags)
- **Bracket array format** (for ingredients and dietary restrictions)
- Arrays with varying spacing
- Mix of array and single-value fields
- Complex ingredient specifications

### 4. `edge_cases.cook`
Tests **edge cases** for array parsing:
```yaml
tags: []
empty_ingredients: [ ]
spaced_tags: [   quick   ,   easy   ]
single_item_array: [ dessert ]
no_spaces: [chocolate,vanilla,strawberry]
```

**Features shown:**
- Empty arrays
- Arrays with excessive spacing
- Single-item arrays
- Arrays without spaces between items

## Testing the Examples

You can test any of these examples with the cooklang parser:

```bash
# Test array tags
go run cmd/cook/main.go examples/array_tags.cook

# Test single tag
go run cmd/cook/main.go examples/single_tag.cook

# Test with JSON output
go run cmd/cook/main.go examples/mixed_arrays.cook --json

# Test edge cases
go run cmd/cook/main.go examples/edge_cases.cook
```

## Expected Behavior

The parser should handle all array formats by:
1. **Detecting bracket syntax**: `[ item1, item2, item3 ]`
2. **Detecting YAML list syntax**: `key:\n  - item1\n  - item2`
3. **Cleaning whitespace**: Removing extra spaces around items
4. **Converting to string**: Joining array items with `", "` for storage
5. **Preserving single values**: Non-array values remain unchanged

### Array Parsing Examples

| Input | Output |
|-------|--------|
| `tags: [ pasta, italian, comfort-food ]` | `tags: pasta, italian, comfort-food` |
| `tags:\n  - pasta\n  - italian\n  - comfort-food` | `tags: pasta, italian, comfort-food` |
| `tags: breakfast` | `tags: breakfast` |
| `tags: []` | `tags: ` (empty string) |
| `tags: [ dessert ]` | `tags: dessert` |
| `tags: [a,b,c]` | `tags: a, b, c` |
