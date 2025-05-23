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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/pointer"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	// Test WithConcurrentRequestLimit
	client := NewClient(baseURL, nil, WithConcurrentRequestLimit(10))
	assert.NotNil(t, client.concurrentRequestLimiter, "Expected concurrentRequestLimiter to be set")

	// Test WithTimeout
	client = NewClient(baseURL, nil, WithTimeout(5*time.Second))
	assert.Equal(t, 5*time.Second, client.timeout, "Expected timeout to be set to 5 seconds")

	// Test WithHTTPClient
	customHTTPClient := &http.Client{}
	client = NewClient(baseURL, customHTTPClient)
	assert.Equal(t, customHTTPClient, client.httpClient, "Expected custom HTTP client to be set")

	// Test WithRetryOptions
	opts := &RetryOptions{}
	client = NewClient(baseURL, nil, WithRetryOptions(opts))
	assert.Equal(t, opts, client.retryOptions, "Expected RetryOptions to be set")
}

func TestClient_CRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Always return 200 OK for GET requests
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil)

	t.Run("GET", func(t *testing.T) {
		resp, err := client.GET(t.Context(), "/test", RequestOptions{})

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")
		assert.Equal(t, http.MethodGet, resp.Request.Method)
		assert.Equal(t, baseURL.String()+"/test", resp.Request.URL.String())
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := client.POST(t.Context(), "/test", nil, RequestOptions{})

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201")
		assert.Equal(t, http.MethodPost, resp.Request.Method)
		assert.Equal(t, baseURL.String()+"/test", resp.Request.URL.String())
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := client.PUT(t.Context(), "/test", nil, RequestOptions{})

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204")
		assert.Equal(t, http.MethodPut, resp.Request.Method)
		assert.Equal(t, baseURL.String()+"/test", resp.Request.URL.String())
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.DELETE(t.Context(), "/test", RequestOptions{})

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204")
		assert.Equal(t, http.MethodDelete, resp.Request.Method)
		assert.Equal(t, baseURL.String()+"/test", resp.Request.URL.String())
	})
}

func TestClient_CRUD_HTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil)

	ctxWithLogger := testutils.ContextWithLogger(t)

	testCases := []struct {
		name      string
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			name:   "GET",
			method: http.MethodGet,
			requestFn: func() (*http.Response, error) {
				return client.GET(ctxWithLogger, "/test", RequestOptions{})
			},
		},
		{
			name:   "POST",
			method: http.MethodPost,
			requestFn: func() (*http.Response, error) {
				return client.POST(ctxWithLogger, "/test", nil, RequestOptions{})
			},
		},
		{
			name:   "POST_WithCustomHeaders",
			method: http.MethodPost,
			requestFn: func() (*http.Response, error) {
				return client.POST(ctxWithLogger, "/test", nil, RequestOptions{})
			},
		},
		{
			name:   "PUT",
			method: http.MethodPut,
			requestFn: func() (*http.Response, error) {
				return client.PUT(ctxWithLogger, "/test", nil, RequestOptions{})
			},
		},
		{
			name:   "DELETE",
			method: http.MethodDelete,
			requestFn: func() (*http.Response, error) {
				return client.DELETE(t.Context(), "/test", RequestOptions{})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.requestFn()

			assert.Nil(t, err, "Expected not error")
			assert.NotNil(t, resp, "Expected response to be not nil")
			assert.Equal(t, tc.method, resp.Request.Method, "Expected request info method to be "+tc.method)
			assert.Equal(t, baseURL.String()+"/test", resp.Request.URL.String(), "Expected request info url to be "+baseURL.String()+"/test")

			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			assert.Equal(t, body, []byte("Internal Server Error\n"))

		})
	}
}

type errorTransport struct{}

func (t *errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}

