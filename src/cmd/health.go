package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
	"github.com/howmanysmall/npm-registry-mcp/src/github"
	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/spf13/cobra"
)

func newHealthCmd(options *cliOptions) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:     "health [package]",
		Aliases: []string{"evaluate", "check"},
		Short:   "Evaluate the health and risk of an NPM package",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			pkgName := args[0]

			npmClient := npm.NewClient()
			ghClient := github.NewClient()
			appCache := cache.New(0, 0) // No cache for CLI single-run

			input := tools.InstallInput{
				Package: pkgName,
			}

			handler := tools.NewInstallHandler(npmClient, ghClient, appCache)

			_, output, err := handler(context.Background(), nil, input)
			if err != nil {
				return err
			}

			if options.jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(output)
			}

			fmt.Printf("Health Assessment for: %s@%s\n", output.Package, output.Version)
			fmt.Printf("Verdict: %s\n", output.Verdict)
			fmt.Printf("Score:   %d/100\n", output.Score)

			fmt.Printf("\nMetrics:\n")
			fmt.Printf("  Maintenance: %s (Last publish: %s)\n", output.Maintenance.Status, output.Maintenance.LastPublish)
			fmt.Printf("  Popularity:  %s (%d weekly downloads, %s)\n", output.Popularity.Status, output.Popularity.WeeklyDownloads, output.Popularity.Trend)
			fmt.Printf("  Security:    %d vulnerabilities (%s)\n", output.Security.Vulnerabilities, output.Security.Status)
			fmt.Printf("  License:     %s (Risk: %s)\n", output.License.SPDX, output.License.Risk)
			fmt.Printf("  TypeScript:  %s (%s)\n", output.TypeScript.Support, output.TypeScript.Status)
			fmt.Printf("  Dependencies: %d direct, %d outdated\n", output.Dependencies.Direct, output.Dependencies.Outdated)
			fmt.Printf("  Size:        %s unpacked\n", output.Size.Unpacked)

			if len(output.Warnings) > 0 {
				fmt.Printf("\nWarnings:\n")

				for _, w := range output.Warnings {
					fmt.Printf("  - %s\n", w)
				}
			}

			return nil
		},
	}

	return healthCmd
}
