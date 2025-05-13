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

// Package rest provides an extended rest.Client with optional behaviours like rate limiting, request/response logging, etc.
package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

// RequestOptions are additional options that should be applied
// to a request
type RequestOptions struct {
	// QueryParams are HTTP query parameters that shall be appended
	// to the endpoint url
	QueryParams url.Values

	// ContentType is the "Content-Type" HTTP header that shall be used
	// during a request. A "Content-Type" header configured for the client
	// will be overwritten for the particular request
	ContentType string

	// CustomShouldRetryFunc optionally overrides the ShouldRetryFunc of
	// the RetryOptions specified for the client.

	CustomShouldRetryFunc RetryFunc

	// DelayAfterRetry optionally overrides the DelayAfterRetry of
	// the RetryOptions specified for the client.
	DelayAfterRetry *time.Duration

	// MaxRetries optionally overrides the MaxRetries of
	// the RetryOptions specified for the client.
	MaxRetries *int
}

// Option represents a functional Option for the Client.
type Option func(*Client)

// WithConcurrentRequestLimit sets the maximum number of concurrent requests allowed.
func WithConcurrentRequestLimit(maxConcurrent int) Option {
	return func(c *Client) {
		c.concurrentRequestLimiter = NewConcurrentRequestLimiter(maxConcurrent)
	}
}

// WithTimeout sets the request timeout for the Client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetryOptions sets the RetryOptions for the Client.
func WithRetryOptions(opts *RetryOptions) Option {
	return func(c *Client) {
		c.retryOptions = opts
	}
}

// WithHTTPListener sets the HTTPListener for the Client.
func WithHTTPListener(listener *HTTPListener) Option {
	return func(c *Client) {
		c.httpListener = listener
	}
}

// WithRateLimiter activates a RateLimiter for the Client.
// The RateLimiter will block subsequent Client calls after a 429 status code is received until the mandated reset time is
// reached. If the server should not reply with an X-RateLimit-Reset header, a default delay is enforced.
// Note that a Client with RateLimiter will not automatically retry an API call after a limit was hit, but return the
// Too Many Requests 429 Response to you.
// If the Client should retry on errors, configure RetryOptions as well.
func WithRateLimiter() Option {
	return func(c *Client) {
		c.rateLimiter = NewRateLimiter()
	}
}

// Client represents a general HTTP client
type Client struct {
	baseURL     *url.URL          // Base URL of the server
	httpClient  *http.Client      // Custom HTTP client support
	headers     map[string]string // Custom headers to be set
	headerMutex *sync.RWMutex

	concurrentRequestLimiter *ConcurrentRequestLimiter // Concurrent request limiter component (optional)
	timeout                  time.Duration             // Request timeout (optional)
	retryOptions             *RetryOptions             // Retry options (optional)
	httpListener             *HTTPListener             // HTTP listener component (optional)
	rateLimiter              *RateLimiter              // Rate limiter component (optional)
}

// NewClient creates a new instance of the Client with specified options.
func NewClient(baseURL *url.URL, httpClient *http.Client, opts ...Option) *Client {
	client := &Client{
		baseURL:     baseURL,
		headers:     make(map[string]string),
		headerMutex: &sync.RWMutex{},
		httpClient:  httpClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Do executes the given request and returns a raw *http.Response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.acquireLockAndSendWithRetries(req.Context(), req, 0, RequestOptions{})
}

// GET sends a GET request to the specified endpoint.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) GET(ctx context.Context, endpoint string, options RequestOptions) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodGet, endpoint, nil, options)
}

// PUT sends a PUT request to the specified endpoint with the given body.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) PUT(ctx context.Context, endpoint string, body io.Reader, options RequestOptions) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPut, endpoint, body, options)
}

// POST sends a POST request to the specified endpoint with the given body.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) POST(ctx context.Context, endpoint string, body io.Reader, options RequestOptions) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPost, endpoint, body, options)
}

// PATCH sends a PATCH request to the specified endpoint with the given body.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) PATCH(ctx context.Context, endpoint string, body io.Reader, options RequestOptions) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPatch, endpoint, body, options)
}

// DELETE sends a DELETE request to the specified endpoint.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) DELETE(ctx context.Context, endpoint string, options RequestOptions) (*http.Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodDelete, endpoint, nil, options)
}

// SetHeader sets a custom header for the HTTP client.
func (c *Client) SetHeader(key, value string) {
	c.headerMutex.Lock()
	defer c.headerMutex.Unlock()
	c.headers[key] = value
}

// BaseURL returns the base url configured for this client
func (c *Client) BaseURL() *url.URL {
	return c.baseURL
}

