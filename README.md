# npm-registry-mcp

A Model Context Protocol (MCP) server for NPM package analysis with health scoring, license risk assessment, and comprehensive package evaluation.

## Features

- **3 MCP Tools** for searching, inspecting, and evaluating NPM packages
- **Health Scoring** with weighted factors (maintenance, popularity, security, dependencies)
- **License Risk Assessment** using SPDX identifiers (Low/Medium/High/Critical)
- **GitHub Integration** for commit activity and repository health
- **In-Memory Caching** with 5-minute TTL for API responses

## Installation

### From Source

```bash
go install github.com/howmanysmall/npm-registry-mcp/src@latest
```

### From Releases

Download the latest binary from [GitHub Releases](https://github.com/howmanysmall/npm-registry-mcp/releases).

### Build Locally

```bash
git clone https://github.com/howmanysmall/npm-registry-mcp.git
cd npm-registry-mcp
go build -o npm-registry-mcp ./src
```

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | No | GitHub Personal Access Token for higher API rate limits (60/hr without, 5000/hr with) |

### .env File Support

Create a `.env` file in the working directory:

```env
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
```

## Usage with Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "npm-registry": {
      "command": "/path/to/npm-registry-mcp",
      "env": {
        "GITHUB_TOKEN": "ghp_xxxxxxxxxxxxxxxxxxxx"
      }
    }
  }
}
```

## Tools

### search-npm-packages

Search the NPM registry for packages.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query |
| `limit` | integer | No | Max results (1-100, default: 10) |

**Example:**

```json
{
  "query": "react",
  "limit": 5
}
```

### get-npm-package

Get detailed information about an NPM package.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Package name |

**Example:**

```json
{
  "name": "lodash"
}
```

**Returns:** Name, version, description, license, homepage, repository, maintainers, keywords, dependencies, and recent versions.

### should-i-install

Comprehensive health check for evaluating whether to install a package.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `package` | string | Yes | Package name to evaluate |

**Example:**

```json
{
  "package": "lodash"
}
```

**Returns:**

- `verdict`: `"yes"` | `"caution"` | `"no"`
- `score`: 0-100 health score
- `maintenance`: Last publish date and status
- `dependencies`: Direct, transitive, and outdated counts
- `security`: Vulnerability count
- `popularity`: Weekly downloads and trend
- `license`: SPDX identifier and risk level
- `warnings`: Array of concern messages

**Verdict Criteria:**

| Verdict | Criteria |
|---------|----------|
| `yes` | Score >= 70, no warnings, no vulnerabilities |
| `caution` | Score 40-69, or warnings present |
| `no` | Score < 40, or vulnerabilities present |

## Health Scoring Algorithm

| Factor | Weight | Description |
|--------|--------|-------------|
| Last Publish | 25% | Time since last release (100 pts if <=30 days) |
| Download Trend | 20% | Growth/decline in weekly downloads |
| Dependencies | 20% | Percentage of outdated dependencies |
| Commit Activity | 15% | Commits in last 90 days (requires GitHub token) |
| Maintainers | 10% | Number of active maintainers |
| Vulnerabilities | 10% | Known security vulnerabilities |

## License Risk Levels

| Risk | Examples | Description |
|------|----------|-------------|
| Low | MIT, Apache-2.0, BSD-3-Clause, ISC | Permissive, safe for any use |
| Medium | LGPL-3.0, MPL-2.0, EPL-2.0 | Weak copyleft, some restrictions |
| High | GPL-3.0, AGPL-3.0 | Strong copyleft, derivative works must share |
| Critical | SSPL-1.0, BUSL-1.1, UNLICENSED | Problematic, review with legal |

## Development

```bash
# Build
go build -o npm-registry-mcp ./src

# Test
go test -v -race ./...

# Lint
golangci-lint run ./...

# Integration tests (requires network)
go test -tags=integration -v ./src
```

## Building & Releasing

### Local Build

```bash
go build -o npm-registry-mcp ./src
```

### Cross-Platform Build (via GoReleaser)

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build snapshot (no publish)
goreleaser build --snapshot --clean

# Binaries output to dist/
```

### Creating a Release

```bash
# 1. Commit all changes
git add .
git commit -m "feat: your changes"

# 2. Create version tag
git tag v0.1.0

# 3. Push to GitHub
git push origin main
git push origin v0.1.0
```

The release workflow triggers automatically on `v*` tags and:

- Builds binaries for Linux, macOS, Windows (amd64 + arm64)
- Creates GitHub release with checksums
- Generates changelog from commits

### Using svu for Versioning (Optional)

```bash
# Install svu
go install github.com/caarlos0/svu@latest

# Get next version based on commit messages
svu next

# Tag and push
git tag $(svu next)
git push origin $(svu next)
```

## License

MIT
