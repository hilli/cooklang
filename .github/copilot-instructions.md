We strongly prefer that Go tests are written in a way that they can be run with `go test` without any additional setup. This means that tests should not require any external files or resources, and should be self-contained.

Adher to the specification by running the following command to test the implementation against the Cooklang specification:

```shell
task test-spec
```

```shell
task test # Run all tests
go test -v ./... # Run all tests with verbose output
```

We are using a [taskfile](../Taskfile.yaml) as a makefile replacement to execute tests.
