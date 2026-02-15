package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "npm-registry-mcp",
	Short: "NPM Registry MCP server and CLI tool",
	Long: `A Model Context Protocol (MCP) server for NPM package analysis that also works as a CLI.
When run without any subcommands, it starts the MCP server using stdio transport.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// This will be handled in main.go to start the MCP server
			return nil
		}
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
}

// GetJSONOutput returns the value of the json-output flag
func GetJSONOutput() bool {
	return jsonOutput
}

// ShouldRunMCP returns true if no subcommand was provided
func ShouldRunMCP() bool {
	// If Execute was called and no subcommand matched, we might want to run MCP.
	// However, Cobra's default behavior is to show help if no subcommand matches.
	// We'll need to check os.Args in main.go or customize the root command.
	return len(os.Args) == 1
}
