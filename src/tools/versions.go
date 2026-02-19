package tools

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// VersionsInput is the input for the list-npm-package-versions tool
type VersionsInput struct {
	Name  string `json:"name" jsonschema:"NPM package name"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of versions (1-1000), default 100"`
}

// VersionsOutput is the output for the list-npm-package-versions tool
type VersionsOutput struct {
	Name          string   `json:"name"`
	TotalVersions int      `json:"totalVersions"`
	VersionsShown int      `json:"versionsShown"`
	LatestVersion string   `json:"latestVersion"`
	Versions      []string `json:"versions"`
}

// VersionsHandler is the handler type for list-npm-package-versions
type VersionsHandler = func(context.Context, *mcp.CallToolRequest, VersionsInput) (*mcp.CallToolResult, VersionsOutput, error)

// NewVersionsHandler creates a new versions handler
func NewVersionsHandler(client *npm.Client) VersionsHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input VersionsInput) (*mcp.CallToolResult, VersionsOutput, error) {
		limit := input.Limit
		if limit <= 0 {
			limit = 100
		}

		if limit > 1000 {
			limit = 1000
		}

		pkg, err := client.GetPackage(ctx, input.Name)
		if err != nil {
			return nil, VersionsOutput{}, err
		}

		allVersions := make([]string, 0, len(pkg.Versions))
		for v := range pkg.Versions {
			allVersions = append(allVersions, v)
		}

		sort.Slice(allVersions, func(i, j int) bool {
			return compareSemver(allVersions[i], allVersions[j]) > 0
		})

		limitedVersions := allVersions
		if len(allVersions) > limit {
			limitedVersions = allVersions[:limit]
		}

		return nil, VersionsOutput{
			Name:          pkg.Name,
			TotalVersions: len(allVersions),
			VersionsShown: len(limitedVersions),
			LatestVersion: pkg.DistTags["latest"],
			Versions:      limitedVersions,
		}, nil
	}
}

func compareSemver(a, b string) int {
	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := range maxLen {
		var aPart, bPart int
		if i < len(aParts) {
			aPart = aParts[i]
		}

		if i < len(bParts) {
			bPart = bParts[i]
		}

		if aPart != bPart {
			return aPart - bPart
		}
	}

	return 0
}

func parseSemverParts(version string) []int {
	version = strings.TrimPrefix(version, "v")

	parts := strings.FieldsFunc(version, func(r rune) bool {
		return r == '.' || r == '-' || r == '+'
	})

	result := make([]int, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(p); err == nil {
			result = append(result, n)
		}
	}

	return result
}

// VersionsTool returns the tool definition for list-package-versions
func VersionsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list-package-versions",
		Description: "List all versions of a specific NPM package",
	}
}
