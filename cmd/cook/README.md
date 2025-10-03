# Cook CLI

A comprehensive command-line tool for parsing, rendering, and managing [Cooklang](https://cooklang.org) recipes.

## Features

- 📖 **Parse** recipes and display detailed information
- 🥕 **Extract ingredients** from single or multiple recipes
- 🛒 **Create shopping lists** with automatic categorization
- 🎨 **Render recipes** in multiple formats (Cooklang, Markdown, HTML)
- ⚖️ **Scale recipes** to different serving sizes
- 🔄 **Unit conversion** between metric and imperial systems

## Installation

### From Source

```bash
cd cmd/cook
go build -o cook .
```

### Install to GOPATH

```bash
go install github.com/hilli/cooklang/cmd/cook@latest
```

## Commands

### `cook parse`

Parse and display a Cooklang recipe with all its components.

```bash
# Basic usage
cook parse recipe.cook

# Show detailed step-by-step breakdown
cook parse recipe.cook --detailed

# Output as JSON
cook parse recipe.cook --json

# JSON output with detailed information
cook parse recipe.cook --json --detailed
```

**Example output:**

```
📄 Recipe: example_recipes/Negroni.cook

📋 Title: Negroni
👥 Servings: 1
🏷️  Tags: [classic bitter aperitif gin vermouth campari]

🥕 Ingredients:
  • gin: 50 ml
  • vermouth: 50 ml
  • Campari: 50 ml
  ...

📖 Instructions:
1. All ingredients are 1:1, so ajust the amount to your liking.
2. Pour gin (50 ml), vermouth (50 ml) and Campari (50 ml) in a...
```

### `cook ingredients`

Extract and optionally consolidate ingredients from one or more recipes.

```bash
# List ingredients from a single recipe
cook ingredients recipe.cook

# Consolidate ingredients from multiple recipes
cook ingredients --consolidate recipe1.cook recipe2.cook

# Convert to specific unit system
cook ingredients --unit metric recipe.cook

# Consolidate and convert units
cook ingredients --consolidate --unit imperial *.cook

# Output as JSON
cook ingredients --json recipe.cook
```

**Options:**

- `--consolidate, -c`: Combine ingredients with the same name
- `--unit, -u`: Convert to unit system (`metric` or `imperial`)
- `--json`: Output as JSON

**Example output:**

```
📚 Ingredients from 2 recipes:
  1. Negroni
  2. Alaska
✓ Consolidated

🥕 Ingredients (11):
   1. gin: 50 ml
   2. vermouth: 50 ml
   ...
```

### `cook shopping-list`

Create a categorized shopping list from multiple recipes.

```bash
# Create shopping list from multiple recipes
cook shopping-list recipe1.cook recipe2.cook recipe3.cook

# Scale recipes in the shopping list
cook shopping-list --scale "recipe1.cook:2,recipe2.cook:4" recipe1.cook recipe2.cook

# Convert to specific unit system
cook shopping-list --unit metric *.cook

# Simple output (no categories)
cook shopping-list --simple recipe.cook

# Output as JSON
cook shopping-list --json recipe.cook
```

**Options:**

- `--scale, -s`: Scale recipes (format: `file:servings,file:servings`)
- `--unit, -u`: Convert to unit system (`metric` or `imperial`)
- `--simple`: Simple list without categories
- `--json`: Output as JSON

**Example output:**

```
🛒 Shopping List

📚 From 2 recipes:
  1. Negroni
  2. Alaska

🥬 Produce:
  ☐ orange zest: 1
  ☐ lemon peel (some)

🥛 Dairy & Refrigerated:
  ☐ ice: some
  ☐ ice cube: 1 cube

🍷 Beverages & Alcohol:
  ☐ gin: 50 ml
  ☐ vermouth: 50 ml
  ☐ Campari: 50 ml
  ...

📊 Total: 11 unique ingredients
```

### `cook render`

Render a recipe in different formats.

```bash
# Render as Markdown (to stdout)
cook render recipe.cook --format markdown

# Render as HTML
cook render recipe.cook --format html

# Render to file
cook render recipe.cook --format html --output recipe.html

# Render as Cooklang (normalized format)
cook render recipe.cook --format cooklang
```

**Supported formats:**

- `cooklang` / `cook`: Cooklang format (normalized)
- `markdown` / `md`: Markdown format
- `html`: HTML format

**Example:**

```bash
cook render Negroni.cook --format markdown --output Negroni.md
✅ Rendered to: Negroni.md
```

### `cook scale`

Scale a recipe's ingredients for different serving sizes.

```bash
# Scale to specific number of servings
cook scale recipe.cook --servings 4

# Scale by a custom factor
cook scale recipe.cook --factor 1.5

# Scale and convert units
cook scale recipe.cook --servings 6 --unit metric

# Scale and save to file
cook scale recipe.cook --servings 2 --output scaled.cook

# Scale and output in different format
cook scale recipe.cook --factor 0.5 --format markdown

# Scale and output as JSON
cook scale recipe.cook --servings 8 --json
```

**Options:**

- `--servings, -s`: Target number of servings
- `--factor, -f`: Scaling factor (e.g., 0.5 for half, 2 for double)
- `--unit, -u`: Convert to unit system (`metric` or `imperial`)
- `--output, -o`: Output file (default: stdout)
- `--format`: Output format (cooklang, markdown, html, json)
- `--json`: Output as JSON

**Example:**

```bash
cook scale Negroni.cook --servings 4
ℹ Scaling from 1 to 4 servings (factor: 4.00x)

Pour @gin{200%ml}, @vermouth{200%ml} and @Campari{200%ml} in a #tumber glass{} with a large @ice cube{4%cube}
...
```

## Usage Examples

### Daily Workflow

```bash
# Browse your recipe
cook parse ~/recipes/pasta-carbonara.cook

# Plan dinner for 6 people
cook scale ~/recipes/pasta-carbonara.cook --servings 6

# Create a shopping list for the week
cook shopping-list ~/recipes/monday.cook ~/recipes/tuesday.cook ~/recipes/wednesday.cook

# Export recipe to share
cook render ~/recipes/pasta-carbonara.cook --format markdown --output carbonara.md
```

### Recipe Organization

```bash
# Extract ingredients from all recipes
cook ingredients ~/recipes/*.cook --consolidate --json > all_ingredients.json

# Create categorized shopping list from favorites
cook shopping-list ~/recipes/favorites/*.cook --unit metric

# Generate HTML versions of all recipes
for recipe in ~/recipes/*.cook; do
  name=$(basename "$recipe" .cook)
  cook render "$recipe" --format html --output "html/${name}.html"
done
```

### Recipe Development

```bash
# Parse and validate a recipe
cook parse new-recipe.cook --detailed

# Test scaling
cook scale new-recipe.cook --servings 2
cook scale new-recipe.cook --servings 8

# Check ingredient list
cook ingredients new-recipe.cook

# Generate multiple formats
cook render new-recipe.cook --format cooklang --output recipe.cook
cook render new-recipe.cook --format markdown --output recipe.md
cook render new-recipe.cook --format html --output recipe.html
```

## Global Flags

- `--help, -h`: Help for any command
- `--version, -v`: Show version information

## Output Formats

### Human-Readable

The default output is designed for terminal viewing with:

- 📋 Unicode icons for visual appeal
- ✅ Color-coded messages (success, info, warning, error)
- 📊 Structured information display
- ☐ Checkbox lists for shopping

### JSON

All commands support JSON output with `--json` flag:

```bash
cook parse recipe.cook --json | jq .
```

**JSON structure:**

```json
{
  "metadata": {
    "title": "Recipe Title",
    "servings": "4",
    "tags": "tag1, tag2"
  },
  "ingredients": [
    {
      "name": "flour",
      "quantity": 500,
      "unit": "g"
    }
  ],
  "steps": ["Step 1", "Step 2"]
}
```

## Tips & Tricks

### Shell Aliases

Add to your `.bashrc` or `.zshrc`:

```bash
alias ckparse='cook parse'
alias cking='cook ingredients'
alias ckshop='cook shopping-list'
alias ckscale='cook scale'
alias ckrender='cook render'
```

### Batch Processing

```bash
# Find all recipes with a specific ingredient
for recipe in *.cook; do
  if cook ingredients "$recipe" --json | jq -e '.ingredients[] | select(.name == "garlic")' > /dev/null; then
    echo "$recipe contains garlic"
  fi
done

# Create shopping lists by category
cook shopping-list breakfast/*.cook --unit metric > breakfast_shopping.txt
cook shopping-list dinner/*.cook --unit metric > dinner_shopping.txt
```

### Integration with Other Tools

```bash
# Convert all recipes to HTML and preview
cook render recipe.cook --format html > /tmp/recipe.html && open /tmp/recipe.html

# Create shopping list and copy to clipboard
cook shopping-list *.cook | pbcopy  # macOS
cook shopping-list *.cook | xclip -selection clipboard  # Linux

# Extract ingredients and import to spreadsheet
cook ingredients *.cook --json > ingredients.json
```

## Error Handling

The CLI provides helpful error messages:

```bash
# Invalid file
$ cook parse nonexistent.cook
❌ Error: failed to read file: open nonexistent.cook: no such file or directory

# Invalid scaling parameters
$ cook scale recipe.cook
❌ Error: must specify either --servings or --factor

# Unsupported format
$ cook render recipe.cook --format pdf
❌ Error: unsupported format: pdf (supported: cooklang, markdown, html)
```

## Contributing

Found a bug or want to add a feature? Please visit the main repository at:
<https://github.com/hilli/cooklang>

## License

See the LICENSE file in the root of the repository.

## About Cooklang

Cooklang is a markup language for recipes that makes them readable by both humans and computers.

Learn more at: <https://cooklang.org>
