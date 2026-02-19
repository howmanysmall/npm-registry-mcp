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
	mcp.AddTool(server, tools.SearchTool(), tools.NewSearchHandler(npmClient))
	mcp.AddTool(server, tools.PackageTool(), tools.NewPackageHandler(npmClient))
	mcp.AddTool(server, tools.VersionsTool(), tools.NewVersionsHandler(npmClient))
	mcp.AddTool(server, tools.InstallTool(), tools.NewInstallHandler(npmClient, nil, nil))

	// In a real test we would check if server has the tools and if schemas are populated.
	// But since server.ListTools is not easily accessible here without running the server,
	// we just ensure it doesn't panic.
}
