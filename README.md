[![CI - Build and Test](https://github.com/hilli/cooklang/actions/workflows/ci.yaml/badge.svg)](https://github.com/hilli/cooklang/actions/workflows/ci.yaml)

# cooklang

Go implementation of a cooklang parser.

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
