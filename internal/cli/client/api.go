package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is an HTTP client wrapper for the EnvHub API
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New creates a new API client
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request makes an authenticated HTTP request
func (c *Client) Request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyBytes = b
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.HTTPClient.Do(req)
}

// Get makes a GET request
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

// Post makes a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

// Put makes a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}

// Delete makes a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}
