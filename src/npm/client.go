// Package npm provides an HTTP client for the NPM registry API.
package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultRegistryURL  = "https://registry.npmjs.org"
	defaultDownloadsURL = "https://api.npmjs.org"
	defaultTimeout      = 30 * time.Second
)

// Client is an HTTP client for the NPM registry API
type Client struct {
	httpClient       *http.Client
	registryBaseURL  string
	downloadsBaseURL string
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(c *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = c
	}
}

// WithBaseURL sets the registry base URL (for testing)
func WithBaseURL(u string) ClientOption {
	return func(client *Client) {
		client.registryBaseURL = u
	}
}

// WithDownloadsBaseURL sets the downloads API base URL (for testing)
func WithDownloadsBaseURL(u string) ClientOption {
	return func(client *Client) {
		client.downloadsBaseURL = u
	}
}

// NewClient creates a new NPM registry client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		registryBaseURL:  defaultRegistryURL,
		downloadsBaseURL: defaultDownloadsURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Search searches the NPM registry
func (c *Client) Search(ctx context.Context, query string, limit int) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("text", query)
	params.Set("size", strconv.Itoa(limit))

	u := fmt.Sprintf("%s/-/v1/search?%s", c.registryBaseURL, params.Encode())

	var result SearchResponse
	if err := c.doJSON(ctx, u, "application/json", &result); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	return &result, nil
}

// GetPackage gets detailed package information
func (c *Client) GetPackage(ctx context.Context, name string) (*PackageResponse, error) {
	u := fmt.Sprintf("%s/%s", c.registryBaseURL, url.PathEscape(name))

	var result PackageResponse
	if err := c.doJSON(ctx, u, "application/json", &result); err != nil {
		return nil, fmt.Errorf("get package %s: %w", name, err)
	}

	return &result, nil
}

// GetAbbreviatedPackage gets abbreviated package information (faster)
func (c *Client) GetAbbreviatedPackage(ctx context.Context, name string) (*AbbreviatedPackageResponse, error) {
	u := fmt.Sprintf("%s/%s", c.registryBaseURL, url.PathEscape(name))

	var result AbbreviatedPackageResponse
	if err := c.doJSON(ctx, u, "application/vnd.npm.install-v1+json", &result); err != nil {
		return nil, fmt.Errorf("get abbreviated package %s: %w", name, err)
	}

	return &result, nil
}

// GetDownloads gets download statistics for a package
func (c *Client) GetDownloads(ctx context.Context, name, period string) (*DownloadPoint, error) {
	u := fmt.Sprintf("%s/downloads/point/%s/%s", c.downloadsBaseURL, period, url.PathEscape(name))

	var result DownloadPoint
	if err := c.doJSON(ctx, u, "application/json", &result); err != nil {
		return nil, fmt.Errorf("get downloads %s: %w", name, err)
	}

	return &result, nil
}

// GetDownloadRange gets daily download statistics for a package
func (c *Client) GetDownloadRange(ctx context.Context, name, period string) (*DownloadRange, error) {
	u := fmt.Sprintf("%s/downloads/range/%s/%s", c.downloadsBaseURL, period, url.PathEscape(name))

	var result DownloadRange
	if err := c.doJSON(ctx, u, "application/json", &result); err != nil {
		return nil, fmt.Errorf("get download range %s: %w", name, err)
	}

	return &result, nil
}

// GetAdvisories gets security advisories for a set of packages and versions
func (c *Client) GetAdvisories(ctx context.Context, packages map[string][]string) (map[string][]Advisory, error) {
	u := fmt.Sprintf("%s/-/npm/v1/security/advisories/bulk", c.registryBaseURL)

	body, err := json.Marshal(packages)
	if err != nil {
		return nil, fmt.Errorf("marshal advisories request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if decErr := json.NewDecoder(resp.Body).Decode(&apiErr); decErr == nil && apiErr.Error != "" {
			return nil, fmt.Errorf("%w: %d: %s", ErrUnexpectedStatus, resp.StatusCode, apiErr.Error)
		}

		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	var result map[string][]Advisory
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func (c *Client) doJSON(ctx context.Context, url, acceptHeader string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if acceptHeader != "" {
		req.Header.Set("Accept", acceptHeader)
	} else {
		req.Header.Set("Accept", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response from NPM API
		var apiErr APIError
		if decErr := json.NewDecoder(resp.Body).Decode(&apiErr); decErr == nil && apiErr.Error != "" {
			return fmt.Errorf("%w: %d: %s", ErrUnexpectedStatus, resp.StatusCode, apiErr.Error)
		}

		return fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// APIError represents an error response from the NPM API
type APIError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Sentinel errors
var (
	ErrUnexpectedStatus = errors.New("unexpected status")
	ErrRateLimited      = errors.New("rate limited")
)
