package tools

import (
	"context"
	"fmt"
	"sort"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
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
func NewPackageHandler(client *npm.Client, appCache *cache.Cache) PackageHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input PackageInput) (*mcp.CallToolResult, PackageOutput, error) {
		// Check cache
		cacheKey := fmt.Sprintf("package:%s", input.Name)
		if appCache != nil {
			if cached, found := cache.Get[PackageOutput](appCache, cacheKey); found {
				return nil, cached, nil
			}
		}

		pkg, err := client.GetPackage(ctx, input.Name)
		if err != nil {
			return nil, PackageOutput{}, err
		}

		latestVersion := pkg.DistTags["latest"]
		latestMeta := pkg.Versions[latestVersion]

		maintainers := getMaintainerNames(pkg.Maintainers)

		var repoURL string
		if pkg.Repository != nil {
			repoURL = pkg.Repository.URL
		}

		versions, total := getRecentVersions(pkg.Versions)

		output := PackageOutput{
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
			TotalVersions:    total,
		}

		// Store in cache
		if appCache != nil {
			appCache.Set(cacheKey, output)
		}

		return nil, output, nil
	}
}

func getMaintainerNames(maintainers []npm.Maintainer) []string {
	names := make([]string, 0, len(maintainers))
	for _, m := range maintainers {
		name := m.Name
		if name == "" {
			name = m.Username
		}

		names = append(names, name)
	}

	return names
}

func getRecentVersions(versions map[string]npm.PackageVersion) ([]string, int) {
	// Get all versions sorted by semver descending
	allVersions := make([]string, 0, len(versions))
	for v := range versions {
		allVersions = append(allVersions, v)
	}

	sort.Slice(allVersions, func(i, j int) bool {
		return compareSemver(allVersions[i], allVersions[j]) > 0
	})

	// Limit to 50
	recent := allVersions
	if len(allVersions) > 50 {
		recent = allVersions[:50]
	}

	return recent, len(allVersions)
}

// PackageTool returns the tool definition for get-npm-package
func PackageTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get-npm-package",
		Description: "Get detailed information about a specific NPM package",
	}
}

// ReadmeInput is the input for the get-package-readme tool
type ReadmeInput struct {
	Name string `json:"name" jsonschema:"NPM package name"`
}

// ReadmeOutput is the output for the get-package-readme tool
type ReadmeOutput struct {
	Name   string `json:"name"`
	Readme string `json:"readme"`
}

// ReadmeHandler is the handler type for get-package-readme
type ReadmeHandler = func(context.Context, *mcp.CallToolRequest, ReadmeInput) (*mcp.CallToolResult, ReadmeOutput, error)

// NewReadmeHandler creates a new readme handler
func NewReadmeHandler(client *npm.Client, appCache *cache.Cache) ReadmeHandler {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ReadmeInput) (*mcp.CallToolResult, ReadmeOutput, error) {
		// Check cache
		cacheKey := fmt.Sprintf("readme:%s", input.Name)
		if appCache != nil {
			if cached, found := cache.Get[ReadmeOutput](appCache, cacheKey); found {
				return nil, cached, nil
			}
		}

		pkg, err := client.GetPackage(ctx, input.Name)
		if err != nil {
			return nil, ReadmeOutput{}, err
		}

		output := ReadmeOutput{
			Name:   pkg.Name,
			Readme: pkg.Readme,
		}

		// Store in cache
		if appCache != nil {
			appCache.Set(cacheKey, output)
		}

		return nil, output, nil
	}
}

// ReadmeTool returns the tool definition for get-package-readme
func ReadmeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get-package-readme",
		Description: "Get the README of a specific NPM package",
	}
}
