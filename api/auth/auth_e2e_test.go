//go:build e2e

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
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"
)

func TestNewApiTokenBasedClient_ValidToken(t *testing.T) {
	t.Parallel()

	apiToken := os.Getenv("API_TOKEN")
	classicUrl := os.Getenv("CLASSIC_URL")

	client := NewAPITokenClient(context.TODO(), apiToken)

	targetUrl, err := url.JoinPath(classicUrl, "/api/v1/config/clusterversion")
	require.NoError(t, err)

	resp, err := client.Get(targetUrl)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewApiTokenBasedClient_InvalidToken(t *testing.T) {
	t.Parallel()

	apiToken := "some-invalid-token"
	classicUrl := os.Getenv("CLASSIC_URL")

	client := NewAPITokenClient(context.TODO(), apiToken)

	targetUrl, err := url.JoinPath(classicUrl, "/api/v1/config/clusterversion")
	require.NoError(t, err)

	resp, err := client.Get(targetUrl)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNewPlatformTokenClient_ValidToken(t *testing.T) {
	t.Parallel()

	platformToken := os.Getenv("PLATFORM_TOKEN")
	platformUrl := os.Getenv("PLATFORM_URL")

	client := NewPlatformTokenClient(context.TODO(), platformToken)

	targetUrl, err := url.JoinPath(platformUrl, "/platform/management/v1/environment")
	require.NoError(t, err)

	resp, err := client.Get(targetUrl)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewPlatformTokenClient_InvalidToken(t *testing.T) {
	t.Parallel()

	platformToken := "some-invalid-token"
	platformUrl := os.Getenv("PLATFORM_URL")

	client := NewPlatformTokenClient(context.TODO(), platformToken)

	targetUrl, err := url.JoinPath(platformUrl, "/platform/management/v1/environment")
	require.NoError(t, err)

	resp, err := client.Get(targetUrl)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNewOAuthClient_ValidToken(t *testing.T) {
	t.Parallel()

	credentials := clientcredentials.Config{
		TokenURL:     os.Getenv("OAUTH_TOKEN_URL"),
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
	}
	platformUrl := os.Getenv("PLATFORM_URL")

	client := NewOAuthClient(context.TODO(), &credentials)

	targetUrl, err := url.JoinPath(platformUrl, "/platform/management/v1/environment")
	require.NoError(t, err)

	resp, err := client.Get(targetUrl)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewOAuthClient_InvalidToken(t *testing.T) {
	t.Parallel()

	credentials := clientcredentials.Config{
		TokenURL:     os.Getenv("OAUTH_TOKEN_URL"),
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: "invalid-token",
	}
	platformUrl := os.Getenv("PLATFORM_URL")

	client := NewOAuthClient(context.TODO(), &credentials)

	targetUrl, err := url.JoinPath(platformUrl, "/platform/management/v1/environment")
	require.NoError(t, err)

	_, err = client.Get(targetUrl)
	require.Error(t, err)
}
