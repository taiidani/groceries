// Package client provides an HTTP client for the groceries REST API.
// It is used by the web server to call the API on behalf of the authenticated
// user, using the Bearer token stored in their session.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is an API client for the groceries REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New creates a new Client targeting baseURL and authenticating with token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
	}
}

// do executes an HTTP request, attaching the Bearer token, and returns the
// response. The caller is responsible for closing resp.Body.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("client: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("client: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: %s %s: %w", method, path, err)
	}

	return resp, nil
}

// decode reads a JSON response body into dst and closes the body.
func decode[T any](resp *http.Response, dst *T) error {
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("client: decode response: %w", err)
	}
	return nil
}

// checkError reads the response body as an API error if the status code is
// not in the 2xx range, and returns a descriptive error. The body is always
// closed.
func checkError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		resp.Body.Close()
		return nil
	}

	defer resp.Body.Close()

	var apiErr struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil || apiErr.Error == "" {
		return fmt.Errorf("client: unexpected status %d", resp.StatusCode)
	}

	return fmt.Errorf("client: %s", apiErr.Error)
}
