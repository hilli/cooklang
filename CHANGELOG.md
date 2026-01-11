# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Updated dependencies (go-yaml v1.19.2, cobra v1.10.2)

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

[Unreleased]: https://github.com/hilli/cooklang/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/hilli/cooklang/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/hilli/cooklang/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/hilli/cooklang/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/hilli/cooklang/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/hilli/cooklang/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/hilli/cooklang/releases/tag/v0.2.0
