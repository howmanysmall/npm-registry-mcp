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

	if err := cmd.Execute(os.Args[1:], RunMCP); err != nil {
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

	mcp.AddTool(server, tools.SearchTool(), tools.NewSearchHandler(npmClient))
	mcp.AddTool(server, tools.PackageTool(), tools.NewPackageHandler(npmClient))
	mcp.AddTool(server, tools.VersionsTool(), tools.NewVersionsHandler(npmClient))
	mcp.AddTool(server, tools.InstallTool(), tools.NewInstallHandler(npmClient, ghClient, appCache))

	log.Printf("Starting %s v%s", serverName, serverVersion)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