func TestClient_CRUD_TransportErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200) // a network error should be forced before the server can ever reply success
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, &http.Client{Transport: &errorTransport{}})

	ctx := testutils.ContextWithLogger(t)

	testCases := []struct {
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			method: "GET",
			requestFn: func() (*http.Response, error) {
				return client.GET(ctx, "/test", RequestOptions{})
			},
		},
		{
			method: "POST",
			requestFn: func() (*http.Response, error) {
				return client.POST(ctx, "/test", nil, RequestOptions{})
			},
		},
		{
			method: "POST_WithCustomHeaders",
			requestFn: func() (*http.Response, error) {
				return client.POST(ctx, "/test", nil, RequestOptions{})
			},
		},
		{
			method: "PUT",
			requestFn: func() (*http.Response, error) {
				return client.PUT(ctx, "/test", nil, RequestOptions{})
			},
		},
		{
			method: "DELETE",
			requestFn: func() (*http.Response, error) {
				return client.DELETE(ctx, "/test", RequestOptions{})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			resp, err := tc.requestFn()

			assert.Error(t, err, "Expected error")
			assert.Zero(t, resp, "Expected response to be zero")
		})
	}
}

func TestClient_CRUD_EOFIsWrappedInUserfriendlyError(t *testing.T) {
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		server.CloseClientConnections() // cause a connection reset on request
	})
	server.Start()
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, server.Client())

	ctx := testutils.ContextWithLogger(t)

	testCases := []struct {
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			method: "GET",
			requestFn: func() (*http.Response, error) {
				return client.GET(ctx, "/test", RequestOptions{})
			},
		},
		{
			method: "POST",
			requestFn: func() (*http.Response, error) {
				return client.POST(ctx, "/test", nil, RequestOptions{})
			},
		},
		{
			method: "POST_WithCustomHeaders",
			requestFn: func() (*http.Response, error) {
				return client.POST(ctx, "/test", nil, RequestOptions{})
			},
		},
		{
			method: "PUT",
			requestFn: func() (*http.Response, error) {
				return client.PUT(ctx, "/test", nil, RequestOptions{})
			},
		},
		{
			method: "DELETE",
			requestFn: func() (*http.Response, error) {
				return client.DELETE(ctx, "/test", RequestOptions{})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			_, err := tc.requestFn()

			assert.Error(t, err, "Expected error")
			assert.ErrorContains(t, err, "connection closed unexpectedly")
		})
	}
}

func TestClient_WithRetries(t *testing.T) {
	apiHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if apiHits == 0 {
			rw.WriteHeader(http.StatusBadRequest)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		apiHits++
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil, WithRetryOptions(&RetryOptions{
		MaxRetries: 1,
		ShouldRetryFunc: func(resp *http.Response) bool {
			return resp.StatusCode != http.StatusOK
		},
	}))

	ctx := testutils.ContextWithLogger(t)

	resp, err := client.GET(ctx, "", RequestOptions{})
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, apiHits)
}

func TestClient_WithCustomRetriesOnRequest(t *testing.T) {
	apiHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusTooManyRequests)
		apiHits++
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil, WithRetryOptions(&RetryOptions{
		MaxRetries:      10,
		DelayAfterRetry: time.Duration(0),
		ShouldRetryFunc: func(resp *http.Response) bool {
			return resp.StatusCode != http.StatusOK
		},
	}))

	ctx := testutils.ContextWithLogger(t)

	startTime := time.Now()
	resp, err := client.GET(ctx, "", RequestOptions{MaxRetries: pointer.Pointer(1), DelayAfterRetry: pointer.Pointer(time.Millisecond * 100)})
	elapsedTime := time.Since(startTime)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, elapsedTime, time.Millisecond*100)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, 2, apiHits)
}

func TestClient_WithIgnoredRetries(t *testing.T) {
	apiHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusForbidden)
		apiHits++
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil, WithRetryOptions(&RetryOptions{
		MaxRetries: 1,
		ShouldRetryFunc: func(resp *http.Response) bool {
			return resp.StatusCode != http.StatusOK
		},
	}))

	ctx := testutils.ContextWithLogger(t)

	resp, err := client.GET(ctx, "", RequestOptions{})
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, 1, apiHits)
}

