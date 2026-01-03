// Package tools provides MCP tool handlers for NPM registry operations.
package tools

import (
	"context"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchInput is the input for the search-npm-packages tool
type SearchInput struct {
	Query string `json:"query" jsonschema:"search query for NPM packages"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of results (1-100), default 10"`
}

// SearchOutput is the output for the search-npm-packages tool
type SearchOutput struct {
	Packages []PackageSummary `json:"packages"`
	Total    int              `json:"total"`
}

// PackageSummary is a summary of a package from search results
type PackageSummary struct {
	Name            string  `json:"name"`
	Version         string  `json:"version"`
	Description     string  `json:"description"`
	WeeklyDownloads int     `json:"weeklyDownloads,omitempty"`
	Score           float64 `json:"score"`
}

// SearchHandler is the handler type for search
type SearchHandler = func(context.Context, *mcp.CallToolRequest, SearchInput) (*mcp.CallToolResult, SearchOutput, error)

// NewSearchHandler creates a new search handler
func NewSearchHandler(client *npm.Client) SearchHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
		// Default limit
		limit := input.Limit
		if limit <= 0 {
			limit = 10
		}

		if limit > 100 {
			limit = 100
		}

		result, err := client.Search(ctx, input.Query, limit)
		if err != nil {
			return nil, SearchOutput{}, err
		}

		packages := make([]PackageSummary, 0, len(result.Objects))
		for _, obj := range result.Objects {
			packages = append(packages, PackageSummary{
				Name:            obj.Package.Name,
				Version:         obj.Package.Version,
				Description:     obj.Package.Description,
				WeeklyDownloads: obj.Downloads.Weekly,
				Score:           obj.Score.Final,
			})
		}

		return nil, SearchOutput{
			Packages: packages,
			Total:    result.Total,
		}, nil
	}
}

// SearchTool returns the tool definition for search-npm-packages
func SearchTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "search-npm-packages",
		Description: "Search the NPM registry for packages matching a query",
	}
}
