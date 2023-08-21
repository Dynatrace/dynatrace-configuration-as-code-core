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
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	// Test WithConcurrentRequestLimit
	client := NewClient(baseURL, nil, testr.New(t), WithConcurrentRequestLimit(10))
	assert.NotNil(t, client.concurrentRequestLimiter, "Expected concurrentRequestLimiter to be set")

	// Test WithTimeout
	client = NewClient(baseURL, nil, testr.New(t), WithTimeout(5*time.Second))
	assert.Equal(t, 5*time.Second, client.timeout, "Expected timeout to be set to 5 seconds")

	// Test WithHTTPClient
	customHTTPClient := &http.Client{}
	client = NewClient(baseURL, customHTTPClient, testr.New(t))
	assert.Equal(t, customHTTPClient, client.httpClient, "Expected custom HTTP client to be set")

	// Test WithRequestRetrier
	retrier := &RequestRetrier{}
	client = NewClient(baseURL, nil, testr.New(t), WithRequestRetrier(retrier))
	assert.Equal(t, retrier, client.requestRetrier, "Expected RequestRetrier to be set")
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
	client := NewClient(baseURL, nil, testr.New(t))

	t.Run("GET", func(t *testing.T) {
		resp, err := client.GET(context.Background(), "/test", RequestOptions{})
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := client.POST(context.Background(), "/test", nil, RequestOptions{})
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201")
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := client.PUT(context.Background(), "/test", nil, RequestOptions{})
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204")
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.DELETE(context.Background(), "/test", RequestOptions{})
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204")
	})
}

func TestClient_CRUD_HTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil, testr.New(t))

	testCases := []struct {
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			method: "GET",
			requestFn: func() (*http.Response, error) {
				return client.GET(context.Background(), "/test", RequestOptions{})
			},
		},
		{
			method: "POST",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil, RequestOptions{})
			},
		},
		{
			method: "POST_WithCustomHeaders",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil, RequestOptions{})
			},
		},
		{
			method: "PUT",
			requestFn: func() (*http.Response, error) {
				return client.PUT(context.Background(), "/test", nil, RequestOptions{})
			},
		},
		{
			method: "DELETE",
			requestFn: func() (*http.Response, error) {
				return client.DELETE(context.Background(), "/test", RequestOptions{})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			resp, err := tc.requestFn()

			assert.Error(t, err, "Expected error")
			assert.Nil(t, resp, "Expected response to be nil")
			assert.IsType(t, HTTPError{}, err)
			assert.Equal(t, err.(HTTPError).Payload, []byte("Internal Server Error\n"))
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
	client := NewClient(baseURL, &http.Client{Transport: &errorTransport{}}, testr.New(t))

	testCases := []struct {
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			method: "GET",
			requestFn: func() (*http.Response, error) {
				return client.GET(context.Background(), "/test", RequestOptions{})
			},
		},
		{
			method: "POST",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil, RequestOptions{})
			},
		},
		{
			method: "POST_WithCustomHeaders",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil, RequestOptions{})
			},
		},
		{
			method: "PUT",
			requestFn: func() (*http.Response, error) {
				return client.PUT(context.Background(), "/test", nil, RequestOptions{})
			},
		},
		{
			method: "DELETE",
			requestFn: func() (*http.Response, error) {
				return client.DELETE(context.Background(), "/test", RequestOptions{})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			resp, err := tc.requestFn()

			assert.Error(t, err, "Expected error")
			assert.Nil(t, resp, "Expected response to be nil")
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
	client := NewClient(baseURL, nil, testr.New(t), WithRequestRetrier(&RequestRetrier{
		MaxRetries: 1,
		ShouldRetryFunc: func(resp *http.Response) bool {
			return resp.StatusCode != http.StatusOK
		},
	}))

	resp, err := client.GET(context.Background(), "", RequestOptions{})
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, apiHits)
}