func TestClient_WithHTTPListener(t *testing.T) {

	ctx := testutils.ContextWithLogger(t)

	tests := []struct {
		name                    string
		httpClient              *http.Client
		handler                 http.HandlerFunc
		restClientCalls         func(client *Client) error
		expectedRecordsRecorded int
		expectError             bool
	}{
		{
			name:       "One GET call - results in two records being recorded (request + response)",
			httpClient: &http.Client{},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("{}"))
			},
			restClientCalls: func(client *Client) error {
				_, err := client.GET(t.Context(), "", RequestOptions{})
				return err
			},
			expectedRecordsRecorded: 2,
		},
		{
			name:       "One Get and one Post call - results in four records being recorded (get request + response + post request + response)",
			httpClient: &http.Client{},
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("{}"))
			},
			restClientCalls: func(client *Client) error {
				_, err := client.GET(ctx, "", RequestOptions{})
				_, err2 := client.POST(ctx, "", nil, RequestOptions{})
				return errors.Join(err, err2)
			},
			expectedRecordsRecorded: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server := httptest.NewServer(test.handler)
			defer server.Close()
			wg := sync.WaitGroup{}
			wg.Add(test.expectedRecordsRecorded)
			httpListener := &HTTPListener{
				Callback: func(response RequestResponse) {
					wg.Done()
				},
			}

			baseURL, _ := url.Parse(server.URL)
			err := test.restClientCalls(NewClient(baseURL, test.httpClient, WithHTTPListener(httpListener)))
			wg.Wait()
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type testClock struct {
	currentTime   time.Time
	requestedWait *time.Duration
}

func (t *testClock) Now() time.Time {
	return t.currentTime
}

func (t *testClock) After(d time.Duration) <-chan time.Time {
	t.requestedWait = &d

	c := make(chan time.Time, 1)
	defer func() {
		c <- t.currentTime.Add(d)
	}()
	return c
}

func TestClient_WithRateLimiting(t *testing.T) {
	now := time.Date(2022, time.August, 14, 9, 30, 0, 0, time.Local)

	tests := []struct {
		name             string
		givenHeaders     map[string]string
		wantBlockingTime time.Duration
	}{
		{
			name: "simple rate limit",
			givenHeaders: map[string]string{
				"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(time.Second).Unix()),
			},
			wantBlockingTime: time.Second,
		},
		{
			name: "long rate limit",
			givenHeaders: map[string]string{
				"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(5*time.Second).Unix()),
			},
			wantBlockingTime: 5 * time.Second,
		},
		{
			name: "missing limit header is ok",
			givenHeaders: map[string]string{
				"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(time.Second).Unix()),
			},
			wantBlockingTime: time.Second,
		},
		{
			name:             "missing reset time header results in default timeout",
			givenHeaders:     map[string]string{},
			wantBlockingTime: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			apiHits := 0
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				for k, v := range tt.givenHeaders {
					rw.Header().Set(k, v)
				}
				rw.WriteHeader(http.StatusTooManyRequests)
				apiHits++
			}))
			defer server.Close()

			clock := testClock{currentTime: now}

			limiter := RateLimiter{
				Clock: &clock,
			}

			baseURL, _ := url.Parse(server.URL)
			client := NewClient(baseURL, nil)
			client.rateLimiter = &limiter
			client.retryOptions = &RetryOptions{MaxRetries: 1, ShouldRetryFunc: RetryIfNotSuccess}

			ctx := testutils.ContextWithVerboseLogger(t)

			resp, err := client.GET(ctx, "irrelevant", RequestOptions{})
			assert.NoError(t, err)
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
			assert.Equal(t, tt.wantBlockingTime.Seconds(), clock.requestedWait.Seconds())
		})
	}
}

func TestClient_WithRateLimiting_HardLimitActuallyBlocks(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping real time (0.5sec) rate limiting test in short test mode")
	}

	expectedWaitTime := 5 * time.Second
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		rw.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%v", time.Now().Add(expectedWaitTime).Unix()))
		rw.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil,
		WithRateLimiter(), // default rate limiter with real clock
		WithRetryOptions(&RetryOptions{MaxRetries: 1, ShouldRetryFunc: RetryIfTooManyRequests}),
	)

	ctx := testutils.ContextWithVerboseLogger(t)
	before := time.Now()
	client.GET(ctx, "", RequestOptions{})
	after := time.Now()
	diff := after.Sub(before)

	assert.InDelta(t, expectedWaitTime.Milliseconds(), diff.Milliseconds(), 1000)
}

