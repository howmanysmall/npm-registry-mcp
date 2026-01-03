package github_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/github"
)

func TestClient_GetRepository(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/lodash/lodash" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"full_name": "lodash/lodash",
			"stargazers_count": 60000,
			"forks_count": 7000,
			"open_issues_count": 100,
			"pushed_at": "2025-12-18T00:00:00Z",
			"created_at": "2012-04-07T00:00:00Z"
		}`))
	}))
	defer server.Close()

	client := github.NewClient(github.WithBaseURL(server.URL))

	result, err := client.GetRepository(context.Background(), "lodash", "lodash")
	if err != nil {
		t.Fatalf("GetRepository failed: %v", err)
	}

	if result.StargazersCount != 60000 {
		t.Errorf("expected 60000 stars, got %d", result.StargazersCount)
	}
}

func TestClient_GetCommits(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/lodash/lodash/commits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"sha": "abc123", "commit": {"message": "fix: something"}},
			{"sha": "def456", "commit": {"message": "feat: new thing"}}
		]`))
	}))
	defer server.Close()

	client := github.NewClient(github.WithBaseURL(server.URL))

	result, err := client.GetCommits(context.Background(), "lodash", "lodash", 90)
	if err != nil {
		t.Fatalf("GetCommits failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 commits, got %d", len(result))
	}
}

func TestClient_GetRepository_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := github.NewClient(github.WithBaseURL(server.URL))

	_, err := client.GetRepository(context.Background(), "nonexistent", "repo")
	if !errors.Is(err, github.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_GetRepository_RateLimited(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := github.NewClient(github.WithBaseURL(server.URL))

	_, err := client.GetRepository(context.Background(), "owner", "repo")
	if !errors.Is(err, github.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}
