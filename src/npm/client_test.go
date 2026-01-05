package npm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
)

const testPackageName = "lodash"

func TestClient_Search(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/-/v1/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.URL.Query().Get("text") != testPackageName {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("text"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"objects": [
				{
					"package": {
						"name": "lodash",
						"version": "4.17.21",
						"description": "Lodash modular utilities"
					},
					"score": {"final": 0.9}
				}
			],
			"total": 1
		}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithBaseURL(server.URL))

	result, err := client.Search(context.Background(), testPackageName, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(result.Objects) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Objects))
	}

	if result.Objects[0].Package.Name != testPackageName {
		t.Errorf("expected lodash, got %s", result.Objects[0].Package.Name)
	}
}

func TestClient_GetPackage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lodash" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"description": "Lodash modular utilities",
			"dist-tags": {"latest": "4.17.21"},
			"versions": {},
			"time": {"created": "2012-04-07T00:00:00.000Z"},
			"license": "MIT"
		}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithBaseURL(server.URL))

	result, err := client.GetPackage(context.Background(), testPackageName)
	if err != nil {
		t.Fatalf("GetPackage failed: %v", err)
	}

	if result.Name != testPackageName {
		t.Errorf("expected lodash, got %s", result.Name)
	}

	if result.License != "MIT" {
		t.Errorf("expected MIT, got %s", result.License)
	}
}

func TestClient_GetDownloads(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/downloads/point/last-week/lodash" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"downloads": 45000000,
			"start": "2025-12-27",
			"end": "2026-01-02",
			"package": "lodash"
		}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithDownloadsBaseURL(server.URL))

	result, err := client.GetDownloads(context.Background(), testPackageName, "last-week")
	if err != nil {
		t.Fatalf("GetDownloads failed: %v", err)
	}

	if result.Downloads != 45000000 {
		t.Errorf("expected 45000000, got %d", result.Downloads)
	}
}

func TestClient_GetDownloadRange(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/downloads/range/last-week/lodash" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"downloads": [
				{"downloads": 6000000, "day": "2025-12-27"},
				{"downloads": 6500000, "day": "2025-12-28"}
			],
			"start": "2025-12-27",
			"end": "2026-01-02",
			"package": "lodash"
		}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithDownloadsBaseURL(server.URL))

	result, err := client.GetDownloadRange(context.Background(), testPackageName, "last-week")
	if err != nil {
		t.Fatalf("GetDownloadRange failed: %v", err)
	}

	if len(result.Downloads) != 2 {
		t.Errorf("expected 2 days, got %d", len(result.Downloads))
	}
}

func TestClient_GetPackage_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Not found"}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithBaseURL(server.URL))

	_, err := client.GetPackage(context.Background(), "nonexistent-package")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestClient_Search_APIErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"The 'text' parameter must be between 2 and 64 characters","code":"ERR_TEXT_LENGTH"}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithBaseURL(server.URL))

	_, err := client.Search(context.Background(), "a", 1)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	// The error should contain the API error message, not just "unexpected status: 400"
	errStr := err.Error()
	if !contains(errStr, "text") || !contains(errStr, "2 and 64 characters") {
		t.Errorf("error should contain API message, got: %s", errStr)
	}
}

func TestClient_GetPackage_WithArrayEngines(t *testing.T) {
	t.Parallel()

	// This simulates old packages like lodash v0.1.0 that have engines as an array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "lodash",
			"description": "Lodash modular utilities",
			"dist-tags": {"latest": "4.17.21"},
			"versions": {
				"0.1.0": {
					"name": "lodash",
					"version": "0.1.0",
					"engines": ["node", "rhino"]
				},
				"4.17.21": {
					"name": "lodash",
					"version": "4.17.21",
					"engines": {"node": ">=0.10.0"}
				}
			},
			"time": {"created": "2012-04-07T00:00:00.000Z"},
			"license": "MIT"
		}`))
	}))
	defer server.Close()

	client := npm.NewClient(npm.WithBaseURL(server.URL))

	result, err := client.GetPackage(context.Background(), "lodash")
	if err != nil {
		t.Fatalf("GetPackage should handle array engines, got error: %v", err)
	}

	if result.Name != "lodash" {
		t.Errorf("expected lodash, got %s", result.Name)
	}

	// Verify both versions are accessible
	if len(result.Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(result.Versions))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
