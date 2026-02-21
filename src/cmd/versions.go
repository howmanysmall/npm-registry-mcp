package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/spf13/cobra"
)

func newVersionsCmd(options *cliOptions) *cobra.Command {
	var versionsLimit int

	versionsCmd := &cobra.Command{
		Use:   "versions [package]",
		Short: "List all versions of an NPM package",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			pkgName := args[0]
			npmClient := npm.NewClient()

			input := tools.VersionsInput{
				Name:  pkgName,
				Limit: versionsLimit,
			}

			handler := tools.NewVersionsHandler(npmClient, nil)

			_, output, err := handler(context.Background(), nil, input)
			if err != nil {
				return err
			}

			if options.jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(output)
			}

			fmt.Printf("Package: %s\n", output.Name)
			fmt.Printf("Latest Version: %s\n", output.LatestVersion)
			fmt.Printf("Total Versions: %d\n", output.TotalVersions)
			fmt.Printf("Versions Shown: %d\n\n", output.VersionsShown)

			fmt.Println(strings.Join(output.Versions, "\n"))

			return nil
		},
	}

	versionsCmd.Flags().IntVarP(&versionsLimit, "limit", "l", 100, "Maximum number of versions (1-1000)")

	return versionsCmd
}
