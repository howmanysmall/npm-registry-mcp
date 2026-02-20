//go:build integration

package main
 
import (
	"context"
	"testing"
	"time"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
)

func TestIntegration_RealNPMAPI(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := npm.NewClient()

	// Test search
	t.Run("Search", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := client.Search(ctx, "lodash", 5)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(result.Objects) == 0 {
			t.Error("expected at least one result")
		}

		found := false

		for _, obj := range result.Objects {
			if obj.Package.Name == "lodash" {
				found = true
				break
			}
		}

		if !found {
			t.Error("expected to find lodash in results")
		}
	})

	// Test get package
	t.Run("GetPackage", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := client.GetPackage(ctx, "lodash")
		if err != nil {
			t.Fatalf("GetPackage failed: %v", err)
		}

		if result.Name != "lodash" {
			t.Errorf("expected lodash, got %s", result.Name)
		}

		if result.License != "MIT" {
			t.Errorf("expected MIT license, got %s", result.License)
		}
	})

	// Test downloads
	t.Run("GetDownloads", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := client.GetDownloads(ctx, "lodash", "last-week")
		if err != nil {
			t.Fatalf("GetDownloads failed: %v", err)
		}

		if result.Downloads < 1000000 {
			t.Errorf("expected lodash to have >1M weekly downloads, got %d", result.Downloads)
		}
	})
}
