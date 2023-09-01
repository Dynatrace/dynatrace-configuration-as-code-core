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
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RequestOptions are additional options that should be applied
// to a request
type RequestOptions struct {
	// QueryParams are HTTP query parameters that shall be appended
	// to the endpoint url
	QueryParams url.Values
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

// WithRequestRetrier sets the RequestRetrier for the Client.
func WithRequestRetrier(retrier *RequestRetrier) Option {
	return func(c *Client) {
		c.requestRetrier = retrier
	}
}

// WithRequestResponseRecorder sets the RequestResponseRecorder for the Client.
// Note, that this is happening sync. so there MUST be a listener gabbing the
// recorded messages, to not block the client
func WithRequestResponseRecorder(recorder *RequestResponseRecorder) Option {
	return func(c *Client) {
		c.requestResponseRecorder = recorder
	}
}

// WithRateLimiter activates a RateLimiter for the Client.
// The RateLimiter will block subsequent Client calls after a 429 status code is received until the mandated reset time is
// reached. If the server should not reply with an X-RateLimit-Reset header, a default delay is enforced.
// Note that a Client with RateLimiter will not automatically retry an API call after a limit was hit, but return the
// Too Many Requests 429 Response to you.
// If wish for the Client to retry on errors configure a RequestRetrier as well.
func WithRateLimiter() Option {
	return func(c *Client) {
		c.rateLimiter = NewRateLimiter()
	}
}

// Client represents a general HTTP client
type Client struct {
	baseURL    *url.URL          // Base URL of the server
	httpClient *http.Client      // Custom HTTP client support
	headers    map[string]string // Custom headers to be set

	concurrentRequestLimiter *ConcurrentRequestLimiter // Concurrent request limiter component (optional)
	timeout                  time.Duration             // Request timeout (optional)
	requestRetrier           *RequestRetrier           // HTTP request retrier component (optional)
	requestResponseRecorder  *RequestResponseRecorder  // Request-response recorder component (optional)
	rateLimiter              *RateLimiter              // Rate limiter component (optional)
}

// NewClient creates a new instance of the Client with specified options.
func NewClient(baseURL *url.URL, httpClient *http.Client, opts ...Option) *Client {
	client := &Client{
		baseURL:    baseURL,
		headers:    make(map[string]string),
		httpClient: httpClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// GET sends a GET request to the specified endpoint.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) GET(ctx context.Context, endpoint string, options RequestOptions) (Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodGet, endpoint, nil, 0, options)
}

// PUT sends a PUT request to the specified endpoint with the given body.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) PUT(ctx context.Context, endpoint string, body io.Reader, options RequestOptions) (Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPut, endpoint, body, 0, options)
}

// POST sends a POST request to the specified endpoint with the given body.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) POST(ctx context.Context, endpoint string, body io.Reader, options RequestOptions) (Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodPost, endpoint, body, 0, options)
}

// DELETE sends a DELETE request to the specified endpoint.
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
func (c *Client) DELETE(ctx context.Context, endpoint string, options RequestOptions) (Response, error) {
	return c.sendRequestWithRetries(ctx, http.MethodDelete, endpoint, nil, 0, options)
}

// SetHeader sets a custom header for the HTTP client.
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// sendRequestWithRetries sends an HTTP request with custom headers and modified request body, with retries if configured.
func (c *Client) sendRequestWithRetries(ctx context.Context, method string, endpoint string, body io.Reader, retryCount int, options RequestOptions) (Response, error) {

	logger := logr.FromContextOrDiscard(ctx)

	if c.rateLimiter != nil {
		c.rateLimiter.Wait(ctx) // If a limit is reached, this blocks until operations are permitted again
	}

	// Apply concurrent request limiting if concurrentRequestLimiter is set
	if c.concurrentRequestLimiter != nil {
		c.concurrentRequestLimiter.Acquire()
		defer c.concurrentRequestLimiter.Release()
	}

	fullURL := c.baseURL.JoinPath(endpoint)
	if options.QueryParams != nil {
		fullURL.RawQuery = options.QueryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Content-type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
		}
	}

	if c.requestResponseRecorder != nil {
		// wrap the body so that it could be read again
		if req.Body, err = ReusableReader(req.Body); err != nil {
			return Response{}, err
		}
		c.requestResponseRecorder.RecordRequest(ctx, req)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		if c.requestResponseRecorder != nil {
			c.requestResponseRecorder.RecordResponse(ctx, nil, err)
		}

		if isConnectionResetErr(err) {
			return Response{}, fmt.Errorf("unable to connect to host %q, connection closed unexpectedly: %w", req.Host, err)
		}

		return Response{}, err
	}

	if c.requestResponseRecorder != nil {
		// wrap the body so that it could be read again
		if response.Body, err = ReusableReader(response.Body); err != nil {
			return Response{}, err
		}
		c.requestResponseRecorder.RecordResponse(ctx, response, nil)
	}

	// Update the rate limiter with the response headers
	if c.rateLimiter != nil {
		c.rateLimiter.Update(ctx, response.StatusCode, response.Header)
	}

	if c.requestRetrier != nil && retryCount < c.requestRetrier.MaxRetries &&
		c.requestRetrier.ShouldRetryFunc != nil && c.requestRetrier.ShouldRetryFunc(response) {
		logger.V(1).Info(fmt.Sprintf("Retrying failed request %q (HTTP %s) after %d ms delay... (try %d/%d)", fullURL, response.Status, 100, retryCount+1, c.requestRetrier.MaxRetries), "statusCode", response.StatusCode, "try", retryCount+1, "maxRetries", c.requestRetrier.MaxRetries)
		time.Sleep(100 * time.Millisecond)
		return c.sendRequestWithRetries(ctx, method, endpoint, body, retryCount+1, options)
	}

	// Read payload
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to read response body of failed request %q (HTTP %s)", fullURL, response.Status))
	}

	if err := response.Body.Close(); err != nil {
		logger.V(1).Error(err, "Failed to close response body of failed request")
	}

	return Response{
		Payload:    payload,
		StatusCode: response.StatusCode,
		RequestInfo: RequestInfo{
			Method: req.Method,
			URL:    req.URL.String(),
		}}, nil
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
