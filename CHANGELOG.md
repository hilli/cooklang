# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2026-01-25

### Changed
- Bartender mode metric rounding threshold increased from 10ml to 30ml
  - Amounts < 30ml now round to nearest 2.5ml (was 5ml for amounts >= 10ml)
  - Amounts >= 30ml continue to round to nearest 5ml
  - This provides more precision for typical cocktail measurements (e.g., 22.5ml stays as 22.5ml instead of becoming 25ml)

## [1.0.1] - 2026-01-12

### Fixed
- HTML output from `render` and `scale` commands now includes complete document structure with `<meta charset="UTF-8">` for proper emoji rendering

### Security
- Updated `golang.org/x/crypto` from v0.37.0 to v0.46.0 (fixes CVE-2025-22869, CVE-2025-22872)
- Updated `github.com/go-viper/mapstructure/v2` from v2.2.1 to v2.4.0 (fixes sensitive data logging issues)
- Updated `golangci-lint` from v2.1.6 to v2.8.0
- Updated `go-task` from v3.43.3 to v3.46.4

## [1.0.0] - 2026-01-11

### Added
- `CreateShoppingListForServings()` to create shopping lists scaled to target servings
- `CreateShoppingListForServingsWithUnit()` for combined servings scaling and unit conversion
- `Recipe.ScaleToServings()` method for convenient servings-based scaling
- Comprehensive Go doc comments for all public API functions
- New examples: `ExampleRecipe_Scale`, `ExampleRecipe_ScaleToServings`, `ExampleCreateShoppingListForServings`
- Known Usages section in README

### Changed
- `Recipe.Scale()` now respects the `Fixed` flag on ingredients (won't scale fixed quantities)
- `ConsolidateByName()` now converts single ingredients to the target unit when specified
- Improved documentation for array field handling in FrontmatterEditor (tags, images)
- Minimum Go version set to 1.24

### Fixed
- Shopping list `--servings` CLI flag now properly scales recipes before consolidation

## [0.4.0] - 2026-01-09

### Added
- Full Cooklang specification v7 compliance
- Fixed quantities support with `=` prefix (e.g., `@salt{=1%tsp}`)
- Note blocks per Cooklang spec proposal (`> Note text`)
- Canonical extensions spec tests for block comments, sections, and notes
- Comprehensive unit tests for token package
- Unit tests for comments, sections, and notes

## [0.3.1] - 2026-01-08

### Fixed
- Parse comments, block comments, and sections after newlines correctly

## [0.3.0] - 2026-01-08

### Added
- `Recipe.Scale()` method for scaling recipes to different serving sizes
- Block comments support (`[- comment -]`) per Cooklang spec
- Section headers (`=== Section Name ===`) for organizing recipe steps
- Improved comment rendering in all output formats

### Changed
- Aligned CLI commands with library features

## [0.2.2] - 2025-12-28

### Fixed
- Include unit in `Timer.RenderDisplay()` output

## [0.2.1] - 2025-12-27

### Fixed
- Remove "some" prefix from `RenderDisplay()` for unspecified quantities

## [0.2.0] - 2025-12-25

### Added
- Bartender mode for cocktail-friendly unit conversions
- `NewIngredient` constructor and exported `CreateTypedUnit` function
- Print-optimized HTML renderer for single-page recipe output
- Multi-line YAML block scalar support in frontmatter
- Windows CRLF and old Mac line ending support
- `RenderDisplay` methods for user-friendly text rendering
- Comprehensive documentation for all public methods
- `GetCookware()` method on Recipe type
- Unicode fraction character support in measurements (e.g., ½, ¼)
- Complex fraction handling (e.g., "1 1/2")
- Image detection for recipes
- Frontmatter CRUD operations
- Unit conversion system (metric, imperial, US)
- Multiple renderers (Cooklang, Markdown, HTML)
- Extended spec mode with multi-word timers, annotations, and comment preservation
- CLI tool with parse, render, scale, ingredients, and shopping-list commands
- Shopping list generation with automatic consolidation
- GitHub Actions CI workflows for testing and linting

### Changed
- Improved recipe data structure with linked-list step/component design

[1.0.1]: https://github.com/hilli/cooklang/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/hilli/cooklang/compare/v0.4.0...v1.0.0
[0.4.0]: https://github.com/hilli/cooklang/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/hilli/cooklang/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/hilli/cooklang/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/hilli/cooklang/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/hilli/cooklang/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/hilli/cooklang/releases/tag/v0.2.0
