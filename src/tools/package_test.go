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

func TestPackageTool(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"description": "Lodash utilities",
			"dist-tags": {"latest": "4.17.21"},
			"versions": {
				"4.17.21": {
					"name": "lodash",
					"version": "4.17.21",
					"dependencies": {}
				}
			},
			"time": {
				"4.17.21": "2021-02-20T00:00:00.000Z"
			},
			"maintainers": [{"name": "jdalton", "email": "john@example.com"}],
			"license": "MIT",
			"repository": {"type": "git", "url": "https://github.com/lodash/lodash"}
		}`))
	}))
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
}
