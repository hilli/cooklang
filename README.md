[![CI - Build and Test](https://github.com/hilli/cooklang/actions/workflows/ci.yaml/badge.svg)](https://github.com/hilli/cooklang/actions/workflows/ci.yaml)

# cooklang

Go implementation of a cooklang parser.

## Features

- ✅ Full Cooklang specification compliance
- 🖼️ **Automatic image detection** - Auto-discovers recipe images matching filename patterns
- 📝 Frontmatter CRUD operations - Programmatically edit recipe metadata
- 🧮 Unit conversion system with metric/imperial/US systems
- 📋 Shopping list generation from multiple recipes
- 🎨 Multiple output formats (Cooklang, Markdown, HTML, JSON)
- 🔧 Extended mode with ingredient/cookware annotations
- ⚖️ Recipe scaling and ingredient consolidation
- 🛠️ Comprehensive CLI tool

## Usage Examples

The library includes comprehensive, runnable examples demonstrating all major features. See [docs/EXAMPLES.md](docs/EXAMPLES.md) for a complete list.

Quick example:

```go
package main

import (
    "fmt"
    "log"
    "github.com/hilli/cooklang"
)

func main() {
    recipeText := `---
title: Pasta Aglio e Olio
servings: 2
---
Cook @pasta{400%g} in salted water for ~{10%minutes}.
Meanwhile, heat @olive oil{4%tbsp} and sauté @garlic{3%cloves}.
Toss everything together and serve.`

    recipe, err := cooklang.ParseString(recipeText)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Recipe: %s (serves %.0f)\n", recipe.Title, recipe.Servings)
    
    // Get shopping list
    shoppingList, _ := recipe.GetCollectedIngredientsMap()
    for ingredient, quantity := range shoppingList {
        fmt.Printf("- %s: %s\n", ingredient, quantity)
    }
}
```

## Known Usages

Projects using this library:

- [cocktailti.me](https://cocktailti.me) - Cocktail recipe sharing website

*Using this library? Open a PR to add your project!*

## Cooklang specification

See the [Cooklang specification](https://github.com/cooklang/spec/) for details.

### Extended Syntax

This parser supports additional syntax extensions beyond the base Cooklang specification:

#### Ingredient Annotations

Ingredients can have annotations in parentheses to specify preparation notes or state:

```cooklang
@milk{1%l}(cold)
```

The annotation `(cold)` is stored as the ingredient's `value` field.

#### Cookware Annotations

Similarly, cookware items can have annotations for usage hints:

```cooklang
#pan{}(for frying)
```

#### Named Timers

Timers can have descriptive multi-word names:

```cooklang
~roast time{4%hours}
```

#### Comment Preservation

Comments are preserved as a distinct type in the parsed output rather than being discarded:

```cooklang
-- This is a comment
```

Comments are accessible with `type: comment` and their text in the `value` field.

## Developing

### Prerequisites (Well, not really)

This project uses a [Taskfile](https://taskfile.dev) for _convenience_. Install by running:

```shell
go install tool
```

### Running Tests

Run all tests with:

```shell
task test
```

Test the specification specifically with:

```shell
task test-spec
```

Lint the stuff:

```shell
task lint
```
