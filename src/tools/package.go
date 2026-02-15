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

// PackageOutput is the output for the get-npm-package-details tool
type PackageOutput struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	LatestVersion    string            `json:"latestVersion"`
	License          string            `json:"license"`
	Homepage         string            `json:"homepage,omitempty"`
	Repository       string            `json:"repository,omitempty"`
	Maintainers      []string          `json:"maintainers"`
	Keywords         []string          `json:"keywords,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	DevDependencies  map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
	Engines          map[string]string `json:"engines,omitempty"`
	Versions         []string          `json:"versions"`
	TotalVersions    int               `json:"totalVersions"`
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

		latestVersion := pkg.DistTags["latest"]
		latestMeta := pkg.Versions[latestVersion]

		maintainers := make([]string, 0, len(pkg.Maintainers))
		for _, m := range pkg.Maintainers {
			name := m.Name
			if name == "" {
				name = m.Username
			}

			maintainers = append(maintainers, name)
		}

		var repoURL string
		if pkg.Repository != nil {
			repoURL = pkg.Repository.URL
		}

		// Get all versions sorted by semver descending
		allVersions := make([]string, 0, len(pkg.Versions))
		for v := range pkg.Versions {
			allVersions = append(allVersions, v)
		}

		sort.Slice(allVersions, func(i, j int) bool {
			return compareSemver(allVersions[i], allVersions[j]) > 0
		})

		// Limit to 50
		versions := allVersions
		if len(allVersions) > 50 {
			versions = allVersions[:50]
		}

		return nil, PackageOutput{
			Name:             pkg.Name,
			Description:      pkg.Description,
			LatestVersion:    latestVersion,
			License:          string(pkg.License),
			Homepage:         pkg.Homepage,
			Repository:       repoURL,
			Maintainers:      maintainers,
			Keywords:         pkg.Keywords,
			Dependencies:     latestMeta.Dependencies,
			DevDependencies:  latestMeta.DevDependencies,
			PeerDependencies: latestMeta.PeerDeps,
			Engines:          latestMeta.Engines,
			Versions:         versions,
			TotalVersions:    len(allVersions),
		}, nil
	}
}

// PackageTool returns the tool definition for get-npm-package-details
func PackageTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get-npm-package-details",
		Description: "Get detailed information about a specific NPM package",
	}
}
