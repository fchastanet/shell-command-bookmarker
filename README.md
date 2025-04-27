# Bash Tools Command Bookmarker

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)

> **_TIP:_** Checkout related projects of this suite
>
> - [My documents](https://fchastanet.github.io/my-documents/)
> - [Bash Tools Framework](https://fchastanet.github.io/bash-tools-framework/)
> - [Bash Tools](https://fchastanet.github.io/bash-tools/)
> - [Bash Dev Env](https://fchastanet.github.io/bash-dev-env/)
> - [Bash Compiler](https://fchastanet.github.io/bash-compiler/)
> - [Bash Tools Command Bookmarker](https://fchastanet.github.io/bash-tools-command-bookmarker/)

<!-- markdownlint-capture -->

<!-- markdownlint-disable MD013 -->

[![GitHub release (latest SemVer)](https://img.shields.io/github/release/fchastanet/bash-tools-command-bookmarker?logo=github&sort=semver)](https://github.com/fchastanet/bash-tools-command-bookmarker/releases)
[![GitHubLicense](https://img.shields.io/github/license/Naereen/StrapDown.js.svg)](https://github.com/fchastanet/bash-tools-command-bookmarker/blob/master/LICENSE)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![CI/CD](https://github.com/fchastanet/bash-tools-command-bookmarker/actions/workflows/main.yml/badge.svg)](https://github.com/fchastanet/bash-tools-command-bookmarker/actions?query=workflow%3A%22Lint+and+test%22+branch%3Amaster)
[![ProjectStatus](http://opensource.box.com/badges/active.svg)](http://opensource.box.com/badges "Project Status")
[![DeepSource](https://deepsource.io/gh/fchastanet/bash-tools-command-bookmarker.svg/?label=active+issues&show_trend=true)](https://deepsource.io/gh/fchastanet/bash-tools-command-bookmarker/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/fchastanet/bash-tools-command-bookmarker.svg/?label=resolved+issues&show_trend=true)](https://deepsource.io/gh/fchastanet/bash-tools-command-bookmarker/?ref=repository-badge)
[![AverageTimeToResolveAnIssue](http://isitmaintained.com/badge/resolution/fchastanet/bash-tools-command-bookmarker.svg)](http://isitmaintained.com/project/fchastanet/bash-tools-command-bookmarker "Average time to resolve an issue")
[![PercentageOfIssuesStillOpen](http://isitmaintained.com/badge/open/fchastanet/bash-tools-command-bookmarker.svg)](http://isitmaintained.com/project/fchastanet/bash-tools-command-bookmarker "Percentage of issues still open")

<!-- markdownlint-restore -->

- [1. Excerpt](#1-excerpt)
- [2. Documentation](#2-documentation)
  - [2.1. Go Libraries used](#21-go-libraries-used)
- [3. Development](#3-development)
  - [3.1. Necessary tools](#31-necessary-tools)
  - [3.2. Pre-commit hook](#32-pre-commit-hook)
  - [3.3. pre-commit external tools install](#33-pre-commit-external-tools-install)
  - [3.4. detect dead code](#34-detect-dead-code)
  - [3.5. Build/run/clean](#35-buildrunclean)
    - [3.5.1. Build](#351-build)
    - [3.5.2. Tests](#352-tests)
    - [3.5.3. Coverage](#353-coverage)
    - [3.5.4. run the binary](#354-run-the-binary)
    - [3.5.5. Clean](#355-clean)
- [4. Commands](#4-commands)

## 1. Excerpt

> [!WARNING]
>
> **Development in progress, not functional yet !**

![application preview](doc/preview.png)

This tool provides a terminal-based user interface (TUI) for managing and
organizing shell commands. It allows users to:

- Save frequently used shell commands as bookmarks
- Categorize commands with tags
- Search through saved commands quickly
- Execute bookmarked commands directly from the interface

The application uses the Bubbletea framework to create an interactive terminal
UI with features like:

- Tab-based navigation
- Keyboard shortcuts
- Focus management between different UI components
- Command organization and filtering

This tool is part of a larger suite of Bash productivity tools designed to
enhance shell workflows and command management.

## 2. Documentation

### 2.1. Go Libraries used

- [slog](https://pkg.go.dev/golang.org/x/exp/slog) is logging system
  - [slog tutorial](https://betterstack.com/community/guides/logging/logging-in-go/#customizing-the-default-logger)
- [Bubbletea](https://github.com/charmbracelet/bubbletea) A powerful little TUI
  framework.
- Not a library, but a lot of snippets, ui logic and design have been taken from
  [PUG - A terminal user interface for terraform power users](https://github.com/leg100/pug).
- snippets from
  [Brandon Fulljames](https://github.com/Evertras/bubble-table/blob/main/table/dimensions.go)

## 3. Development

### 3.1. Necessary tools

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

### 3.2. Pre-commit hook

This repository uses pre-commit software to ensure every commits respects a set
of rules specified by the `.pre-commit-config.yaml` file. It supposes pre-commit
software is [installed](https://pre-commit.com/#install) in your environment.

You also have to execute the following command to enable it:

```bash
pre-commit install --hook-type pre-commit --hook-type pre-push
```

Now each time you commit or push, some linters/compilation tools are launched
automatically

### 3.3. pre-commit external tools install

```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/OpenPeeDeeP/depguard/cmd/depguard@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/dkorunic/betteralign/cmd/betteralign@latest
go install github.com/go-critic/go-critic/cmd/go-critic@latest
go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
```

### 3.4. detect dead code

```bash
go install golang.org/x/tools/cmd/deadcode@latest
deadcode -filter "github.com/fchastanet/shell-command-bookmarker" ./app/main.go
```

### 3.5. Build/run/clean

Formatting is managed exclusively by pre-commit hooks.

#### 3.5.1. Build

```bash
.build/build-docker.sh
```

```bash
.build/build-local.sh
```

#### 3.5.2. Tests

```bash
.build/test.sh
```

#### 3.5.3. Coverage

```bash
.build/coverage.sh
```

#### 3.5.4. run the binary

```bash
.build/run.sh
```

#### 3.5.5. Clean

```bash
.build/clean.sh
```

## 4. Commands

Run the project

```bash
go run ./main
```
