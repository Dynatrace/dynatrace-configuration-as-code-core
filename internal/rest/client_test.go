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

package rest_test

import (
	"context"
	"errors"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/v1/internal/rest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Test WithConcurrentRequestLimit
	client := rest.NewClient("https://example.com", rest.WithConcurrentRequestLimit(10))
	assert.NotNil(t, client.ConcurrentRequestLimiter, "Expected ConcurrentRequestLimiter to be set")

	// Test WithTimeout
	client = rest.NewClient("https://example.com", rest.WithTimeout(5*time.Second))
	assert.Equal(t, 5*time.Second, client.Timeout, "Expected timeout to be set to 5 seconds")

	// Test WithHTTPClient
	customHTTPClient := &http.Client{}
	client = rest.NewClient("https://example.com", rest.WithHTTPClient(customHTTPClient))
	assert.Equal(t, customHTTPClient, client.HTTPClient, "Expected custom HTTP client to be set")

	// Test WithRequestRetrier
	retrier := &rest.RequestRetrier{}
	client = rest.NewClient("https://example.com", rest.WithRequestRetrier(retrier))
	assert.Equal(t, retrier, client.RequestRetrier, "Expected RequestRetrier to be set")
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

	client := rest.NewClient(server.URL)

	t.Run("GET", func(t *testing.T) {
		resp, err := client.GET(context.Background(), "/test")
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := client.POST(context.Background(), "/test", nil)
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201")
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := client.PUT(context.Background(), "/test", nil)
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204")
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.DELETE(context.Background(), "/test")
		defer resp.Body.Close()

		assert.NoError(t, err, "Unexpected error")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204")
	})
}

func TestClient_CRUD_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()
	client := rest.NewClient(server.URL)

	testCases := []struct {
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			method: "GET",
			requestFn: func() (*http.Response, error) {
				return client.GET(context.Background(), "/test")
			},
		},
		{
			method: "POST",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil)
			},
		},
		{
			method: "POST_WithCustomHeaders",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil)
			},
		},
		{
			method: "PUT",
			requestFn: func() (*http.Response, error) {
				return client.PUT(context.Background(), "/test", nil)
			},
		},
		{
			method: "DELETE",
			requestFn: func() (*http.Response, error) {
				return client.DELETE(context.Background(), "/test")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			resp, err := tc.requestFn()

			assert.Error(t, err, "Expected error")
			assert.Nil(t, resp, "Expected response to be nil")
			assert.IsType(t, rest.HTTPError{}, err)
			assert.Equal(t, err.(rest.HTTPError).Payload, []byte("Internal Server Error\n"))
		})
	}
}

func TestClient_CRUD_Errors_2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()
	client := rest.NewClient(server.URL, rest.WithHTTPClient(&http.Client{Transport: &errorTransport{}}))

	testCases := []struct {
		method    string
		requestFn func() (*http.Response, error)
	}{
		{
			method: "GET",
			requestFn: func() (*http.Response, error) {
				return client.GET(context.Background(), "/test")
			},
		},
		{
			method: "POST",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil)
			},
		},
		{
			method: "POST_WithCustomHeaders",
			requestFn: func() (*http.Response, error) {
				return client.POST(context.Background(), "/test", nil)
			},
		},
		{
			method: "PUT",
			requestFn: func() (*http.Response, error) {
				return client.PUT(context.Background(), "/test", nil)
			},
		},
		{
			method: "DELETE",
			requestFn: func() (*http.Response, error) {
				return client.DELETE(context.Background(), "/test")
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

	client := rest.NewClient(server.URL, rest.WithRequestRetrier(&rest.RequestRetrier{
		MaxRetries: 1,
		ShouldRetryFunc: func(resp *http.Response) bool {
			return resp.StatusCode != http.StatusOK
		},
	}))

	resp, err := client.GET(context.Background(), "")
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
		restClientCalls         func(client *rest.Client) error
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
			restClientCalls: func(client *rest.Client) error {
				_, err := client.GET(context.TODO(), "")
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
			restClientCalls: func(client *rest.Client) error {
				_, err := client.GET(context.TODO(), "")
				_, err2 := client.POST(context.TODO(), "", nil)
				return errors.Join(err, err2)
			},
			expectedRecordsRecorded: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := rest.NewRequestResponseRecorder()
			server := httptest.NewServer(test.handler)
			defer server.Close()
			wg := sync.WaitGroup{}
			wg.Add(test.expectedRecordsRecorded)
			go func() {
				for range recorder.Channel {
					wg.Done()
				}
			}()
			err := test.restClientCalls(rest.NewClient(server.URL, rest.WithHTTPClient(test.httpClient), rest.WithRequestResponseRecorder(recorder)))
			wg.Wait()
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type errorTransport struct{}

func (t *errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}
