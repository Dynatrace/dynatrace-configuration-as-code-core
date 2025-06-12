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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

	client := NewOAuthBasedClient(t.Context(), config)

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

func TestNewPlatformTokenClient(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer platform-token", auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	client := NewPlatformTokenClient(t.Context(), "platform-token")

	// Make a request to the mock API server
	resp, err := client.Get(apiServer.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
}

func TestNewTokenBasedClient(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Api-Token api-token", auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	client := NewTokenBasedClient(t.Context(), "api-token")

	// Make a request to the mock API server
	resp, err := client.Get(apiServer.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
}
