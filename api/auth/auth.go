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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const DynatraceSSOTokenURL = "https://sso.dynatrace.com/sso/oauth2/token" //nolint:gosec

// NewAPITokenClient creates a new [http.Client] that sets the Authentication header to use [Dynatrace API tokens].
//
// [Dynatrace API tokens]: https://docs.dynatrace.com/docs/discover-dynatrace/references/dynatrace-api/basics/dynatrace-api-authentication
func NewAPITokenClient(ctx context.Context, apiToken string) *http.Client {
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Api-Token", AccessToken: apiToken}))
}

// NewPlatformTokenClient creates a new [http.Client] that sets the Authentication header to use [Dynatrace platform tokens].
//
// [Dynatrace platform tokens]: https://docs.dynatrace.com/docs/manage/identity-access-management/access-tokens-and-oauth-clients/platform-tokens
func NewPlatformTokenClient(ctx context.Context, platformToken string) *http.Client {
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: platformToken}))
}

// NewOAuthClient creates a new [http.Client] with OAuth2 client credentials authentication.
// If the [credentials.TokenURL] is not provided, we fall back to the Dynatrace SSO token URL.
// For more information see the [Dynatrace OAuth client documentation].
//
// [Dynatrace OAuth client documentation]: https://docs.dynatrace.com/docs/manage/identity-access-management/access-tokens-and-oauth-clients/oauth-clients
func NewOAuthClient(ctx context.Context, credentials *clientcredentials.Config) *http.Client {
	if credentials.TokenURL == "" {
		credentials.TokenURL = DynatraceSSOTokenURL
	}

	return credentials.Client(ctx)
}