func TestClient_WithRequestResponseRecorder(t *testing.T) {
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
				_, err := client.GET(context.TODO(), "", RequestOptions{})
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
				_, err := client.GET(context.TODO(), "", RequestOptions{})
				_, err2 := client.POST(context.TODO(), "", nil, RequestOptions{})
				return errors.Join(err, err2)
			},
			expectedRecordsRecorded: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := NewRequestResponseRecorder()
			server := httptest.NewServer(test.handler)
			defer server.Close()
			wg := sync.WaitGroup{}
			wg.Add(test.expectedRecordsRecorded)
			go func() {
				for range recorder.Channel {
					wg.Done()
				}
			}()

			baseURL, _ := url.Parse(server.URL)
			err := test.restClientCalls(NewClient(baseURL, test.httpClient, testr.New(t), WithRequestResponseRecorder(recorder)))
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

	now := time.Now()

	tests := []struct {
		name             string
		givenHeaders     map[string]string
		wantBlockingTime time.Duration
	}{
		{
			name: "simple rate limit",
			givenHeaders: map[string]string{
				"X-RateLimit-Limit": "42",
				"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(100*time.Millisecond).UnixMicro()), // 100 ms after current time
			},
			wantBlockingTime: 100 * time.Millisecond,
		},
		{
			name: "long rate limit",
			givenHeaders: map[string]string{
				"X-RateLimit-Limit": "42",
				"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(5*time.Second).UnixMicro()), // 5 sec after current time
			},
			wantBlockingTime: 5 * time.Second,
		},
		{
			name: "missing limit header is ok",
			givenHeaders: map[string]string{
				"X-RateLimit-Reset": fmt.Sprintf("%v", now.Add(100*time.Millisecond).UnixMicro()), // 100 ms after current time
			},
			wantBlockingTime: 100 * time.Millisecond,
		},
		{
			name:             "missing reset time header results in default timeout",
			givenHeaders:     map[string]string{},
			wantBlockingTime: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			apiHits := 0
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if apiHits == 0 {

					for k, v := range tt.givenHeaders {
						rw.Header().Set(k, v)
					}
					rw.WriteHeader(http.StatusTooManyRequests)

				} else {
					rw.WriteHeader(http.StatusOK)
				}
				apiHits++
			}))
			defer server.Close()

			clock := testClock{currentTime: now}

			limiter := RateLimiter{
				Clock: &clock,
			}

			baseURL, _ := url.Parse(server.URL)
			client := NewClient(baseURL, nil, testr.New(t))
			client.rateLimiter = &limiter

			_, err := client.GET(context.Background(), "", RequestOptions{})

			assert.Error(t, err)

			var httpErr HTTPError
			assert.ErrorAs(t, err, &httpErr)
			errors.As(err, &httpErr)

			assert.Equal(t, http.StatusTooManyRequests, httpErr.Code)
			assert.Nil(t, clock.requestedWait)
			resp, err := client.GET(context.Background(), "", RequestOptions{})
			if err != nil {
				t.Fatalf("failed to send GET request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.NotNil(t, clock.requestedWait)
			assert.InDelta(t, tt.wantBlockingTime.Milliseconds(), clock.requestedWait.Milliseconds(), 5, "expected limiter to clock to have blocked for time from from HTTP headers")
		})
	}
}

func TestClient_WithRateLimiting_HardLimitActuallyBlocks(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping real time (0.5sec) rate limiting test in short test mode")
	}

	expectedWaitTime := 500 * time.Millisecond

	apiHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if apiHits == 0 {
			rw.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%v", time.Now().Add(expectedWaitTime).UnixMicro()))
			rw.WriteHeader(http.StatusTooManyRequests)

		} else {
			rw.WriteHeader(http.StatusOK)
		}
		apiHits++
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil, testr.New(t),
		WithRateLimiter(), // default rate limiter with real clock
	)

	_, err := client.GET(context.Background(), "", RequestOptions{})
	assert.Error(t, err)

	before := time.Now()
	resp, err := client.GET(context.Background(), "", RequestOptions{})
	after := time.Now()

	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	diff := after.Sub(before)

	assert.GreaterOrEqual(t, diff, expectedWaitTime, "expected to rate limited rest call to take at least as long as the rate limit timeout")
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
	client := NewClient(baseURL, nil, testr.New(t),
		WithRateLimiter(), // default rate limiter with real clock
	)

	_, _ = client.GET(context.Background(), "", RequestOptions{}) // first call initializes rate limiter

	before := time.Now()
	for i := 0; i < rps+1; i++ {
		_, _ = client.GET(context.Background(), "", RequestOptions{})
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
	client := NewClient(baseURL, nil, testr.New(t),
		WithRateLimiter(), // default rate limiter with real clock
	)

	_, _ = client.GET(context.Background(), "", RequestOptions{}) // first call initializes rate limiter

	before := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = client.GET(context.Background(), "", RequestOptions{})
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
	client := NewClient(baseURL, nil, testr.New(t),
		WithRequestRetrier(&RequestRetrier{
			MaxRetries:      5,
			ShouldRetryFunc: RetryIfNotSuccess,
		}),
		WithRateLimiter(), // default rate limiter with real clock
	)

	resp, err := client.GET(context.Background(), "/sample/endpoint", RequestOptions{url.Values{"type": {"car", "bike"}}})
	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{ "hello": "there" }`, string(b))
}

func TestClient_RequestOptionsQueryParams(t *testing.T) {
	expectedQueryParams := url.Values{"foo": {"bar", "baz"}}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, expectedQueryParams, req.URL.Query())
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, nil, testr.New(t))
	client.GET(context.TODO(), "", RequestOptions{expectedQueryParams})
}
