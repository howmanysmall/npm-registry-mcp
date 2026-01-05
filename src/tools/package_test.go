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

func getPackageServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"description": "Lodash utilities",
			"dist-tags": {"latest": "4.17.21"},
			"versions": {
				"4.17.21": {
					"name": "lodash",
					"version": "4.17.21",
					"dependencies": {"dep1": "^1.0.0"},
					"devDependencies": {"dev1": "^2.0.0"},
					"peerDependencies": {"peer1": "^3.0.0"},
					"engines": {"node": ">=14"}
				},
				"4.17.20": {"version": "4.17.20"},
				"4.17.19": {"version": "4.17.19"}
			},
			"time": {
				"4.17.21": "2021-02-20T00:00:00.000Z"
			},
			"maintainers": [{"name": "jdalton", "email": "john@example.com"}],
			"license": "MIT",
			"repository": {"type": "git", "url": "https://github.com/lodash/lodash"},
			"bugs": {"url": "https://github.com/lodash/lodash/issues"},
			"author": {"name": "John-David Dalton"}
		}`))
	}))
}

func TestPackageTool(t *testing.T) {
	t.Parallel()

	server := getPackageServer()
	defer server.Close()

	npmClient := npm.NewClient(npm.WithBaseURL(server.URL))
	handler := tools.NewPackageHandler(npmClient)

	req := &mcp.CallToolRequest{}
	input := tools.PackageInput{
		Name: "lodash",
	}

	result, output, err := handler(context.Background(), req, input)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if result != nil && result.IsError {
		t.Fatalf("unexpected error result")
	}

	if output.Name != "lodash" {
		t.Errorf("expected lodash, got %s", output.Name)
	}

	if output.License != "MIT" {
		t.Errorf("expected MIT, got %s", output.License)
	}

	if output.TotalVersions != 3 {
		t.Errorf("expected 3 total versions, got %d", output.TotalVersions)
	}

	if len(output.Versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(output.Versions))
	}

	if output.DevDependencies == nil || output.DevDependencies["dev1"] != "^2.0.0" {
		t.Errorf("expected devDependencies with dev1, got %v", output.DevDependencies)
	}

	if output.PeerDependencies == nil || output.PeerDependencies["peer1"] != "^3.0.0" {
		t.Errorf("expected peerDependencies with peer1, got %v", output.PeerDependencies)
	}

	if output.Engines == nil || output.Engines["node"] != ">=14" {
		t.Errorf("expected engines with node, got %v", output.Engines)
	}
}
