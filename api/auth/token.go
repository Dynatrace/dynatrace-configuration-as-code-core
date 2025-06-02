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

package auth

import (
	"fmt"
	"net/http"
)

type tokenType = string

const (
	apiToken      tokenType = "Api-Token"
	platformToken tokenType = "Bearer"
)

// NewTokenBasedClient creates a new HTTP client with token-based authentication.
// It takes a token string as an argument and returns an instance of *http.Client.
func NewTokenBasedClient(token string) *http.Client {
	// Create a new tokenAuthTransport and initialize it with the provided token.
	return &http.Client{Transport: newTokenAuthTransport(nil, token, apiToken)}
}

// NewTokenBasedClient creates a new HTTP client with platform token-based authentication.
// It takes a token string as an argument and returns an instance of *http.Client.
func NewPlatformTokenBasedClient(token string) *http.Client {
	// Create a new tokenAuthTransport and initialize it with the provided token.
	return &http.Client{Transport: newTokenAuthTransport(nil, token, platformToken)}
}

// tokenAuthTransport is a custom transport that adds token-based authentication headers to HTTP requests.
type tokenAuthTransport struct {
	http.RoundTripper
	header http.Header
}

// newTokenAuthTransport creates a new instance of tokenAuthTransport.
// It takes a baseTransport (an existing HTTP transport) and a token string as arguments,
// and returns a pointer to the newly created tokenAuthTransport instance.
func newTokenAuthTransport(baseTransport http.RoundTripper, token string, tType tokenType) *tokenAuthTransport {
	// If no baseTransport is provided, use the default HTTP transport.
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	// Create a new tokenAuthTransport instance and initialize it.
	t := &tokenAuthTransport{
		RoundTripper: baseTransport,
		header:       http.Header{},
	}

	// Set the "Authorization" header with the provided token.
	t.header.Set("Authorization", fmt.Sprintf("%s %s", tType, token))
	return t
}

// RoundTrip implements the http.RoundTripper interface's RoundTrip method.
// It adds the authentication headers from the tokenAuthTransport instance to the request's headers
// and delegates the actual round trip to the underlying transport.
func (t *tokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Copy authentication headers from tokenAuthTransport to the request.
	for k, v := range t.header {
		req.Header[k] = v
	}

	// Perform the actual HTTP request using the underlying transport.
	return t.RoundTripper.RoundTrip(req)
}
