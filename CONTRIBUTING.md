# Contributing to Cooklang Go

Thank you for your interest in contributing to the Cooklang Go parser! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.24 or later
- [Task](https://taskfile.dev/) (optional, but recommended)

### Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/cooklang.git
   cd cooklang
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```

## Development Workflow

### Running Tests

```bash
# Using Task (recommended)
task test

# Or directly with Go
go test ./...
```

### Running Linter

```bash
# Using Task (recommended)
task lint

# Or directly with golangci-lint
golangci-lint run
```

### Building

```bash
# Build the library
go build ./...

# Build the CLI
go build -o bin/cook ./cmd/cook
```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/description` - for new features
- `fix/description` - for bug fixes
- `docs/description` - for documentation changes

### Code Style

- Follow standard Go conventions
- Run `gofmt` or `goimports` on your code
- Ensure `golangci-lint` passes without errors
- Add godoc comments for all exported types, functions, and methods

### Testing Requirements

- Add tests for new functionality
- Ensure all existing tests pass
- Aim to maintain or improve code coverage

### Commit Messages

Write clear, concise commit messages:
- Use the imperative mood ("Add feature" not "Added feature")
- Keep the first line under 72 characters
- Reference issues when applicable (e.g., "Fix #123")

Examples:
```
feat: add support for recipe notes
fix: handle empty ingredient quantities
docs: update README with new CLI options
test: add tests for unit conversion edge cases
```

## Pull Request Process

1. **Create a feature branch** from `main`
2. **Make your changes** following the guidelines above
3. **Run tests and linter** to ensure everything passes:
   ```bash
   task test
   task lint
   ```
4. **Push your branch** and create a Pull Request
5. **Fill out the PR template** with a clear description of your changes
6. **Wait for review** - maintainers will review your PR and may request changes

### PR Checklist

Before submitting, ensure:
- [ ] Tests pass (`task test` or `go test ./...`)
- [ ] Linter passes (`task lint` or `golangci-lint run`)
- [ ] New code has appropriate test coverage
- [ ] Documentation is updated if needed
- [ ] Commit messages are clear and descriptive

## Reporting Issues

### Bug Reports

When reporting bugs, please include:
- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Sample recipe file if applicable

### Feature Requests

When requesting features:
- Describe the use case
- Explain how it relates to the Cooklang specification
- Provide examples if possible

## Code of Conduct

Be respectful and constructive in all interactions. We're all here to build great software together.

## Questions?

If you have questions about contributing, feel free to open an issue for discussion.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
