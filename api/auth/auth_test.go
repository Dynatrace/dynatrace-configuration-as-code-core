/*
 * @license
 * Copyright 2025 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/testutils"
)

func TestNewOAuthBasedClient_TokenSetCorrectly(t *testing.T) {
	// Mock OAuth2 token server
	tokenServer := testutils.OAuthMockServer(t, "mocked-token")
	defer tokenServer.Close()

	// OAuth2 client credentials config
	config := &clientcredentials.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     tokenServer.URL,
	}

	client := NewOAuthClient(t.Context(), config)

	// Mock API server to verify Authorization header
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer mocked-token", auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	// Make a request to the mock API server
	resp, err := client.Get(apiServer.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
}

func TestNewPlatformTokenSourceClient(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer token-from-source", auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	client := NewPlatformTokenSourceClient(t.Context(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token-from-source"}))

	// Make a request to the mock API server
	resp, err := client.Get(apiServer.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
}

// tokenSequenceSource hands out a different, unexpired token on every call. A client that caches tokens -
// which is what oauth2.NewClient would do by wrapping the source in an oauth2.ReuseTokenSource - therefore
// produces a different sequence of Authorization headers than one that consults the source per request.
type tokenSequenceSource struct {
	calls int
}

func (s *tokenSequenceSource) Token() (*oauth2.Token, error) {
	s.calls++
	return &oauth2.Token{AccessToken: fmt.Sprintf("token-%d", s.calls), Expiry: time.Now().Add(time.Hour)}, nil
}

func TestNewPlatformTokenSourceClient_SourceIsConsultedOnEveryRequest(t *testing.T) {
	// The server echoes the Authorization header back so that the test goroutine, rather than the
	// handler goroutine, collects it.
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Header.Get("Authorization")))
	}))
	defer apiServer.Close()

	client := NewPlatformTokenSourceClient(t.Context(), &tokenSequenceSource{})

	var authorizations []string
	for range 3 {
		resp, err := client.Get(apiServer.URL)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		resp.Body.Close()

		authorizations = append(authorizations, string(body))
	}

	assert.Equal(t, []string{"Bearer token-1", "Bearer token-2", "Bearer token-3"}, authorizations)
}

// recordingTransport counts the requests it forwards to its base transport.
type recordingTransport struct {
	base  http.RoundTripper
	calls int
}

func (t *recordingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.calls++
	return t.base.RoundTrip(request)
}

func TestNewPlatformTokenSourceClient_UsesTransportOfContextClient(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	transport := &recordingTransport{base: http.DefaultTransport}
	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, &http.Client{Transport: transport})

	client := NewPlatformTokenSourceClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token-from-source"}))

	resp, err := client.Get(apiServer.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 1, transport.calls)
}

func TestNewTokenBasedClient(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Api-Token api-token", auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	client := NewAPITokenClient(t.Context(), "api-token")

	// Make a request to the mock API server
	resp, err := client.Get(apiServer.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
}
