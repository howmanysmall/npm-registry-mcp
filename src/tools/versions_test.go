package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const latestVersion = "4.17.21"

func getVersionServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"dist-tags": {"latest": "4.17.21"},
			"versions": {
				"4.17.21": {"version": "4.17.21"},
				"4.17.20": {"version": "4.17.20"},
				"4.17.19": {"version": "4.17.19"},
				"4.16.0": {"version": "4.16.0"},
				"3.10.1": {"version": "3.10.1"}
			},
			"time": {}
		}`))
	}))
}

func TestVersionsTool(t *testing.T) {
	t.Parallel()

	server := getVersionServer()
	defer server.Close()

	npmClient := npm.NewClient(npm.WithBaseURL(server.URL))
	handler := tools.NewVersionsHandler(npmClient, nil)

	req := &mcp.CallToolRequest{}
	input := tools.VersionsInput{
		Name:  testPackageName,
		Limit: 100,
	}

	result, output, err := handler(context.Background(), req, input)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if result != nil && result.IsError {
		t.Fatalf("unexpected error result")
	}

	if output.Name != testPackageName {
		t.Errorf("expected name %s, got %s", testPackageName, output.Name)
	}

	if output.LatestVersion != latestVersion {
		t.Errorf("expected latest %s, got %s", latestVersion, output.LatestVersion)
	}

	if output.TotalVersions != 5 {
		t.Errorf("expected 5 total versions, got %d", output.TotalVersions)
	}

	if output.VersionsShown != 5 {
		t.Errorf("expected 5 versions shown, got %d", output.VersionsShown)
	}

	// Verify sorted descending by semver
	if len(output.Versions) < 2 {
		t.Fatal("expected at least 2 versions")
	}

	if output.Versions[0] != latestVersion {
		t.Errorf("expected first version %s, got %s", latestVersion, output.Versions[0])
	}

	if output.Versions[len(output.Versions)-1] != "3.10.1" {
		t.Errorf("expected last version 3.10.1, got %s", output.Versions[len(output.Versions)-1])
	}
}

func TestVersionsTool_Limit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"dist-tags": {"latest": "4.17.21"},
			"versions": {
				"4.17.21": {"version": "4.17.21"},
				"4.17.20": {"version": "4.17.20"},
				"4.17.19": {"version": "4.17.19"},
				"4.16.0": {"version": "4.16.0"},
				"3.10.1": {"version": "3.10.1"}
			},
			"time": {}
		}`))
	}))
	defer server.Close()

	npmClient := npm.NewClient(npm.WithBaseURL(server.URL))
	handler := tools.NewVersionsHandler(npmClient, nil)

	req := &mcp.CallToolRequest{}
	input := tools.VersionsInput{
		Name:  testPackageName,
		Limit: 2,
	}

	_, output, err := handler(context.Background(), req, input)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if output.TotalVersions != 5 {
		t.Errorf("expected 5 total versions, got %d", output.TotalVersions)
	}

	if output.VersionsShown != 2 {
		t.Errorf("expected 2 versions shown, got %d", output.VersionsShown)
	}

	if len(output.Versions) != 2 {
		t.Errorf("expected 2 versions in list, got %d", len(output.Versions))
	}
}

func TestTagsTool(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"dist-tags": {"latest": "4.17.21", "next": "5.0.0"}
		}`))
	}))
	defer server.Close()

	npmClient := npm.NewClient(npm.WithBaseURL(server.URL))
	handler := tools.NewTagsHandler(npmClient, nil)

	req := &mcp.CallToolRequest{}
	input := tools.TagsInput{
		Name: testPackageName,
	}

	_, output, err := handler(context.Background(), req, input)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if output.Name != testPackageName {
		t.Errorf("expected %s, got %s", testPackageName, output.Name)
	}

	if output.Tags["latest"] != latestVersion || output.Tags["next"] != "5.0.0" {
		t.Errorf("expected tags, got %v", output.Tags)
	}
}
