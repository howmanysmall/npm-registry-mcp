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

func newInfoCmd(options *cliOptions) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info [package]",
		Short: "Get detailed information about an NPM package",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			pkgName := args[0]
			npmClient := npm.NewClient()

			input := tools.PackageInput{
				Name: pkgName,
			}

			handler := tools.NewPackageHandler(npmClient, nil)

			_, output, err := handler(context.Background(), nil, input)
			if err != nil {
				return err
			}

			if options.jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(output)
			}

			fmt.Printf("Package: %s\n", output.Name)
			fmt.Printf("Description: %s\n", output.Description)
			fmt.Printf("Latest Version: %s\n", output.LatestVersion)
			fmt.Printf("License: %s\n", output.License)

			if output.Homepage != "" {
				fmt.Printf("Homepage: %s\n", output.Homepage)
			}

			if output.Repository != "" {
				fmt.Printf("Repository: %s\n", output.Repository)
			}

			fmt.Printf("Maintainers: %s\n", strings.Join(output.Maintainers, ", "))

			if len(output.Keywords) > 0 {
				fmt.Printf("Keywords: %s\n", strings.Join(output.Keywords, ", "))
			}

			fmt.Printf("\nDependencies (%d):\n", len(output.Dependencies))

			for name, version := range output.Dependencies {
				fmt.Printf("  %s: %s\n", name, version)
			}

			if len(output.Versions) > 0 {
				fmt.Printf("\nRecent Versions (%d total):\n", output.TotalVersions)
				fmt.Printf("  %s\n", strings.Join(output.Versions, ", "))
			}

			return nil
		},
	}

	return infoCmd
}
