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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestServer is a wrapper around httptest.Server that provides utility methods for testing.
type TestServer struct {
	calls            int
	*httptest.Server // Embedded httptest.Server for underlying server functionality.
}

// URL returns the URL of the test server.
func (t TestServer) URL() *url.URL {
	u, _ := url.Parse(t.Server.URL) //nolint:errcheck
	return u
}

// Calls returns the number of calls invoked on the test server
func (t TestServer) Calls() int {
	return t.calls
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

// NewHTTPTestServer creates a new HTTP test server with the specified responses for each HTTP method.
func NewHTTPTestServer(t *testing.T, responses []ResponseDef) *TestServer {
	t.Helper()
	for i, r := range responses {
		// it's only allowed to set ONE handler per response
		if !checkExactlyOneHandlerSet(r) {
			panic(fmt.Sprintf("Response nr. %d has more than one handler defined", i))
		}
	}

	testServer := &TestServer{}
	handler := func(rw http.ResponseWriter, req *http.Request) {
		t.Helper()
		testServer.calls++
		if len(responses) <= testServer.calls-1 {
			t.Fatalf("Exceeded number of calls to test server (expected: %d), request: %s %s - %s", len(responses), req.Method, req.URL, req.Body)
		}

		responseDef := responses[testServer.calls-1]
		handlers := map[string]func(*testing.T, *http.Request) Response{
			http.MethodGet:    responseDef.Get,
			http.MethodPost:   responseDef.Post,
			http.MethodPut:    responseDef.Put,
			http.MethodDelete: responseDef.Delete,
			http.MethodPatch:  responseDef.Patch,
		}

		handlerFunc, found := handlers[req.Method]
		if !found {
			panic(fmt.Sprintf("No %s method defined for server call nr. %d", req.Method, testServer.calls))
		}
		response := handlerFunc(t, req)
		if response.ContentType != "" {
			rw.Header().Set("Content-Type", response.ContentType)
		} else {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		rw.WriteHeader(response.ResponseCode)
		_, err := rw.Write([]byte(response.ResponseBody)) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		if err != nil {
			panic(err)
		}
		responseDef.Validate(t, req)
	}
	testServer.Server = httptest.NewServer(http.HandlerFunc(handler))
	return testServer
}

type Response struct {
	ResponseCode int
	ResponseBody string
	ContentType  string
}

type ResponseDef struct {
	GET             func(*testing.T, *http.Request) Response
	PUT             func(*testing.T, *http.Request) Response
	POST            func(*testing.T, *http.Request) Response
	DELETE          func(*testing.T, *http.Request) Response
	PATCH           func(*testing.T, *http.Request) Response
	ValidateRequest func(*testing.T, *http.Request)
}

func (r ResponseDef) Get(t *testing.T, req *http.Request) Response {
	t.Helper()
	if r.GET == nil {
		fatal(t, req)
	}
	return r.GET(t, req)
}
func (r ResponseDef) Put(t *testing.T, req *http.Request) Response {
	t.Helper()
	if r.PUT == nil {
		fatal(t, req)
	}
	return r.PUT(t, req)
}
func (r ResponseDef) Post(t *testing.T, req *http.Request) Response {
	t.Helper()
	if r.POST == nil {
		fatal(t, req)
	}
	return r.POST(t, req)
}

func (r ResponseDef) Delete(t *testing.T, req *http.Request) Response {
	t.Helper()
	if r.DELETE == nil {
		fatal(t, req)
	}
	return r.DELETE(t, req)
}
func (r ResponseDef) Patch(t *testing.T, req *http.Request) Response {
	t.Helper()
	if r.PATCH == nil {
		fatal(t, req)
	}
	return r.PATCH(t, req)
}
func fatal(t testing.TB, req *http.Request) {
	t.Helper()
	t.Fatalf("%s invoked, but not defined\n\trequested: %s %s", req.Method, req.Method, req.RequestURI)
}

func (r ResponseDef) Validate(t *testing.T, req *http.Request) {
	if r.ValidateRequest != nil {
		r.ValidateRequest(t, req)
	}
}

// ErrorTransport is custom transport that always produces a simulated network error.
type ErrorTransport struct{}

// RoundTrip implements the RoundTripper interface and returns a simulated network error.
func (t *ErrorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}

func checkExactlyOneHandlerSet(def ResponseDef) bool {
	var count int
	if def.GET != nil {
		count++
	}
	if def.POST != nil {
		count++
	}
	if def.PUT != nil {
		count++
	}
	if def.DELETE != nil {
		count++
	}
	if def.PATCH != nil {
		count++
	}
	return count == 1
}