func (c *Client) acquireLockAndSendWithRetries(ctx context.Context, req *http.Request, retryCount int, options RequestOptions) (*http.Response, error) {
	// Apply concurrent request limiting if concurrentRequestLimiter is set
	if c.concurrentRequestLimiter != nil {
		c.concurrentRequestLimiter.Acquire()
		defer c.concurrentRequestLimiter.Release()
	}

	c.setHeadersOnRequest(req, options)
	return c.sendWithRetries(ctx, req, retryCount, options)
}

func (c *Client) setHeadersOnRequest(req *http.Request, options RequestOptions) {
	// Set Content-Type header accordingly
	if options.ContentType != "" {
		req.Header.Set("Content-type", options.ContentType)
	} else {
		req.Header.Set("Content-type", "application/json")
	}

	// set fixed headers
	c.headerMutex.RLock()
	defer c.headerMutex.RUnlock()
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
}

func (c *Client) sendWithRetries(ctx context.Context, req *http.Request, retryCount int, options RequestOptions) (*http.Response, error) {
	logger := logr.FromContextOrDiscard(ctx)
	if c.rateLimiter != nil {
		c.rateLimiter.Wait(ctx) // If a limit is reached, this blocks until operations are permitted again
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
		}
	}

	var reqID string
	var err error
	// wrap the body so that it could be read again
	if req.Body, err = ReusableReader(req.Body); err != nil {
		return nil, err
	}
	if c.httpListener != nil {
		reqID = uuid.NewString()
		c.httpListener.onRequest(reqID, req)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		if c.httpListener != nil {
			c.httpListener.onResponse(reqID, nil, err)
		}

		if isConnectionResetErr(err) {
			return nil, fmt.Errorf("unable to connect to host %q, connection closed unexpectedly: %w", req.Host, err)
		}

		return nil, err
	}

	// wrap the body so that it could be read again
	if response.Body, err = ReusableReader(response.Body); err != nil {
		return nil, err
	}

	if c.httpListener != nil {
		c.httpListener.onResponse(reqID, response, nil)
	}

	// Update the rate limiter with the response headers
	if c.rateLimiter != nil {
		c.rateLimiter.Update(ctx, response.StatusCode, response.Header)
	}

	// merge client retry options with request retry options
	retryOptions := mergeRetryOptions(c.retryOptions, options.CustomShouldRetryFunc, options.DelayAfterRetry, options.MaxRetries)

	if ShouldRetry(response.StatusCode) && retryOptions.ShouldRetryFunc != nil && retryCount < retryOptions.MaxRetries && retryOptions.ShouldRetryFunc(response) {
		logger.V(1).Info(fmt.Sprintf("Retrying failed request %q (HTTP %s) after %d ms delay... (try %d/%d)", req.URL, response.Status, retryOptions.DelayAfterRetry.Milliseconds(), retryCount+1, retryOptions.MaxRetries), "statusCode", response.StatusCode, "try", retryCount+1, "maxRetries", retryOptions.MaxRetries)
		time.Sleep(retryOptions.DelayAfterRetry)
		return c.sendWithRetries(ctx, req, retryCount+1, options)
	}
	return response, nil
}

// mergeRetryOptions merges the client-set retry options with the options specified for a specific request.
// The options for the request are preferred over the ones for the client
func mergeRetryOptions(clientOptions *RetryOptions, retryFunc RetryFunc, delay *time.Duration, maxRetries *int) RetryOptions {
	mergedOptions := RetryOptions{}
	if clientOptions != nil {
		mergedOptions = *clientOptions
	}
	if retryFunc != nil {
		mergedOptions.ShouldRetryFunc = retryFunc
	}
	if delay != nil {
		mergedOptions.DelayAfterRetry = *delay
	}
	if maxRetries != nil {
		mergedOptions.MaxRetries = *maxRetries
	}
	return mergedOptions
}

// sendRequestWithRetries sends an HTTP request with custom headers and modified request body, with retries if configured.
func (c *Client) sendRequestWithRetries(ctx context.Context, method string, endpoint string, body io.Reader, options RequestOptions) (*http.Response, error) {
	fullURL := c.baseURL.JoinPath(endpoint)
	if options.QueryParams != nil {
		fullURL.RawQuery = options.QueryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return nil, err
	}

	return c.acquireLockAndSendWithRetries(ctx, req, 0, options)
}

func isConnectionResetErr(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && errors.Is(urlErr, io.EOF) {
		// there is no direct way to discern a connection reset error, but if it's an url.Error wrapping an io.EOF we can be relatively certain it is
		// unless net/http stops reporting this as io.EOF
		return true
	}
	return false
}
