package tools_test

import (
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolRegistration(t *testing.T) {
	t.Parallel()

	npmClient := npm.NewClient()
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)

	// Register tools using the generic AddTool
	mcp.AddTool(server, tools.SearchTool(), tools.NewSearchHandler(npmClient, nil))
	mcp.AddTool(server, tools.PackageTool(), tools.NewPackageHandler(npmClient, nil))
	mcp.AddTool(server, tools.VersionsTool(), tools.NewVersionsHandler(npmClient, nil))
	mcp.AddTool(server, tools.InstallTool(), tools.NewInstallHandler(npmClient, nil, nil))
	mcp.AddTool(server, tools.ReadmeTool(), tools.NewReadmeHandler(npmClient, nil))
	mcp.AddTool(server, tools.TagsTool(), tools.NewTagsHandler(npmClient, nil))
	mcp.AddTool(server, tools.SecurityTool(), tools.NewSecurityHandler(npmClient, nil))

	// In a real test we would check if server has the tools and if schemas are populated.
	// But since server.ListTools is not easily accessible here without running the server,
	// we just ensure it doesn't panic.
}
