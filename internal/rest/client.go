// @license
// Copyright 2023 Dynatrace LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"context"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"time"
)

// Option represents a functional Option for the Client.
type Option func(*Client)

// WithConcurrentRequestLimit sets the maximum number of concurrent requests allowed.
func WithConcurrentRequestLimit(maxConcurrent int) Option {
	return func(c *Client) {
		c.ConcurrentRequestLimiter = NewConcurrentRequestLimiter(maxConcurrent)
	}
}

// WithTimeout sets the request timeout for the Client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithHTTPClient sets the HTTP client for the Client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = client
	}
}

// WithRequestRetrier sets the RequestRetrier for the Client.
func WithRequestRetrier(retrier *RequestRetrier) Option {
	return func(c *Client) {
		c.RequestRetrier = retrier
	}
}

// WithRequestResponseRecorder sets the RequestResponseRecorder for the Client.
// Note, that this is happening sync. so there MUST be a listener gabbing the
// recorded messages, to not block the client
func WithRequestResponseRecorder(recorder *RequestResponseRecorder) Option {
	return func(c *Client) {
		c.RequestResponseRecorder = recorder
	}
}

// Client represents a general HTTP client
type Client struct {
	BaseURL                  string                    // base URL of the server
	Headers                  map[string]string         // Custom headers to be set
	Logger                   logr.Logger               // Logger interface to be used
	ConcurrentRequestLimiter *ConcurrentRequestLimiter // Concurrent request limiter component (optional)
	Timeout                  time.Duration             // Request timeout (optional)
	HTTPClient               *http.Client              // Custom HTTP client support (optional)
	RequestRetrier           *RequestRetrier           // HTTP request retrier component (optional)
	RequestResponseRecorder  *RequestResponseRecorder  // Request-response recorder component (optional)
	RateLimiter              *RateLimiter              // Rate limiter component (optional)
}

// NewClient creates a new instance of the Client with specified options.
func NewClient(baseURL string, opts ...Option) *Client {
	client := &Client{
		BaseURL: baseURL,
		Headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// GET sends a GET request to the specified endpoint.
func (c *Client) GET(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodGet, endpoint, nil, 0)
}

// PUT sends a PUT request to the specified endpoint with the given body.
func (c *Client) PUT(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPut, endpoint, body, 0)
}

// POST sends a POST request to the specified endpoint with the given body.
func (c *Client) POST(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPost, endpoint, body, 0)
}

// DELETE sends a DELETE request to the specified endpoint.
func (c *Client) DELETE(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodDelete, endpoint, nil, 0)
}

// SetHeader sets a custom header for the HTTP client.
func (c *Client) SetHeader(key, value string) {
	c.Headers[key] = value
}

// sendRequestWithRetries sends an HTTP request with custom headers and modified request body, with retries if configured.
func (c *Client) sendRequestWithRetries(ctx context.Context, method, endpoint string, body io.Reader, retryCount int) (*http.Response, error) {

	if c.RateLimiter != nil {
		if !c.RateLimiter.Allow() {
			c.RateLimiter.Allow() // This should block until the reset time is reached
		}
	}

	// Apply concurrent request limiting if ConcurrentRequestLimiter is set
	if c.ConcurrentRequestLimiter != nil {
		c.ConcurrentRequestLimiter.Acquire()
		defer c.ConcurrentRequestLimiter.Release()
	}

	url := c.BaseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout,
		}
	}

	if c.RequestResponseRecorder != nil {
		// wrap the body so that it could be read again
		req.Body = ReusableReader(req.Body)
		c.RequestResponseRecorder.RecordRequest(ctx, req)
	}

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		if c.RequestResponseRecorder != nil {
			c.RequestResponseRecorder.RecordResponse(ctx, nil, err)
		}
		return nil, err
	}

	if c.RequestResponseRecorder != nil {
		// wrap the body so that it could be read again
		response.Body = ReusableReader(response.Body)
		c.RequestResponseRecorder.RecordResponse(ctx, response, nil)
	}

	// Update the rate limiter with the response headers
	if c.RateLimiter != nil {
		c.RateLimiter.Update(response.Header)
	}

	if c.RequestRetrier != nil && retryCount < c.RequestRetrier.MaxRetries {
		if c.RequestRetrier.ShouldRetryFunc != nil && c.RequestRetrier.ShouldRetryFunc(response) {
			time.Sleep(100 * time.Millisecond)
			return c.sendRequestWithRetries(ctx, method, endpoint, body, retryCount+1)
		}
	}

	if !isSuccess(response) {
		// If the response code is not in the success range, read the response payload for the error details
		payload, _ := io.ReadAll(response.Body)
		response.Body.Close()
		return nil, HTTPError{Code: response.StatusCode, Payload: payload}
	}

	return response, nil
}

// isSuccess checks if the HTTP response is in the success range (2xx).
func isSuccess(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode <= 299
}
