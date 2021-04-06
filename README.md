# Go Repository Template for Google Cloud Run

[![codecov](https://codecov.io/gh/dev-templates/go-cloud-run/branch/main/graph/badge.svg)](https://codecov.io/gh/dev-templates/go-cloud-run)
[![Build Status](https://github.com/dev-templates/go-cloud-run/workflows/build/badge.svg)](https://github.com/dev-templates/go-cloud-run)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/dev-templates/go-cloud-run)
[![go.mod](https://img.shields.io/github/go-mod/go-version/dev-templates/go-cloud-run)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/dev-templates/go-cloud-run)](https://goreportcard.com/report/github.com/dev-templates/go-cloud-run)
[![LICENSE](https://img.shields.io/github/license/dev-templates/go-cloud-run)](LICENSE)

This is a GitHub repository template for Go

It includes:

- continous integration via [GitHub Actions](https://github.com/features/actions),
- dependency management using [Go Modules](https://github.com/golang/go/wiki/Modules),
- code formatting using [gofumpt](https://github.com/mvdan/gofumpt),
- linting with [golangci-lint](https://github.com/golangci/golangci-lint),
- [Codecov report](https://codecov.io/),
- dependencies scanning and updating thanks to [Dependabot](https://dependabot.com),
- [Visual Studio Code](https://code.visualstudio.com) configuration with [Go](https://code.visualstudio.com/docs/languages/go)

## Usage

1. Sign up on [Codecov](https://codecov.io/) and configure [Codecov GitHub Application](https://github.com/apps/codecov) for all repositories.
2. Click the `Use this template` button (alt. clone or download this repository).
3. Replace all occurences of `dev-templates/go-cloud-run` to `your_org/repo_name` in all files.
4. Rename folder `cmd/go-cloud-run` to `cmd/app_name`.
5. Replace `gcr.io/app/server` to `gcr.io/<PROJECT_ID>/<IMAGE_ID>/repo_name` in `docker-compose.yml`.
5. Update [LICENSE](LICENSE) and [README.md](README.md).

## Build

- Visual Studio Code: `Terminal` → `Run Build Task... (Ctrl+Shift+P => Tasks: Run Task)` to execute a fast build or run docker-compose.

## Maintainance

Notable files:
- [.github/workflows](.github/workflows) - GitHub Actions workflows,
- [.github/dependabot.yml](.github/dependabot.yml) - Dependabot configuration,
- [.vscode](.vscode) - Visual Studio Code configuration files,
- [go.mod](go.mod) - [Go module definition](https://github.com/golang/go/wiki/Modules#gomod),
