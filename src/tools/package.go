package tools

import (
	"context"
	"sort"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PackageInput is the input for the get-npm-package tool
type PackageInput struct {
	Name string `json:"name" jsonschema:"NPM package name"`
}

// PackageOutput is the output for the get-npm-package tool
type PackageOutput struct {
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Description    string            `json:"description"`
	License        string            `json:"license"`
	Homepage       string            `json:"homepage,omitempty"`
	Repository     string            `json:"repository,omitempty"`
	Maintainers    []string          `json:"maintainers"`
	Keywords       []string          `json:"keywords,omitempty"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
	RecentVersions []VersionInfo     `json:"recentVersions"`
}

// VersionInfo contains version metadata
type VersionInfo struct {
	Version     string `json:"version"`
	PublishedAt string `json:"publishedAt"`
}

// PackageHandler is the handler type for get-npm-package
type PackageHandler = func(context.Context, *mcp.CallToolRequest, PackageInput) (*mcp.CallToolResult, PackageOutput, error)

// NewPackageHandler creates a new package handler
func NewPackageHandler(client *npm.Client) PackageHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input PackageInput) (*mcp.CallToolResult, PackageOutput, error) {
		pkg, err := client.GetPackage(ctx, input.Name)
		if err != nil {
			return nil, PackageOutput{}, err
		}

		// Get latest version
		latestVersion := pkg.DistTags["latest"]
		latestMeta := pkg.Versions[latestVersion]

		// Extract maintainers
		maintainers := make([]string, 0, len(pkg.Maintainers))
		for _, m := range pkg.Maintainers {
			name := m.Name
			if name == "" {
				name = m.Username
			}

			maintainers = append(maintainers, name)
		}

		// Get repository URL
		var repoURL string
		if pkg.Repository != nil {
			repoURL = pkg.Repository.URL
		}

		// Get recent versions (last 10)
		versions := make([]VersionInfo, 0, len(pkg.Time))
		for v, t := range pkg.Time {
			if v == "created" || v == "modified" {
				continue
			}

			versions = append(versions, VersionInfo{
				Version:     v,
				PublishedAt: t,
			})
		}

		// Sort by date descending
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].PublishedAt > versions[j].PublishedAt
		})

		// Limit to 10
		if len(versions) > 10 {
			versions = versions[:10]
		}

		return nil, PackageOutput{
			Name:           pkg.Name,
			Version:        latestVersion,
			Description:    pkg.Description,
			License:        pkg.License,
			Homepage:       pkg.Homepage,
			Repository:     repoURL,
			Maintainers:    maintainers,
			Keywords:       pkg.Keywords,
			Dependencies:   latestMeta.Dependencies,
			RecentVersions: versions,
		}, nil
	}
}

// PackageTool returns the tool definition for get-npm-package
func PackageTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get-npm-package",
		Description: "Get detailed information about an NPM package",
	}
}
