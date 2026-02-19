package cmd

import "github.com/spf13/cobra"

type cliOptions struct {
	jsonOutput bool
}

// Execute builds the root command, applies args, and executes it.
func Execute(args []string) error {
	rootCmd := newRootCmd()
	rootCmd.SetArgs(args)

	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	options := &cliOptions{}

	rootCmd := &cobra.Command{
		Use:           "npm-registry-mcp",
		Short:         "NPM Registry MCP server and CLI tool",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `A Model Context Protocol (MCP) server for NPM package analysis that also works as a CLI.
When run without any subcommands, it starts the MCP server using stdio transport.`,
	}

	rootCmd.PersistentFlags().BoolVar(&options.jsonOutput, "json", false, "Output results in JSON format")
	rootCmd.AddCommand(
		newSearchCmd(options),
		newInfoCmd(options),
		newVersionsCmd(options),
		newHealthCmd(options),
	)

	return rootCmd
}