func TestClient_WithRateLimiting_SoftLimitActuallyBlocks(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping real time (1sec) rate limiting test in short test mode")
	}

	rps := 2

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-RateLimit-Limit", strconv.Itoa(rps))
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil,
		WithRateLimiter(), // default rate limiter with real clock
	)

	ctx := testutils.ContextWithVerboseLogger(t)

	_, _ = client.GET(ctx, "", RequestOptions{}) // first call initializes rate limiter

	before := time.Now()
	for i := 0; i < rps+1; i++ {
		_, _ = client.GET(ctx, "", RequestOptions{})
	}

	after := time.Now()
	diff := after.Sub(before)

	assert.Greater(t, diff, time.Second, "expected client to be rate limited to %d calls per second and need more than 1sec for %d calls", rps, rps+1)
}

func TestClient_WithRateLimiting_SoftLimitCanBeUpdated(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping real time (1-2sec) rate limiting test in short test mode")
	}

	apiCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if apiCalls > 2 {
			rw.Header().Set("X-RateLimit-Limit", "500")
		} else {
			rw.Header().Set("X-RateLimit-Limit", "2")
		}

		rw.WriteHeader(http.StatusOK)
		apiCalls++
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil,
		WithRateLimiter(), // default rate limiter with real clock
	)

	ctx := testutils.ContextWithVerboseLogger(t)

	_, _ = client.GET(ctx, "", RequestOptions{}) // first call initializes rate limiter

	before := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = client.GET(ctx, "", RequestOptions{})
	}

	after := time.Now()
	diff := after.Sub(before)

	assert.Greater(t, diff, time.Second, "expected client to be limited to 2 calls per second for the first two calls, thus take at least 1 sec to complete 100 calls")
	assert.Less(t, diff, 2*time.Second, "expected only the first 2 calls to be limited to 2 call per second, and the other 98 to complete in less than 1 sec under a new limit of 500rps")
}

func TestClient_WithRetriesAndRateLimit(t *testing.T) {

	now := time.Now()

	responses := []struct {
		code    int
		headers map[string]string
		body    string
	}{
		{400, map[string]string{}, "{}"},
		{503, map[string]string{}, "{}"},
		{429, map[string]string{
			"X-RateLimit-Limit": "42",
			"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(5*time.Millisecond).UnixMicro()), // 5 ms after current time
		}, "{}"},
		{200, map[string]string{}, `{ "hello": "there" }`},
	}
	apiCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if apiCalls > len(responses) {
			t.Fatal("test server received more calls than expected")
		}

		resp := responses[apiCalls]
		apiCalls++

		rw.WriteHeader(resp.code)
		for k, v := range resp.headers {
			rw.Header().Set(k, v)
		}
		_, _ = rw.Write([]byte(resp.body))
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil,
		WithRetryOptions(&RetryOptions{
			MaxRetries:      5,
			ShouldRetryFunc: RetryIfNotSuccess,
		}),
		WithRateLimiter(), // default rate limiter with real clock
	)

	ctx := testutils.ContextWithVerboseLogger(t)

	resp, err := client.GET(ctx, "/sample/endpoint", RequestOptions{QueryParams: url.Values{"type": {"car", "bike"}}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Equal(t, `{ "hello": "there" }`, string(body))
}

func TestClient_RequestOptionsQueryParams(t *testing.T) {
	expectedQueryParams := url.Values{"foo": {"bar", "baz"}}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, expectedQueryParams, req.URL.Query())
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil)

	ctx := testutils.ContextWithLogger(t)

	client.GET(ctx, "", RequestOptions{QueryParams: expectedQueryParams})
}

func TestClient_RequestOptionsCustomShouldRetryFunc(t *testing.T) {
	apiHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if apiHits == 0 {
			rw.WriteHeader(http.StatusBadRequest)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		apiHits++
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, server.Client(), WithRetryOptions(&RetryOptions{MaxRetries: 1, ShouldRetryFunc: func(resp *http.Response) bool { return false }}))

	ctx := testutils.ContextWithLogger(t)

	resp, err := client.GET(ctx, "", RequestOptions{CustomShouldRetryFunc: func(resp *http.Response) bool { return resp.StatusCode != http.StatusOK }})
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, apiHits)
}
