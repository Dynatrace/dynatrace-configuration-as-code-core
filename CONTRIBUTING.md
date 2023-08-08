# Contributing to the Dynatrace Configuration as Code Core Libraries

- [Contributing to the Dynatrace Configuration as Code Core Libraries](#contributing-to-the-dynatrace-configuration-as-code-core-libraries)
  - [What to contribute](#what-to-contribute)
  - [How to contribute](#how-to-contribute)
    - [Examples of Commit Style Messages](#examples-of-commit-style-messages)
  - [Code of Conduct and Shared Values](#code-of-conduct-and-shared-values)
  - [Building the Dynatrace Configuration as Code Core libraries](#building-the-dynatrace-configuration-as-code-core-libraries)
  - [Testing the Dynatrace Configuration as Code Core libraries](#testing-the-dynatrace-configuration-as-code-core-libraries)
    - [Writing Tests](#writing-tests)
  - [Checking in go mod and sum files](#checking-in-go-mod-and-sum-files)
  - [General information on code](#general-information-on-code)
    - [Test Mocks](#test-mocks)
    - [Formatting](#formatting)
  - [Pre-Commit Hook](#pre-commit-hook)


## What to contribute

Dynatrace Configuration as Code Core provides libraries simplifying the development of configuration as code tooling for Dynatrace.

It provides Go libraries for things like API clients, which are shared between several Dynatrace configuration as code tools.

## How to contribute

The easiest way to start contributing or helping with the Configuration as Code Core project is to pick an existing [issue/bug](https://github.com/dynatrace/dynatrace-configuration-as-code-core/issues) and [get to work](#building-the-dynatrace-configuration-as-code-core-libraries).

For proposing a change, we seek to discuss potential changes in GitHub issues in advance before implementation. 
That will allow us to give design feedback up front and set expectations about the scope of the change and, for more significant changes, 
how best to approach the work such that the Configuration as Code team can review it and merge it with other concurrent work. 
This allows being respectful of the time of community contributors.

The repo follows a relatively standard branching & PR workflow.

Branches naming follows the `feature/{Issue}/{description}` or `bugfix/{Issue}/{description}` pattern.

Branches are rebased, and only fast-forward merges to main are permitted. No merge commits.

By default, commits are not auto-squashed when merging a PR, so please ensure your commits are fit to go into main.

For convenience auto-squashing all PR commits into a single one is an optional merge strategy - but we strive for [atomic commits](https://www.freshconsulting.com/insights/blog/atomic-commits/)
with [good commit messages](https://cbea.ms/git-commit/) in main so not auto-squashing is recommended.

Commits should conform to  [Conventional Commit](https://www.conventionalcommits.org/) standard.

### Examples of Commit Style Messages

New Feature Changes
``` 
feat: allow provided config object to extend other configs
```

Bug Fix Changes
```
fix: change function call

see the issue for details

on typos fixed.

Reviewed-by: Z
Refs #133 
```

Documentation Changes
```
docs: correct getting started guide 
```

More examples can be found [here](https://www.conventionalcommits.org/en/v1.0.0/#examples)


## Code of Conduct and Shared Values

Before contributing, please read and approve [our Code Of Conduct](https://github.com/dynatrace/dynatrace-configuration-as-code-core/blob/main/CODE_OF_CONDUCT.md) outlining our shared values and expectations. 

## Building the Dynatrace Configuration as Code Core libraries

The libraries are written in [Go](https://golang.org/), so you will need to have [installed Go](https://golang.org/dl/) to build it.

As the libraries themselves provide no exectuable you may compile them using `make compile`, but there are no executables to be built and used.

Generally make sure to take a look at the [Makefile](./Makefile) to see available build targets.

## Testing the Dynatrace Configuration as Code Core libraries

Run the unit tests for the whole module with `make test` in the root folder.

### Writing Tests

Take a look at [Go Testing](https://golang.org/pkg/testing/) for more info on testing in Go.

We use [github.com/stretchr/testify](github.com/stretchr/testify) as a testing/assert library.

In general, we aim for test coverage above 80% but do not enfore hard limits, please make sure your code is reasonably well covered by tests.

## Checking in go mod and sum files

Go module files `go.mod` and `go.sum` are checked-in in the root folder of the repo, so generally run `go` from there.

`mod` and `sum` may change while building the project.
To keep those files clean off unnecessary changes, please always run `go mod tidy` before committing changes to these files!

## General information on code

### Test Mocks

Go Mockgen is used for some generated mock files.
You will have to generate them.
To explicitly generate the mocked files, run `make mocks` in the root folder.

### Formatting

This project uses the default go formatting tool `go fmt`.

To format all files, you can use the Make target `make format`.

## Pre-Commit Hook

Before committing changes, please make sure you've added the `pre-commit` hook from the hooks folder.
On Unix, you can use the `setup-git-hooks.sh` to symlink that file into your `.git/hooks` folder.
