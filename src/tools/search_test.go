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

func TestSearchTool(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"objects": [
				{
					"package": {
						"name": "lodash",
						"version": "4.17.21",
						"description": "Lodash utilities"
					},
					"score": {"final": 0.9, "detail": {"popularity": 0.95}}
				}
			],
			"total": 1
		}`))
	}))
	defer server.Close()

	npmClient := npm.NewClient(npm.WithBaseURL(server.URL))
	handler := tools.NewSearchHandler(npmClient, nil)

	req := &mcp.CallToolRequest{}
	input := tools.SearchInput{
		Query: "lodash",
		Limit: 10,
	}

	result, output, err := handler(context.Background(), req, input)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if result != nil && result.IsError {
		t.Fatalf("unexpected error result")
	}

	if len(output.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(output.Packages))
	}

	if output.Packages[0].Name != "lodash" {
		t.Errorf("expected lodash, got %s", output.Packages[0].Name)
	}
}
