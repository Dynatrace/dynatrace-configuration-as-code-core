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

package testutils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestServer is a wrapper around httptest.Server that provides utility methods for testing.
type TestServer struct {
	*httptest.Server // Embedded httptest.Server for underlying server functionality.
}

// URL returns the URL of the test server.
func (t TestServer) URL() *url.URL {
	u, _ := url.Parse(t.Server.URL) //nolint:errcheck
	return u
}

// Client returns an HTTP client associated with the test server.
func (t TestServer) Client() *http.Client {
	return t.Server.Client()
}

// FaultyClient returns an HTTP client associated with the test server that always produces a network error.
func (t TestServer) FaultyClient() *http.Client {
	client := t.Server.Client()
	client.Transport = &ErrorTransport{}
	return client
}

// HTTPMethod is an alias for string, representing an HTTP method.
type HTTPMethod = string

// ServerResponses is a map of HTTP methods to expected responses for testing.
type ServerResponses map[HTTPMethod]struct {
	ResponseCode        int                 // HTTP response status code.
	ResponseBody        string              // HTTP response body.
	ValidateRequestFunc func(*http.Request) // Function to validate incoming requests.
}

// NewHTTPTestServer creates a new HTTP test server with the specified responses for each HTTP method.
func NewHTTPTestServer(t *testing.T, arg ServerResponses) *TestServer {
	return &TestServer{Server: httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if res, found := arg[req.Method]; found {
			if res.ValidateRequestFunc != nil {
				res.ValidateRequestFunc(req)
			}
			rw.WriteHeader(res.ResponseCode)
			_, _ = rw.Write([]byte(res.ResponseBody)) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		} else {
			t.Errorf("unexpected HTTP method call: %s", req.Method)
		}
	}))}
}

// ErrorTransport is custom transport that always produces a simulated network error.
type ErrorTransport struct{}

// RoundTrip implements the RoundTripper interface and returns a simulated network error.
func (t *ErrorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}
