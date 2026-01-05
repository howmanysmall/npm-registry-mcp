// Package github provides a client for the GitHub API.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	defaultBaseURL = "https://api.github.com"
	defaultTimeout = 30 * time.Second
)

// Client is an HTTP client for the GitHub API
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithBaseURL sets the base URL (for testing)
func WithBaseURL(url string) ClientOption {
	return func(client *Client) {
		client.baseURL = url
	}
}

// WithToken sets the GitHub token
func WithToken(token string) ClientOption {
	return func(client *Client) {
		client.token = token
	}
}

// NewClient creates a new GitHub API client
func NewClient(clientOptions ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: defaultBaseURL,
		token:   os.Getenv("GITHUB_TOKEN"),
	}

	for _, option := range clientOptions {
		option(client)
	}

	return client
}

// GetRepository gets repository information
func (client *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	u := fmt.Sprintf("%s/repos/%s/%s", client.baseURL, owner, repo)

	var result Repository
	if err := client.doJSON(ctx, u, &result); err != nil {
		return nil, fmt.Errorf("get repository %s/%s: %w", owner, repo, err)
	}

	return &result, nil
}

// GetCommits gets recent commits (within days)
func (client *Client) GetCommits(ctx context.Context, owner, repo string, days int) ([]Commit, error) {
	since := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)
	u := fmt.Sprintf("%s/repos/%s/%s/commits?since=%s&per_page=100", client.baseURL, owner, repo, since)

	var result []Commit
	if err := client.doJSON(ctx, u, &result); err != nil {
		return nil, fmt.Errorf("get commits %s/%s: %w", owner, repo, err)
	}

	return result, nil
}

// HasToken returns true if a GitHub token is configured
func (client *Client) HasToken() bool {
	return client.token != ""
}

func (client *Client) doJSON(ctx context.Context, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if client.token != "" {
		req.Header.Set("Authorization", "Bearer "+client.token)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// Sentinel errors
var (
	ErrNotFound         = errors.New("not found")
	ErrRateLimited      = errors.New("rate limited")
	ErrUnexpectedStatus = errors.New("unexpected status")
)
