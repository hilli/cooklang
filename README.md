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

## Cooklang specification

See the [Cooklang specification](https://github.com/cooklang/spec/) for details.

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
