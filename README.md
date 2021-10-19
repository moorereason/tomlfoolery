# TOMLFoolery

This proof-of-concept package provides a rudimentary framework for comparing
TOML decoder implementations' handling of fuzzing input.

## Requirements

- A Go 1.18 development branch or [gotip](https://pkg.go.dev/golang.org/dl/gotip)
  environment is needed since we're using the new [integrated fuzzing
  capabilities](https://go.dev/blog/fuzz-beta).

- Each decoder must implement the [toml-test](https://github.com/BurntSushi/toml-test/) decoder interface.

- The fuzzing function requires two environment variables, `TOML_A` and `TOML_B`, to be set to the paths to two decoder executables.

## Usage

```
TOML_A=/path/to/toml-a TOML_B=/path/to/toml-b gotip test -fuzz=.
```

## Fuzzing

We currently have use a single `FuzzUnmarshal` function.

Failed tests will be written to `testdata/fuzz/FuzzUnmarshal`.  Once a failed
test case is created, it will be used upon each subsequent run of the tests.  If
you want to throw away a fuzz case, simply the delete the test file.

### Seeding with toml-test

The `FuzzUnmarshal` function has a `addTomlTestCases` boolean that controls
whether the tests cases from `toml-test` will be added as seeds to the fuzzer.

## Thanks

A special thanks goes out to Martin Tournoij for creating the
[toml-test](https://github.com/BurntSushi/toml-test/) framework, which made this
proof-of-concept a simple endeavor.
