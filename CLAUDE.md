# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build -v ./...                    # Build all packages
go test -v -race ./...               # Run tests with race detector
go test -v -race -coverprofile=coverage.txt ./...  # With coverage
golangci-lint run ./...              # Lint (5min timeout)
```

Run a single test:

```bash
go test -v -race ./src/npm -run TestSearch
```

## Architecture

This is an MCP (Model Context Protocol) server that exposes NPM registry analysis tools to AI models.

### Entry Point

`src/main.go` - Initializes clients, registers tools, runs stdio transport.

### MCP Tools

- **search-npm-packages** - Search NPM registry with text query
- **get-npm-package** - Get detailed package metadata
- **should-i-install** - Comprehensive health assessment with verdict (yes/caution/no)

### Package Structure

| Package | Purpose |
|---------|---------|
| `src/npm` | NPM registry API client |
| `src/github` | GitHub API client for repo metadata |
| `src/tools` | MCP tool definitions and handlers |
| `src/health` | Package health scoring algorithm |
| `src/license` | License risk classification |
| `src/cache` | Type-safe generic cache wrapper |

### Key Patterns

- **Functional options** for client configuration (see `npm.NewClient()`, `github.NewClient()`)
- **Generics** in cache package for type-safe retrieval
- All API methods accept `context.Context` for cancellation

## Environment

- `GITHUB_TOKEN` - Optional. Authenticated: 5000 req/hour, unauthenticated: 60 req/hour
- `.env` file auto-loaded via godotenv

## Linting

Uses golangci-lint with gofumpt formatter. Key enabled linters: govet, staticcheck, errcheck, gosec, errorlint, bodyclose, paralleltest. See `.golangci.json` for full config.
