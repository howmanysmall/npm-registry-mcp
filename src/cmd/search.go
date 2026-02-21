package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newSearchCmd(options *cliOptions) *cobra.Command {
	var searchLimit int

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search the NPM registry for packages",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			query := args[0]
			npmClient := npm.NewClient()

			input := tools.SearchInput{
				Query: query,
				Limit: searchLimit,
			}

			handler := tools.NewSearchHandler(npmClient, nil)

			_, output, err := handler(context.Background(), nil, input)
			if err != nil {
				return err
			}

			if options.jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(output)
			}

			if len(output.Packages) == 0 {
				fmt.Println("No packages found.")
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.Header("Name", "Version", "Score", "Description")

			for _, p := range output.Packages {
				_ = table.Append(
					p.Name,
					p.Version,
					fmt.Sprintf("%.2f", p.Score),
					p.Description,
				)
			}

			_ = table.Render()

			fmt.Printf("\nTotal results: %d\n", output.Total)

			return nil
		},
	}

	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results (1-100)")

	return searchCmd
}
