package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/github"
	"github.com/howmanysmall/npm-registry-mcp/src/npm"
	"github.com/howmanysmall/npm-registry-mcp/src/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const testPackageName = "lodash"

func getNpmServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/lodash":
			_, _ = w.Write([]byte(`{
				"name": "lodash",
				"description": "Lodash utilities",
				"dist-tags": {"latest": "4.17.21"},
				"versions": {
					"4.17.21": {
						"name": "lodash",
						"version": "4.17.21",
						"dependencies": {},
						"dist": {"unpackedSize": 1400000}
					}
				},
				"time": {"4.17.21": "2024-01-15T00:00:00.000Z"},
				"maintainers": [{"name": "jdalton"}],
				"license": "MIT",
				"repository": {"url": "https://github.com/lodash/lodash"}
			}`))
		case "/downloads/point/last-week/lodash":
			_, _ = w.Write([]byte(`{"downloads": 45000000, "package": "lodash"}`))
		case "/downloads/point/last-month/lodash":
			_, _ = w.Write([]byte(`{"downloads": 180000000, "package": "lodash"}`))
		}
	}))
}

func TestInstallTool(t *testing.T) {
	t.Parallel()

	npmServer := getNpmServer()
	defer npmServer.Close()

	ghServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"stargazers_count": 60000,
			"open_issues_count": 100,
			"pushed_at": "2024-12-01T00:00:00Z"
		}`))
	}))
	defer ghServer.Close()

	npmClient := npm.NewClient(
		npm.WithBaseURL(npmServer.URL),
		npm.WithDownloadsBaseURL(npmServer.URL),
	)
	ghClient := github.NewClient(github.WithBaseURL(ghServer.URL))

	handler := tools.NewInstallHandler(npmClient, ghClient, nil)

	req := &mcp.CallToolRequest{}
	input := tools.InstallInput{
		Package: testPackageName,
	}

	result, output, err := handler(context.Background(), req, input)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if result != nil && result.IsError {
		t.Fatalf("unexpected error result")
	}

	if output.Package != testPackageName {
		t.Errorf("expected lodash, got %s", output.Package)
	}

	if output.License.SPDX != "MIT" {
		t.Errorf("expected MIT license, got %s", output.License.SPDX)
	}

	if output.Score < 50 {
		t.Errorf("expected score >= 50 for healthy package, got %d", output.Score)
	}
}
