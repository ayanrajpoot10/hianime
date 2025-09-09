package httpclient

import (
	"fmt"
	"net/http"
	"time"
)

// Client wraps http.Client with additional functionality
type Client struct {
	client    *http.Client
	userAgent string
	baseURL   string
	retries   int
}

// Config holds configuration for the HTTP client
type Config struct {
	Timeout   time.Duration
	UserAgent string
	BaseURL   string
	Retries   int
}

// New creates a new HTTP client with the provided configuration
func New(cfg Config) *Client {
	return &Client{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		userAgent: cfg.UserAgent,
		baseURL:   cfg.BaseURL,
		retries:   cfg.Retries,
	}
}

// Get performs a GET request with default headers
func (c *Client) Get(url string) (*http.Response, error) {
	return c.GetWithHeaders(url, nil)
}

// GetWithHeaders performs a GET request with custom headers
func (c *Client) GetWithHeaders(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL+"/")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		resp, lastErr = c.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Close response body if there was an error
		if resp != nil {
			resp.Body.Close()
		}

		// Don't retry on the last attempt
		if attempt < c.retries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to make request after %d retries: %w", c.retries+1, lastErr)
	}

	return resp, fmt.Errorf("unexpected status code after %d retries: %d", c.retries+1, resp.StatusCode)
}

// Post performs a POST request
func (c *Client) Post(url string, body []byte, contentType string) (*http.Response, error) {
	return c.PostWithHeaders(url, body, contentType, nil)
}

// PostWithHeaders performs a POST request with custom headers
func (c *Client) PostWithHeaders(url string, body []byte, contentType string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL+"/")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		resp, lastErr = c.client.Do(req)
		if lastErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Close response body if there was an error
		if resp != nil {
			resp.Body.Close()
		}

		// Don't retry on the last attempt
		if attempt < c.retries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to make request after %d retries: %w", c.retries+1, lastErr)
	}

	return resp, fmt.Errorf("unexpected status code after %d retries: %d", c.retries+1, resp.StatusCode)
}

// Do performs a custom HTTP request
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Set default headers if not already set
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if req.Header.Get("Referer") == "" {
		req.Header.Set("Referer", c.baseURL+"/")
	}

	return c.client.Do(req)
}

// SetUserAgent updates the user agent string
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// SetBaseURL updates the base URL
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// GetUnderlyingClient returns the underlying http.Client for advanced usage
func (c *Client) GetUnderlyingClient() *http.Client {
	return c.client
}
