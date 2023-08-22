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
	"context"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
)

// NewOAuthBasedClient creates a new HTTP client with OAuth2 client credentials authentication.
// It takes a context and a clientcredentials.Config as arguments and returns an instance of *http.Client.
// The client credentials are used to authenticate requests made by the returned HTTP client.
func NewOAuthBasedClient(ctx context.Context, credentials clientcredentials.Config) *http.Client {
	if credentials.TokenURL == "" {
		credentials.TokenURL = "https://sso.dynatrace.com/sso/oauth2/token" //nolint:gosec
	}

	// Create an HTTP client using the provided OAuth2 client credentials configuration.
	// This client will automatically manage the token acquisition and inclusion in requests.
	return credentials.Client(ctx)
}
