// Package main provides the entry point for the npm-registry-mcp server.
// This MCP server exposes tools for interacting with the NPM registry API.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
	"github.com/howmanysmall/npm-registry-mcp/src/cmd"
	"github.com/howmanysmall/npm-registry-mcp/src/github"
	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "npm-registry-mcp"
	serverVersion = "0.1.0"
)

func main() {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	// Ensure logs go to stderr to avoid corrupting MCP stdio transport
	log.SetOutput(os.Stderr)

	// If no arguments are provided, default to starting the MCP server
	if len(os.Args) == 1 {
		RunMCP()
		return
	}

	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// RunMCP starts the MCP server using stdio transport
func RunMCP() {
	npmClient := npm.NewClient()
	ghClient := github.NewClient()
	appCache := cache.New(5*time.Minute, 10*time.Minute)

	if ghClient.HasToken() {
		log.Println("GitHub token configured (5000 req/hour)")
	} else {
		log.Println("No GitHub token - using unauthenticated API (60 req/hour)")
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	mcp.AddTool(server, tools.SearchTool(), tools.NewSearchHandler(npmClient, appCache))
	mcp.AddTool(server, tools.PackageTool(), tools.NewPackageHandler(npmClient, appCache))
	mcp.AddTool(server, tools.ReadmeTool(), tools.NewReadmeHandler(npmClient, appCache))
	mcp.AddTool(server, tools.VersionsTool(), tools.NewVersionsHandler(npmClient, appCache))
	mcp.AddTool(server, tools.TagsTool(), tools.NewTagsHandler(npmClient, appCache))
	mcp.AddTool(server, tools.InstallTool(), tools.NewInstallHandler(npmClient, ghClient, appCache))
	mcp.AddTool(server, tools.SecurityTool(), tools.NewSecurityHandler(npmClient, appCache))

	log.Printf("Starting %s v%s", serverName, serverVersion)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
