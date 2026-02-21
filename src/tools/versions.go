package tools

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
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
func NewVersionsHandler(client *npm.Client, appCache *cache.Cache) VersionsHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input VersionsInput) (*mcp.CallToolResult, VersionsOutput, error) {
		limit := input.Limit
		if limit <= 0 {
			limit = 100
		}

		if limit > 1000 {
			limit = 1000
		}

		// Check cache
		cacheKey := fmt.Sprintf("versions:%s:%d", input.Name, limit)
		if appCache != nil {
			if cached, found := cache.Get[VersionsOutput](appCache, cacheKey); found {
				return nil, cached, nil
			}
		}

		pkg, err := client.GetAbbreviatedPackage(ctx, input.Name)
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

		output := VersionsOutput{
			Name:          pkg.Name,
			TotalVersions: len(allVersions),
			VersionsShown: len(limitedVersions),
			LatestVersion: pkg.DistTags["latest"],
			Versions:      limitedVersions,
		}

		// Store in cache
		if appCache != nil {
			appCache.Set(cacheKey, output)
		}

		return nil, output, nil
	}
}

func compareSemver(a, b string) int {
	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)

	maxLen := max(len(bParts), len(aParts))

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

// TagsInput is the input for the list-package-tags tool
type TagsInput struct {
	Name string `json:"name" jsonschema:"NPM package name"`
}

// TagsOutput is the output for the list-package-tags tool
type TagsOutput struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
}

// TagsHandler is the handler type for list-package-tags
type TagsHandler = func(context.Context, *mcp.CallToolRequest, TagsInput) (*mcp.CallToolResult, TagsOutput, error)

// NewTagsHandler creates a new tags handler
func NewTagsHandler(client *npm.Client, appCache *cache.Cache) TagsHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input TagsInput) (*mcp.CallToolResult, TagsOutput, error) {
		// Check cache
		cacheKey := fmt.Sprintf("tags:%s", input.Name)
		if appCache != nil {
			if cached, found := cache.Get[TagsOutput](appCache, cacheKey); found {
				return nil, cached, nil
			}
		}

		pkg, err := client.GetAbbreviatedPackage(ctx, input.Name)
		if err != nil {
			return nil, TagsOutput{}, err
		}

		output := TagsOutput{
			Name: pkg.Name,
			Tags: pkg.DistTags,
		}

		// Store in cache
		if appCache != nil {
			appCache.Set(cacheKey, output)
		}

		return nil, output, nil
	}
}

// TagsTool returns the tool definition for list-package-tags
func TagsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list-package-tags",
		Description: "List all distribution tags (e.g., latest, beta) for an NPM package",
	}
}
